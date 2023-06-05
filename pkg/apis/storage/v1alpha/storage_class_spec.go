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

package v1alpha

import (
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// StorageClassSpec contains specification for create StorageClass.
type StorageClassSpec struct {
	Name          string                              `json:"name,omitempty"`
	IsDefault     bool                                `json:"isDefault,omitempty"`
	ReclaimPolicy *core.PersistentVolumeReclaimPolicy `json:"reclaimPolicy,omitempty"`
}

// Validate the given spec, returning an error on validation
// problems or nil if all ok.
func (s StorageClassSpec) Validate() error {
	if err := shared.ValidateResourceName(s.Name); err != nil {
		return errors.WithStack(err)
	}

	switch r := s.GetReclaimPolicy(); r {
	case core.PersistentVolumeReclaimRetain, core.PersistentVolumeReclaimDelete:
	default:
		return errors.Newf("Unsupported ReclaimPolicy: %s", r)
	}

	return nil
}

// SetDefaults fills empty field with default values.
func (s *StorageClassSpec) SetDefaults(localStorageName string) {
	if s.Name == "" {
		s.Name = localStorageName
	}
}

// GetReclaimPolicy returns StorageClass Reclaim Policy
func (s *StorageClassSpec) GetReclaimPolicy() core.PersistentVolumeReclaimPolicy {
	return util.TypeOrDefault(s.ReclaimPolicy, core.PersistentVolumeReclaimRetain)
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
	if s.GetReclaimPolicy() != target.GetReclaimPolicy() {
		target.ReclaimPolicy = s.ReclaimPolicy
		result = append(result, fieldPrefix+"reclaimPolicy")
	}
	return result
}
