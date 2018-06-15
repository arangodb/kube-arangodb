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
	"github.com/rs/zerolog"
	"k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// createRotateServerStoragePlan creates plan to rotate a server and its volume because of a
// different storage class or a difference in storage resource requirements.
func createRotateServerStoragePlan(log zerolog.Logger, spec api.DeploymentSpec, status api.DeploymentStatus,
	getPVC func(pvcName string) (*v1.PersistentVolumeClaim, error)) api.Plan {
	if spec.GetMode() == api.DeploymentModeSingle {
		// Storage cannot be changed in single server deployments
		return nil
	}
	var plan api.Plan
	status.Members.ForeachServerGroup(func(group api.ServerGroup, members api.MemberStatusList) error {
		for _, m := range members {
			if len(plan) > 0 {
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
			if group == api.ServerGroupSyncWorkers {
				// SyncWorkers have no externally created TLS keyfile
				continue
			}
			groupSpec := spec.GetServerGroupSpec(group)
			storageClassName := groupSpec.GetStorageClassName()
			if storageClassName == "" {
				// Using default storage class name
				continue
			}
			// Load PVC
			pvc, err := getPVC(m.PersistentVolumeClaimName)
			if err != nil {
				log.Warn().Err(err).
					Str("role", group.AsRole()).
					Str("id", m.ID).
					Msg("Failed to get PVC")
				continue
			}
			replacementNeeded := false
			if util.StringOrDefault(pvc.Spec.StorageClassName) != storageClassName {
				// Storageclass has changed
				replacementNeeded = true
			}
			if replacementNeeded {
				if group != api.ServerGroupAgents {
					plan = append(plan,
						// Scale up, so we're sure that a new member is available
						api.NewAction(api.ActionTypeAddMember, group, ""),
						api.NewAction(api.ActionTypeWaitForMemberUp, group, api.MemberIDPreviousAction),
					)
				}
				if group == api.ServerGroupDBServers {
					plan = append(plan,
						// Cleanout
						api.NewAction(api.ActionTypeCleanOutMember, group, m.ID),
					)
				}
				plan = append(plan,
					// Shutdown & remove the server
					api.NewAction(api.ActionTypeShutdownMember, group, m.ID),
					api.NewAction(api.ActionTypeRemoveMember, group, m.ID),
				)
				if group == api.ServerGroupAgents {
					plan = append(plan,
						// Scale up, so we're adding the remove agent
						api.NewAction(api.ActionTypeAddMember, group, ""),
						api.NewAction(api.ActionTypeWaitForMemberUp, group, api.MemberIDPreviousAction),
					)
				}
			}
		}
		return nil
	})
	return plan
}
