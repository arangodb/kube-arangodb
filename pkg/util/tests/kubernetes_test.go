//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	mlApi "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func NewMetaObjectRun[T meta.Object](t *testing.T) {
	var obj T
	t.Run(reflect.TypeOf(obj).String(), func(t *testing.T) {
		t.Run("Item", func(t *testing.T) {
			NewItem(t, operation.Update, NewMetaObject[T](t, "test", "test"))
		})
		t.Run("K8S", func(t *testing.T) {
			c := kclient.NewFakeClient()

			obj := NewMetaObject[T](t, "test", "test")

			require.NotNil(t, obj)

			refresh := CreateObjects(t, c.Kubernetes(), c.Arango(), &obj)

			refresh(t)
		})
	})
}

func Test_NewMetaObject(t *testing.T) {
	NewMetaObjectRun[*core.Secret](t)
	NewMetaObjectRun[*api.ArangoDeployment](t)
	NewMetaObjectRun[*api.ArangoClusterSynchronization](t)
	NewMetaObjectRun[*backupApi.ArangoBackup](t)
	NewMetaObjectRun[*mlApi.ArangoMLExtension](t)
}
