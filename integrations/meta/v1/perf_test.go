//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	pbMetaV1 "github.com/arangodb/kube-arangodb/integrations/meta/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func duration(t *testing.T) func() {
	start := time.Now()
	return func() {
		elapsed := time.Since(start)
		t.Logf("Elapsed time %s", elapsed)
	}
}

func testWithSize(t *testing.T, threads, size int) {
	t.Run(fmt.Sprintf("%d T %d", size, threads), func(t *testing.T) {
		defer duration(t)()
		ctx, c := context.WithCancel(context.Background())
		defer c()

		client := Client(t, GetInternalRemoteCache(t), ctx)

		t.Run("Insert Data", func(t *testing.T) {
			defer duration(t)()

			util.ParallelProcess(func(in int) {
				_, err := client.Set(t.Context(), &pbMetaV1.SetRequest{
					Key: fmt.Sprintf("data.%09d", in),
				})
				require.NoError(t, err)
			}, threads, util.IntInput(size))
		})

		t.Run("Read Data", func(t *testing.T) {
			defer duration(t)()

			z := GetAllKeys(t, client, &pbMetaV1.ListRequest{
				Batch: util.NewType[int32](1024),
			})

			t.Logf("Read Data: %d", len(z))
		})

		t.Run("Read All Data", func(t *testing.T) {
			defer duration(t)()

			z := GetAllKeys(t, client, &pbMetaV1.ListRequest{
				Batch: util.NewType[int32](1024),
			})

			util.ParallelProcess(func(in string) {
				_, err := client.Get(t.Context(), &pbMetaV1.ObjectRequest{
					Key: in,
				})
				require.NoError(t, err)
			}, threads, z)
		})

		t.Run("Read All Data in Batches", func(t *testing.T) {
			defer duration(t)()

			z := GetAllKeys(t, client, &pbMetaV1.ListRequest{
				Batch: util.NewType[int32](1024),
			})

			for _, in := range util.BatchList(128, z) {
				_, err := client.GetBatch(t.Context(), &pbMetaV1.ObjectBatchRequest{
					Items: util.FormatList(in, func(s string) *pbMetaV1.ObjectRequest {
						return &pbMetaV1.ObjectRequest{
							Key: s,
						}
					}),
				})
				require.NoError(t, err)
			}
		})
	})
}

func Test_Performance(t *testing.T) {
	testWithSize(t, 32, 128)
	testWithSize(t, 32, 12000)
}
