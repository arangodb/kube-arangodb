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

	"github.com/dchest/uniuri"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// createInitialMembers creates all members needed for the initial state of the deployment.
// Note: This does not create any pods of PVCs
func (d *Deployment) createInitialMembers(apiObject *api.ArangoDeployment) error {
	log := d.deps.Log
	log.Debug().Msg("creating initial members...")

	// Go over all groups and create members
	var events []*v1.Event
	if err := apiObject.ForeachServerGroup(func(group api.ServerGroup, spec api.ServerGroupSpec, status *api.MemberStatusList) error {
		for len(*status) < spec.GetCount() {
			id, err := d.createMember(group, "", apiObject)
			if err != nil {
				return maskAny(err)
			}
			events = append(events, k8sutil.NewMemberAddEvent(id, group.AsRole(), apiObject))
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
	// Save events
	for _, evt := range events {
		d.CreateEvent(evt)
	}

	return nil
}

// createMember creates member and adds it to the applicable member list.
// Note: This does not create any pods of PVCs
// Note: The updated status is not yet written to the apiserver.
func (d *Deployment) createMember(group api.ServerGroup, id string, apiObject *api.ArangoDeployment) (string, error) {
	log := d.deps.Log
	if id == "" {
		idPrefix := getArangodIDPrefix(group)
		for {
			id = idPrefix + strings.ToLower(uniuri.NewLen(8)) // K8s accepts only lowercase, so we use it here as well
			if !d.status.Members.ContainsID(id) {
				break
			}
			// Duplicate, try again
		}
	}
	deploymentName := apiObject.GetName()
	role := group.AsRole()

	switch group {
	case api.ServerGroupSingle:
		log.Debug().Str("id", id).Msg("Adding single server")
		if err := d.status.Members.Single.Add(api.MemberStatus{
			ID:        id,
			CreatedAt: metav1.Now(),
			Phase:     api.MemberPhaseNone,
			PersistentVolumeClaimName: k8sutil.CreatePersistentVolumeClaimName(deploymentName, role, id),
			PodName:                   "",
		}); err != nil {
			return "", maskAny(err)
		}
	case api.ServerGroupAgents:
		log.Debug().Str("id", id).Msg("Adding agent")
		if err := d.status.Members.Agents.Add(api.MemberStatus{
			ID:        id,
			CreatedAt: metav1.Now(),
			Phase:     api.MemberPhaseNone,
			PersistentVolumeClaimName: k8sutil.CreatePersistentVolumeClaimName(deploymentName, role, id),
			PodName:                   "",
		}); err != nil {
			return "", maskAny(err)
		}
	case api.ServerGroupDBServers:
		log.Debug().Str("id", id).Msg("Adding dbserver")
		if err := d.status.Members.DBServers.Add(api.MemberStatus{
			ID:        id,
			CreatedAt: metav1.Now(),
			Phase:     api.MemberPhaseNone,
			PersistentVolumeClaimName: k8sutil.CreatePersistentVolumeClaimName(deploymentName, role, id),
			PodName:                   "",
		}); err != nil {
			return "", maskAny(err)
		}
	case api.ServerGroupCoordinators:
		log.Debug().Str("id", id).Msg("Adding coordinator")
		if err := d.status.Members.Coordinators.Add(api.MemberStatus{
			ID:        id,
			CreatedAt: metav1.Now(),
			Phase:     api.MemberPhaseNone,
			PersistentVolumeClaimName: "",
			PodName:                   "",
		}); err != nil {
			return "", maskAny(err)
		}
	case api.ServerGroupSyncMasters:
		log.Debug().Str("id", id).Msg("Adding syncmaster")
		if err := d.status.Members.SyncMasters.Add(api.MemberStatus{
			ID:        id,
			CreatedAt: metav1.Now(),
			Phase:     api.MemberPhaseNone,
			PersistentVolumeClaimName: "",
			PodName:                   "",
		}); err != nil {
			return "", maskAny(err)
		}
	case api.ServerGroupSyncWorkers:
		log.Debug().Str("id", id).Msg("Adding syncworker")
		if err := d.status.Members.SyncWorkers.Add(api.MemberStatus{
			ID:        id,
			CreatedAt: metav1.Now(),
			Phase:     api.MemberPhaseNone,
			PersistentVolumeClaimName: "",
			PodName:                   "",
		}); err != nil {
			return "", maskAny(err)
		}
	default:
		return "", maskAny(fmt.Errorf("Unknown server group %d", group))
	}

	return id, nil
}

// getArangodIDPrefix returns the prefix required ID's of arangod servers
// in the given group.
func getArangodIDPrefix(group api.ServerGroup) string {
	switch group {
	case api.ServerGroupCoordinators:
		return "CRDN-"
	case api.ServerGroupDBServers:
		return "PRMR-"
	case api.ServerGroupAgents:
		return "AGNT-"
	default:
		return ""
	}
}
