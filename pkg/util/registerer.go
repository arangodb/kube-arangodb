//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package util

import "sync"

func NewRegisterer[K comparable, V any]() Registerer[K, V] {
	return &registerer[K, V]{
		items: make(map[K]V),
	}
}

type Registerer[K comparable, V any] interface {
	Register(key K, value V) bool
	MustRegister(key K, value V)

	Items() []KV[K, V]
}

type registerer[K comparable, V any] struct {
	lock sync.Mutex

	items map[K]V
}

func (r *registerer[K, V]) Register(key K, value V) bool {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, ok := r.items[key]; ok {
		return false
	}

	r.items[key] = value

	return true
}

func (r *registerer[K, V]) MustRegister(key K, value V) {
	if !r.Register(key, value) {
		panic("Unable to register item")
	}
}

func (r *registerer[K, V]) Items() []KV[K, V] {
	r.lock.Lock()
	defer r.lock.Unlock()

	return Extract(r.items)
}
