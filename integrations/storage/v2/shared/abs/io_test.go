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

package abs

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/uuid"

	pbImplStorageV2Shared "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test(t *testing.T) {
	var config = Configuration{
		BucketName:   tests.GetAzureBlobStorageContainer(t),
		BucketPrefix: fmt.Sprintf("tmp/unit-test/%s/", uuid.NewUUID()),
		MaxListKeys:  nil,
		Client:       tests.GetAzureConfig(t),
	}

	client, err := config.New()
	require.NoError(t, err)

	require.NoError(t, client.Init(shutdown.Context(), &pbImplStorageV2Shared.InitOptions{}))

	// List Done
	{
		objs, err := client.List(shutdown.Context(), "")
		require.NoError(t, err)
		data, err := objs.Next(shutdown.Context())
		require.NoError(t, err)
		require.Len(t, data, 0)
		_, err = objs.Next(shutdown.Context())
		require.ErrorIs(t, err, io.EOF)
	}

	{
		// Write
		data, err := client.Write(shutdown.Context(), "my-file.txt")
		require.NoError(t, err)

		require.False(t, data.Closed())

		_, err = data.Write([]byte("hello world"))
		require.NoError(t, err)

		checksum, bytes, err := data.Close(shutdown.Context())
		require.NoError(t, err)
		require.Equal(t, checksum, "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9")
		require.EqualValues(t, bytes, 11)

		require.True(t, data.Closed())
	}

	// List Done
	{
		objs, err := client.List(shutdown.Context(), "")
		require.NoError(t, err)
		data, err := objs.Next(shutdown.Context())
		require.NoError(t, err)
		require.Len(t, data, 1)
		_, err = objs.Next(shutdown.Context())
		require.ErrorIs(t, err, io.EOF)
	}

	{
		// Read
		_, err := client.Read(shutdown.Context(), "my-file2.txt")
		require.ErrorIs(t, err, os.ErrNotExist)

	}

	{
		// Read
		data, err := client.Read(shutdown.Context(), "my-file.txt")
		require.NoError(t, err)

		require.False(t, data.Closed())

		z, err := io.ReadAll(data)
		require.NoError(t, err)

		checksum, bytes, err := data.Close(shutdown.Context())
		require.NoError(t, err)
		require.Equal(t, checksum, "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9")
		require.EqualValues(t, bytes, 11)
		require.Len(t, z, 11)

		require.True(t, data.Closed())
	}
}
