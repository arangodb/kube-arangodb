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

package constants

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/require"
)

func withFile(t *testing.T, ns string) func(in func(file string)) {
	return func(in func(key string)) {
		p := t.TempDir()

		ret := path.Join(p, strings.ToLower(uniuri.NewLen(32)))

		require.NoError(t, os.WriteFile(ret, []byte(ns), 0644))

		in(ret)
	}
}

func withEnv(t *testing.T, ns string) func(in func(env string)) {
	return func(in func(key string)) {
		key := fmt.Sprintf("MY_NS_ENV_%s", strings.ToUpper(uniuri.NewLen(8)))

		require.NoError(t, os.Setenv(key, ns))
		defer func() {
			require.NoError(t, os.Unsetenv(key))
		}()

		in(key)
	}
}

func Test_Namespace(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		n, ok := namespaceWithSAAndEnv(PathMountServiceAccountNamespace, EnvOperatorPodNamespace)
		require.False(t, ok)
		require.EqualValues(t, "", n)
	})
	t.Run("With Env", func(t *testing.T) {
		withEnv(t, "myNs1")(func(env string) {
			n, ok := namespaceWithSAAndEnv(PathMountServiceAccountNamespace, env)
			require.True(t, ok)
			require.EqualValues(t, "myNs1", n)
		})
	})
	t.Run("With Empty Env", func(t *testing.T) {
		withEnv(t, "")(func(env string) {
			n, ok := namespaceWithSAAndEnv(PathMountServiceAccountNamespace, env)
			require.False(t, ok)
			require.EqualValues(t, "", n)
		})
	})
	t.Run("With Whitespace Env", func(t *testing.T) {
		withEnv(t, " \n ")(func(env string) {
			n, ok := namespaceWithSAAndEnv(PathMountServiceAccountNamespace, env)
			require.False(t, ok)
			require.EqualValues(t, "", n)
		})
	})
	t.Run("With File", func(t *testing.T) {
		withFile(t, "myNs2")(func(file string) {
			n, ok := namespaceWithSAAndEnv(file, EnvOperatorPodNamespace)
			require.True(t, ok)
			require.EqualValues(t, "myNs2", n)
		})
	})
	t.Run("With Missing File", func(t *testing.T) {
		withFile(t, "myNs2")(func(file string) {
			n, ok := namespaceWithSAAndEnv(fmt.Sprintf("%s.missing", file), EnvOperatorPodNamespace)
			require.False(t, ok)
			require.EqualValues(t, "", n)
		})
	})
	t.Run("With Empty File", func(t *testing.T) {
		withFile(t, "")(func(file string) {
			n, ok := namespaceWithSAAndEnv(file, EnvOperatorPodNamespace)
			require.False(t, ok)
			require.EqualValues(t, "", n)
		})
	})
	t.Run("With Whitespace File", func(t *testing.T) {
		withFile(t, " \n ")(func(file string) {
			n, ok := namespaceWithSAAndEnv(file, EnvOperatorPodNamespace)
			require.False(t, ok)
			require.EqualValues(t, "", n)
		})
	})
	t.Run("With File & Env", func(t *testing.T) {
		withFile(t, "myNs2")(func(file string) {
			withEnv(t, "myNs1")(func(env string) {
				n, ok := namespaceWithSAAndEnv(file, env)
				require.True(t, ok)
				require.EqualValues(t, "myNs1", n)
			})
		})
	})
}
