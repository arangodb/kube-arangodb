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
}

// Validate the given spec, returning an error on validation
// problems or nil if all ok.
func (s DeploymentReplicationSpec) Validate() error {
	/*	if err := s.StorageClass.Validate(); err != nil {
			return maskAny(err)
		}
		if len(s.LocalPath) == 0 {
			return maskAny(errors.Wrapf(ValidationError, "localPath cannot be empty"))
		}
		for _, p := range s.LocalPath {
			if len(p) == 0 {
				return maskAny(errors.Wrapf(ValidationError, "localPath cannot contain empty strings"))
			}
		}*/
	return nil
}

// SetDefaults fills empty field with default values.
func (s *DeploymentReplicationSpec) SetDefaults() {
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to `spec.`.
func (s DeploymentReplicationSpec) ResetImmutableFields(target *DeploymentReplicationSpec) []string {
	var result []string
	return result
}
