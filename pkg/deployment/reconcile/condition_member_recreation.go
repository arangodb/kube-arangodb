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
	"fmt"
	"strings"

	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
)

func createMemberRecreationConditionsPlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	var p api.Plan

	for _, m := range status.Members.AsList() {
		message, recreate := EvaluateMemberRecreationCondition(ctx, log, apiObject, spec, status, m.Group, m.Member,
			cachedStatus, context, isStorageClassChanged, isVolumeSizeChanged)

		if !recreate {
			if _, ok := m.Member.Conditions.Get(api.MemberReplacementRequired); ok {
				// Unset condition
				p = append(p, removeMemberConditionActionV2("Member replacement not required", api.MemberReplacementRequired, m.Group, m.Member.ID))
			}
		} else {
			if c, ok := m.Member.Conditions.Get(api.MemberReplacementRequired); !ok || !c.IsTrue() || c.Message != message {
				// Update condition
				p = append(p, updateMemberConditionActionV2("Member replacement required", api.MemberReplacementRequired, m.Group, m.Member.ID, true, "Member replacement required", message, ""))
			}
		}
	}

	return p
}

type MemberRecreationConditionEvaluator func(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	group api.ServerGroup, member api.MemberStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) (bool, string, error)

func EvaluateMemberRecreationCondition(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	group api.ServerGroup, member api.MemberStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext, evaluators ...MemberRecreationConditionEvaluator) (string, bool) {
	args := make([]string, 0, len(evaluators))

	for _, e := range evaluators {
		ok, s, err := e(ctx, log, apiObject, spec, status, group, member, cachedStatus, context)
		if err != nil {
			// When one of an evaluator requires pod's replacement then it should be done.
			continue
		}

		if ok {
			args = append(args, s)
		}
	}

	return strings.Join(args, ", "), len(args) > 0
}

// isStorageClassChanged returns true and reason when the member should be replaced.
func isStorageClassChanged(_ context.Context, log zerolog.Logger, apiObject k8sutil.APIObject, spec api.DeploymentSpec,
	_ api.DeploymentStatus, group api.ServerGroup, member api.MemberStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) (bool, string, error) {
	if spec.GetMode() == api.DeploymentModeSingle {
		// Storage cannot be changed in single server deployments.
		return false, "", nil
	}

	if member.Phase != api.MemberPhaseCreated {
		// Only make changes when phase is created.
		return false, "", nil
	}

	if member.PersistentVolumeClaimName == "" {
		// Plan is irrelevant without PVC.
		return false, "", nil
	}

	groupSpec := spec.GetServerGroupSpec(group)
	storageClassName := groupSpec.GetStorageClassName()
	if storageClassName == "" {
		// A storage class is not set.
		return false, "", nil
	}

	// Check if a storage class changed.
	if pvc, ok := cachedStatus.PersistentVolumeClaim(member.PersistentVolumeClaimName); !ok {
		log.Warn().Str("role", group.AsRole()).Str("id", member.ID).Msg("Failed to get PVC")
		return false, "", fmt.Errorf("failed to get PVC %s", member.PersistentVolumeClaimName)
	} else {
		pvcClassName := util.StringOrDefault(pvc.Spec.StorageClassName)
		if pvcClassName == storageClassName {
			// A storage class has not been changed.
			return false, "", nil
		}
		if pvcClassName == "" {
			// TODO what to do here?
			return false, "", nil
		}
	}

	// From here on, it is known that a storage class has changed.
	if group != api.ServerGroupDBServers && group != api.ServerGroupAgents {
		// Only agents & DB servers are allowed to change their storage class.
		context.CreateEvent(k8sutil.NewCannotChangeStorageClassEvent(apiObject, member.ID, group.AsRole(), "Not supported"))
		return false, "", nil
	}

	// From here on it is known that the member requires replacement, so `true` must be returned.
	// If pod does not exist then it will try next time.
	if pod, ok := cachedStatus.Pod(member.PodName); ok {
		if _, ok := pod.GetAnnotations()[deployment.ArangoDeploymentPodReplaceAnnotation]; !ok {
			log.Warn().
				Str("pod-name", member.PodName).
				Str("server-group", group.AsRole()).
				Msgf("try changing a storage class name, but %s", getRequiredReplaceMessage(member.PodName))
			// No return here.
		}
	} else {
		return false, "", fmt.Errorf("failed to get pod %s", member.PodName)
	}

	return true, "Storage class has changed", nil
}

