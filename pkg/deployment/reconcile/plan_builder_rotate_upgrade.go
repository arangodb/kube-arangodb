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
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"reflect"
	"strings"

	"github.com/arangodb/go-driver"
	upgraderules "github.com/arangodb/go-upgrade-rules"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// createRotateOrUpgradePlan goes over all pods to check if an upgrade or rotate is needed.
func createRotateOrUpgradePlan(log zerolog.Logger, apiObject k8sutil.APIObject, spec api.DeploymentSpec,
	status api.DeploymentStatus, context PlanBuilderContext, pods []core.Pod) api.Plan {

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

			pod, found := k8sutil.GetPodByName(pods, m.PodName)
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
				// Upgrade is not needed, see if rotation is needed
				rotNeeded, reason := podNeedsRotation(log, pod, apiObject, spec, group, status, m.ID, context)
				if rotNeeded {
					newPlan = createRotateMemberPlan(log, m, group, reason)
				}
			}
		}
		return nil
	})

	if upgradeNotAllowed {
		context.CreateEvent(k8sutil.NewUpgradeNotAllowedEvent(apiObject, fromVersion, toVersion, fromLicense, toLicense))
	} else if !newPlan.IsEmpty() {
		if clusterReadyForUpgrade(context) {
			// Use the new plan
			return newPlan
		} else {
			log.Info().Msg("Pod needs upgrade but cluster is not ready. Either some shards are not in sync or some member is not ready.")
		}
	}
	return nil
}

// podNeedsUpgrading decides if an upgrade of the pod is needed (to comply with
// the given spec) and if that is allowed.
func podNeedsUpgrading(log zerolog.Logger, p core.Pod, spec api.DeploymentSpec, images api.ImageInfoList) upgradeDecision {
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
func podNeedsRotation(log zerolog.Logger, p core.Pod, apiObject metav1.Object, spec api.DeploymentSpec,
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

	if group.IsExportMetrics() {
		e, hasExporter := k8sutil.GetContainerByName(&p, k8sutil.ExporterContainerName)

		if spec.Metrics.IsEnabled() {
			if !hasExporter {
				return true, "Exporter configuration changed"
			}

			if spec.Metrics.HasImage() {
				if e.Image != spec.Metrics.GetImage() {
					return true, "Exporter image changed"
				}
			}

			if k8sutil.IsResourceRequirementsChanged(spec.Metrics.Resources, e.Resources) {
				return true, "Resources requirements have been changed for exporter"
			}
		} else if hasExporter {
			return true, "Exporter was disabled"
		}
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

	// Check priorities
	if groupSpec.PriorityClassName != p.Spec.PriorityClassName {
		return true, "Pod priority changed"
	}

	// Check resource requirements
	var resources core.ResourceRequirements
	if groupSpec.HasVolumeClaimTemplate() {
		resources = groupSpec.Resources // If there is a volume claim template compare all resources
	} else {
		resources = k8sutil.ExtractPodResourceRequirement(groupSpec.Resources)
	}

	if k8sutil.IsResourceRequirementsChanged(resources, k8sutil.GetArangoDBContainerFromPod(&p).Resources) {
		return true, "Resource Requirements changed"
	}

	var memberStatus, _, _ = status.Members.MemberStatusByPodName(p.GetName())
	if memberStatus.SideCarSpecs == nil {
		memberStatus.SideCarSpecs = make(map[string]core.Container)
	}

	// Check for missing side cars in
	for _, specSidecar := range groupSpec.GetSidecars() {
		var stateSidecar core.Container
		if stateSidecar, found = memberStatus.SideCarSpecs[specSidecar.Name]; !found {
			return true, "Sidecar " + specSidecar.Name + " not found in running pod " + p.GetName()
		}
		if sideCarRequireRotation(specSidecar.DeepCopy(), &stateSidecar) {
			return true, "Sidecar " + specSidecar.Name + " requires rotation"
		}
	}

	for name := range memberStatus.SideCarSpecs {
		var found = false
		for _, specSidecar := range groupSpec.GetSidecars() {
			if name == specSidecar.Name {
				found = true
				break
			}
		}
		if !found {
			return true, "Sidecar " + name + " no longer in specification"
		}
	}

	// Check for probe changes

	// Readiness
	if rotate, reason := compareProbes(pod.ReadinessSpec(group),
		groupSpec.GetProbesSpec().GetReadinessProbeDisabled(),
		groupSpec.GetProbesSpec().ReadinessProbeSpec,
		c.ReadinessProbe); rotate {
		return rotate, reason
	}

	// Liveness
	if rotate, reason := compareProbes(pod.LivenessSpec(group),
		groupSpec.GetProbesSpec().LivenessProbeDisabled,
		groupSpec.GetProbesSpec().LivenessProbeSpec,
		c.LivenessProbe); rotate {
		return rotate, reason
	}

	return false, ""
}

func compareProbes(probe pod.Probe, groupProbeDisabled *bool, groupProbeSpec *api.ServerGroupProbeSpec, containerProbe *core.Probe) (bool, string) {
	if !probe.CanBeEnabled {
		if containerProbe != nil {
			return true, "Probe needs to be disabled"
		}
		// Nothing to do - we cannot enable probe and probe is disabled
		return false, ""
	}

	enabled := probe.EnabledByDefault

	if groupProbeDisabled != nil {
		enabled = !*groupProbeDisabled
	}

	if enabled && containerProbe == nil {
		// We expected probe to be enabled but it is disabled in container
		return true, "Enabling probe"
	}

	if !enabled {
		if containerProbe != nil {
			// We expected probe to be disabled but it is enabled in container
			return true, "Disabling probe"
		}

		// Nothing to do - probe is disabled and it should stay as it is
		return false, ""
	}

	if groupProbeSpec.GetTimeoutSeconds(containerProbe.TimeoutSeconds) != containerProbe.TimeoutSeconds {
		return true, "Timeout seconds does not match"
	}

	if groupProbeSpec.GetPeriodSeconds(containerProbe.PeriodSeconds) != containerProbe.PeriodSeconds {
		return true, "Period seconds does not match"
	}

	// Recreate probe if timeout seconds are different
	if groupProbeSpec.GetInitialDelaySeconds(containerProbe.InitialDelaySeconds) != containerProbe.InitialDelaySeconds {
		return true, "Initial delay seconds does not match"
	}

	// Recreate probe if timeout seconds are different
	if groupProbeSpec.GetFailureThreshold(containerProbe.FailureThreshold) != containerProbe.FailureThreshold {
		return true, "Failure threshold does not match"
	}

	// Recreate probe if timeout seconds are different
	if groupProbeSpec.GetSuccessThreshold(containerProbe.SuccessThreshold) != containerProbe.SuccessThreshold {
		return true, "Success threshold does not match"
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

// sideCarRequireRotation checks if side car requires rotation including default parameters
func sideCarRequireRotation(wanted, given *core.Container) bool {
	return !reflect.DeepEqual(wanted, given)
}

// normalizeServiceAccountName replaces default with empty string, otherwise returns the input.
func normalizeServiceAccountName(name string) string {
	if name == "default" {
		return ""
	}
	return ""
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

func getContainerArgs(c core.Container) []string {
	if len(c.Command) >= 1 {
		return c.Command[1:]
	}
	return c.Args
}
