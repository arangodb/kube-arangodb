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

const DefaultTopologySpecLabel = "topology.kubernetes.io/zone"

type TopologySpec struct {
	Enabled bool    `json:"enabled,omitempty"`
	Zones   int     `json:"zones,omitempty"`
	Label   *string `json:"label,omitempty"`
}

func (t *TopologySpec) IsEnabled() bool {
	if t == nil {
		return false
	}

	return t.Enabled && t.Zones > 0
}

func (t *TopologySpec) GetZones() int {
	if t == nil {
		return 0
	}

	return t.Zones
}

func (t *TopologySpec) GetLabel() string {
	if t == nil || t.Label == nil {
		return DefaultTopologySpecLabel
	}

	return *t.Label
}
