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

type keyArrayReplace struct {
	KeyChanger
	newValue     any
	currentValue any
}

func NewKeyReplace(key Key, currentValue, newValue any) KeyChanger {
	return &keyArrayReplace{
		KeyChanger:   &keyCommon{key: key},
		newValue:     newValue,
		currentValue: currentValue,
	}
}
func (k *keyArrayReplace) GetNew() any {
	return k.newValue
}
func (k *keyArrayReplace) GetVal() any {
	return k.currentValue
}
func (k *keyArrayReplace) GetOperation() Operation {
	return OperationReplace
}
