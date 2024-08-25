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

package tests

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type FileGenerator interface {
	Parent(t *testing.T) FileGenerator

	Directory(t *testing.T, name string) FileGenerator

	File(t *testing.T, name string, data []byte) FileGenerator

	FileR(t *testing.T, name string, size int) FileGenerator
}

type fileGenerator struct {
	root string
	path string
}

func (f fileGenerator) FileR(t *testing.T, name string, size int) FileGenerator {
	var data = make([]byte, size)

	_, err := util.Rand().Read(data)
	require.NoError(t, err)

	return f.File(t, name, data)
}

func (f fileGenerator) Parent(t *testing.T) FileGenerator {
	require.NotEqual(t, f.root, f.path, "Unable to jump above root")

	return fileGenerator{
		root: f.root,
		path: path.Dir(f.path),
	}
}

func (f fileGenerator) Directory(t *testing.T, name string) FileGenerator {
	np := path.Join(f.path, name)

	if err := os.Mkdir(np, 0755); err != nil {
		if !errors.Is(err, os.ErrExist) {
			require.NoError(t, err)
		}
	}

	return fileGenerator{
		root: f.root,
		path: np,
	}
}

func (f fileGenerator) File(t *testing.T, name string, data []byte) FileGenerator {
	require.NoError(t, os.WriteFile(path.Join(f.path, name), data, 0644))
	return f
}

func NewFileGenerator(t *testing.T, root string) FileGenerator {
	if err := os.Mkdir(root, 0755); err != nil {
		if !errors.Is(err, os.ErrExist) {
			require.NoError(t, err)
		}
	}

	return fileGenerator{root: root, path: root}
}
