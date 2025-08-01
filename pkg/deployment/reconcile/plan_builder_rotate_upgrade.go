//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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
	"fmt"

	"github.com/arangodb/go-driver"
	upgraderules "github.com/arangodb/go-upgrade-rules"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	sharedReconcile "github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/deployment/rotation"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/compare"
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

	Hold bool
}

// createRotateOrUpgradePlan goes over all pods to check if an upgrade or rotate is needed.
func (r *Reconciler) createRotateOrUpgradePlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	var plan api.Plan

	newPlan, idle := r.createRotateOrUpgradePlanInternal(apiObject, spec, status, context)
	if idle {
		plan = append(plan,
			actions.NewClusterAction(api.ActionTypeIdle))
	} else {
		plan = append(plan, newPlan...)
	}
	return plan
}

func (r *Reconciler) createMarkToRemovePlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	var plan api.Plan

	for _, e := range status.Members.AsListInGroups(spec.Rotate.GetOrder(nil).Groups()...) {
		m := e.Member
		group := e.Group
		if m.Phase != api.MemberPhaseCreated || m.Pod.GetName() == "" {
			// Only rotate when phase is created
			continue
		}

		cache, ok := context.ACS().ClusterCache(m.ClusterID)
		if !ok {
			continue
		}

		pod, found := cache.Pod().V1().GetSimple(m.Pod.GetName())
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
	return plan
}

func (r *Reconciler) createRotateOrUpgradePlanInternal(apiObject k8sutil.APIObject, spec api.DeploymentSpec, status api.DeploymentStatus, context PlanBuilderContext) (api.Plan, bool) {
	decision := r.createRotateOrUpgradeDecision(spec, status, context)

	agencyCache, ok := context.GetAgencyCache()
	if !ok {
		// Unable to get agency state, do not restart
		return nil, true
	}

	if decision.IsUpgrade() {
		return r.createUpgradePlanInternalCondition(apiObject, spec, status, context, decision, agencyCache)
	} else if decision.IsUpdate() {
		return r.createUpdatePlanInternalCondition(apiObject, spec, status, decision, context, agencyCache)
	} else {
		upgradeCondition := status.Conditions.IsTrue(api.ConditionTypeUpgradeInProgress)
		updateCondition := status.Conditions.IsTrue(api.ConditionTypeUpdateInProgress)

		if upgradeCondition || updateCondition {
			p := make(api.Plan, 0, 2)

			if upgradeCondition {
				p = append(p, sharedReconcile.RemoveConditionActionV2("Upgrade done", api.ConditionTypeUpgradeInProgress))
			}

			if updateCondition {
				p = append(p, sharedReconcile.RemoveConditionActionV2("Update done", api.ConditionTypeUpdateInProgress))
			}

			return p, false
		}
	}

	return nil, false
}

func (r *Reconciler) createUpdatePlanInternalCondition(apiObject k8sutil.APIObject, spec api.DeploymentSpec, status api.DeploymentStatus, decision updateUpgradeDecisionMap, context PlanBuilderContext, agencyCache state.State) (api.Plan, bool) {
	plan, idle := r.createUpdatePlanInternal(apiObject, spec, status, decision, context, agencyCache)

	if idle || len(plan) > 0 {
		if !status.Conditions.IsTrue(api.ConditionTypeUpdateInProgress) {
			plan = append(api.Plan{
				sharedReconcile.UpdateConditionActionV2("Update in progress", api.ConditionTypeUpdateInProgress, true, "", "", ""),
			}, plan...)
		}
	}

	return plan, idle
}

