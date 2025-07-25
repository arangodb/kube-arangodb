//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package v1beta1

import (
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

type ArangoPlatformStorageSpec struct {
	// Backend defines how storage is implemented
	Backend *ArangoPlatformStorageSpecBackend `json:"backend,omitempty"`
}

func (s *ArangoPlatformStorageSpec) GetBackend() *ArangoPlatformStorageSpecBackend {
	if s == nil || s.Backend == nil {
		return nil
	}
	return s.Backend
}

func (s *ArangoPlatformStorageSpec) Validate() error {
	if s == nil {
		s = &ArangoPlatformStorageSpec{}
	}

	if err := shared.WithErrors(shared.PrefixResourceErrors("spec",
		shared.PrefixResourceError("backend", s.Backend.Validate()),
	)); err != nil {
		return err
	}

	return nil
}
