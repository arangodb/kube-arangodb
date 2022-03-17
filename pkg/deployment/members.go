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

package deployment

import (
	"context"

	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/names"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

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
func createMember(log zerolog.Logger, status *api.DeploymentStatus, group api.ServerGroup, id string, apiObject *api.ArangoDeployment, mods ...reconcile.CreateMemberMod) (string, error) {
	m, err := renderMember(log, status, group, id, apiObject)
	if err != nil {
		return "", err
	}

	for _, mod := range mods {
		if err := mod(status, group, m); err != nil {
			return "", err
		}
	}

	if err := status.Members.Add(*m, group); err != nil {
		return "", err
	}

	return m.ID, nil
}

func renderMember(log zerolog.Logger, status *api.DeploymentStatus, group api.ServerGroup, id string, apiObject *api.ArangoDeployment) (*api.MemberStatus, error) {
	if group == api.ServerGroupAgents {
		if status.Agency == nil {
			return nil, errors.New("Agency is not yet defined")
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
		return nil, errors.New("Unable to get ID")
	}
	deploymentName := apiObject.GetName()
	role := group.AsRole()
	arch := apiObject.Spec.Architecture.GetDefault()

	switch group {
	case api.ServerGroupSingle:
		log.Debug().Str("id", id).Msg("Adding single server")
		return &api.MemberStatus{
			ID:                        id,
			UID:                       uuid.NewUUID(),
			CreatedAt:                 metav1.Now(),
			Phase:                     api.MemberPhaseNone,
			PersistentVolumeClaimName: k8sutil.CreatePersistentVolumeClaimName(deploymentName, role, id),
			PodName:                   "",
			Image:                     apiObject.Status.CurrentImage,
			Architecture:              &arch,
		}, nil
	case api.ServerGroupAgents:
		log.Debug().Str("id", id).Msg("Adding agent")
		return &api.MemberStatus{
			ID:                        id,
			UID:                       uuid.NewUUID(),
			CreatedAt:                 metav1.Now(),
			Phase:                     api.MemberPhaseNone,
			PersistentVolumeClaimName: k8sutil.CreatePersistentVolumeClaimName(deploymentName, role, id),
			PodName:                   "",
			Image:                     apiObject.Status.CurrentImage,
			Architecture:              &arch,
		}, nil
	case api.ServerGroupDBServers:
		log.Debug().Str("id", id).Msg("Adding dbserver")
		return &api.MemberStatus{
			ID:                        id,
			UID:                       uuid.NewUUID(),
			CreatedAt:                 metav1.Now(),
			Phase:                     api.MemberPhaseNone,
			PersistentVolumeClaimName: k8sutil.CreatePersistentVolumeClaimName(deploymentName, role, id),
			PodName:                   "",
			Image:                     apiObject.Status.CurrentImage,
			Architecture:              &arch,
		}, nil
	case api.ServerGroupCoordinators:
		log.Debug().Str("id", id).Msg("Adding coordinator")
		return &api.MemberStatus{
			ID:                        id,
			UID:                       uuid.NewUUID(),
			CreatedAt:                 metav1.Now(),
			Phase:                     api.MemberPhaseNone,
			PersistentVolumeClaimName: "",
			PodName:                   "",
			Image:                     apiObject.Status.CurrentImage,
			Architecture:              &arch,
		}, nil
	case api.ServerGroupSyncMasters:
		log.Debug().Str("id", id).Msg("Adding syncmaster")
		return &api.MemberStatus{
			ID:                        id,
			UID:                       uuid.NewUUID(),
			CreatedAt:                 metav1.Now(),
			Phase:                     api.MemberPhaseNone,
			PersistentVolumeClaimName: "",
			PodName:                   "",
			Image:                     apiObject.Status.CurrentImage,
			Architecture:              &arch,
		}, nil
	case api.ServerGroupSyncWorkers:
		log.Debug().Str("id", id).Msg("Adding syncworker")
		return &api.MemberStatus{
			ID:                        id,
			UID:                       uuid.NewUUID(),
			CreatedAt:                 metav1.Now(),
			Phase:                     api.MemberPhaseNone,
			PersistentVolumeClaimName: "",
			PodName:                   "",
			Image:                     apiObject.Status.CurrentImage,
			Architecture:              &arch,
		}, nil
	default:
		return nil, errors.WithStack(errors.Newf("Unknown server group %d", group))
	}
}
