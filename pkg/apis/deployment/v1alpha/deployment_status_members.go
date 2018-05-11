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

package v1alpha

import (
	"fmt"

	"github.com/pkg/errors"
)

// DeploymentStatusMembers holds the member status of all server groups
type DeploymentStatusMembers struct {
	Single       MemberStatusList `json:"single,omitempty"`
	Agents       MemberStatusList `json:"agents,omitempty"`
	DBServers    MemberStatusList `json:"dbservers,omitempty"`
	Coordinators MemberStatusList `json:"coordinators,omitempty"`
	SyncMasters  MemberStatusList `json:"syncmasters,omitempty"`
	SyncWorkers  MemberStatusList `json:"syncworkers,omitempty"`
}

// ContainsID returns true if the given set of members contains a member with given ID.
func (ds DeploymentStatusMembers) ContainsID(id string) bool {
	return ds.Single.ContainsID(id) ||
		ds.Agents.ContainsID(id) ||
		ds.DBServers.ContainsID(id) ||
		ds.Coordinators.ContainsID(id) ||
		ds.SyncMasters.ContainsID(id) ||
		ds.SyncWorkers.ContainsID(id)
}

// ElementByID returns the element in the given list that has the given ID and true.
// If no such element exists, false is returned.
func (ds DeploymentStatusMembers) ElementByID(id string) (MemberStatus, ServerGroup, bool) {
	if result, found := ds.Single.ElementByID(id); found {
		return result, ServerGroupSingle, true
	}
	if result, found := ds.Agents.ElementByID(id); found {
		return result, ServerGroupAgents, true
	}
	if result, found := ds.DBServers.ElementByID(id); found {
		return result, ServerGroupDBServers, true
	}
	if result, found := ds.Coordinators.ElementByID(id); found {
		return result, ServerGroupCoordinators, true
	}
	if result, found := ds.SyncMasters.ElementByID(id); found {
		return result, ServerGroupSyncMasters, true
	}
	if result, found := ds.SyncWorkers.ElementByID(id); found {
		return result, ServerGroupSyncWorkers, true
	}
	return MemberStatus{}, 0, false
}

// ForeachServerGroup calls the given callback for all server groups.
// If the callback returns an error, this error is returned and the callback is
// not called for the remaining groups.
func (ds DeploymentStatusMembers) ForeachServerGroup(cb func(group ServerGroup, list *MemberStatusList) error) error {
	if err := cb(ServerGroupSingle, &ds.Single); err != nil {
		return maskAny(err)
	}
	if err := cb(ServerGroupAgents, &ds.Agents); err != nil {
		return maskAny(err)
	}
	if err := cb(ServerGroupDBServers, &ds.DBServers); err != nil {
		return maskAny(err)
	}
	if err := cb(ServerGroupCoordinators, &ds.Coordinators); err != nil {
		return maskAny(err)
	}
	if err := cb(ServerGroupSyncMasters, &ds.SyncMasters); err != nil {
		return maskAny(err)
	}
	if err := cb(ServerGroupSyncWorkers, &ds.SyncWorkers); err != nil {
		return maskAny(err)
	}
	return nil
}

// MemberStatusByPodName returns a reference to the element in the given set of lists that has the given pod name.
// If no such element exists, nil is returned.
func (ds DeploymentStatusMembers) MemberStatusByPodName(podName string) (MemberStatus, ServerGroup, bool) {
	if result, found := ds.Single.ElementByPodName(podName); found {
		return result, ServerGroupSingle, true
	}
	if result, found := ds.Agents.ElementByPodName(podName); found {
		return result, ServerGroupAgents, true
	}
	if result, found := ds.DBServers.ElementByPodName(podName); found {
		return result, ServerGroupDBServers, true
	}
	if result, found := ds.Coordinators.ElementByPodName(podName); found {
		return result, ServerGroupCoordinators, true
	}
	if result, found := ds.SyncMasters.ElementByPodName(podName); found {
		return result, ServerGroupSyncMasters, true
	}
	if result, found := ds.SyncWorkers.ElementByPodName(podName); found {
		return result, ServerGroupSyncWorkers, true
	}
	return MemberStatus{}, 0, false
}

// MemberStatusByPVCName returns a reference to the element in the given set of lists that has the given PVC name.
// If no such element exists, nil is returned.
func (ds DeploymentStatusMembers) MemberStatusByPVCName(pvcName string) (MemberStatus, ServerGroup, bool) {
	if result, found := ds.Single.ElementByPVCName(pvcName); found {
		return result, ServerGroupSingle, true
	}
	if result, found := ds.Agents.ElementByPVCName(pvcName); found {
		return result, ServerGroupAgents, true
	}
	if result, found := ds.DBServers.ElementByPVCName(pvcName); found {
		return result, ServerGroupDBServers, true
	}
	// Note: Other server groups do not have PVC's so we can skip them.
	return MemberStatus{}, 0, false
}

// UpdateMemberStatus updates the given status in the given group.
func (ds *DeploymentStatusMembers) UpdateMemberStatus(status MemberStatus, group ServerGroup) error {
	var err error
	switch group {
	case ServerGroupSingle:
		err = ds.Single.Update(status)
	case ServerGroupAgents:
		err = ds.Agents.Update(status)
	case ServerGroupDBServers:
		err = ds.DBServers.Update(status)
	case ServerGroupCoordinators:
		err = ds.Coordinators.Update(status)
	case ServerGroupSyncMasters:
		err = ds.SyncMasters.Update(status)
	case ServerGroupSyncWorkers:
		err = ds.SyncWorkers.Update(status)
	default:
		return maskAny(errors.Wrapf(NotFoundError, "ServerGroup %d is not known", group))
	}
	if err != nil {
		return maskAny(err)
	}
	return nil
}

// RemoveByID a member with given ID from the given group.
// Returns a NotFoundError if the ID of the given member or group cannot be found.
func (ds *DeploymentStatusMembers) RemoveByID(id string, group ServerGroup) error {
	var err error
	switch group {
	case ServerGroupSingle:
		err = ds.Single.RemoveByID(id)
	case ServerGroupAgents:
		err = ds.Agents.RemoveByID(id)
	case ServerGroupDBServers:
		err = ds.DBServers.RemoveByID(id)
	case ServerGroupCoordinators:
		err = ds.Coordinators.RemoveByID(id)
	case ServerGroupSyncMasters:
		err = ds.SyncMasters.RemoveByID(id)
	case ServerGroupSyncWorkers:
		err = ds.SyncWorkers.RemoveByID(id)
	default:
		return maskAny(errors.Wrapf(NotFoundError, "ServerGroup %d is not known", group))
	}
	if err != nil {
		return maskAny(err)
	}
	return nil
}

// AllMembersReady returns true when all members are in the Ready state.
func (ds DeploymentStatusMembers) AllMembersReady() bool {
	if err := ds.ForeachServerGroup(func(group ServerGroup, list *MemberStatusList) error {
		for _, x := range *list {
			if !x.Conditions.IsTrue(ConditionTypeReady) {
				return fmt.Errorf("not ready")
			}
		}
		return nil
	}); err != nil {
		return false
	}
	return true
}
