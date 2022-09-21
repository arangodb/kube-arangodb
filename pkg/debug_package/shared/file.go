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

package shared

import (
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/json"
)

type GenFunc func(cmd *cobra.Command, logger zerolog.Logger, files chan<- File) error

type File interface {
	Path() string
	Write() ([]byte, error)
}

func NewJSONFile(path string, write func() (interface{}, error)) File {
	return NewFile(path, func() ([]byte, error) {
		obj, err := write()
		if err != nil {
			return nil, err
		}

		return json.Marshal(obj)
	})
}

func NewFile(path string, write func() ([]byte, error)) File {
	return file{
		name:  path,
		write: write,
	}
}

type file struct {
	name  string
	write func() ([]byte, error)
}

func (f file) Path() string {
	return f.name
}

func (f file) Write() ([]byte, error) {
	return f.write()
}

type Factory interface {
	Name() string
	Generate(cmd *cobra.Command, logger zerolog.Logger, files chan<- File) error

	Init(command *cobra.Command)
}

func NewFactory(name string, cmd func(cmd *cobra.Command), gen GenFunc) Factory {
	return factory{
		name:     name,
		generate: gen,
		cmd:      cmd,
	}
}

type factory struct {
	name     string
	generate GenFunc
	cmd      func(cmd *cobra.Command)
}

func (f factory) Init(command *cobra.Command) {
	if c := f.cmd; c != nil {
		c(command)
	}
}

func (f factory) Name() string {
	return f.name
}

func (f factory) Generate(cmd *cobra.Command, logger zerolog.Logger, files chan<- File) error {
	return f.generate(cmd, logger, files)
}
