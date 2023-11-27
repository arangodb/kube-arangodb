//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ArangoMLStorageSpecMode struct {
	// Sidecar mode runs the storage implementation as a sidecar
	Sidecar *ArangoMLStorageSpecModeSidecar `json:"sidecar,omitempty"`
}

func (s *ArangoMLStorageSpecMode) GetSidecar() *ArangoMLStorageSpecModeSidecar {
	if s == nil || s.Sidecar == nil {
		return &ArangoMLStorageSpecModeSidecar{}
	}
	return s.Sidecar
}

func (s *ArangoMLStorageSpecMode) Validate() error {
	if s == nil {
		return errors.Newf("Mode is not defined")
	}
	return shared.WithErrors(shared.PrefixResourceError("sidecar", s.Sidecar.Validate()))
}
