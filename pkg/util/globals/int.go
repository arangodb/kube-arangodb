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

package globals

type Int interface {
	Set(in int)
	Get() int
}

func NewInt(def int) Int {
	return &intObj{i: def}
}

type intObj struct {
	i int
}

func (i *intObj) Set(in int) {
	i.i = in
}

func (i *intObj) Get() int {
	return i.i
}

type Int64 interface {
	Set(in int64)
	Get() int64
}

func NewInt64(def int64) Int64 {
	return &int64Obj{i: def}
}

type int64Obj struct {
	i int64
}

func (i *int64Obj) Set(in int64) {
	i.i = in
}

func (i *int64Obj) Get() int64 {
	return i.i
}
