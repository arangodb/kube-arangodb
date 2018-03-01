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
	"github.com/pkg/errors"
)

// DeploymentMode specifies the type of ArangoDB deployment to create.
type DeploymentMode string

const (
	// DeploymentModeSingle yields a single server
	DeploymentModeSingle DeploymentMode = "single"
	// DeploymentModeResilientSingle yields an agency and a resilient-single server pair
	DeploymentModeResilientSingle DeploymentMode = "resilientsingle"
	// DeploymentModeCluster yields an full cluster (agency, dbservers & coordinators)
	DeploymentModeCluster DeploymentMode = "cluster"
)

// Validate the mode.
// Return errors when validation fails, nil on success.
func (m DeploymentMode) Validate() error {
	switch m {
	case DeploymentModeSingle, DeploymentModeResilientSingle, DeploymentModeCluster:
		return nil
	default:
		return maskAny(errors.Wrapf(ValidationError, "Unknown deployment mode: '%s'", string(m)))
	}
}

// HasSingleServers returns true when the given mode is "single" or "resilientsingle".
func (m DeploymentMode) HasSingleServers() bool {
	return m == DeploymentModeSingle || m == DeploymentModeResilientSingle
}

// HasAgents returns true when the given mode is "resilientsingle" or "cluster".
func (m DeploymentMode) HasAgents() bool {
	return m == DeploymentModeResilientSingle || m == DeploymentModeCluster
}

// HasDBServers returns true when the given mode is "cluster".
func (m DeploymentMode) HasDBServers() bool {
	return m == DeploymentModeCluster
}

// HasCoordinators returns true when the given mode is "cluster".
func (m DeploymentMode) HasCoordinators() bool {
	return m == DeploymentModeCluster
}

// SupportsSync returns true when the given mode supports dc2dc replication.
func (m DeploymentMode) SupportsSync() bool {
	return m == DeploymentModeCluster
}
