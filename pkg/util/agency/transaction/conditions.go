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

type KeyConditioner interface {
	// GetName returns a name of condition.
	GetName() string
	// GetValue returns the value for which condition must be met
	GetValue() any
}
type Conditions map[string]KeyConditioner

func NewConditionIfEqual(value any) KeyConditioner {
	return &keyConditionIfEqual{
		value: value,
	}
}
func NewConditionIfNotEqual(value any) KeyConditioner {
	return &keyConditionIfNotEqual{
		value: value,
	}
}
func NewConditionOldEmpty(value bool) KeyConditioner {
	return &keyConditionOldEmpty{
		value: value,
	}
}
func NewConditionIsArray(value bool) KeyConditioner {
	return &keyConditionIsArray{
		value: value,
	}
}

type keyConditionIfEqual struct {
	value any
}
type keyConditionIfNotEqual struct {
	value any
}
type keyConditionOldEmpty struct {
	value bool
}
type keyConditionIsArray struct {
	value bool
}

func (k *keyConditionIfEqual) GetName() string {
	return "old"
}
func (k *keyConditionIfEqual) GetValue() any {
	return k.value
}
func (k *keyConditionIfNotEqual) GetName() string {
	return "oldNot"
}
func (k *keyConditionIfNotEqual) GetValue() any {
	return k.value
}
func (k *keyConditionOldEmpty) GetName() string {
	return "oldEmpty"
}
func (k *keyConditionOldEmpty) GetValue() any {
	return k.value
}
func (k *keyConditionIsArray) GetName() string {
	return "isArray"
}
func (k *keyConditionIsArray) GetValue() any {
	return k.value
}
