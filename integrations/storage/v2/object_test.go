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

package v2

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	pbImplStorageV2Shared "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

type testObjectFunc func(t *testing.T, mods ...util.Mod[platformApi.ArangoPlatformStorage]) (string, string, kclient.Client)

func listAll(t *testing.T, in util.NextIterator[[]pbImplStorageV2Shared.File]) []pbImplStorageV2Shared.File {
	var r []pbImplStorageV2Shared.File

	for {
		obj, err := in.Next(shutdown.Context())
		if err != nil {
			if err == io.EOF {
				break
			}

			require.NoError(t, err)
		}

		r = append(r, obj...)
	}

	return r
}

func testObject(t *testing.T, gen testObjectFunc, mods ...util.Mod[platformApi.ArangoPlatformStorage]) {
	name, namespace, client := gen(t, mods...)

	storage, err := client.Arango().PlatformV1alpha1().ArangoPlatformStorages(namespace).Get(shutdown.Context(), name, meta.GetOptions{})
	require.NoError(t, err)

	require.NoError(t, storage.Spec.Validate())

	ios, err := NewIOFromObject(shutdown.Context(), client, storage)
	require.NoError(t, err)

	t.Run("Init", func(t *testing.T) {
		require.NoError(t, ios.Init(shutdown.Context(), &pbImplStorageV2Shared.InitOptions{}))
	})

	t.Run("Data IO", func(t *testing.T) {
		data := make([]byte, 1024*1024*4)

		for id := range data {
			data[id] = 0
		}

		q, err := ios.Write(shutdown.Context(), "test.data")
		require.NoError(t, err)

		_, err = util.WriteAll(q, data)
		require.NoError(t, err)

		checksum, size, err := q.Close(shutdown.Context())
		require.NoError(t, err)

		t.Logf("Write Checksum: %s", checksum)

		require.EqualValues(t, 1024*1024*4, size)

		r, err := ios.Read(shutdown.Context(), "test.data")
		require.NoError(t, err)

		data, err = io.ReadAll(r)
		require.NoError(t, err)

		echecksum, esize, err := r.Close(shutdown.Context())
		require.NoError(t, err)

		require.EqualValues(t, 1024*1024*4, esize)
		require.Len(t, data, 1024*1024*4)

		t.Logf("Read Checksum: %s", echecksum)

		require.EqualValues(t, echecksum, checksum)

		removed, err := ios.Delete(shutdown.Context(), "test.data")
		require.NoError(t, err)
		require.True(t, removed)
	})

	t.Run("Lister", func(t *testing.T) {
		t.Run("Ensure empty", func(t *testing.T) {
			objectIter, err := ios.List(shutdown.Context(), "")
			require.NoError(t, err)

			objects := listAll(t, objectIter)
			require.Len(t, objects, 0)
		})

		// Create Files
		files := map[string]int{
			"test.data":           128,
			"path/test.data":      64,
			"paths/test.data":     127,
			"path/test/test.data": 256,
		}

		t.Run("Create Files", func(t *testing.T) {
			for k, s := range files {
				t.Run(k, func(t *testing.T) {
					data := make([]byte, s)

					q, err := ios.Write(shutdown.Context(), k)
					require.NoError(t, err)

					_, err = util.WriteAll(q, data)
					require.NoError(t, err)

					checksum, size, err := q.Close(shutdown.Context())
					require.NoError(t, err)

					t.Logf("Write Checksum: %s", checksum)

					require.EqualValues(t, s, size)

					r, err := ios.Read(shutdown.Context(), k)
					require.NoError(t, err)

					data, err = io.ReadAll(r)
					require.NoError(t, err)

					echecksum, esize, err := r.Close(shutdown.Context())
					require.NoError(t, err)

					require.EqualValues(t, s, esize)
					require.Len(t, data, s)

					t.Logf("Read Checksum: %s", echecksum)

					require.EqualValues(t, echecksum, checksum)
				})
			}
		})

		t.Run("Ensure Created", func(t *testing.T) {
			objectIter, err := ios.List(shutdown.Context(), "")
			require.NoError(t, err)

			objects := listAll(t, objectIter)
			require.Len(t, objects, len(files))
		})

		t.Run("Ensure Unknown Path List", func(t *testing.T) {
			objectIter, err := ios.List(shutdown.Context(), "unknown/")
			require.NoError(t, err)

			objects := listAll(t, objectIter)
			require.Len(t, objects, 0)
		})

		t.Run("Ensure Known Path List", func(t *testing.T) {
			objectIter, err := ios.List(shutdown.Context(), "path/")
			require.NoError(t, err)

			objects := listAll(t, objectIter)
			require.Len(t, objects, 2)
		})

		t.Run("Ensure Known Path With File List", func(t *testing.T) {
			objectIter, err := ios.List(shutdown.Context(), "path/test")
			require.NoError(t, err)

			objects := listAll(t, objectIter)
			require.Len(t, objects, 2)
		})

		t.Run("Ensure Known Path With Subpath List", func(t *testing.T) {
			objectIter, err := ios.List(shutdown.Context(), "path/test/")
			require.NoError(t, err)

			objects := listAll(t, objectIter)
			require.Len(t, objects, 1)
		})

		t.Run("Ensure File List", func(t *testing.T) {
			objectIter, err := ios.List(shutdown.Context(), "path/test/test.data")
			require.NoError(t, err)

			objects := listAll(t, objectIter)
			require.Len(t, objects, 1)
		})

		t.Run("Delete Files", func(t *testing.T) {
			for k := range files {
				t.Run(k, func(t *testing.T) {
					removed, err := ios.Delete(shutdown.Context(), k)
					require.NoError(t, err)
					require.True(t, removed)
				})
			}
		})

		t.Run("Re-Ensure empty", func(t *testing.T) {
			objectIter, err := ios.List(shutdown.Context(), "")
			require.NoError(t, err)

			objects := listAll(t, objectIter)
			require.Len(t, objects, 0)
		})
	})
}
