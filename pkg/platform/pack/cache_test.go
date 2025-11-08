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

package pack

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func TestCache(t *testing.T) {
	c := NewCache(t.TempDir())

	var d = make([]byte, 1024)

	checksum := util.SHA256(d)

	_, err := c.Get(checksum, "some/file")
	require.ErrorAs(t, err, &os.ErrNotExist)

	t.Run("Invalid Sha", func(t *testing.T) {
		out, err := c.CacheObject("ABCD", "some/file")
		require.NoError(t, err)

		z, err := out.Write(d)
		require.NoError(t, err)
		require.Len(t, d, z)

		require.Error(t, out.Close())

		require.EqualValues(t, 0, c.Saved())
	})

	t.Run("Valid Sha", func(t *testing.T) {
		out, err := c.CacheObject(util.SHA256(d), "some/file")
		require.NoError(t, err)

		z, err := out.Write(d)
		require.NoError(t, err)
		require.Len(t, d, z)

		require.NoError(t, out.Close())

		require.EqualValues(t, 1, c.Saved())
	})

	t.Run("Multi Upload Sha", func(t *testing.T) {
		out, err := c.CacheObject(util.SHA256(d), "some/file2")
		require.NoError(t, err)

		nout, err := c.CacheObject(util.SHA256(d), "some/file2")
		require.NoError(t, err)
		require.Nil(t, nout)

		z, err := out.Write(d)
		require.NoError(t, err)
		require.Len(t, d, z)

		require.NoError(t, out.Close())

		require.EqualValues(t, 1, c.Saved())
	})
}
