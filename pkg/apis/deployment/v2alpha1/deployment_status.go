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

package v2alpha1

import (
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

// DeploymentStatus contains the status part of a Cluster resource.
type DeploymentStatus struct {
	// Phase holds the current lifetime phase of the deployment
	Phase DeploymentPhase `json:"phase,omitempty"`
	// Reason contains a human readable reason for reaching the current state (can be empty)
	Reason string `json:"reason,omitempty"` // Reason for current state

	// AppliedVersion defines checksum of applied spec
	AppliedVersion string `json:"appliedVersion"`

	// AcceptedSpecVersion defines checksum of accepted spec
	AcceptedSpecVersion *string `json:"acceptedSpecVersion,omitempty"`

	// ServiceName holds the name of the Service a client can use (inside the k8s cluster)
	// to access ArangoDB.
	ServiceName string `json:"serviceName,omitempty"`
	// SyncServiceName holds the name of the Service a client can use (inside the k8s cluster)
	// to access syncmasters (only set when dc2dc synchronization is enabled).
	SyncServiceName string `json:"syncServiceName,omitempty"`

	ExporterServiceName string `json:"exporterServiceName,omitempty"`

	ExporterServiceMonitorName string `json:"exporterServiceMonitorName,omitempty"`

	Restore *DeploymentRestoreResult `json:"restore,omitempty"`

	// Images holds a list of ArangoDB images with their ID and ArangoDB version.
	Images ImageInfoList `json:"arangodb-images,omitempty"`
	// Image that is currently being used when new pods are created
	CurrentImage *ImageInfo `json:"current-image,omitempty"`

	// Members holds the status for all members in all server groups
	Members DeploymentStatusMembers `json:"members"`

	// Conditions specific to the entire deployment
	Conditions ConditionList `json:"conditions,omitempty"`

	// Plan to update this deployment
	Plan Plan `json:"plan,omitempty"`

	// HighPriorityPlan to update this deployment. Executed before plan
	HighPriorityPlan Plan `json:"highPriorityPlan,omitempty"`

	// ResourcesPlan to update this deployment. Executed before plan, after highPlan
	ResourcesPlan Plan `json:"resourcesPlan,omitempty"`

	// AcceptedSpec contains the last specification that was accepted by the operator.
	AcceptedSpec *DeploymentSpec `json:"accepted-spec,omitempty"`

	// SecretHashes keeps a sha256 hash of secret values, so we can
	// detect changes in secret values.
	SecretHashes *SecretHashes `json:"secret-hashes,omitempty"`

	// Hashes keep status of hashes in deployment
	Hashes DeploymentStatusHashes `json:"hashes,omitempty"`

	// ForceStatusReload if set to true forces a reload of the status from the custom resource.
	ForceStatusReload *bool `json:"force-status-reload,omitempty"`

	// Agency keeps information about agency
	Agency *DeploymentStatusAgencyInfo `json:"agency,omitempty"`

	Topology *TopologyStatus `json:"topology,omitempty"`

	Rebalancer *ArangoDeploymentRebalancerStatus `json:"rebalancer,omitempty"`

	BackOff BackOff `json:"backoff,omitempty"`

	Version *Version `json:"version,omitempty"`

	Timezone *string `json:"timezone,omitempty"`

	Single       *ServerGroupStatus `json:"single,omitempty"`
	Agents       *ServerGroupStatus `json:"agents,omitempty"`
	DBServers    *ServerGroupStatus `json:"dbservers,omitempty"`
	Coordinators *ServerGroupStatus `json:"coordinators,omitempty"`
	SyncMasters  *ServerGroupStatus `json:"syncmasters,omitempty"`
	SyncWorkers  *ServerGroupStatus `json:"syncworkers,omitempty"`
}

// Equal checks for equality
func (ds *DeploymentStatus) Equal(other DeploymentStatus) bool {
	return ds.Phase == other.Phase &&
		ds.Reason == other.Reason &&
		ds.ServiceName == other.ServiceName &&
		ds.SyncServiceName == other.SyncServiceName &&
		ds.ExporterServiceName == other.ExporterServiceName &&
		ds.ExporterServiceMonitorName == other.ExporterServiceMonitorName &&
		ds.Images.Equal(other.Images) &&
		ds.Restore.Equal(other.Restore) &&
		ds.CurrentImage.Equal(other.CurrentImage) &&
		ds.Members.Equal(other.Members) &&
		ds.Conditions.Equal(other.Conditions) &&
		ds.Plan.Equal(other.Plan) &&
		ds.HighPriorityPlan.Equal(other.HighPriorityPlan) &&
		ds.ResourcesPlan.Equal(other.ResourcesPlan) &&
		strings.CompareStringPointers(ds.AcceptedSpecVersion, other.AcceptedSpecVersion) &&
		ds.AcceptedSpec.Equal(other.AcceptedSpec) &&
		ds.SecretHashes.Equal(other.SecretHashes) &&
		ds.Agency.Equal(other.Agency) &&
		ds.Topology.Equal(other.Topology) &&
		ds.BackOff.Equal(other.BackOff) &&
		ds.Version.Equal(other.Version) &&
		ds.Single.Equal(other.Single) &&
		ds.Agents.Equal(other.Agents) &&
		ds.DBServers.Equal(other.DBServers) &&
		ds.Coordinators.Equal(other.Coordinators) &&
		ds.SyncMasters.Equal(other.SyncMasters) &&
		ds.SyncWorkers.Equal(other.SyncWorkers) &&
		strings.CompareStringPointers(ds.Timezone, other.Timezone)
}

// IsForceReload returns true if ForceStatusReload is set to true
func (ds *DeploymentStatus) IsForceReload() bool {
	return util.TypeOrDefault[bool](ds.ForceStatusReload, false)
}

func (ds *DeploymentStatus) IsPlanEmpty() bool {
	return ds.Plan.IsEmpty() && ds.HighPriorityPlan.IsEmpty()
}

func (ds *DeploymentStatus) NonInternalActions() int {
	return ds.Plan.NonInternalActions() + ds.HighPriorityPlan.NonInternalActions()
}

// GetServerGroupStatus returns the server group status (from this
// deployment status) for the given group.
func (ds DeploymentStatus) GetServerGroupStatus(group ServerGroup) ServerGroupStatus {
	if v := ds.getServerGroupStatus(group); v == nil {
		return ServerGroupStatus{}
	} else {
		return *v
	}
}

func (ds DeploymentStatus) getServerGroupStatus(group ServerGroup) *ServerGroupStatus {
	switch group {
	case ServerGroupSingle:
		return ds.Single.DeepCopy()
	case ServerGroupAgents:
		return ds.Agents.DeepCopy()
	case ServerGroupDBServers:
		return ds.DBServers.DeepCopy()
	case ServerGroupCoordinators:
		return ds.Coordinators.DeepCopy()
	case ServerGroupSyncMasters:
		return ds.SyncMasters.DeepCopy()
	case ServerGroupSyncWorkers:
		return ds.SyncWorkers.DeepCopy()
	default:
		return nil
	}
}

// UpdateServerGroupStatus returns the server group status (from this
// deployment status) for the given group.
func (ds *DeploymentStatus) UpdateServerGroupStatus(group ServerGroup, gspec ServerGroupStatus) {
	switch group {
	case ServerGroupSingle:
		ds.Single = gspec.DeepCopy()
	case ServerGroupAgents:
		ds.Agents = &gspec
	case ServerGroupDBServers:
		ds.DBServers = &gspec
	case ServerGroupCoordinators:
		ds.Coordinators = &gspec
	case ServerGroupSyncMasters:
		ds.SyncMasters = &gspec
	case ServerGroupSyncWorkers:
		ds.SyncWorkers = &gspec
	}
}
