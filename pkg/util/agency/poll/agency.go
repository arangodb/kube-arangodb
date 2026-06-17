//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package poll

import (
	"encoding/json"

	"github.com/arangodb-helper/go-helper/pkg/refs"

	"github.com/arangodb/kube-arangodb/pkg/util/agency/transaction"
)

type Response[T interface{}] struct {
	Result Result[T] `json:"result"`
}
type Result[T interface{}] struct {
	CommitIndex uint64  `json:"commitIndex,omitempty"`
	FirstIndex  *uint64 `json:"firstIndex,omitempty"`
	ReadDB      *T      `json:"readDB,omitempty"`
	Log         []Patch `json:"log,omitempty"`
}
type Patch struct {
	Index uint64  `json:"index"`
	Items ItemSet `json:"query"`
}
type ItemSet map[string]Item
type ItemLazy struct {
	Operation *transaction.Operation `json:"op,omitempty"`
	Data      *json.RawMessage       `json:"new,omitempty"`
	Position  *json.RawMessage       `json:"pos,omitempty"`
	Value     *json.RawMessage       `json:"val,omitempty"`
}

func (i ItemLazy) GetData() []byte {
	if d := i.Data; d != nil {
		return *d
	}
	return nil
}
func (i ItemLazy) GetPosition() []byte {
	if d := i.Position; d != nil {
		return *d
	}
	return nil
}
func (i ItemLazy) GetValue() []byte {
	if d := i.Value; d != nil {
		return *d
	}
	return nil
}

type Item struct {
	ItemLazy
}

func (i *Item) UnmarshalJSON(bytes []byte) error {
	var l ItemLazy
	// Try to do first unmarshal
	if err := json.Unmarshal(bytes, &l); err == nil {
		if l.Operation != nil || l.Data != nil {
			i.ItemLazy = l
			return nil
		}
	}
	data := make([]byte, len(bytes))
	copy(data, bytes)
	// Fallback to old agency set
	i.ItemLazy = ItemLazy{
		Operation: refs.NewType(transaction.OperationSet),
		Data:      refs.NewType[json.RawMessage](data),
	}
	return nil
}
