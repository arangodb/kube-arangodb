//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package reconcile

import (
	"strings"

	driver "github.com/arangodb/go-driver"
	upgraderules "github.com/arangodb/go-upgrade-rules"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// upgradeDecision is the result of an upgrade check.
type upgradeDecision struct {
	FromVersion       driver.Version
	FromLicense       upgraderules.License
	ToVersion         driver.Version
	ToLicense         upgraderules.License
	UpgradeNeeded     bool // If set, the image version has changed
	UpgradeAllowed    bool // If set, it is an allowed version change
	AutoUpgradeNeeded bool // If set, the database must be started with `--database.auto-upgrade` once
}

// CreatePlan considers the current specification & status of the deployment creates a plan to
// get the status in line with the specification.
// If a plan already exists, nothing is done.
func (d *Reconciler) CreatePlan() error {
	// Get all current pods
	pods, err := d.context.GetOwnedPods()
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get owned pods")
		return maskAny(err)
	}

	// Create plan
	apiObject := d.context.GetAPIObject()
	spec := d.context.GetSpec()
	status, lastVersion := d.context.GetStatus()
	ctx := newPlanBuilderContext(d.context)
	newPlan, changed := createPlan(d.log, apiObject, status.Plan, spec, status, pods, ctx)

	// If not change, we're done
	if !changed {
		return nil
	}

	// Save plan
	if len(newPlan) == 0 {
		// Nothing to do
		return nil
	}
	status.Plan = newPlan
	if err := d.context.UpdateStatus(status, lastVersion); err != nil {
		return maskAny(err)
	}
	return nil
}

