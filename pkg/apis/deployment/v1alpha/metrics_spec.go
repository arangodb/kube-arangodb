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
//

package v1alpha

import "github.com/arangodb/kube-arangodb/pkg/util"

// MetricsSpec contains spec for arangodb exporter
type MetricsSpec struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// IsEnabled returns whether metrics are enabled or not
func (s *MetricsSpec) IsEnabled() bool {
	return util.BoolOrDefault(s.Enabled, false)
}
