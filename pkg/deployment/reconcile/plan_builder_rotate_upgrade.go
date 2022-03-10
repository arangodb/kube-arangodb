//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package reconcile

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/deployment/rotation"

	"github.com/arangodb/kube-arangodb/pkg/deployment/features"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"

	"github.com/arangodb/go-driver"
	upgraderules "github.com/arangodb/go-upgrade-rules"
	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
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

// upgradeDecision is the result of an upgrade check.
type upgradeDecision struct {
	FromVersion       driver.Version
	FromLicense       upgraderules.License
	ToVersion         driver.Version
	ToLicense         upgraderules.License
	UpgradeNeeded     bool // If set, the image version has changed
	UpgradeAllowed    bool // If set, it is an allowed version change
	AutoUpgradeNeeded bool // If set, the database must be started with `--database.auto-upgrade` once

	Hold bool
}

// createRotateOrUpgradePlan goes over all pods to check if an upgrade or rotate is needed.
func createRotateOrUpgradePlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	var plan api.Plan

	newPlan, idle := createRotateOrUpgradePlanInternal(log, apiObject, spec, status, cachedStatus, context)
	if idle {
		plan = append(plan,
			actions.NewClusterAction(api.ActionTypeIdle))
	} else {
		plan = append(plan, newPlan...)
	}
	return plan
}

func createMarkToRemovePlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	var plan api.Plan

	status.Members.ForeachServerInGroups(func(group api.ServerGroup, members api.MemberStatusList) error {
		for _, m := range members {
			if m.Phase != api.MemberPhaseCreated || m.PodName == "" {
				// Only rotate when phase is created
				continue
			}

			pod, found := cachedStatus.Pod(m.PodName)
			if !found {
				continue
			}

			if pod.Annotations != nil {
				if _, ok := pod.Annotations[deployment.ArangoDeploymentPodReplaceAnnotation]; ok && (group == api.ServerGroupDBServers || group == api.ServerGroupAgents || group == api.ServerGroupCoordinators) {
					if !m.Conditions.IsTrue(api.ConditionTypeMarkedToRemove) {
						plan = append(plan, actions.NewAction(api.ActionTypeMarkToRemoveMember, group, m, "Replace flag present"))
						continue
					}
				}
			}
		}

		return nil
	}, rotationByAnnotationOrder...)

	return plan
}

