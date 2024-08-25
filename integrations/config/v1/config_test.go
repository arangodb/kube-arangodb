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

package v1

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ConfigCreation(t *testing.T) {
	create := func(config Config) error {
		_, err := New(config)
		return err
	}

	dir := t.TempDir()

	require.NoError(t, os.WriteFile(fmt.Sprintf("%s/file", dir), []byte{}, 0644))

	require.EqualError(t, create(Config{}), "Requires at least 1 module")

	require.EqualError(t, create(Config{
		Modules: map[string]ModuleDefinition{
			"test": {},
		},
	}), "Path for module `test` cannot be empty")

	require.EqualError(t, create(Config{
		Modules: map[string]ModuleDefinition{
			"test": {
				Path: "some/relative/path",
			},
		},
	}), "Path `some/relative/path` for module `test` needs to be absolute")

	require.EqualError(t, create(Config{
		Modules: map[string]ModuleDefinition{
			"test": {
				Path: fmt.Sprintf("%s/non-existent", dir),
			},
		},
	}), fmt.Sprintf("Path `%s/non-existent` for module `test` does not exists", dir))

	require.EqualError(t, create(Config{
		Modules: map[string]ModuleDefinition{
			"test": {
				Path: fmt.Sprintf("%s/file", dir),
			},
		},
	}), fmt.Sprintf("Path `%s/file` for module `test` is not a directory", dir))

	require.NoError(t, create(Config{
		Modules: map[string]ModuleDefinition{
			"test": {
				Path: dir,
			},
		},
	}))
}
