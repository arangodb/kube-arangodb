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

// DeploymentMode specifies the type of ArangoDB deployment to create.
type DeploymentMode string

const (
	// DeploymentModeSingle yields a single server
	DeploymentModeSingle DeploymentMode = "Single"
	// DeploymentModeActiveFailover yields an agency and a active-failover server pair
	DeploymentModeActiveFailover DeploymentMode = "ActiveFailover"
	// DeploymentModeCluster yields an full cluster (agency, dbservers & coordinators)
	DeploymentModeCluster DeploymentMode = "Cluster"
)

// Validate the mode.
// Return errors when validation fails, nil on success.
func (m DeploymentMode) Validate() error {
	switch m {
	case DeploymentModeSingle, DeploymentModeActiveFailover, DeploymentModeCluster:
		return nil
	default:
		return errors.WithStack(errors.Wrapf(ValidationError, "Unknown deployment mode: '%s'", string(m)))
	}
}

// Get mode or default value
func (m *DeploymentMode) Get() DeploymentMode {
	if m == nil {
		return DeploymentModeCluster
	}

	return *m
}

// String return string from mode
func (m *DeploymentMode) String() string {
	return string(m.Get())
}

// Nww return pointer to mode
func (m DeploymentMode) New() *DeploymentMode {
	return &m
}

// HasSingleServers returns true when the given mode is "Single" or "ActiveFailover".
func (m DeploymentMode) HasSingleServers() bool {
	return m == DeploymentModeSingle || m == DeploymentModeActiveFailover
}

// HasAgents returns true when the given mode is "ActiveFailover" or "Cluster".
func (m DeploymentMode) HasAgents() bool {
	return m == DeploymentModeActiveFailover || m == DeploymentModeCluster
}

// HasDBServers returns true when the given mode is "Cluster".
func (m DeploymentMode) HasDBServers() bool {
	return m == DeploymentModeCluster
}

// HasCoordinators returns true when the given mode is "Cluster".
func (m DeploymentMode) HasCoordinators() bool {
	return m == DeploymentModeCluster
}

// SupportsSync returns true when the given mode supports dc2dc replication.
func (m DeploymentMode) SupportsSync() bool {
	return m == DeploymentModeCluster
}

// IsCluster returns true if the deployment mode is cluster
func (m DeploymentMode) IsCluster() bool {
	return m == DeploymentModeCluster
}

// ServingGroup returns mode serving group
func (m DeploymentMode) ServingGroup() ServerGroup {
	switch m {
	case DeploymentModeCluster:
		return ServerGroupCoordinators
	case DeploymentModeSingle, DeploymentModeActiveFailover:
		return ServerGroupSingle
	default:
		return ServerGroupUnknown
	}
}

// NewMode returns a reference to a string with given value.
func NewMode(input DeploymentMode) *DeploymentMode {
	return &input
}

// NewModeOrNil returns nil if input is nil, otherwise returns a clone of the given value.
func NewModeOrNil(input *DeploymentMode) *DeploymentMode {
	if input == nil {
		return nil
	}
	return NewMode(*input)
}

// ModeOrDefault returns the default value (or empty string) if input is nil, otherwise returns the referenced value.
func ModeOrDefault(input *DeploymentMode, defaultValue ...DeploymentMode) DeploymentMode {
	if input == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return *input
}
