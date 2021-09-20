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
	"context"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/names"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// createInitialMembers creates all members needed for the initial state of the deployment.
// Note: This does not create any pods of PVCs
func (d *Deployment) createInitialMembers(ctx context.Context, apiObject *api.ArangoDeployment) error {
	log := d.deps.Log
	log.Debug().Msg("creating initial members...")

	// Go over all groups and create members
	var events []*k8sutil.Event
	status, lastVersion := d.GetStatus()
	if err := apiObject.ForeachServerGroup(func(group api.ServerGroup, spec api.ServerGroupSpec, members *api.MemberStatusList) error {
		for len(*members) < spec.GetCount() {
			id, err := createMember(log, &status, group, "", apiObject)
			if err != nil {
				return errors.WithStack(err)
			}
			events = append(events, k8sutil.NewMemberAddEvent(id, group.AsRole(), apiObject))
		}
		return nil
	}, &status); err != nil {
		return errors.WithStack(err)
	}

	// Save status
	log.Debug().Msg("saving initial members...")
	if err := d.UpdateStatus(ctx, status, lastVersion); err != nil {
		return errors.WithStack(err)
	}
	// Save events
	for _, evt := range events {
		d.CreateEvent(evt)
	}

	return nil
}

func (d *Deployment) createAgencyMapping(ctx context.Context) error {
	spec := d.GetSpec()
	status, _ := d.GetStatus()

	if !spec.Mode.HasAgents() {
		return nil
	}

	if status.Agency != nil {
		return nil
	}

	var i api.DeploymentStatusAgencyInfo

	if spec.Agents.Count == nil {
		return nil
	}

	agents := status.Members.Agents

	if len(agents) > *spec.Agents.Count {
		return errors.Newf("Agency size is bigger than requested size")
	}

	c := api.DeploymentStatusAgencySize(*spec.Agents.Count)

	i.Size = &c

	for id := range agents {
		i.IDs = append(i.IDs, agents[id].ID)
	}

	for len(i.IDs) < *spec.Agents.Count {
		i.IDs = append(i.IDs, names.GetArangodID(api.ServerGroupAgents))
	}

	return d.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		s.Agency = &i
		return true
	})
}

// createMember creates member and adds it to the applicable member list.
// Note: This does not create any pods of PVCs
// Note: The updated status is not yet written to the apiserver.
func createMember(log zerolog.Logger, status *api.DeploymentStatus, group api.ServerGroup, id string, apiObject *api.ArangoDeployment) (string, error) {
	if group == api.ServerGroupAgents {
		if status.Agency == nil {
			return "", errors.New("Agency is not yet defined")
		}
		// In case of agents we need to use hardcoded ids
		if id == "" {
			for _, nid := range status.Agency.IDs {
				if !status.Members.ContainsID(nid) {
					id = nid
					break
				}
			}
		}
	} else {
		if id == "" {
			for {
				id = names.GetArangodID(group)
				if !status.Members.ContainsID(id) {
					break
				}
				// Duplicate, try again
			}
		}
	}
	if id == "" {
		return "nil", errors.New("Unable to get ID")
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
			Image:                     apiObject.Status.CurrentImage,
		}, group); err != nil {
			return "", errors.WithStack(err)
		}
	case api.ServerGroupAgents:
		log.Debug().Str("id", id).Msg("Adding agent")
		if err := status.Members.Add(api.MemberStatus{
			ID:                        id,
			CreatedAt:                 metav1.Now(),
			Phase:                     api.MemberPhaseNone,
			PersistentVolumeClaimName: k8sutil.CreatePersistentVolumeClaimName(deploymentName, role, id),
			PodName:                   "",
			Image:                     apiObject.Status.CurrentImage,
		}, group); err != nil {
			return "", errors.WithStack(err)
		}
	case api.ServerGroupDBServers:
		log.Debug().Str("id", id).Msg("Adding dbserver")
		if err := status.Members.Add(api.MemberStatus{
			ID:                        id,
			CreatedAt:                 metav1.Now(),
			Phase:                     api.MemberPhaseNone,
			PersistentVolumeClaimName: k8sutil.CreatePersistentVolumeClaimName(deploymentName, role, id),
			PodName:                   "",
			Image:                     apiObject.Status.CurrentImage,
		}, group); err != nil {
			return "", errors.WithStack(err)
		}
	case api.ServerGroupCoordinators:
		log.Debug().Str("id", id).Msg("Adding coordinator")
		if err := status.Members.Add(api.MemberStatus{
			ID:                        id,
			CreatedAt:                 metav1.Now(),
			Phase:                     api.MemberPhaseNone,
			PersistentVolumeClaimName: "",
			PodName:                   "",
			Image:                     apiObject.Status.CurrentImage,
		}, group); err != nil {
			return "", errors.WithStack(err)
		}
	case api.ServerGroupSyncMasters:
		log.Debug().Str("id", id).Msg("Adding syncmaster")
		if err := status.Members.Add(api.MemberStatus{
			ID:                        id,
			CreatedAt:                 metav1.Now(),
			Phase:                     api.MemberPhaseNone,
			PersistentVolumeClaimName: "",
			PodName:                   "",
			Image:                     apiObject.Status.CurrentImage,
		}, group); err != nil {
			return "", errors.WithStack(err)
		}
	case api.ServerGroupSyncWorkers:
		log.Debug().Str("id", id).Msg("Adding syncworker")
		if err := status.Members.Add(api.MemberStatus{
			ID:                        id,
			CreatedAt:                 metav1.Now(),
			Phase:                     api.MemberPhaseNone,
			PersistentVolumeClaimName: "",
			PodName:                   "",
			Image:                     apiObject.Status.CurrentImage,
		}, group); err != nil {
			return "", errors.WithStack(err)
		}
	default:
		return "", errors.WithStack(errors.Newf("Unknown server group %d", group))
	}

	return id, nil
}
