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

import "github.com/arangodb/kube-arangodb/pkg/util/errors"

// DeploymentReplicationSpec contains the specification part of
// an ArangoDeploymentReplication.
type DeploymentReplicationSpec struct {
	Source      EndpointSpec `json:"source"`
	Destination EndpointSpec `json:"destination"`
	// Cancellation describes what to do during cancellation process.
	Cancellation DeploymentReplicationCancel `json:"cancellation"`
}

// DeploymentReplicationCancel describes what to do during cancellation process.
type DeploymentReplicationCancel struct {
	// EnsureInSync if it is true then during cancellation process data consistency is required.
	// Default value is true.
	EnsureInSync *bool `json:"ensureInSync"`
	// SourceReadOnly if it true then after cancellation source data center should be in read-only mode.
	// Default value is false.
	SourceReadOnly *bool `json:"sourceReadOnly"`
}

// Validate the given spec, returning an error on validation
// problems or nil if all ok.
func (s DeploymentReplicationSpec) Validate() error {
	if err := s.Source.Validate(true); err != nil {
		return errors.WithStack(err)
	}
	if err := s.Destination.Validate(false); err != nil {
		return errors.WithStack(err)
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
