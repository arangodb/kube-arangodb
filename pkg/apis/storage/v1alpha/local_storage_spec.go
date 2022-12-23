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

package v1alpha

import (
	"strings"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// LocalStorageSpec contains the specification part of
// an ArangoLocalStorage.
type LocalStorageSpec struct {
	StorageClass StorageClassSpec `json:"storageClass"`
	LocalPath    []string         `json:"localPath,omitempty"`

	Tolerations  []core.Toleration `json:"tolerations,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Privileged   *bool             `json:"privileged,omitempty"`

	PodCustomization *LocalStoragePodCustomization `json:"podCustomization,omitempty"`
}

// Validate the given spec, returning an error on validation
// problems or nil if all ok.
func (s LocalStorageSpec) Validate() error {
	if err := s.StorageClass.Validate(); err != nil {
		return errors.WithStack(err)
	}
	if len(s.LocalPath) == 0 {
		return errors.WithStack(errors.Wrapf(ValidationError, "localPath cannot be empty"))
	}
	for _, p := range s.LocalPath {
		if len(p) == 0 {
			return errors.WithStack(errors.Wrapf(ValidationError, "localPath cannot contain empty strings"))
		}
	}
	return nil
}

// SetDefaults fills empty field with default values.
func (s *LocalStorageSpec) SetDefaults(localStorageName string) {
	s.StorageClass.SetDefaults(localStorageName)
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to `spec.`.
func (s LocalStorageSpec) ResetImmutableFields(target *LocalStorageSpec) []string {
	var result []string
	if list := s.StorageClass.ResetImmutableFields("storageClass.", &target.StorageClass); len(list) > 0 {
		result = append(result, list...)
	}
	if strings.Join(s.LocalPath, ",") != strings.Join(target.LocalPath, ",") {
		target.LocalPath = s.LocalPath
		result = append(result, "localPath")
	}
	// TODO NodeSelector
	return result
}

func (s LocalStorageSpec) GetPrivileged() bool {
	if s.Privileged == nil {
		return false
	}

	return *s.Privileged
}
