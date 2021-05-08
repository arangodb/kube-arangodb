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

	json "github.com/json-iterator/go"

	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"

	"github.com/arangodb/go-driver"
	upgraderules "github.com/arangodb/go-upgrade-rules"
	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"
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
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	var plan api.Plan

	newPlan, idle := createRotateOrUpgradePlanInternal(ctx, log, apiObject, spec, status, cachedStatus, context)
	if idle {
		plan = append(plan,
			api.NewAction(api.ActionTypeIdle, api.ServerGroupUnknown, ""))
	} else {
		plan = append(plan, newPlan...)
	}
	return plan
}

func createRotateOrUpgradePlanInternal(ctx context.Context, log zerolog.Logger, apiObject k8sutil.APIObject, spec api.DeploymentSpec,
	status api.DeploymentStatus, cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) (api.Plan, bool) {

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
			decision := podNeedsUpgrading(log, m, spec, status.Images)
			if decision.Hold {
				return nil
			}

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
				newPlan = createUpgradeMemberPlan(log, m, group, "Version upgrade", spec, status,
					!decision.AutoUpgradeNeeded)
			} else {
				// Use new level of rotate logic
				rotNeeded, reason := podNeedsRotation(ctx, log, apiObject, pod, spec, group, status, m, cachedStatus, context)
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
				if _, ok := pod.Annotations[deployment.ArangoDeploymentPodReplaceAnnotation]; ok && group == api.ServerGroupDBServers {
					newPlan = api.Plan{api.NewAction(api.ActionTypeMarkToRemoveMember, group, m.ID, "Replace flag present")}
					continue
				}

				if _, ok := pod.Annotations[deployment.ArangoDeploymentPodRotateAnnotation]; ok {
					newPlan = createRotateMemberPlan(log, m, group, "Rotation flag present")
					continue
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
func podNeedsUpgrading(log zerolog.Logger, status api.MemberStatus, spec api.DeploymentSpec, images api.ImageInfoList) upgradeDecision {
	currentImage, found := currentImageInfo(spec, images)
	if !found {
		// Hold rotation tasks - we do not know image
		return upgradeDecision{Hold: true}
	}

	memberImage, found := memberImageInfo(spec, status, images)
	if !found {
		// Member info not found
		return upgradeDecision{UpgradeNeeded: false}
	}

	if currentImage.Image == memberImage.Image {
		// No change
		return upgradeDecision{UpgradeNeeded: false}
	}
	// Image changed, check if change is allowed
	specVersion := currentImage.ArangoDBVersion
	memberVersion := memberImage.ArangoDBVersion
	asLicense := func(info api.ImageInfo) upgraderules.License {
		if info.Enterprise {
			return upgraderules.LicenseEnterprise
		}
		return upgraderules.LicenseCommunity
	}
	specLicense := asLicense(currentImage)
	memberLicense := asLicense(memberImage)
	if err := upgraderules.CheckUpgradeRulesWithLicense(memberVersion, specVersion, memberLicense, specLicense); err != nil {
		// E.g. 3.x -> 4.x, we cannot allow automatically
		return upgradeDecision{
			FromVersion:    memberVersion,
			FromLicense:    memberLicense,
			ToVersion:      specVersion,
			ToLicense:      specLicense,
			UpgradeNeeded:  true,
			UpgradeAllowed: false,
		}
	}
	if specVersion.Major() != memberVersion.Major() || specVersion.Minor() != memberVersion.Minor() {
		// Is allowed, with `--database.auto-upgrade`
		log.Info().Str("spec-version", string(specVersion)).Str("pod-version", string(memberVersion)).
			Int("spec-version.major", specVersion.Major()).Int("spec-version.minor", specVersion.Minor()).
			Int("pod-version.major", memberVersion.Major()).Int("pod-version.minor", memberVersion.Minor()).
			Msg("Deciding to do a upgrade with --auto-upgrade")
		return upgradeDecision{
			FromVersion:       memberVersion,
			FromLicense:       memberLicense,
			ToVersion:         specVersion,
			ToLicense:         specLicense,
			UpgradeNeeded:     true,
			UpgradeAllowed:    true,
			AutoUpgradeNeeded: true,
		}
	}
	// Patch version change, rotate only
	return upgradeDecision{
		FromVersion:       memberVersion,
		FromLicense:       memberLicense,
		ToVersion:         specVersion,
		ToLicense:         specLicense,
		UpgradeNeeded:     true,
		UpgradeAllowed:    true,
		AutoUpgradeNeeded: true,
	}
}

func currentImageInfo(spec api.DeploymentSpec, images api.ImageInfoList) (api.ImageInfo, bool) {
	if i, ok := images.GetByImage(spec.GetImage()); ok {
		return i, true
	}
	if i, ok := images.GetByImageID(spec.GetImage()); ok {
		return i, true
	}

	return api.ImageInfo{}, false
}

func memberImageInfo(spec api.DeploymentSpec, status api.MemberStatus, images api.ImageInfoList) (api.ImageInfo, bool) {
	if status.Image != nil {
		return *status.Image, true
	}

	if i, ok := images.GetByImage(spec.GetImage()); ok {
		return i, true
	}

	if i, ok := images.GetByImageID(spec.GetImage()); ok {
		return i, true
	}

	return api.ImageInfo{}, false
}

// podNeedsRotation returns true when the specification of the
// given pod differs from what it should be according to the
// given deployment spec.
// When true is returned, a reason for the rotation is already returned.
func podNeedsRotation(ctx context.Context, log zerolog.Logger, apiObject k8sutil.APIObject, p *core.Pod, spec api.DeploymentSpec,
	group api.ServerGroup, status api.DeploymentStatus, m api.MemberStatus,
	cachedStatus inspectorInterface.Inspector, planCtx PlanBuilderContext) (bool, string) {

	if m.PodUID != p.UID {
		return true, "Pod UID does not match, this pod is not managed by Operator. Recreating"
	}

	if m.PodSpecVersion == "" {
		return true, "Pod Spec Version is nil - recreating pod"
	}

	imageInfo, imageFound := planCtx.SelectImage(spec, status)
	if !imageFound {
		// Image is not found, so rotation is not needed
		return false, ""
	}

	if m.Image != nil {
		imageInfo = *m.Image
	}

	groupSpec := spec.GetServerGroupSpec(group)

	renderedPod, err := planCtx.RenderPodForMember(ctx, cachedStatus, spec, status, m.ID, imageInfo)
	if err != nil {
		log.Err(err).Msg("Error while rendering pod")
		return false, ""
	}

	checksum, err := resources.ChecksumArangoPod(groupSpec, renderedPod)
	if err != nil {
		log.Err(err).Msg("Error while getting pod checksum")
		return false, ""
	}

	if m.PodSpecVersion != checksum {
		if _, err := json.Marshal(renderedPod); err == nil {
			log.Info().Str("id", m.ID).Str("Before", m.PodSpecVersion).Str("After", checksum).Msgf("XXXXXXXXXXX Pod needs rotation - checksum does not match")
		}
		return true, "Pod needs rotation - checksum does not match"
	}

	endpoint, err := pod.GenerateMemberEndpoint(cachedStatus, apiObject, spec, group, m)
	if err != nil {
		log.Err(err).Msg("Error while getting pod endpoint")
		return false, ""
	}

	if e := m.Endpoint; e == nil {
		if spec.CommunicationMethod == nil {
			// TODO: Remove in 1.2.0 release to allow rotation
			return false, "Pod endpoint is not set and CommunicationMethod is not set, do not recreate"
		}

		return true, "Communication method has been set - ensure endpoint"
	} else {
		if *e != endpoint {
			return true, "Pod endpoint changed"
		}
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
	group api.ServerGroup, reason string, spec api.DeploymentSpec, status api.DeploymentStatus, rotateStatefull bool) api.Plan {
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
	var plan = api.Plan{
		api.NewAction(api.ActionTypeCleanTLSKeyfileCertificate, group, member.ID, "Remove server keyfile and enforce renewal/recreation"),
	}
	if status.CurrentImage == nil || status.CurrentImage.Image != spec.GetImage() {
		plan = append(plan,
			api.NewAction(api.ActionTypeSetCurrentImage, group, "", reason).SetImage(spec.GetImage()),
		)
	}
	if member.Image == nil || member.Image.Image != spec.GetImage() {
		plan = append(plan,
			api.NewAction(api.ActionTypeSetMemberCurrentImage, group, member.ID, reason).SetImage(spec.GetImage()),
		)
	}
	plan = append(plan,
		api.NewAction(api.ActionTypeResignLeadership, group, member.ID, reason),
		api.NewAction(upgradeAction, group, member.ID, reason),
		api.NewAction(api.ActionTypeWaitForMemberUp, group, member.ID),
	)
	return withMaintenance(plan...)
}