func (r *Reconciler) createUpdatePlanInternal(apiObject k8sutil.APIObject, spec api.DeploymentSpec, status api.DeploymentStatus, decision updateUpgradeDecisionMap, context PlanBuilderContext, agencyCache state.State) (api.Plan, bool) {
	// Update phase
	for _, m := range status.Members.AsListInGroups(getRotateOrder(spec).Groups()...) {
		d := decision[newUpdateUpgradeDecisionItemFromElement(m)]
		if !d.update {
			continue
		}

		if !d.updateAllowed {
			// Update is not allowed due to constraint
			if !d.unsafeUpdateAllowed {
				r.planLogger.Str("member", m.Member.ID).Str("Reason", d.updateMessage).Info("Pod needs restart but cluster is not ready. Either some shards are not in sync or some member is not ready.")
				continue
			}
			r.planLogger.Str("member", m.Member.ID).Str("Reason", d.updateMessage).Info("Pod needs restart but cluster is not ready. Either some shards are not in sync or some member is not ready, but unsafe upgrade is allowed")
		}

		if m.Member.Conditions.IsTrue(api.ConditionTypeRestart) {
			return r.createRotateMemberPlan(m.Member, m.Group, spec, "Restart flag present", util.CheckConditionalP1Nil(agencyCache.GetRebootID, state.Server(m.Member.ID))), false
		}

		arangoMember, ok := context.ACS().CurrentClusterCache().ArangoMember().V1().GetSimple(m.Member.ArangoMemberName(apiObject.GetName(), m.Group))
		if !ok {
			continue
		}

		cache, ok := context.ACS().ClusterCache(m.Member.ClusterID)
		if !ok {
			continue
		}

		p, ok := cache.Pod().V1().GetSimple(m.Member.Pod.GetName())
		if !ok {
			p = nil
		}

		if svc, ok := cache.Service().V1().GetSimple(arangoMember.GetName()); ok {
			if k8sutil.IsServiceRotationRequired(spec, svc) {
				return api.Plan{
					sharedReconcile.RemoveMemberConditionActionV2("Cleaning update", api.ConditionTypePendingUpdate, m.Group, m.Member.ID),
					sharedReconcile.UpdateMemberConditionActionV2("Cleaning update", api.ConditionTypeUpdating, m.Group, m.Member.ID, true, "Cleaning update", "", ""),
				}, false
			}
		}

		if mode, p, checksum, reason, err := rotation.IsRotationRequired(context.ACS(), spec, m.Member, m.Group, p, arangoMember.Spec.Template, arangoMember.Status.Template); err != nil {
			r.planLogger.Err(err).Str("member", m.Member.ID).Error("Error while generating update plan")
			continue
		} else if mode != compare.InPlaceRotation {
			return api.Plan{
				sharedReconcile.RemoveMemberConditionActionV2(reason, api.ConditionTypePendingUpdate, m.Group, m.Member.ID),
				sharedReconcile.UpdateMemberConditionActionV2(reason, api.ConditionTypeUpdating, m.Group, m.Member.ID, true, reason, "", ""),
			}, false
		} else {
			p = append(p, waitForMemberActions(m.Group, m.Member)...)

			p = append(p, actions.NewAction(api.ActionTypeArangoMemberUpdatePodStatus, m.Group, m.Member, "Propagating status of pod").AddParam(ActionTypeArangoMemberUpdatePodStatusChecksum, checksum))
			p = p.WrapWithPlan(api.Plan{
				sharedReconcile.RemoveMemberConditionActionV2(reason, api.ConditionTypePendingUpdate, m.Group, m.Member.ID),
				sharedReconcile.UpdateMemberConditionActionV2(reason, api.ConditionTypeUpdating, m.Group, m.Member.ID, true, reason, "", ""),
			}, api.Plan{
				sharedReconcile.RemoveMemberConditionActionV2(reason, api.ConditionTypeUpdating, m.Group, m.Member.ID),
			})

			return p, false
		}
	}
	return nil, true
}

func (r *Reconciler) createUpgradePlanInternalCondition(apiObject k8sutil.APIObject, spec api.DeploymentSpec, status api.DeploymentStatus, context PlanBuilderContext, decision updateUpgradeDecisionMap, agencyCache state.State) (api.Plan, bool) {
	plan, idle := r.createUpgradePlanInternal(apiObject, spec, status, context, decision, agencyCache)

	if idle || len(plan) > 0 {
		if !status.Conditions.IsTrue(api.ConditionTypeUpgradeInProgress) {
			plan = append(api.Plan{
				sharedReconcile.UpdateConditionActionV2("Upgrade in progress", api.ConditionTypeUpgradeInProgress, true, "", "", ""),
			}, plan...)
		}
	}

	return plan, idle
}

