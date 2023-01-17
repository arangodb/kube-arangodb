//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func (d *Deployment) createAgencyMapping(ctx context.Context) error {
	spec := d.GetSpec()
	status := d.GetStatus()

	if !spec.Mode.Get().HasAgents() {
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

	agentsStatus := status.GetServerGroupStatus(api.ServerGroupAgents)

	for len(i.IDs) < *spec.Agents.Count {
		i.IDs = append(i.IDs, d.renderMemberID(spec, &status, &agentsStatus, api.ServerGroupAgents))
	}

	return d.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		s.Agency = &i
		s.UpdateServerGroupStatus(api.ServerGroupAgents, agentsStatus)
		return true
	})
}

// createMember creates member and adds it to the applicable member list.
// Note: This does not create any pods of PVCs
// Note: The updated status is not yet written to the apiserver.
func (d *Deployment) createMember(spec api.DeploymentSpec, status *api.DeploymentStatus, group api.ServerGroup, id string, apiObject *api.ArangoDeployment, mods ...reconcile.CreateMemberMod) (string, error) {
	gs := status.GetServerGroupStatus(group)

	m, err := d.renderMember(spec, status, &gs, group, id, apiObject)
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

	status.UpdateServerGroupStatus(group, gs)

	return m.ID, nil
}

func (d *Deployment) renderMember(spec api.DeploymentSpec, status *api.DeploymentStatus, groupStatus *api.ServerGroupStatus, group api.ServerGroup, id string, apiObject *api.ArangoDeployment) (*api.MemberStatus, error) {
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
			id = d.renderMemberID(spec, status, groupStatus, group)
		}
	}
	if id == "" {
		return nil, errors.New("Unable to get ID")
	}
	deploymentName := apiObject.GetName()
	role := group.AsRole()

	arch := apiObject.GetAcceptedSpec().Architecture.GetDefault()
	if arch != api.ArangoDeploymentArchitectureAMD64 && apiObject.Status.CurrentImage != nil &&
		apiObject.Status.CurrentImage.ArangoDBVersion.CompareTo("3.10.0") < 0 {
		arch = api.ArangoDeploymentArchitectureAMD64
		d.log.Str("arch", string(arch)).Warn("Cannot render pod with requested arch. It's not supported in ArangoDB < 3.10.0. Defaulting architecture to AMD64")
		d.CreateEvent(k8sutil.NewCannotSetArchitectureEvent(d.GetAPIObject(), string(arch), id))
	}

	switch group {
	case api.ServerGroupSingle:
		d.log.Str("id", id).Debug("Adding single server")
		return &api.MemberStatus{
			ID:        id,
			UID:       uuid.NewUUID(),
			CreatedAt: meta.Now(),
			Phase:     api.MemberPhaseNone,
			PersistentVolumeClaim: &api.MemberPersistentVolumeClaimStatus{
				Name: shared.CreatePersistentVolumeClaimName(deploymentName, role, id),
			},
			PersistentVolumeClaimName: shared.CreatePersistentVolumeClaimName(deploymentName, role, id),
			Image:                     apiObject.Status.CurrentImage,
			Architecture:              &arch,
		}, nil
	case api.ServerGroupAgents:
		d.log.Str("id", id).Debug("Adding agent")
		return &api.MemberStatus{
			ID:        id,
			UID:       uuid.NewUUID(),
			CreatedAt: meta.Now(),
			Phase:     api.MemberPhaseNone,
			PersistentVolumeClaim: &api.MemberPersistentVolumeClaimStatus{
				Name: shared.CreatePersistentVolumeClaimName(deploymentName, role, id),
			},
			PersistentVolumeClaimName: shared.CreatePersistentVolumeClaimName(deploymentName, role, id),
			Image:                     apiObject.Status.CurrentImage,
			Architecture:              &arch,
		}, nil
	case api.ServerGroupDBServers:
		d.log.Str("id", id).Debug("Adding dbserver")
		return &api.MemberStatus{
			ID:        id,
			UID:       uuid.NewUUID(),
			CreatedAt: meta.Now(),
			Phase:     api.MemberPhaseNone,
			PersistentVolumeClaim: &api.MemberPersistentVolumeClaimStatus{
				Name: shared.CreatePersistentVolumeClaimName(deploymentName, role, id),
			},
			PersistentVolumeClaimName: shared.CreatePersistentVolumeClaimName(deploymentName, role, id),
			Image:                     apiObject.Status.CurrentImage,
			Architecture:              &arch,
		}, nil
	case api.ServerGroupCoordinators:
		d.log.Str("id", id).Debug("Adding coordinator")
		return &api.MemberStatus{
			ID:           id,
			UID:          uuid.NewUUID(),
			CreatedAt:    meta.Now(),
			Phase:        api.MemberPhaseNone,
			Image:        apiObject.Status.CurrentImage,
			Architecture: &arch,
		}, nil
	case api.ServerGroupSyncMasters:
		d.log.Str("id", id).Debug("Adding syncmaster")
		return &api.MemberStatus{
			ID:           id,
			UID:          uuid.NewUUID(),
			CreatedAt:    meta.Now(),
			Phase:        api.MemberPhaseNone,
			Image:        apiObject.Status.CurrentImage,
			Architecture: &arch,
		}, nil
	case api.ServerGroupSyncWorkers:
		d.log.Str("id", id).Debug("Adding syncworker")
		return &api.MemberStatus{
			ID:           id,
			UID:          uuid.NewUUID(),
			CreatedAt:    meta.Now(),
			Phase:        api.MemberPhaseNone,
			Image:        apiObject.Status.CurrentImage,
			Architecture: &arch,
		}, nil
	default:
		return nil, errors.WithStack(errors.Newf("Unknown server group %d", group))
	}
}
