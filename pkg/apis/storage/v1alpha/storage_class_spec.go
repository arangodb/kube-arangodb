//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// StorageClassSpec contains specification for create StorageClass.
type StorageClassSpec struct {
	Name      string `json:"name,omitempty"`
	IsDefault bool   `json:"isDefault,omitempty"`
}

// Validate the given spec, returning an error on validation
// problems or nil if all ok.
func (s StorageClassSpec) Validate() error {
	if err := k8sutil.ValidateResourceName(s.Name); err != nil {
		return maskAny(err)
	}
	return nil
}

// SetDefaults fills empty field with default values.
func (s *StorageClassSpec) SetDefaults(localStorageName string) {
	if s.Name == "" {
		s.Name = localStorageName
	}
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to `spec.`.
func (s StorageClassSpec) ResetImmutableFields(fieldPrefix string, target *StorageClassSpec) []string {
	var result []string
	if s.Name != target.Name {
		target.Name = s.Name
		result = append(result, fieldPrefix+"name")
	}
	return result
}
