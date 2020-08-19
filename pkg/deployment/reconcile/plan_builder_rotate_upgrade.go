//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//
// Author Tomasz Mielech <tomasz@arangodb.com>
//

package reconcile

import (
	"context"

	"github.com/arangodb/go-driver"
	upgraderules "github.com/arangodb/go-upgrade-rules"
	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// rotationByAnnotationOrder - Change order of execution - Coordinators and Agents should be executed before DBServer to save time
	rotationByAnnotationOrder = []api.ServerGroup{
		api.ServerGroupAgents,
		api.ServerGroupSingle,
		api.ServerGroupCoordinators,
		api.ServerGroupDBServers,
		api.ServerGroupSyncMasters,
		api.ServerGroupSyncWorkers,
	}
)

// createRotateOrUpgradePlan goes over all pods to check if an upgrade or rotate is needed.
func createRotateOrUpgradePlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspector.Inspector, context PlanBuilderContext) api.Plan {
	var plan api.Plan

	newPlan, idle := createRotateOrUpgradePlanInternal(log, apiObject, spec, status, cachedStatus, context)
	if idle {
		plan = append(plan,
			api.NewAction(api.ActionTypeIdle, api.ServerGroupUnknown, ""))
	} else {
		plan = append(plan, newPlan...)
	}
	return plan
}

func createRotateOrUpgradePlanInternal(log zerolog.Logger, apiObject k8sutil.APIObject, spec api.DeploymentSpec,
	status api.DeploymentStatus, cachedStatus inspector.Inspector, context PlanBuilderContext) (api.Plan, bool) {

	var newPlan api.Plan
	var upgradeNotAllowed bool
	var fromVersion, toVersion driver.Version
	var fromLicense, toLicense upgraderules.License

	status.Members.ForeachServerGroup(func(group api.ServerGroup, members api.MemberStatusList) error {

		for _, m := range members {
			if m.Phase != api.MemberPhaseCreated || m.PodName == "" {
				// Only rotate when phase is created
				continue
			}

			pod, found := cachedStatus.Pod(m.PodName)
			if !found {
				continue
			}

			// Got pod, compare it with what it should be
			decision := podNeedsUpgrading(log, pod, spec, status.Images)
			if decision.UpgradeNeeded && !decision.UpgradeAllowed {
				// Oops, upgrade is not allowed
				upgradeNotAllowed = true
				fromVersion = decision.FromVersion
				fromLicense = decision.FromLicense
				toVersion = decision.ToVersion
				toLicense = decision.ToLicense
				return nil
			}

			if !newPlan.IsEmpty() {
				// Only rotate/upgrade 1 pod at a time
				continue
			}

			if decision.UpgradeNeeded {
				// Yes, upgrade is needed (and allowed)
				newPlan = createUpgradeMemberPlan(log, m, group, "Version upgrade", spec.GetImage(), status,
					!decision.AutoUpgradeNeeded)
			} else {
				// Use new level of rotate logic
				rotNeeded, reason := podNeedsRotation(log, pod, apiObject, spec, group, status, m, cachedStatus, context)
				if rotNeeded {
					newPlan = createRotateMemberPlan(log, m, group, reason)
				}
			}

			if !newPlan.IsEmpty() {
				// Only rotate/upgrade 1 pod at a time
				continue
			}
		}
		return nil
	})

	status.Members.ForeachServerInGroups(func(group api.ServerGroup, members api.MemberStatusList) error {
		for _, m := range members {
			if m.Phase != api.MemberPhaseCreated || m.PodName == "" {
				// Only rotate when phase is created
				continue
			}

			if !newPlan.IsEmpty() {
				// Only rotate/upgrade 1 pod at a time
				continue
			}

			pod, found := cachedStatus.Pod(m.PodName)
			if !found {
				continue
			}

			if pod.Annotations != nil {
				if _, ok := pod.Annotations[deployment.ArangoDeploymentPodRotateAnnotation]; ok {
					newPlan = createRotateMemberPlan(log, m, group, "Rotation flag present")
				}
			}
		}

		return nil
	}, rotationByAnnotationOrder...)

	if upgradeNotAllowed {
		context.CreateEvent(k8sutil.NewUpgradeNotAllowedEvent(apiObject, fromVersion, toVersion, fromLicense, toLicense))
	} else if !newPlan.IsEmpty() {
		if clusterReadyForUpgrade(context) {
			// Use the new plan
			return newPlan, false
		} else {
			if util.BoolOrDefault(spec.AllowUnsafeUpgrade, false) {
				log.Info().Msg("Pod needs upgrade but cluster is not ready. Either some shards are not in sync or some member is not ready, but unsafe upgrade is allowed")
				// Use the new plan
				return newPlan, false
			} else {
				log.Info().Msg("Pod needs upgrade but cluster is not ready. Either some shards are not in sync or some member is not ready.")
				return nil, true
			}
		}
	}
	return nil, false
}