// createPlan considers the given specification & status and creates a plan to get the status in line with the specification.
// If a plan already exists, the given plan is returned with false.
// Otherwise the new plan is returned with a boolean true.
func createPlan(log zerolog.Logger, apiObject k8sutil.APIObject,
	currentPlan api.Plan, spec api.DeploymentSpec,
	status api.DeploymentStatus, pods []v1.Pod,
	context PlanBuilderContext) (api.Plan, bool) {
	if len(currentPlan) > 0 {
		// Plan already exists, complete that first
		return currentPlan, false
	}

	// Check for various scenario's
	var plan api.Plan

	// Check for members in failed state
	status.Members.ForeachServerGroup(func(group api.ServerGroup, members api.MemberStatusList) error {
		for _, m := range members {
			if m.Phase == api.MemberPhaseFailed && len(plan) == 0 {
				log.Debug().
					Str("id", m.ID).
					Str("role", group.AsRole()).
					Msg("Creating member replacement plan because member has failed")
				newID := ""
				if group == api.ServerGroupAgents {
					newID = m.ID // Agents cannot (yet) be replaced with new IDs
				}
				plan = append(plan,
					api.NewAction(api.ActionTypeRemoveMember, group, m.ID),
					api.NewAction(api.ActionTypeAddMember, group, newID),
				)
			}
		}
		return nil
	})

	// Check for cleaned out dbserver in created state
	for _, m := range status.Members.DBServers {
		if len(plan) == 0 && m.Phase.IsCreatedOrDrain() && m.Conditions.IsTrue(api.ConditionTypeCleanedOut) {
			log.Debug().
				Str("id", m.ID).
				Str("role", api.ServerGroupDBServers.AsRole()).
				Msg("Creating dbserver replacement plan because server is cleanout in created phase")
			plan = append(plan,
				api.NewAction(api.ActionTypeRemoveMember, api.ServerGroupDBServers, m.ID),
				api.NewAction(api.ActionTypeAddMember, api.ServerGroupDBServers, ""),
			)
		}
	}

	// Check for scale up/down
	if len(plan) == 0 {
		switch spec.GetMode() {
		case api.DeploymentModeSingle:
			// Never scale down
		case api.DeploymentModeActiveFailover:
			// Only scale singles
			plan = append(plan, createScalePlan(log, status.Members.Single, api.ServerGroupSingle, spec.Single.GetCount())...)
		case api.DeploymentModeCluster:
			// Scale dbservers, coordinators
			plan = append(plan, createScalePlan(log, status.Members.DBServers, api.ServerGroupDBServers, spec.DBServers.GetCount())...)
			plan = append(plan, createScalePlan(log, status.Members.Coordinators, api.ServerGroupCoordinators, spec.Coordinators.GetCount())...)
		}
		if spec.GetMode().SupportsSync() {
			// Scale syncmasters & syncworkers
			plan = append(plan, createScalePlan(log, status.Members.SyncMasters, api.ServerGroupSyncMasters, spec.SyncMasters.GetCount())...)
			plan = append(plan, createScalePlan(log, status.Members.SyncWorkers, api.ServerGroupSyncWorkers, spec.SyncWorkers.GetCount())...)
		}
	}

	// Check for the need to rotate one or more members
	if len(plan) == 0 {
		getPod := func(podName string) *v1.Pod {
			for _, p := range pods {
				if p.GetName() == podName {
					return &p
				}
			}
			return nil
		}
		// createRotateOrUpgradePlan goes over all pods to check if an upgrade or rotate
		// is needed. If an upgrade is needed but not allowed, the second return value
		// will be true.
		// Returns: (newPlan, upgradeNotAllowed)
		createRotateOrUpgradePlan := func() (api.Plan, bool, driver.Version, driver.Version, upgraderules.License, upgraderules.License) {
			var newPlan api.Plan
			upgradeNotAllowed := false
			var fromVersion, toVersion driver.Version
			var fromLicense, toLicense upgraderules.License
			status.Members.ForeachServerGroup(func(group api.ServerGroup, members api.MemberStatusList) error {
				for _, m := range members {
					if m.Phase != api.MemberPhaseCreated {
						// Only rotate when phase is created
						continue
					}
					if podName := m.PodName; podName != "" {
						if p := getPod(podName); p != nil {
							// Got pod, compare it with what it should be
							decision := podNeedsUpgrading(*p, spec, status.Images)
							if decision.UpgradeNeeded && !decision.UpgradeAllowed {
								// Oops, upgrade is not allowed
								upgradeNotAllowed = true
								fromVersion = decision.FromVersion
								fromLicense = decision.FromLicense
								toVersion = decision.ToVersion
								toLicense = decision.ToLicense
								return nil
							} else if len(newPlan) == 0 {
								// Only rotate/upgrade 1 pod at a time
								if decision.UpgradeNeeded {
									// Yes, upgrade is needed (and allowed)
									newPlan = createUpgradeMemberPlan(log, m, group, "Version upgrade", spec.GetImage(), status)
								} else {
									// Upgrade is not needed, see if rotation is needed
									if rotNeeded, reason := podNeedsRotation(log, *p, apiObject, spec, group, status, m.ID, context); rotNeeded {
										newPlan = createRotateMemberPlan(log, m, group, reason)
									}
								}
							}
						}
					}
				}
				return nil
			})
			return newPlan, upgradeNotAllowed, fromVersion, toVersion, fromLicense, toLicense
		}
		if newPlan, upgradeNotAllowed, fromVersion, toVersion, fromLicense, toLicense := createRotateOrUpgradePlan(); upgradeNotAllowed {
			// Upgrade is needed, but not allowed
			context.CreateEvent(k8sutil.NewUpgradeNotAllowedEvent(apiObject, fromVersion, toVersion, fromLicense, toLicense))
		} else {
			// Use the new plan
			plan = newPlan
		}
	}

	// Check for the need to rotate TLS certificate of a members
	if len(plan) == 0 {
		plan = createRotateTLSServerCertificatePlan(log, spec, status, context.GetTLSKeyfile)
	}

	// Check for changes storage classes or requirements
	if len(plan) == 0 {
		plan = createRotateServerStoragePlan(log, apiObject, spec, status, context.GetPvc, context.CreateEvent)
	}

	// Check for the need to rotate TLS CA certificate and all members
	if len(plan) == 0 {
		plan = createRotateTLSCAPlan(log, apiObject, spec, status, context.GetTLSCA, context.CreateEvent)
	}

	// Return plan
	return plan, true
}

