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

package deployment

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func Test_SpecAccept_Initial(t *testing.T) {
	spec := api.DeploymentSpec{}

	spec.SetDefaults(testDeploymentName)

	depl := &api.ArangoDeployment{
		ObjectMeta: meta.ObjectMeta{
			Name: testDeploymentName,
		},
		Spec: spec,
		Status: api.DeploymentStatus{
			AcceptedSpec: spec.DeepCopy(),
		},
	}

	d, _ := createTestDeployment(t, Config{}, depl)

	if _, err := d.deps.Client.Arango().DatabaseV1().ArangoDeployments(testNamespace).Create(context.Background(), depl, meta.CreateOptions{}); err != nil {
		require.NoError(t, err)
	}

	_, _, err := d.acceptNewSpec(context.Background(), depl)
	require.NoError(t, err)

	depl, err = d.deps.Client.Arango().DatabaseV1().ArangoDeployments(testNamespace).Get(context.Background(), testDeploymentName, meta.GetOptions{})
	require.NoError(t, err)

	checksum, err := spec.Checksum()
	require.NoError(t, err)

	require.NotNil(t, depl.Status.AcceptedSpecVersion)
	require.Equal(t, checksum, *depl.Status.AcceptedSpecVersion)
}