// isVolumeSizeChanged returns true and reason when the member should be replaced.
func isVolumeSizeChanged(_ context.Context, log zerolog.Logger, _ k8sutil.APIObject, spec api.DeploymentSpec,
	_ api.DeploymentStatus, group api.ServerGroup, member api.MemberStatus,
	cachedStatus inspectorInterface.Inspector, _ PlanBuilderContext) (bool, string, error) {
	if spec.GetMode() == api.DeploymentModeSingle {
		// Storage cannot be changed in single server deployments.
		return false, "", nil
	}

	if member.Phase != api.MemberPhaseCreated {
		// Only make changes when phase is created.
		return false, "", nil
	}

	if member.PersistentVolumeClaimName == "" {
		// Plan is irrelevant without PVC.
		return false, "", nil
	}

	pvc, ok := cachedStatus.PersistentVolumeClaim(member.PersistentVolumeClaimName)
	if !ok {
		log.Warn().
			Str("role", group.AsRole()).
			Str("id", member.ID).
			Msg("Failed to get PVC")

		return false, "", fmt.Errorf("failed to get PVC %s", member.PersistentVolumeClaimName)
	}

	groupSpec := spec.GetServerGroupSpec(group)
	ok, volumeSize, requestedSize := shouldVolumeResize(groupSpec, pvc)
	if !ok {
		return false, "", nil
	}

	if group != api.ServerGroupDBServers {
		log.Error().
			Str("pvc-storage-size", volumeSize.String()).
			Str("requested-size", requestedSize.String()).
			Msgf("Volume size should not shrink, because it is not possible for \"%s\"", group.AsRole())

		return false, "", nil
	}

	// From here on it is known that the member requires replacement, so `true` must be returned.
	// If pod does not exist then it will try next time.
	if pod, ok := cachedStatus.Pod(member.PodName); ok {
		if _, ok := pod.GetAnnotations()[deployment.ArangoDeploymentPodReplaceAnnotation]; !ok {
			log.Warn().Str("pod-name", member.PodName).
				Msgf("try shrinking volume size, but %s", getRequiredReplaceMessage(member.PodName))
			// No return here.
		}
	} else {
		return false, "", fmt.Errorf("failed to get pod %s", member.PodName)
	}

	return true, "Volume is shrunk", nil
}

// shouldVolumeResize returns false when a volume should not resize.
// Currently, it is only possible to shrink a volume size.
// When return true then the actual and required volume size are returned.
func shouldVolumeResize(groupSpec api.ServerGroupSpec,
	pvc *core.PersistentVolumeClaim) (bool, resource.Quantity, resource.Quantity) {
	var res core.ResourceList
	if groupSpec.HasVolumeClaimTemplate() {
		res = groupSpec.GetVolumeClaimTemplate().Spec.Resources.Requests
	} else {
		res = groupSpec.Resources.Requests
	}

	if requestedSize, ok := res[core.ResourceStorage]; ok {
		if volumeSize, ok := pvc.Spec.Resources.Requests[core.ResourceStorage]; ok {
			if volumeSize.Cmp(requestedSize) > 0 {
				// The actual PVC's volume size is greater than requested size, so it can be shrunk to the requested size.
				return true, volumeSize, requestedSize
			}
		}
	}

	return false, resource.Quantity{}, resource.Quantity{}
}

func getRequiredReplaceMessage(podName string) string {
	return fmt.Sprintf("%s annotation is required to be set on the pod %s",
		deployment.ArangoDeploymentPodReplaceAnnotation, podName)
}
