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

import (
	"sync"

	"k8s.io/apimachinery/pkg/util/yaml"
)

type Loader[T any] interface {
	MustGet() T

	Get() (T, error)
}

type LoaderGen[T any] func() (T, error)

type loader[T any] struct {
	lock sync.Mutex

	gen LoaderGen[T]

	object T

	generated bool
}

func (l *loader[T]) MustGet() T {
	obj, err := l.Get()
	if err != nil {
		panic(err.Error())
	}

	return obj
}

func (l *loader[T]) Get() (T, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.generated {
		return l.object, nil
	}

	obj, err := l.gen()
	if err != nil {
		return Default[T](), err
	}

	l.generated = true
	l.object = obj

	return l.object, nil
}

func NewLoader[T any](gen LoaderGen[T]) Loader[T] {
	return &loader[T]{
		gen: gen,
	}
}

func NewYamlLoader[T any](data []byte) Loader[T] {
	return NewLoader[T](func() (T, error) {
		var obj T

		if err := yaml.Unmarshal(data, &obj); err != nil {
			return Default[T](), err
		}

		return obj, nil
	})
}