// podNeedsUpgrading decides if an upgrade of the pod is needed (to comply with
// the given spec) and if that is allowed.
func podNeedsUpgrading(p v1.Pod, spec api.DeploymentSpec, images api.ImageInfoList) upgradeDecision {
	if c, found := k8sutil.GetContainerByName(&p, k8sutil.ServerContainerName); found {
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
func podNeedsRotation(log zerolog.Logger, p v1.Pod, apiObject metav1.Object, spec api.DeploymentSpec,
	group api.ServerGroup, status api.DeploymentStatus, id string,
	context PlanBuilderContext) (bool, string) {
	groupSpec := spec.GetServerGroupSpec(group)

	// Check image pull policy
	c, found := k8sutil.GetContainerByName(&p, k8sutil.ServerContainerName)
	if found {
		if c.ImagePullPolicy != spec.GetImagePullPolicy() {
			return true, "Image pull policy changed"
		}
	} else {
		return true, "Server container not found"
	}

	podImageInfo, found := status.Images.GetByImageID(c.Image)
	if !found {
		return false, "Server Image not found"
	}

	// Check arguments
	expectedArgs := strings.Join(context.GetExpectedPodArguments(apiObject, spec, group, status.Members.Agents, id, podImageInfo.ArangoDBVersion), " ")
	actualArgs := strings.Join(getContainerArgs(c), " ")
	if expectedArgs != actualArgs {
		log.Debug().
			Str("actual-args", actualArgs).
			Str("expected-args", expectedArgs).
			Msg("Arguments changed. Rotation needed.")
		return true, "Arguments changed"
	}

	// Check service account
	if normalizeServiceAccountName(p.Spec.ServiceAccountName) != normalizeServiceAccountName(groupSpec.GetServiceAccountName()) {
		return true, "ServiceAccountName changed"
	}

	return false, ""
}

// normalizeServiceAccountName replaces default with empty string, otherwise returns the input.
func normalizeServiceAccountName(name string) string {
	if name == "default" {
		return ""
	}
	return ""
}

// createScalePlan creates a scaling plan for a single server group
func createScalePlan(log zerolog.Logger, members api.MemberStatusList, group api.ServerGroup, count int) api.Plan {
	var plan api.Plan
	if len(members) < count {
		// Scale up
		toAdd := count - len(members)
		for i := 0; i < toAdd; i++ {
			plan = append(plan, api.NewAction(api.ActionTypeAddMember, group, ""))
		}
		log.Debug().
			Int("count", count).
			Int("actual-count", len(members)).
			Int("delta", toAdd).
			Str("role", group.AsRole()).
			Msg("Creating scale-up plan")
	} else if len(members) > count {
		// Note, we scale down 1 member at a time
		if m, err := members.SelectMemberToRemove(); err != nil {
			log.Warn().Err(err).Str("role", group.AsRole()).Msg("Failed to select member to remove")
		} else {

			log.Debug().
				Str("member-id", m.ID).
				Str("phase", string(m.Phase)).
				Msg("Found member to remove")
			if group == api.ServerGroupDBServers {
				plan = append(plan,
					api.NewAction(api.ActionTypeCleanOutMember, group, m.ID),
				)
			}
			plan = append(plan,
				api.NewAction(api.ActionTypeShutdownMember, group, m.ID),
				api.NewAction(api.ActionTypeRemoveMember, group, m.ID),
			)
			log.Debug().
				Int("count", count).
				Int("actual-count", len(members)).
				Str("role", group.AsRole()).
				Str("member-id", m.ID).
				Msg("Creating scale-down plan")
		}
	}
	return plan
}

// createRotateMemberPlan creates a plan to rotate (stop-recreate-start) an existing
// member.
func createRotateMemberPlan(log zerolog.Logger, member api.MemberStatus,
	group api.ServerGroup, reason string) api.Plan {
	log.Debug().
		Str("id", member.ID).
		Str("role", group.AsRole()).
		Str("reason", reason).
		Msg("Creating rotation plan")
	plan := api.Plan{
		api.NewAction(api.ActionTypeRotateMember, group, member.ID, reason),
		api.NewAction(api.ActionTypeWaitForMemberUp, group, member.ID),
	}
	return plan
}

// createUpgradeMemberPlan creates a plan to upgrade (stop-recreateWithAutoUpgrade-stop-start) an existing
// member.
func createUpgradeMemberPlan(log zerolog.Logger, member api.MemberStatus,
	group api.ServerGroup, reason string, imageName string, status api.DeploymentStatus) api.Plan {
	upgradeAction := api.ActionTypeUpgradeMember
	if group.IsStateless() {
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

func getContainerArgs(c v1.Container) []string {
	if len(c.Command) >= 1 {
		return c.Command[1:]
	}
	return c.Args
}
