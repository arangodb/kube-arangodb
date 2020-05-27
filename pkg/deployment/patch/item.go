//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package patch

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Operation string

const (
	AddOperation     Operation = "add"
	ReplaceOperation Operation = "replace"
	RemoveOperation  Operation = "remove"
)

var _ json.Marshaler = &Path{}

func NewPath(items ...string) Path {
	return items
}

type Path []string

func (p Path) MarshalJSON() ([]byte, error) {
	if len(p) == 0 {
		return json.Marshal("/")
	}
	v := fmt.Sprintf("/%s", strings.Join(p, "/"))
	return json.Marshal(v)
}

type Item struct {
	Op    Operation   `json:"op"`
	Path  Path        `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func ItemAdd(path Path, value interface{}) Item {
	return Item{
		Op:    AddOperation,
		Path:  path,
		Value: value,
	}
}

func ItemReplace(path Path, value interface{}) Item {
	return Item{
		Op:    ReplaceOperation,
		Path:  path,
		Value: value,
	}
}

func ItemRemove(path Path) Item {
	return Item{
		Op:    RemoveOperation,
		Path:  path,
		Value: nil,
	}
}