// podNeedsUpgrading decides if an upgrade of the pod is needed (to comply with
// the given spec) and if that is allowed.
func podNeedsUpgrading(log zerolog.Logger, p *core.Pod, spec api.DeploymentSpec, images api.ImageInfoList) upgradeDecision {
	if c, found := k8sutil.GetContainerByName(p, k8sutil.ServerContainerName); found {
		specImageInfo, found := images.GetByImage(spec.GetImage())
		if !found {
			return upgradeDecision{UpgradeNeeded: false}
		}
		podImageInfo, found := images.GetByImageID(c.Image)
		if !found {
			return upgradeDecision{UpgradeNeeded: false}
		}
		if specImageInfo.ImageID == podImageInfo.ImageID {
			// No change
			return upgradeDecision{UpgradeNeeded: false}
		}
		// Image changed, check if change is allowed
		specVersion := specImageInfo.ArangoDBVersion
		podVersion := podImageInfo.ArangoDBVersion
		asLicense := func(info api.ImageInfo) upgraderules.License {
			if info.Enterprise {
				return upgraderules.LicenseEnterprise
			}
			return upgraderules.LicenseCommunity
		}
		specLicense := asLicense(specImageInfo)
		podLicense := asLicense(podImageInfo)
		if err := upgraderules.CheckUpgradeRulesWithLicense(podVersion, specVersion, podLicense, specLicense); err != nil {
			// E.g. 3.x -> 4.x, we cannot allow automatically
			return upgradeDecision{
				FromVersion:    podVersion,
				FromLicense:    podLicense,
				ToVersion:      specVersion,
				ToLicense:      specLicense,
				UpgradeNeeded:  true,
				UpgradeAllowed: false,
			}
		}
		if specVersion.Major() != podVersion.Major() || specVersion.Minor() != podVersion.Minor() {
			// Is allowed, with `--database.auto-upgrade`
			log.Info().Str("spec-version", string(specVersion)).Str("pod-version", string(podVersion)).
				Int("spec-version.major", specVersion.Major()).Int("spec-version.minor", specVersion.Minor()).
				Int("pod-version.major", podVersion.Major()).Int("pod-version.minor", podVersion.Minor()).
				Str("pod", p.GetName()).Msg("Deciding to do a upgrade with --auto-upgrade")
			return upgradeDecision{
				FromVersion:       podVersion,
				FromLicense:       podLicense,
				ToVersion:         specVersion,
				ToLicense:         specLicense,
				UpgradeNeeded:     true,
				UpgradeAllowed:    true,
				AutoUpgradeNeeded: true,
			}
		}
		// Patch version change, rotate only
		return upgradeDecision{
			FromVersion:       podVersion,
			FromLicense:       podLicense,
			ToVersion:         specVersion,
			ToLicense:         specLicense,
			UpgradeNeeded:     true,
			UpgradeAllowed:    true,
			AutoUpgradeNeeded: false,
		}
	}
	return upgradeDecision{UpgradeNeeded: false}
}

// podNeedsRotation returns true when the specification of the
// given pod differs from what it should be according to the
// given deployment spec.
// When true is returned, a reason for the rotation is already returned.
func podNeedsRotation(log zerolog.Logger, p *core.Pod, apiObject metav1.Object, spec api.DeploymentSpec,
	group api.ServerGroup, status api.DeploymentStatus, m api.MemberStatus,
	cachedStatus inspector.Inspector, context PlanBuilderContext) (bool, string) {
	if m.PodUID != p.UID {
		return true, "Pod UID does not match, this pod is not managed by Operator. Recreating"
	}

	if m.PodSpecVersion == "" {
		return true, "Pod Spec Version is nil - recreating pod"
	}

	imageInfo, imageFound := context.SelectImage(spec, status)
	if !imageFound {
		// Image is not found, so rotation is not needed
		return false, ""
	}

	renderedPod, err := context.RenderPodForMember(cachedStatus, spec, status, m.ID, imageInfo)
	if err != nil {
		log.Err(err).Msg("Error while rendering pod")
		return false, ""
	}

	checksum, err := k8sutil.GetPodSpecChecksum(renderedPod.Spec)
	if err != nil {
		log.Err(err).Msg("Error while getting pod checksum")
		return false, ""
	}

	if m.PodSpecVersion != checksum {
		return true, "Pod needs rotation - checksum does not match"
	}

	return false, ""
}

// clusterReadyForUpgrade returns true if the cluster is ready for the next update, that is:
// 	- all shards are in sync
// 	- all members are ready and fine
func clusterReadyForUpgrade(context PlanBuilderContext) bool {
	status, _ := context.GetStatus()
	allInSync := context.GetShardSyncStatus()
	return allInSync && status.Conditions.IsTrue(api.ConditionTypeReady)
}

// createUpgradeMemberPlan creates a plan to upgrade (stop-recreateWithAutoUpgrade-stop-start) an existing
// member.
func createUpgradeMemberPlan(log zerolog.Logger, member api.MemberStatus,
	group api.ServerGroup, reason string, imageName string, status api.DeploymentStatus, rotateStatefull bool) api.Plan {
	upgradeAction := api.ActionTypeUpgradeMember
	if rotateStatefull || group.IsStateless() {
		upgradeAction = api.ActionTypeRotateMember
	}
	log.Debug().
		Str("id", member.ID).
		Str("role", group.AsRole()).
		Str("reason", reason).
		Str("action", string(upgradeAction)).
		Msg("Creating upgrade plan")
	plan := api.Plan{
		api.NewAction(upgradeAction, group, member.ID, reason),
		api.NewAction(api.ActionTypeWaitForMemberUp, group, member.ID),
	}
	if status.CurrentImage == nil || status.CurrentImage.Image != imageName {
		plan = append(api.Plan{
			api.NewAction(api.ActionTypeSetCurrentImage, group, "", reason).SetImage(imageName),
		}, plan...)
	}
	return plan
}
