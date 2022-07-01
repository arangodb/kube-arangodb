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

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func NewArangoDeployment(name string) *api.ArangoDeployment {
	return &api.ArangoDeployment{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: FakeNamespace,
			UID:       uuid.NewUUID(),
		},
	}
}

func NewArangoClusterSynchronization(name string) *api.ArangoClusterSynchronization {
	return &api.ArangoClusterSynchronization{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: FakeNamespace,
			UID:       uuid.NewUUID(),
		},
	}
}

func RefreshArangoClusterSynchronization(t *testing.T, client kclient.Client, acs *api.ArangoClusterSynchronization) *api.ArangoClusterSynchronization {
	nacs, err := client.Arango().DatabaseV1().ArangoClusterSynchronizations(acs.GetNamespace()).Get(context.Background(), acs.GetName(), meta.GetOptions{})
	require.NoError(t, err)
	return nacs
}
