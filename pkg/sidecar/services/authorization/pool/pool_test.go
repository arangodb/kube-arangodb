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

package pool

import (
	"fmt"
	"testing"
	"time"

	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/require"

	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/db"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_Pool(t *testing.T) {
	client := db.NewClient(cache.NewObject(tests.TestArangoDBConfig(t).ClientCache())).Database(fmt.Sprintf("db-%s", uniuri.NewLen(8)))

	pz := NewPooler[*sidecarSvcAuthzTypes.Policy](client.
		CreateCollection("_policies", db.SourceCollectionProps("_users")).
		WithUniqueIndex("policies_unique_sequence_index", "sequence").
		WithTTLIndex("policies_deleted_index", 30*24*time.Hour, "deleted").
		Get(), DefaultPoolerTimeout)

	_, _, ok := pz.Item("test")
	require.False(t, ok)
}
