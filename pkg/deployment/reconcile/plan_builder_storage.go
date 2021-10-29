//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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
	"context"

	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
)

// createRotateServerStoragePlan creates plan to rotate a server and its volume because of a
// different storage class or a difference in storage resource requirements.
func createRotateServerStoragePlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	if spec.GetMode() == api.DeploymentModeSingle {
		// Storage cannot be changed in single server deployments
		return nil
	}
	var plan api.Plan
	status.Members.ForeachServerGroup(func(group api.ServerGroup, members api.MemberStatusList) error {
		for _, m := range members {
			if !plan.IsEmpty() {
				// Only 1 change at a time
				continue
			}
			if m.Phase != api.MemberPhaseCreated {
				// Only make changes when phase is created
				continue
			}
			if m.PersistentVolumeClaimName == "" {
				// Plan is irrelevant without PVC
				continue
			}
			groupSpec := spec.GetServerGroupSpec(group)
			storageClassName := groupSpec.GetStorageClassName()

			// Load PVC
			pvc, exists := cachedStatus.PersistentVolumeClaim(m.PersistentVolumeClaimName)
			if !exists {
				log.Warn().
					Str("role", group.AsRole()).
					Str("id", m.ID).
					Msg("Failed to get PVC")
				continue
			}

			if util.StringOrDefault(pvc.Spec.StorageClassName) != storageClassName && storageClassName != "" {
				// Storageclass has changed
				log.Info().Str("pod-name", m.PodName).
					Str("pvc-storage-class", util.StringOrDefault(pvc.Spec.StorageClassName)).
					Str("group-storage-class", storageClassName).Msg("Storage class has changed - pod needs replacement")

				if group == api.ServerGroupDBServers {
					plan = append(plan,
						api.NewAction(api.ActionTypeMarkToRemoveMember, group, m.ID))
				} else if group == api.ServerGroupAgents {
					plan = append(plan,
						api.NewAction(api.ActionTypeKillMemberPod, group, m.ID),
						api.NewAction(api.ActionTypeShutdownMember, group, m.ID),
						api.NewAction(api.ActionTypeRemoveMember, group, m.ID),
						api.NewAction(api.ActionTypeAddMember, group, m.ID),
						api.NewAction(api.ActionTypeWaitForMemberUp, group, m.ID),
					)
				} else {
					// Only agents & dbservers are allowed to change their storage class.
					context.CreateEvent(k8sutil.NewCannotChangeStorageClassEvent(apiObject, m.ID, group.AsRole(), "Not supported"))
				}
			} else {
				var res core.ResourceList
				if groupSpec.HasVolumeClaimTemplate() {
					res = groupSpec.GetVolumeClaimTemplate().Spec.Resources.Requests
				} else {
					res = groupSpec.Resources.Requests
				}
				if requestedSize, ok := res[core.ResourceStorage]; ok {
					if volumeSize, ok := pvc.Spec.Resources.Requests[core.ResourceStorage]; ok {
						cmp := volumeSize.Cmp(requestedSize)
						// Only schrink is possible
						if cmp > 0 {

							if groupSpec.GetVolumeAllowShrink() && group == api.ServerGroupDBServers && !m.Conditions.IsTrue(api.ConditionTypeMarkedToRemove) {
								plan = append(plan, api.NewAction(api.ActionTypeMarkToRemoveMember, group, m.ID))
							} else {
								log.Error().Str("server-group", group.AsRole()).Str("pvc-storage-size", volumeSize.String()).Str("requested-size", requestedSize.String()).
									Msg("Volume size should not shrink")
							}
						}
					}
				}
			}
		}
		return nil
	})

	return plan
}

// createRotateServerStorageResizePlan creates plan to resize storage
func createRotateServerStorageResizePlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	var plan api.Plan

	status.Members.ForeachServerGroup(func(group api.ServerGroup, members api.MemberStatusList) error {
		for _, m := range members {
			if m.Phase != api.MemberPhaseCreated {
				// Only make changes when phase is created
				continue
			}
			if m.PersistentVolumeClaimName == "" {
				// Plan is irrelevant without PVC
				continue
			}
			groupSpec := spec.GetServerGroupSpec(group)

			if !plan.IsEmpty() && groupSpec.VolumeResizeMode.Get() == api.PVCResizeModeRotate {
				// Only 1 change at a time
				return nil
			}

			// Load PVC
			pvc, exists := cachedStatus.PersistentVolumeClaim(m.PersistentVolumeClaimName)
			if !exists {
				log.Warn().
					Str("role", group.AsRole()).
					Str("id", m.ID).
					Msg("Failed to get PVC")
				continue
			}

			var res core.ResourceList
			if groupSpec.HasVolumeClaimTemplate() {
				res = groupSpec.GetVolumeClaimTemplate().Spec.Resources.Requests
			} else {
				res = groupSpec.Resources.Requests
			}
			if requestedSize, ok := res[core.ResourceStorage]; ok {
				if volumeSize, ok := pvc.Spec.Resources.Requests[core.ResourceStorage]; ok {
					cmp := volumeSize.Cmp(requestedSize)
					if cmp < 0 {
						plan = append(plan, pvcResizePlan(log, group, groupSpec, m.ID)...)
					}
				}
			}
		}
		return nil
	})

	return plan
}

func pvcResizePlan(log zerolog.Logger, group api.ServerGroup, groupSpec api.ServerGroupSpec, memberID string) api.Plan {
	mode := groupSpec.VolumeResizeMode.Get()
	switch mode {
	case api.PVCResizeModeRuntime:
		return api.Plan{
			api.NewAction(api.ActionTypePVCResize, group, memberID),
		}
	case api.PVCResizeModeRotate:
		return api.Plan{
			api.NewAction(api.ActionTypeResignLeadership, group, memberID),
			api.NewAction(api.ActionTypeKillMemberPod, group, memberID),
			api.NewAction(api.ActionTypeRotateStartMember, group, memberID),
			api.NewAction(api.ActionTypePVCResize, group, memberID),
			api.NewAction(api.ActionTypePVCResized, group, memberID),
			api.NewAction(api.ActionTypeRotateStopMember, group, memberID),
			api.NewAction(api.ActionTypeWaitForMemberUp, group, memberID),
		}
	default:
		log.Error().Str("server-group", group.AsRole()).Str("mode", mode.String()).
			Msg("Requested mode is not supported")
		return nil
	}
}
