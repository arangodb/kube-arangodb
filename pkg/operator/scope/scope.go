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

package scope

func AsScope(s string) (Scope, bool) {
	switch s {
	case LegacyScope.String():
		return LegacyScope, true
	case NamespacedScope.String():
		return NamespacedScope, true
	case ClusterScope.String():
		return ClusterScope, true
	}

	return "", false
}

type Scope string

func (s Scope) String() string {
	return string(s)
}

func (s Scope) IsCluster() bool {
	return s == ClusterScope
}

func (s Scope) IsNamespaced() bool {
	return s.IsCluster() || s == NamespacedScope
}

const (
	LegacyScope     Scope = "legacy"
	NamespacedScope Scope = "namespaced"
	ClusterScope    Scope = "cluster"

	DefaultScope = LegacyScope
)
