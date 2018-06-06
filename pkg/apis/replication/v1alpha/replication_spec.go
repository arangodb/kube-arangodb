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

// DeploymentReplicationSpec contains the specification part of
// an ArangoDeploymentReplication.
type DeploymentReplicationSpec struct {
	Source      EndpointSpec `json:"source"`
	Destination EndpointSpec `json:"destination"`
}

// Validate the given spec, returning an error on validation
// problems or nil if all ok.
func (s DeploymentReplicationSpec) Validate() error {
	if err := s.Source.Validate(true); err != nil {
		return maskAny(err)
	}
	if err := s.Destination.Validate(false); err != nil {
		return maskAny(err)
	}
	return nil
}

// SetDefaults fills empty field with default values.
func (s *DeploymentReplicationSpec) SetDefaults() {
	s.Source.SetDefaults()
	s.Destination.SetDefaults()
}

// SetDefaultsFrom fills empty field with default values from the given source.
func (s *DeploymentReplicationSpec) SetDefaultsFrom(source DeploymentReplicationSpec) {
	s.Source.SetDefaultsFrom(source.Source)
	s.Destination.SetDefaultsFrom(source.Destination)
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to `spec.`.
func (s DeploymentReplicationSpec) ResetImmutableFields(target *DeploymentReplicationSpec) []string {
	var result []string
	if list := s.Source.ResetImmutableFields(&target.Source, "source."); len(list) > 0 {
		result = append(result, list...)
	}
	if list := s.Destination.ResetImmutableFields(&target.Destination, "destination."); len(list) > 0 {
		result = append(result, list...)
	}
	return result
}
