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
	"math/rand"

	"github.com/pkg/errors"
)

// DeploymentState is a strongly typed state of a deployment
type DeploymentState string

const (
	// DeploymentStateNone indicates that the state is not set yet
	DeploymentStateNone DeploymentState = ""
	// DeploymentStateCreating indicates that the deployment is being created
	DeploymentStateCreating DeploymentState = "Creating"
	// DeploymentStateRunning indicates that all servers are running and reachable
	DeploymentStateRunning DeploymentState = "Running"
	// DeploymentStateScaling indicates that servers are being added or removed
	DeploymentStateScaling DeploymentState = "Scaling"
	// DeploymentStateUpgrading indicates that a version upgrade is in progress
	DeploymentStateUpgrading DeploymentState = "Upgrading"
	// DeploymentStateFailed indicates that a deployment is in a failed state
	// from which automatic recovery is impossible. Inspect `Reason` for more info.
	DeploymentStateFailed DeploymentState = "Failed"
)

// IsFailed returns true if given state is DeploymentStateFailed
func (cs DeploymentState) IsFailed() bool {
	return cs == DeploymentStateFailed
}

// DeploymentStatus contains the status part of a Cluster resource.
type DeploymentStatus struct {
	// State holds the current state of the deployment
	State DeploymentState `json:"state"`
	// Reason contains a human readable reason for reaching the current state (can be empty)
	Reason string `json:"reason,omitempty"` // Reason for current state

	// ServiceName holds the name of the Service a client can use (inside the k8s cluster)
	// to access ArangoDB.
	ServiceName string `json:"serviceName,omitempty"`
	// SyncServiceName holds the name of the Service a client can use (inside the k8s cluster)
	// to access syncmasters (only set when dc2dc synchronization is enabled).
	SyncServiceName string `json:"syncServiceName,omitempty"`

	// Members holds the status for all members in all server groups
	Members DeploymentStatusMembers `json:"members"`

	// Conditions specific to the entire deployment
	Conditions ConditionList `json:"conditions,omitempty"`

	// Plan to update this deployment
	Plan Plan `json:"plan,omitempty"`
}

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

// MemberStatusList is a list of MemberStatus entries
type MemberStatusList []MemberStatus

// ContainsID returns true if the given list contains a member with given ID.
func (l MemberStatusList) ContainsID(id string) bool {
	for _, x := range l {
		if x.ID == id {
			return true
		}
	}
	return false
}

// ElementByID returns the element in the given list that has the given ID and true.
// If no such element exists, false is returned.
func (l MemberStatusList) ElementByID(id string) (MemberStatus, bool) {
	for i, x := range l {
		if x.ID == id {
			return l[i], true
		}
	}
	return MemberStatus{}, false
}

// ElementByPodName returns the element in the given list that has the given pod name and true.
// If no such element exists, false is returned.
func (l MemberStatusList) ElementByPodName(podName string) (MemberStatus, bool) {
	for i, x := range l {
		if x.PodName == podName {
			return l[i], true
		}
	}
	return MemberStatus{}, false
}

// Add a member to the list.
// Returns an AlreadyExistsError if the ID of the given member already exists.
func (l *MemberStatusList) Add(m MemberStatus) error {
	src := *l
	for _, x := range src {
		if x.ID == m.ID {
			return maskAny(errors.Wrapf(AlreadyExistsError, "Member '%s' already exists", m.ID))
		}
	}
	*l = append(src, m)
	return nil
}

// Update a member in the list.
// Returns a NotFoundError if the ID of the given member cannot be found.
func (l MemberStatusList) Update(m MemberStatus) error {
	for i, x := range l {
		if x.ID == m.ID {
			l[i] = m
			return nil
		}
	}
	return maskAny(errors.Wrapf(NotFoundError, "Member '%s' is not a member", m.ID))
}

// RemoveByID a member with given ID from the list.
// Returns a NotFoundError if the ID of the given member cannot be found.
func (l *MemberStatusList) RemoveByID(id string) error {
	src := *l
	for i, x := range src {
		if x.ID == id {
			*l = append(src[:i], src[i+1:]...)
			return nil
		}
	}
	return maskAny(errors.Wrapf(NotFoundError, "Member '%s' is not a member", id))
}

// SelectMemberToRemove selects a member from the given list that should
// be removed in a scale down action.
// Returns an error if the list is empty.
func (l MemberStatusList) SelectMemberToRemove() (MemberStatus, error) {
	if len(l) > 0 {
		// Try to find a not ready member
		for _, m := range l {
			if m.State == MemberStateNone {
				return m, nil
			}
		}
		// Pick a random member that is in created state
		perm := rand.Perm(len(l))
		for _, idx := range perm {
			m := l[idx]
			if m.State == MemberStateCreated {
				return m, nil
			}
		}
	}
	return MemberStatus{}, maskAny(errors.Wrap(NotFoundError, "No member available for removal"))
}

// MemberState is a strongly typed state of a deployment member
type MemberState string

const (
	// MemberStateNone indicates that the state is not set yet
	MemberStateNone MemberState = ""
	// MemberStateCreated indicates that all resources needed for the member have been created
	MemberStateCreated MemberState = "Created"
	// MemberStateCleanOut indicates that a dbserver is in the process of being cleaned out
	MemberStateCleanOut MemberState = "CleanOut"
	// MemberStateShuttingDown indicates that a member is shutting down
	MemberStateShuttingDown MemberState = "ShuttingDown"
)

// MemberStatus holds the current status of a single member (server)
type MemberStatus struct {
	// ID holds the unique ID of the member.
	// This id is also used within the ArangoDB cluster to identify this server.
	ID string `json:"id"`
	// State holds the current state of this member
	State MemberState `json:"state"`
	// PersistentVolumeClaimName holds the name of the persistent volume claim used for this member (if any).
	PersistentVolumeClaimName string `json:"persistentVolumeClaimName,omitempty"`
	// PodName holds the name of the Pod that currently runs this member
	PodName string `json:"podName,omitempty"`
	// Conditions specific to this member
	Conditions ConditionList `json:"conditions,omitempty"`
}
