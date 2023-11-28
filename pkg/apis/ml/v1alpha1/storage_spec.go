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
)

type ArangoMLStorageSpec struct {
	// Mode defines how storage implementation should be deployed
	Mode *ArangoMLStorageSpecMode `json:"mode,omitempty"`
	// Backend defines how storage is implemented
	Backend *ArangoMLStorageSpecBackend `json:"backend,omitempty"`
}

func (s *ArangoMLStorageSpec) GetMode() *ArangoMLStorageSpecMode {
	if s == nil || s.Mode == nil {
		return &ArangoMLStorageSpecMode{}
	}
	return s.Mode
}

func (s *ArangoMLStorageSpec) GetBackend() *ArangoMLStorageSpecBackend {
	if s == nil || s.Backend == nil {
		return &ArangoMLStorageSpecBackend{}
	}
	return s.Backend
}

func (s *ArangoMLStorageSpec) Validate() error {
	if s == nil {
		s = &ArangoMLStorageSpec{}
	}

	if err := shared.WithErrors(shared.PrefixResourceErrors("spec",
		shared.PrefixResourceError("backend", s.Backend.Validate()),
		shared.PrefixResourceError("mode", s.Mode.Validate()),
	)); err != nil {
		return err
	}

	return nil
}
