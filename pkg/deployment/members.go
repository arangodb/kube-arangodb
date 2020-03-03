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
// Author Ewout Prangsma
//

package deployment

import (
	"fmt"
	"strings"

	"github.com/dchest/uniuri"
	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// createInitialMembers creates all members needed for the initial state of the deployment.
// Note: This does not create any pods of PVCs
func (d *Deployment) createInitialMembers(apiObject *api.ArangoDeployment) error {
	log := d.deps.Log
	log.Debug().Msg("creating initial members...")

	// Go over all groups and create members
	var events []*k8sutil.Event
	status, lastVersion := d.GetStatus()
	if err := apiObject.ForeachServerGroup(func(group api.ServerGroup, spec api.ServerGroupSpec, members *api.MemberStatusList) error {
		for len(*members) < spec.GetCount() {
			id, err := createMember(log, &status, group, "", apiObject)
			if err != nil {
				return maskAny(err)
			}
			events = append(events, k8sutil.NewMemberAddEvent(id, group.AsRole(), apiObject))
		}
		return nil
	}, &status); err != nil {
		return maskAny(err)
	}

	// Save status
	log.Debug().Msg("saving initial members...")
	if err := d.UpdateStatus(status, lastVersion); err != nil {
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
func createMember(log zerolog.Logger, status *api.DeploymentStatus, group api.ServerGroup, id string, apiObject *api.ArangoDeployment) (string, error) {
	if id == "" {
		idPrefix := getArangodIDPrefix(group)
		for {
			id = idPrefix + strings.ToLower(uniuri.NewLen(8)) // K8s accepts only lowercase, so we use it here as well
			if !status.Members.ContainsID(id) {
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
		if err := status.Members.Add(api.MemberStatus{
			ID:                        id,
			CreatedAt:                 metav1.Now(),
			Phase:                     api.MemberPhaseNone,
			PersistentVolumeClaimName: k8sutil.CreatePersistentVolumeClaimName(deploymentName, role, id),
			PodName:                   "",
		}, group); err != nil {
			return "", maskAny(err)
		}
	case api.ServerGroupAgents:
		log.Debug().Str("id", id).Msg("Adding agent")
		if err := status.Members.Add(api.MemberStatus{
			ID:                        id,
			CreatedAt:                 metav1.Now(),
			Phase:                     api.MemberPhaseNone,
			PersistentVolumeClaimName: k8sutil.CreatePersistentVolumeClaimName(deploymentName, role, id),
			PodName:                   "",
		}, group); err != nil {
			return "", maskAny(err)
		}
	case api.ServerGroupDBServers:
		log.Debug().Str("id", id).Msg("Adding dbserver")
		if err := status.Members.Add(api.MemberStatus{
			ID:                        id,
			CreatedAt:                 metav1.Now(),
			Phase:                     api.MemberPhaseNone,
			PersistentVolumeClaimName: k8sutil.CreatePersistentVolumeClaimName(deploymentName, role, id),
			PodName:                   "",
		}, group); err != nil {
			return "", maskAny(err)
		}
	case api.ServerGroupCoordinators:
		log.Debug().Str("id", id).Msg("Adding coordinator")
		if err := status.Members.Add(api.MemberStatus{
			ID:                        id,
			CreatedAt:                 metav1.Now(),
			Phase:                     api.MemberPhaseNone,
			PersistentVolumeClaimName: "",
			PodName:                   "",
		}, group); err != nil {
			return "", maskAny(err)
		}
	case api.ServerGroupSyncMasters:
		log.Debug().Str("id", id).Msg("Adding syncmaster")
		if err := status.Members.Add(api.MemberStatus{
			ID:                        id,
			CreatedAt:                 metav1.Now(),
			Phase:                     api.MemberPhaseNone,
			PersistentVolumeClaimName: "",
			PodName:                   "",
		}, group); err != nil {
			return "", maskAny(err)
		}
	case api.ServerGroupSyncWorkers:
		log.Debug().Str("id", id).Msg("Adding syncworker")
		if err := status.Members.Add(api.MemberStatus{
			ID:                        id,
			CreatedAt:                 metav1.Now(),
			Phase:                     api.MemberPhaseNone,
			PersistentVolumeClaimName: "",
			PodName:                   "",
		}, group); err != nil {
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
	case api.ServerGroupSingle:
		return "SNGL-"
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
