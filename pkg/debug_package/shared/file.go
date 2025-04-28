//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"sigs.k8s.io/yaml"
)

type GenFunc func(logger zerolog.Logger, files chan<- File) error
type DataFunc func() ([]byte, error)

type File interface {
	Path() string
	Write() ([]byte, error)
}

func NewJSONFile[T interface{}](path string, write func() ([]T, error)) File {
	return NewFile(path, func() ([]byte, error) {
		obj, err := write()
		if err != nil {
			return nil, err
		}

		return json.Marshal(obj)
	})
}

func NewYAMLFile[T interface{}](path string, write func() ([]T, error)) File {
	return NewFile(path, func() ([]byte, error) {
		obj, err := write()
		if err != nil {
			return nil, err
		}

		buff := bytes.NewBuffer(nil)

		for z := range obj {
			d, err := yaml.Marshal(obj[z])
			if err != nil {
				return nil, err
			}

			buff.Write(d)

			if z+1 < len(obj) {
				buff.Write([]byte("\n---\n\n"))
			}
		}

		return buff.Bytes(), nil
	})
}

func NewFile(path string, write DataFunc) File {
	return file{
		name:  path,
		write: write,
	}
}

type file struct {
	name  string
	write DataFunc
}

func (f file) Path() string {
	return f.name
}

func (f file) Write() ([]byte, error) {
	return f.write()
}

type Factory interface {
	Name() string
	Generate(logger zerolog.Logger, files chan<- File) error
	Enabled() bool
}

func NewFactory(name string, enabled bool, gens ...GenFunc) Factory {
	return factory{
		name:     name,
		enabled:  enabled,
		generate: gens,
	}
}

type factory struct {
	name     string
	enabled  bool
	generate []GenFunc
}

func (f factory) Enabled() bool {
	return f.enabled
}

func (f factory) Name() string {
	return f.name
}

func (f factory) Generate(logger zerolog.Logger, files chan<- File) error {
	for _, gen := range f.generate {
		if err := gen(logger, files); err != nil {
			return err
		}
	}

	return nil
}

func GenerateDataFuncP1[P1 any](call func(p1 P1) ([]byte, error), p1 P1) DataFunc {
	return func() ([]byte, error) {
		return call(p1)
	}
}

func GenerateDataFuncP2[P1, P2 any](call func(p1 P1, p2 P2) ([]byte, error), p1 P1, p2 P2) DataFunc {
	return func() ([]byte, error) {
		return call(p1, p2)
	}
}

func NewFactoryGen() FactoryGen {
	return &rootFactoryGen{}
}

type FactoryGen interface {
	AddSection(name string) FactoryGen

	Register(name string, enabled bool, gens GenFunc) FactoryGen

	Extend(in func(f FactoryGen)) FactoryGen

	Get() []Factory
}

type rootFactoryGen struct {
	lock sync.Mutex

	factories []Factory
}

func (r *rootFactoryGen) Extend(in func(f FactoryGen)) FactoryGen {
	in(r)
	return r
}

func (r *rootFactoryGen) Get() []Factory {
	r.lock.Lock()
	defer r.lock.Unlock()

	q := make([]Factory, len(r.factories))
	copy(q, r.factories)

	return q
}

func (r *rootFactoryGen) AddSection(name string) FactoryGen {
	return rootFactorySection{
		parent:  r,
		section: name,
	}
}

func (r *rootFactoryGen) Register(name string, enabled bool, gens GenFunc) FactoryGen {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.factories = append(r.factories, NewFactory(name, enabled, gens))

	return r
}

type rootFactorySection struct {
	parent FactoryGen

	section string
}

func (r rootFactorySection) Get() []Factory {
	return r.parent.Get()
}

func (r rootFactorySection) AddSection(name string) FactoryGen {
	return rootFactorySection{
		parent:  r,
		section: name,
	}
}

func (r rootFactorySection) Register(name string, enabled bool, gens GenFunc) FactoryGen {
	r.parent.Register(fmt.Sprintf("%s-%s", r.section, name), enabled, gens)
	return r
}

func (r rootFactorySection) Extend(in func(f FactoryGen)) FactoryGen {
	in(r)
	return r
}
