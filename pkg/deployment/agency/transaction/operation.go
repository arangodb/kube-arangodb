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

package transaction

import (
	goStrings "strings"

	"github.com/arangodb-helper/go-helper/pkg/refs"
)

type Operation string

const (
	OperationSet       Operation = "set"
	OperationIncrement Operation = "increment"
	OperationDecrement Operation = "decrement"
	OperationDelete    Operation = "delete"
	OperationPush      Operation = "push"
	OperationPop       Operation = "pop"
	OperationReplace   Operation = "replace"
	OperationErase     Operation = "erase"
)

type Key []string
type KeyChanger interface {
	// GetKey returns which key must be changed
	GetKey() string
	// GetNew returns new value for a key in the agency
	GetNew() any
	// GetOperation returns what type of operation must be performed on a key
	GetOperation() Operation
	// GetVal returns new value for a key in the agency
	GetVal() any
}
type keyCommon struct {
	key Key
}

func (o *Operation) Get() Operation {
	return refs.TypeOrDefault(o, OperationSet)
}
func (k Key) CreateSubKey(elements ...string) Key {
	NewKey := make(Key, 0, len(k)+len(elements))
	NewKey = append(NewKey, k...)
	NewKey = append(NewKey, elements...)
	return NewKey
}
func (k *keyCommon) GetKey() string {
	return CreateFullKey(k.key)
}
func (k *keyCommon) GetNew() any {
	return nil
}
func (k *keyCommon) GetOperation() Operation {
	return ""
}
func (k *keyCommon) GetVal() any {
	return nil
}
func CreateFullKey(key Key) string {
	return "/" + goStrings.Join(key, "/")
}
