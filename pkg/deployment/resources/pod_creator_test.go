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

package resources

import (
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func Test_CreatePodSuffix(t *testing.T) {
	spec := api.DeploymentSpec{
		Image: util.NewType[string]("arangodb/arangodb:latest"),
	}

	t.Run("Length is 6 hex characters", func(t *testing.T) {
		require.Len(t, CreatePodSuffix(spec), 6)
	})

	t.Run("Deterministic for the same spec", func(t *testing.T) {
		require.Equal(t, CreatePodSuffix(spec), CreatePodSuffix(spec))
	})

	t.Run("Differs for a different spec", func(t *testing.T) {
		other := api.DeploymentSpec{
			Image: util.NewType[string]("arangodb/arangodb:other"),
		}
		require.NotEqual(t, CreatePodSuffix(spec), CreatePodSuffix(other))
	})
}
