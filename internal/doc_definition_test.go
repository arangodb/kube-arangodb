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

package internal

import (
	"sort"
	"strings"
)

type DocDefinitions []DocDefinition

type DocDefinition struct {
	Path string
	Type string

	File string
	Line int

	Docs []string

	Links []string

	Important *string

	Enum []string

	Immutable *string

	Default *string
	Example []string
}

func (d DocDefinitions) Sort() {
	sort.Slice(d, func(i, j int) bool {
		a, b := strings.ToLower(d[i].Path), strings.ToLower(d[j].Path)
		if a == b {
			return d[i].Path < d[j].Path
		}
		return a < b
	})
}
