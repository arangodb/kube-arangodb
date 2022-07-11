//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package upgrade

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/uuid"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func testMemberCIDAppendPrepare(t *testing.T, obj *api.ArangoDeployment) {
	t.Run("Member CID Append", func(t *testing.T) {
		obj.UID = uuid.NewUUID()

		obj.Status.Members.Agents = append(obj.Status.Members.Agents, api.MemberStatus{
			ID:        "CIDAppend",
			ClusterID: "",
		})
	})
}

func testMemberCIDAppendCheck(t *testing.T, obj api.ArangoDeployment) {
	t.Run("Member CID Append", func(t *testing.T) {
		m, g, ok := obj.Status.Members.ElementByID("CIDAppend")
		require.True(t, ok)
		require.Equal(t, api.ServerGroupAgents, g)
		require.Equal(t, obj.GetUID(), m.ClusterID)
	})
}