func (r *Reconciler) createUpgradePlanInternal(apiObject k8sutil.APIObject, spec api.DeploymentSpec, status api.DeploymentStatus, context PlanBuilderContext, decision updateUpgradeDecisionMap, agencyCache state.State) (api.Plan, bool) {
	from, to := decision.GetFromToVersion()
	upgradeOrder := getUpgradeOrder(spec, from, to).Groups()

	for _, group := range upgradeOrder {
		for _, m := range status.Members.AsListInGroup(group) {
			// Pre-check
			d := decision[newUpdateUpgradeDecisionItemFromElement(m)]
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
	}

	// Upgrade phase
	// During upgrade always get first member which needs to be upgraded
	for id, group := range upgradeOrder {
		for _, m := range status.Members.AsListInGroup(group) {
			d := decision[newUpdateUpgradeDecisionItemFromElement(m)]
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

			if m.Member.Conditions.IsTrue(api.ConditionTypeMarkedToRemove) {
				// Server is supposed to be removed, lets not continue
				return nil, false
			}

			if decision.AreGroupsPendingUpgrade(upgradeOrder[:id]...) {
				// Not all members before were upgraded
				return nil, false
			}

			dum := util.BoolSwitch(features.IsUpgradeIndexOrderIssueEnabled(group, d.upgradeDecision.FromVersion, d.upgradeDecision.ToVersion), api.ServerGroupUpgradeModeReplace, api.ServerGroupUpgradeModeInplace)

			um := spec.GetServerGroupSpec(group).UpgradeMode.Default(dum)

			mum := spec.GetServerGroupSpec(group).ManualUpgradeMode.Default(dum)

			switch um.Get() {
			case api.ServerGroupUpgradeModeManual:
				if !m.Member.Conditions.IsTrue(api.ConditionTypeUpgradeAllowed) {
					continue
				}

				// Once we pick manual mode, ensure that we still follow planned scenario
				um = mum
			}

			switch um.Get() {
			case api.ServerGroupUpgradeModeReplace:
				switch group {
				case api.ServerGroupDBServers:
					// Members are supposed to be replaced
					return api.Plan{actions.NewAction(api.ActionTypeMarkToRemoveMember, m.Group, m.Member, "Replace by Upgrade")}, false
				}
			}

			compact := um.Get() == api.ServerGroupUpgradeModeOptionalReplace

			if d.updateAllowed {
				// We are fine, group is alive so we can proceed
				r.planLogger.Str("member", m.Member.ID).Str("Reason", d.updateMessage).Info("Upgrade allowed")
				return r.createUpgradeMemberPlan(m.Member, m.Group, "Version upgrade", spec, status, !d.upgradeDecision.AutoUpgradeNeeded, compact, agencyCache), false
			} else if d.unsafeUpdateAllowed {
				r.planLogger.Str("member", m.Member.ID).Str("Reason", d.updateMessage).Info("Pod needs upgrade but cluster is not ready. Either some shards are not in sync or some member is not ready, but unsafe upgrade is allowed")
				return r.createUpgradeMemberPlan(m.Member, m.Group, "Version upgrade", spec, status, !d.upgradeDecision.AutoUpgradeNeeded, compact, agencyCache), false
			} else {
				r.planLogger.Str("member", m.Member.ID).Str("Reason", d.updateMessage).Info("Pod needs upgrade but cluster is not ready. Either some shards are not in sync or some member is not ready.")
				return nil, true
			}
		}
	}

	r.planLogger.Warn("Pod upgrade plan has been made, but it has been dropped due to missing flag")
	return nil, false
}

// podNeedsUpgrading decides if an upgrade of the pod is needed (to comply with
// the given spec) and if that is allowed.
func (r *Reconciler) podNeedsUpgrading(mode api.DeploymentMode, status api.MemberStatus, spec api.DeploymentSpec, images api.ImageInfoList) upgradeDecision {
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

	// Check if ActiveFailover is supported on new version
	if mode == api.DeploymentModeActiveFailover && !features.ActiveFailover().ImageSupported(&currentImage) {
		// Upgrade to 3.12 in Active Failover mode is not supported!
		// No change
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
		r.planLogger.Str("spec-version", string(specVersion)).Str("pod-version", string(memberVersion)).
			Int("spec-version.major", specVersion.Major()).Int("spec-version.minor", specVersion.Minor()).
			Int("pod-version.major", memberVersion.Major()).Int("pod-version.minor", memberVersion.Minor()).
			Info("Deciding to do a upgrade with --auto-upgrade")
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

func (r *Reconciler) getPodDetails(ctx context.Context, apiObject k8sutil.APIObject, spec api.DeploymentSpec,
	group api.ServerGroup, status api.DeploymentStatus, m api.MemberStatus,
	planCtx PlanBuilderContext) (string, *api.ArangoMember, bool) {
	imageInfo, imageFound := planCtx.SelectImageForMember(spec, status, m)
	if !imageFound {
		// Image is not found, so rotation is not needed
		return "", nil, false
	}

	member, ok := planCtx.ACS().CurrentClusterCache().ArangoMember().V1().GetSimple(m.ArangoMemberName(apiObject.GetName(), group))
	if !ok {
		return "", nil, false
	}

	groupSpec := spec.GetServerGroupSpec(group)

	renderedPod, err := planCtx.RenderPodForMember(ctx, planCtx.ACS(), spec, status, m.ID, imageInfo)
	if err != nil {
		r.planLogger.Err(err).Error("Error while rendering pod")
		return "", nil, false
	}

	checksum, err := resources.ChecksumArangoPod(groupSpec, renderedPod)
	if err != nil {
		r.planLogger.Err(err).Error("Error while getting pod checksum")
		return "", nil, false
	}

	return checksum, member, true
}

// arangoMemberPodTemplateNeedsUpdate returns true when the specification of the
// given pod differs from what it should be according to the
// given deployment spec.
// When true is returned, a reason for the rotation is already returned.
func (r *Reconciler) arangoMemberPodTemplateNeedsUpdate(ctx context.Context, apiObject k8sutil.APIObject, spec api.DeploymentSpec,
	group api.ServerGroup, status api.DeploymentStatus, m api.MemberStatus,
	planCtx PlanBuilderContext) (string, bool) {
	checksum, member, valid := r.getPodDetails(ctx, apiObject, spec, group, status, m, planCtx)
	if valid && !member.Spec.Template.EqualPodSpecChecksum(checksum) {
		return "Pod Spec changed", true
	}

	return "", false
}

// groupReadyForRestart returns true if the cluster is ready for the next update, that is:
//   - all shards are in sync
//   - all members are ready and fine
func groupReadyForRestart(context PlanBuilderContext, status api.DeploymentStatus, member api.MemberStatus, group api.ServerGroup) (bool, string) {
	if group == api.ServerGroupSingle {
		return true, "Restart always in single mode"
	}

	if !status.Conditions.IsTrue(api.ConditionTypeBootstrapCompleted) {
		// Restart is allowed always when bootstrap is not yet completed
		return true, "Bootstrap not completed, restart is allowed"
	}

	members := status.Members.MembersOfGroup(group)

	// If current member did not become ready even once. Kill it
	if !member.Conditions.IsTrue(api.ConditionTypeStarted) {
		return true, "Member is not started"
	}

	// If current core containers are dead kill it.
	if !member.Conditions.IsTrue(api.ConditionTypeServing) {
		return true, "Member is not serving"
	}

	if !members.AllMembersServing() {
		return false, "Not all members are serving"
	}

	if member.Conditions.IsTrue(api.ConditionTypeReady) {
		// Our pod is ready, lets check other pods
		if !members.AllMembersReady() {
			return false, "Not all members are ready"
		}
	}

	switch group {
	case api.ServerGroupDBServers:
		agencyState, ok := context.GetAgencyCache()
		if !ok {
			// Unable to get agency state, do not restart
			return false, "Unable to get agency cache"
		}

		blockingRestartShards := state.GetDBServerBlockingRestartShards(agencyState, state.Server(member.ID))

		if s := len(blockingRestartShards); s > 0 {
			return false, fmt.Sprintf("There are %d shards which are blocking restart", s)
		}
	case api.ServerGroupAgents:
		agencyHealth, ok := context.GetAgencyHealth()
		if !ok {
			// Unable to get agency state, do not restart
			return false, "Unable to get agency cache"
		}

		if err := agencyHealth.Healthy(); err != nil {
			return false, fmt.Sprintf("Restart of agent is not allowed due to: %s", err.Error())
		}
	}

	return true, "Restart allowed"
}

func (r *Reconciler) createMemberAllowUpgradeConditionPlan(ctx context.Context,
	apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	var p api.Plan

	for _, m := range status.Members.AsList() {
		if m.Member.Conditions.IsTrue(api.ConditionTypeUpgradeAllowed) {
			continue
		}

		cache, ok := context.ACS().ClusterCache(m.Member.ClusterID)
		if !ok {
			continue
		}

		if pod, ok := cache.Pod().V1().GetSimple(m.Member.Pod.GetName()); ok {
			if _, ok := pod.GetAnnotations()[deployment.ArangoDeploymentPodAllowUpgradeAnnotation]; ok {
				r.log.
					Str("pod-name", m.Member.Pod.GetName()).
					Str("server-group", m.Group.AsRole()).
					Info("Upgrade Allowed by annotation")

				return api.Plan{
					sharedReconcile.UpdateMemberConditionActionV2("Upgrade allowed by Annotation", api.ConditionTypeUpgradeAllowed, m.Group, m.Member.ID, true, "Upgrade allowed by Annotation", "", ""),
				}
			}
		}
	}

	return p
}

// createUpgradeMemberPlan creates a plan to upgrade (stop-recreateWithAutoUpgrade-stop-start) an existing
// member.
func (r *Reconciler) createUpgradeMemberPlan(member api.MemberStatus,
	group api.ServerGroup, reason string, spec api.DeploymentSpec, status api.DeploymentStatus, rotateStatefull, compact bool, agencyCache state.State) api.Plan {
	upgradeAction := api.ActionTypeUpgradeMember
	if rotateStatefull || group.IsStateless() {
		upgradeAction = api.ActionTypeRotateMember
	}
	r.planLogger.
		Str("id", member.ID).
		Str("role", group.AsRole()).
		Str("reason", reason).
		Str("action", string(upgradeAction)).
		Info("Creating upgrade plan")

	plan := createRotateMemberPlanWithAction(member, group, upgradeAction, spec, reason, util.CheckConditionalP1Nil(agencyCache.GetRebootID, state.Server(member.ID)), compact)

	if member.Image == nil || member.Image.Image != spec.GetImage() {
		plan = plan.Before(actions.NewAction(api.ActionTypeSetMemberCurrentImage, group, member, reason).SetImage(spec.GetImage()))
	}
	if status.CurrentImage == nil || status.CurrentImage.Image != spec.GetImage() {
		plan = plan.Before(actions.NewClusterAction(api.ActionTypeSetCurrentImage, reason).SetImage(spec.GetImage()))
	}

	return plan
}

func withSecureWrap(member api.MemberStatus,
	group api.ServerGroup, spec api.DeploymentSpec, rebootID *int, plan ...api.Action) api.Plan {
	image := member.Image
	if image == nil {
		return plan
	}

	if skipResignLeadership(spec.GetMode(), image.ArangoDBVersion) {
		// In this case we skip resign leadership but we enable maintenance
		return withMaintenanceStart(plan...)
	} else {
		return withResignLeadership(group, member, "ResignLeadership", plan, rebootID)
	}
}

func skipResignLeadership(mode api.DeploymentMode, v driver.Version) bool {
	return mode == api.DeploymentModeCluster && features.Maintenance().Enabled() && ((v.CompareTo("3.6.0") >= 0 && v.CompareTo("3.6.14") <= 0) ||
		(v.CompareTo("3.7.0") >= 0 && v.CompareTo("3.7.12") <= 0))
}

func withWaitForMember(plan api.Plan, group api.ServerGroup, member api.MemberStatus) api.Plan {
	return plan.AfterFirst(func(a api.Action) bool {
		return a.Type != api.ActionTypeAddMember
	}, waitForMemberActions(group, member)...)

}

func waitForMemberActions(group api.ServerGroup, member api.MemberStatus) api.Plan {
	return api.Plan{
		actions.NewAction(api.ActionTypeWaitForMemberUp, group, member, "Wait for member to be up after creation"),
		actions.NewAction(api.ActionTypeWaitForMemberReady, group, member, "Wait for member pod to be ready after creation"),
		actions.NewAction(api.ActionTypeWaitForMemberInSync, group, member, "Wait for member to be in sync after creation"),
	}
}

func getUpgradeOrder(spec api.DeploymentSpec, from, to driver.Version) api.DeploymentSpecOrder {
	if upgrade := spec.Upgrade; upgrade == nil || upgrade.Order == nil {
		if to.CompareTo("3.12.4") >= 0 && from.CompareTo("3.12.4") < 0 && from.CompareTo("3.12.0") >= 0 && from != "" && to != "" {
			return api.DeploymentSpecOrderCoordinatorFirst
		}
	}

	return util.BoolSwitch(features.UpgradeAlternativeOrder().Enabled(), api.DeploymentSpecOrderCoordinatorFirst, spec.Upgrade.GetOrder(util.NewType(api.DeploymentSpecOrderStandard)))
}

func getRotateOrder(spec api.DeploymentSpec) api.DeploymentSpecOrder {
	return spec.Rotate.GetOrder(nil)
}
