//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package s3

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/uuid"

	pbImplStorageV2Shared "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func getClient(t *testing.T) pbImplStorageV2Shared.IO {
	var cfg Configuration

	cfg.Client = tests.GetAWSClientConfig(t)
	cfg.BucketName = tests.GetAWSS3Bucket(t)
	cfg.BucketPrefix = fmt.Sprintf("test/%s/", uuid.NewUUID())

	z, err := cfg.New()
	require.NoError(t, err)

	return z
}

func Test(t *testing.T) {
	t.Skipf("DATA")

	w := getClient(t)

	data := make([]byte, 1024*1024*64)

	for id := range data {
		data[id] = 0
	}

	ctx, c := context.WithCancel(context.Background())
	defer c()

	q, err := w.Write(ctx, "test.data")
	require.NoError(t, err)

	_, err = util.WriteAll(q, data)
	require.NoError(t, err)

	checksum, size, err := q.Close(context.Background())
	require.NoError(t, err)

	t.Logf("Write Checksum: %s", checksum)

	require.EqualValues(t, 1024*1024*64, size)

	r, err := w.Read(context.Background(), "test.data")
	require.NoError(t, err)

	data, err = io.ReadAll(r)
	require.NoError(t, err)

	echecksum, esize, err := r.Close(context.Background())
	require.NoError(t, err)

	require.EqualValues(t, 1024*1024*64, esize)
	require.Len(t, data, 1024*1024*64)

	t.Logf("Read Checksum: %s", echecksum)

	require.EqualValues(t, echecksum, checksum)

	removed, err := w.Delete(ctx, "test.data")
	require.NoError(t, err)
	require.True(t, removed)
}
