//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

import (
	core "k8s.io/api/core/v1"
)

// SidecarSpec contains spec for arangodb sidecar
type SidecarSpec struct {
	// Enabled if this is set to `true`, the operator runs a sidecar container for
	// every Agent, DB-Server, Coordinator and Single server.
	// +doc/default: false
	// +doc/link: Metrics collection|../metrics.md
	Enabled *bool `json:"enabled,omitempty"`

	// Resources holds resource requests & limits
	// +doc/type: core.ResourceRequirements
	// +doc/link: Documentation of core.ResourceRequirements|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core
	Resources core.ResourceRequirements `json:"resources,omitempty"`
}

func (s *SidecarSpec) IsEnabled(def bool) bool {
	if s == nil || s.Enabled == nil {
		return def
	}

	return *s.Enabled
}

// Validate the given spec
func (s *SidecarSpec) Validate() error {
	return nil
}
