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

package v2alpha1

import (
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// MemberStatusFunc is a callback which is used to traverse a specific group of servers and check their status.
type MemberStatusFunc func(group ServerGroup, list MemberStatusList) error

// DeploymentStatusMembers holds the member status of all server groups
type DeploymentStatusMembers struct {
	Single       MemberStatusList `json:"single,omitempty"`
	Agents       MemberStatusList `json:"agents,omitempty"`
	DBServers    MemberStatusList `json:"dbservers,omitempty"`
	Coordinators MemberStatusList `json:"coordinators,omitempty"`
	SyncMasters  MemberStatusList `json:"syncmasters,omitempty"`
	SyncWorkers  MemberStatusList `json:"syncworkers,omitempty"`
}

// Equal checks for equality
func (ds DeploymentStatusMembers) Equal(other DeploymentStatusMembers) bool {
	return ds.Single.Equal(other.Single) &&
		ds.Agents.Equal(other.Agents) &&
		ds.DBServers.Equal(other.DBServers) &&
		ds.Coordinators.Equal(other.Coordinators) &&
		ds.SyncMasters.Equal(other.SyncMasters) &&
		ds.SyncWorkers.Equal(other.SyncWorkers)
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
// Deprecated. Use AsList instead
func (ds DeploymentStatusMembers) ForeachServerGroup(cb MemberStatusFunc) error {
	return ds.ForeachServerInGroups(cb, AllServerGroups...)
}

// ForeachServerInGroups calls the given callback for specified server groups.
// Deprecated. Use AsListInGroups instead
func (ds DeploymentStatusMembers) ForeachServerInGroups(cb MemberStatusFunc, groups ...ServerGroup) error {
	for _, group := range groups {
		if err := ds.ForServerGroup(cb, group); err != nil {
			return err
		}
	}

	return nil
}

// ForServerGroup calls the given callback for specified server group.
// Deprecated. Use AsListInGroup or MembersOfGroup
func (ds DeploymentStatusMembers) ForServerGroup(cb MemberStatusFunc, group ServerGroup) error {
	switch group {
	case ServerGroupSingle:
		if err := cb(ServerGroupSingle, ds.Single); err != nil {
			return errors.WithStack(err)
		}
	case ServerGroupAgents:
		if err := cb(ServerGroupAgents, ds.Agents); err != nil {
			return errors.WithStack(err)
		}
	case ServerGroupDBServers:
		if err := cb(ServerGroupDBServers, ds.DBServers); err != nil {
			return errors.WithStack(err)
		}
	case ServerGroupCoordinators:
		if err := cb(ServerGroupCoordinators, ds.Coordinators); err != nil {
			return errors.WithStack(err)
		}
	case ServerGroupSyncMasters:
		if err := cb(ServerGroupSyncMasters, ds.SyncMasters); err != nil {
			return errors.WithStack(err)
		}
	case ServerGroupSyncWorkers:
		if err := cb(ServerGroupSyncWorkers, ds.SyncWorkers); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// MemberStatusByPodName returns a reference to the element in the given set of lists that has the given pod name.
// Returns member status and group which the pod belong to.
// If no such element exists, false is returned.
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

// Add adds the given status in the given group.
func (ds *DeploymentStatusMembers) Add(status MemberStatus, group ServerGroup) error {
	var err error
	switch group {
	case ServerGroupSingle:
		err = ds.Single.add(status)
	case ServerGroupAgents:
		err = ds.Agents.add(status)
	case ServerGroupDBServers:
		err = ds.DBServers.add(status)
	case ServerGroupCoordinators:
		err = ds.Coordinators.add(status)
	case ServerGroupSyncMasters:
		err = ds.SyncMasters.add(status)
	case ServerGroupSyncWorkers:
		err = ds.SyncWorkers.add(status)
	default:
		return errors.WithStack(errors.Wrapf(NotFoundError, "ServerGroup %d is not known", group))
	}
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Update updates the given status in the given group.
func (ds *DeploymentStatusMembers) Update(status MemberStatus, group ServerGroup) error {
	var err error
	switch group {
	case ServerGroupSingle:
		err = ds.Single.update(status)
	case ServerGroupAgents:
		err = ds.Agents.update(status)
	case ServerGroupDBServers:
		err = ds.DBServers.update(status)
	case ServerGroupCoordinators:
		err = ds.Coordinators.update(status)
	case ServerGroupSyncMasters:
		err = ds.SyncMasters.update(status)
	case ServerGroupSyncWorkers:
		err = ds.SyncWorkers.update(status)
	default:
		return errors.WithStack(errors.Wrapf(NotFoundError, "ServerGroup %d is not known", group))
	}
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// RemoveByID a member with given ID from the given group.
// Returns a NotFoundError if the ID of the given member or group cannot be found.
func (ds *DeploymentStatusMembers) RemoveByID(id string, group ServerGroup) error {
	var err error
	switch group {
	case ServerGroupSingle:
		err = ds.Single.removeByID(id)
	case ServerGroupAgents:
		err = ds.Agents.removeByID(id)
	case ServerGroupDBServers:
		err = ds.DBServers.removeByID(id)
	case ServerGroupCoordinators:
		err = ds.Coordinators.removeByID(id)
	case ServerGroupSyncMasters:
		err = ds.SyncMasters.removeByID(id)
	case ServerGroupSyncWorkers:
		err = ds.SyncWorkers.removeByID(id)
	default:
		return errors.WithStack(errors.Wrapf(NotFoundError, "ServerGroup %d is not known", group))
	}
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// AllMembersReady returns true when all members, that must be ready for the given mode, are in the Ready state.
func (ds DeploymentStatusMembers) AllMembersReady(mode DeploymentMode, syncEnabled bool) bool {
	syncReady := func() bool {
		if syncEnabled {
			return ds.SyncMasters.AllMembersReady() && ds.SyncWorkers.AllMembersReady()
		}
		return true
	}
	switch mode {
	case DeploymentModeSingle:
		return ds.Single.MembersReady() > 0
	case DeploymentModeActiveFailover:
		return ds.Agents.AllMembersReady() && ds.Single.MembersReady() > 0
	case DeploymentModeCluster:
		return ds.Agents.AllMembersReady() &&
			ds.DBServers.AllMembersReady() &&
			ds.Coordinators.AllMembersReady() &&
			syncReady()
	default:
		return false
	}
}

// MembersOfGroup returns the member list of the given group
func (ds DeploymentStatusMembers) MembersOfGroup(group ServerGroup) MemberStatusList {
	switch group {
	case ServerGroupSingle:
		return ds.Single
	case ServerGroupAgents:
		return ds.Agents
	case ServerGroupDBServers:
		return ds.DBServers
	case ServerGroupCoordinators:
		return ds.Coordinators
	case ServerGroupSyncMasters:
		return ds.SyncMasters
	case ServerGroupSyncWorkers:
		return ds.SyncWorkers
	default:
		return MemberStatusList{}
	}
}

// PodNames returns all members pod names
func (ds DeploymentStatusMembers) PodNames() []string {
	var n []string

	for _, m := range ds.AsList() {
		if name := m.Member.Pod.GetName(); name != "" {
			n = append(n, name)
		}
	}

	return n
}
