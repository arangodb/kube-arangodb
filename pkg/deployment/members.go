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

package deployment

import (
	"fmt"
	"strings"

	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
	"github.com/dchest/uniuri"

	"github.com/arangodb/k8s-operator/pkg/util/k8sutil"
)

// createInitialMembers creates all members needed for the initial state of the deployment.
// Note: This does not create any pods of PVCs
func (d *Deployment) createInitialMembers(apiObject *api.ArangoDeployment) error {
	log := d.deps.Log
	log.Debug().Msg("creating initial members...")

	// Go over all groups and create members
	if err := apiObject.ForeachServerGroup(func(group api.ServerGroup, spec api.ServerGroupSpec, status *api.MemberStatusList) error {
		for len(*status) < spec.Count {
			if err := d.createMember(group, apiObject); err != nil {
				return maskAny(err)
			}
		}
		return nil
	}, &d.status); err != nil {
		return maskAny(err)
	}

	// Save status
	log.Debug().Msg("saving initial members...")
	if err := d.updateCRStatus(); err != nil {
		return maskAny(err)
	}

	return nil
}

// createMember creates member and adds it to the applicable member list.
// Note: This does not create any pods of PVCs
// Note: The updated status is not yet written to the apiserver.
func (d *Deployment) createMember(group api.ServerGroup, apiObject *api.ArangoDeployment) error {
	log := d.deps.Log
	var id string
	for {
		id = strings.ToLower(uniuri.NewLen(8)) // K8s accepts only lowercase, so we use it here as well
		if !d.status.Members.ContainsID(id) {
			break
		}
		// Duplicate, try again
	}
	deploymentName := apiObject.GetName()
	role := group.AsRole()

	switch group {
	case api.ServerGroupSingle:
		log.Debug().Str("id", id).Msg("Adding single server")
		if err := d.status.Members.Single.Add(api.MemberStatus{
			ID:    id,
			State: api.MemberStateNone,
			PersistentVolumeClaimName: k8sutil.CreatePersistentVolumeClaimName(deploymentName, role, id),
			PodName:                   k8sutil.CreatePodName(deploymentName, role, id),
		}); err != nil {
			return maskAny(err)
		}
	case api.ServerGroupAgents:
		log.Debug().Str("id", id).Msg("Adding agent")
		if err := d.status.Members.Agents.Add(api.MemberStatus{
			ID:    id,
			State: api.MemberStateNone,
			PersistentVolumeClaimName: k8sutil.CreatePersistentVolumeClaimName(deploymentName, role, id),
			PodName:                   k8sutil.CreatePodName(deploymentName, role, id),
		}); err != nil {
			return maskAny(err)
		}
	case api.ServerGroupDBServers:
		log.Debug().Str("id", id).Msg("Adding dbserver")
		if err := d.status.Members.DBServers.Add(api.MemberStatus{
			ID:    id,
			State: api.MemberStateNone,
			PersistentVolumeClaimName: k8sutil.CreatePersistentVolumeClaimName(deploymentName, role, id),
			PodName:                   k8sutil.CreatePodName(deploymentName, role, id),
		}); err != nil {
			return maskAny(err)
		}
	case api.ServerGroupCoordinators:
		log.Debug().Str("id", id).Msg("Adding coordinator")
		if err := d.status.Members.Coordinators.Add(api.MemberStatus{
			ID:    id,
			State: api.MemberStateNone,
			PersistentVolumeClaimName: "",
			PodName:                   k8sutil.CreatePodName(deploymentName, role, id),
		}); err != nil {
			return maskAny(err)
		}
	case api.ServerGroupSyncMasters:
		log.Debug().Str("id", id).Msg("Adding syncmaster")
		if err := d.status.Members.SyncMasters.Add(api.MemberStatus{
			ID:    id,
			State: api.MemberStateNone,
			PersistentVolumeClaimName: "",
			PodName:                   k8sutil.CreatePodName(deploymentName, role, id),
		}); err != nil {
			return maskAny(err)
		}
	case api.ServerGroupSyncWorkers:
		log.Debug().Str("id", id).Msg("Adding syncworker")
		if err := d.status.Members.SyncWorkers.Add(api.MemberStatus{
			ID:    id,
			State: api.MemberStateNone,
			PersistentVolumeClaimName: "",
			PodName:                   k8sutil.CreatePodName(deploymentName, role, id),
		}); err != nil {
			return maskAny(err)
		}
	default:
		return maskAny(fmt.Errorf("Unknown server group %d", group))
	}

	return nil
}

// ensurePVCs creates a PVC's listed in member status
func (d *Deployment) ensurePVCs(apiObject *api.ArangoDeployment) error {
	kubecli := d.deps.KubeCli
	deploymentName := apiObject.GetName()
	ns := apiObject.GetNamespace()
	owner := apiObject.AsOwner()
	if err := apiObject.ForeachServerGroup(func(group api.ServerGroup, spec api.ServerGroupSpec, status *api.MemberStatusList) error {
		for _, m := range *status {
			if m.PersistentVolumeClaimName != "" {
				storageClassName := spec.StorageClassName
				role := group.AsRole()
				resources := spec.Resources
				if err := k8sutil.CreatePersistentVolumeClaim(kubecli, m.PersistentVolumeClaimName, deploymentName, ns, storageClassName, role, resources, owner); err != nil {
					return maskAny(err)
				}
			}
		}
		return nil
	}, &d.status); err != nil {
		return maskAny(err)
	}
	return nil
}
