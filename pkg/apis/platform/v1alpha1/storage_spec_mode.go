//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package v1alpha1

import (
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ArangoPlatformStorageSpecModeType int

const (
	ArangoPlatformStorageSpecModeTypeUnknown ArangoPlatformStorageSpecModeType = iota
	ArangoPlatformStorageSpecModeTypeSidecar
)

type ArangoPlatformStorageSpecMode struct {
	// Sidecar mode runs the storage implementation as a sidecar
	Sidecar *ArangoPlatformStorageSpecModeSidecar `json:"sidecar,omitempty"`
}

func (s *ArangoPlatformStorageSpecMode) GetSidecar() *ArangoPlatformStorageSpecModeSidecar {
	if s == nil || s.Sidecar == nil {
		return nil
	}
	return s.Sidecar
}

func (s *ArangoPlatformStorageSpecMode) GetType() ArangoPlatformStorageSpecModeType {
	return ArangoPlatformStorageSpecModeTypeSidecar
}

func (s *ArangoPlatformStorageSpecMode) Validate() error {
	if s == nil {
		s = &ArangoPlatformStorageSpecMode{}
	}

	if s.GetType() == ArangoPlatformStorageSpecModeTypeUnknown {
		return errors.Errorf("Unknown mode type")
	}

	return nil
}