func createRotateOrUpgradePlanInternal(log zerolog.Logger, apiObject k8sutil.APIObject, spec api.DeploymentSpec, status api.DeploymentStatus, cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) (api.Plan, bool) {
	decision := createRotateOrUpgradeDecision(log, spec, status, context)

	if decision.IsUpgrade() {

		for _, m := range status.Members.AsList() {
			// Pre-check
			d := decision[m.Member.ID]
			if !d.upgrade {
				continue
			}

			// We have member to upgrade
			if d.upgradeDecision.Hold {
				// Holding upgrade
				continue
			}

			if !d.upgradeDecision.UpgradeAllowed {
				context.CreateEvent(k8sutil.NewUpgradeNotAllowedEvent(apiObject, d.upgradeDecision.FromVersion, d.upgradeDecision.ToVersion, d.upgradeDecision.FromLicense, d.upgradeDecision.ToLicense))
				return nil, false
			}
		}

		// Upgrade phase
		// During upgrade always get first member which needs to be upgraded
		for _, m := range status.Members.AsList() {
			d := decision[m.Member.ID]
			if !d.upgrade {
				continue
			}

			// We have member to upgrade
			if d.upgradeDecision.Hold {
				// Holding upgrade
				return nil, false
			}

			if !d.upgradeDecision.UpgradeNeeded {
				// In upgrade scenario but upgrade is not needed
				return nil, false
			}

			if !d.upgradeDecision.UpgradeAllowed {
				context.CreateEvent(k8sutil.NewUpgradeNotAllowedEvent(apiObject, d.upgradeDecision.FromVersion, d.upgradeDecision.ToVersion, d.upgradeDecision.FromLicense, d.upgradeDecision.ToLicense))
				return nil, false
			}

			if d.updateAllowed {
				// We are fine, group is alive so we can proceed
				return createUpgradeMemberPlan(log, m.Member, m.Group, "Version upgrade", spec, status, !d.upgradeDecision.AutoUpgradeNeeded), false
			} else if d.unsafeUpdateAllowed {
				log.Info().Str("member", m.Member.ID).Msg("Pod needs upgrade but cluster is not ready. Either some shards are not in sync or some member is not ready, but unsafe upgrade is allowed")
				return createUpgradeMemberPlan(log, m.Member, m.Group, "Version upgrade", spec, status, !d.upgradeDecision.AutoUpgradeNeeded), false
			} else {
				log.Info().Str("member", m.Member.ID).Msg("Pod needs upgrade but cluster is not ready. Either some shards are not in sync or some member is not ready.")
				return nil, true
			}
		}

		log.Warn().Msg("Pod upgrade plan has been made, but it has been dropped due to missing flag")
		return nil, false
	} else if decision.IsUpdate() {
		// Update phase
		for _, m := range status.Members.AsList() {
			d := decision[m.Member.ID]
			if !d.update {
				continue
			}

			if !d.updateAllowed {
				// Update is not allowed due to constraint
				if !d.unsafeUpdateAllowed {
					log.Info().Str("member", m.Member.ID).Msg("Pod needs restart but cluster is not ready. Either some shards are not in sync or some member is not ready.")
					continue
				}
				log.Info().Str("member", m.Member.ID).Msg("Pod needs restart but cluster is not ready. Either some shards are not in sync or some member is not ready, but unsafe upgrade is allowed")
			}

			if m.Member.Conditions.IsTrue(api.ConditionTypeRestart) {
				return createRotateMemberPlan(log, m.Member, m.Group, spec, "Restart flag present"), false
			}
			arangoMember, ok := cachedStatus.ArangoMember(m.Member.ArangoMemberName(apiObject.GetName(), m.Group))
			if !ok {
				continue
			}

			p, ok := cachedStatus.Pod(m.Member.PodName)
			if !ok {
				p = nil
			}

			if mode, p, reason, err := rotation.IsRotationRequired(log, cachedStatus, spec, m.Member, m.Group, p, arangoMember.Spec.Template, arangoMember.Status.Template); err != nil {
				log.Err(err).Str("member", m.Member.ID).Msgf("Error while generating update plan")
				continue
			} else if mode != rotation.InPlaceRotation {
				return api.Plan{actions.NewAction(api.ActionTypeSetMemberCondition, m.Group, m.Member, "Cleaning update").
					AddParam(api.ConditionTypePendingUpdate.String(), "").
					AddParam(api.ConditionTypeUpdating.String(), "T")}, false
			} else {
				p = p.After(
					actions.NewAction(api.ActionTypeWaitForMemberUp, m.Group, m.Member),
					actions.NewAction(api.ActionTypeWaitForMemberInSync, m.Group, m.Member))

				p = p.Wrap(actions.NewAction(api.ActionTypeSetMemberCondition, m.Group, m.Member, reason).
					AddParam(api.ConditionTypePendingUpdate.String(), "").AddParam(api.ConditionTypeUpdating.String(), "T"),
					actions.NewAction(api.ActionTypeSetMemberCondition, m.Group, m.Member, reason).
						AddParam(api.ConditionTypeUpdating.String(), ""))

				return p, false
			}
		}
		return nil, true
	}

	return nil, false
}

