//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package cli

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func testFlag[T any](t *testing.T, f Flag[T]) {
	cmd := &cobra.Command{}

	require.NoError(t, f.Register(cmd))

	cmd.Run = func(cmd *cobra.Command, args []string) {
		z, err := f.Get(cmd)
		require.NoError(t, err)

		require.EqualValues(t, f.Default, z)
	}

	require.NoError(t, cmd.Execute())
}

func testFlagType[T any](t *testing.T) {
	t.Run(reflect.TypeOf(util.Default[T]()).String(), func(t *testing.T) {
		t.Run("Local", func(t *testing.T) {
			testFlag[T](t, Flag[T]{
				Name:        "flag",
				Description: "",
			})
		})
		t.Run("Persistent", func(t *testing.T) {
			testFlag[T](t, Flag[T]{
				Name:        "flag",
				Description: "",
				Persistent:  true,
			})
		})
	})
}

func Test_FlagTypes(t *testing.T) {
	testFlagType[string](t)
}