// podNeedsUpgrading decides if an upgrade of the pod is needed (to comply with
// the given spec) and if that is allowed.
func podNeedsUpgrading(log zerolog.Logger, status api.MemberStatus, spec api.DeploymentSpec, images api.ImageInfoList) upgradeDecision {
	currentImage, found := currentImageInfo(spec, images, status.Architecture)
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

func currentImageInfo(spec api.DeploymentSpec, images api.ImageInfoList, arch api.ArangoDeploymentArchitectureType) (api.ImageInfo, bool) {
	if i, ok := images.GetByImage(spec.GetImage()); ok {
		return i, true
	}

	if i, ok := images.GetByImageID(spec.GetImage(), arch); ok {
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

	if i, ok := images.GetByImageID(spec.GetImage(), status.Architecture); ok {
		return i, true
	}

	return api.ImageInfo{}, false
}

func getPodDetails(ctx context.Context, log zerolog.Logger, apiObject k8sutil.APIObject, spec api.DeploymentSpec,
	group api.ServerGroup, status api.DeploymentStatus, m api.MemberStatus,
	cachedStatus inspectorInterface.Inspector, planCtx PlanBuilderContext) (string, *core.Pod, *api.ArangoMember, bool) {
	imageInfo, imageFound := planCtx.SelectImageForMember(spec, status, m)
	if !imageFound {
		// Image is not found, so rotation is not needed
		return "", nil, nil, false
	}

	member, ok := cachedStatus.ArangoMember(m.ArangoMemberName(apiObject.GetName(), group))
	if !ok {
		return "", nil, nil, false
	}

	groupSpec := spec.GetServerGroupSpec(group)

	renderedPod, err := planCtx.RenderPodForMember(ctx, cachedStatus, spec, status, m.ID, imageInfo)
	if err != nil {
		log.Err(err).Msg("Error while rendering pod")
		return "", nil, nil, false
	}

	checksum, err := resources.ChecksumArangoPod(groupSpec, renderedPod)
	if err != nil {
		log.Err(err).Msg("Error while getting pod checksum")
		return "", nil, nil, false
	}

	return checksum, renderedPod, member, true
}

// arangoMemberPodTemplateNeedsUpdate returns true when the specification of the
// given pod differs from what it should be according to the
// given deployment spec.
// When true is returned, a reason for the rotation is already returned.
func arangoMemberPodTemplateNeedsUpdate(ctx context.Context, log zerolog.Logger, apiObject k8sutil.APIObject, spec api.DeploymentSpec,
	group api.ServerGroup, status api.DeploymentStatus, m api.MemberStatus,
	cachedStatus inspectorInterface.Inspector, planCtx PlanBuilderContext) (string, bool) {
	checksum, _, member, valid := getPodDetails(ctx, log, apiObject, spec, group, status, m, cachedStatus, planCtx)
	if valid && !member.Spec.Template.EqualPodSpecChecksum(checksum) {
		return "Pod Spec changed", true
	}

	return "", false
}

// clusterReadyForUpgrade returns true if the cluster is ready for the next update, that is:
// 	- all shards are in sync
// 	- all members are ready and fine
func groupReadyForRestart(context PlanBuilderContext, status api.DeploymentStatus, member api.MemberStatus, group api.ServerGroup) bool {
	if group == api.ServerGroupSingle {
		return true
	}

	if !status.Conditions.IsTrue(api.ConditionTypeBootstrapCompleted) {
		// Restart is allowed always when bootstrap is not yet completed
		return true
	}

	// If current member did not become ready even once. Kill it
	if !member.Conditions.IsTrue(api.ConditionTypeStarted) {
		return true
	}

	// If current core containers are dead kill it.
	if !member.Conditions.IsTrue(api.ConditionTypeServing) {
		return true
	}

	switch group {
	case api.ServerGroupDBServers:
		// TODO: Improve shard placement discovery and keep WriteConcern
		return context.GetShardSyncStatus() && status.Members.MembersOfGroup(group).AllMembersServing()
	default:
		// In case of agents we can kill only one agent at same time
		return status.Members.MembersOfGroup(group).AllMembersServing()
	}
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

	plan := createRotateMemberPlanWithAction(member, group, upgradeAction, spec, reason)

	if member.Image == nil || member.Image.Image != spec.GetImage() {
		plan = plan.Before(actions.NewAction(api.ActionTypeSetMemberCurrentImage, group, member, reason).SetImage(spec.GetImage()))
	}
	if status.CurrentImage == nil || status.CurrentImage.Image != spec.GetImage() {
		plan = plan.Before(actions.NewClusterAction(api.ActionTypeSetCurrentImage, reason).SetImage(spec.GetImage()))
	}

	return withSecureWrap(member, group, spec, plan...)
}

func withSecureWrap(member api.MemberStatus,
	group api.ServerGroup, spec api.DeploymentSpec, plan ...api.Action) api.Plan {
	image := member.Image
	if image == nil {
		return plan
	}

	if skipResignLeadership(spec.GetMode(), image.ArangoDBVersion) {
		// In this case we skip resign leadership but we enable maintenance
		return withMaintenanceStart(plan...)
	} else {
		return withResignLeadership(group, member, "ResignLeadership", plan...)
	}
}

func skipResignLeadership(mode api.DeploymentMode, v driver.Version) bool {
	return mode == api.DeploymentModeCluster && features.Maintenance().Enabled() && ((v.CompareTo("3.6.0") >= 0 && v.CompareTo("3.6.14") <= 0) ||
		(v.CompareTo("3.7.0") >= 0 && v.CompareTo("3.7.12") <= 0))
}
