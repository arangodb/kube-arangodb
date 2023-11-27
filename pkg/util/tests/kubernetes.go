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
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"

	"github.com/arangodb/kube-arangodb/pkg/apis/backup"
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/ml"
	mlApi "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1alpha1"
	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func Handle(handler operator.Handler, item operation.Item) error {
	return HandleWithMax(handler, item, 128)
}

func HandleWithMax(handler operator.Handler, item operation.Item, max int) error {
	for i := 0; i < max; i++ {
		if err := handler.Handle(context.Background(), item); err != nil {
			if operator.IsReconcile(err) {
				continue
			}

			return err
		}

		return nil
	}

	return errors.Newf("Max retries reached")
}

type KubernetesObject interface {
	meta.Object
	meta.Type
}

func CreateObjects(t *testing.T, k8s kubernetes.Interface, arango arangoClientSet.Interface, objects ...interface{}) func(t *testing.T) {
	for _, object := range objects {
		switch v := object.(type) {
		case **core.Secret:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.CoreV1().Secrets(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **api.ArangoDeployment:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.DatabaseV1().ArangoDeployments(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **api.ArangoClusterSynchronization:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.DatabaseV1().ArangoClusterSynchronizations(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **backupApi.ArangoBackup:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.BackupV1().ArangoBackups(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **mlApi.ArangoMLExtension:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.MlV1alpha1().ArangoMLExtensions(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		default:
			require.Fail(t, fmt.Sprintf("Unable to create object: %s", reflect.TypeOf(v).String()))
		}
	}

	return func(t *testing.T) {
		RefreshObjects(t, k8s, arango, objects...)
	}
}

func RefreshObjects(t *testing.T, k8s kubernetes.Interface, arango arangoClientSet.Interface, objects ...interface{}) {
	for _, object := range objects {
		switch v := object.(type) {
		case **core.Secret:
			require.NotNil(t, v)

			vl := *v

			vn, err := k8s.CoreV1().Secrets(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			require.NoError(t, err)

			*v = vn
		case **api.ArangoDeployment:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.DatabaseV1().ArangoDeployments(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			require.NoError(t, err)

			*v = vn
		case **api.ArangoClusterSynchronization:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.DatabaseV1().ArangoClusterSynchronizations(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			require.NoError(t, err)

			*v = vn
		case **backupApi.ArangoBackup:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.BackupV1().ArangoBackups(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			require.NoError(t, err)

			*v = vn
		case **mlApi.ArangoMLExtension:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.MlV1alpha1().ArangoMLExtensions(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			require.NoError(t, err)

			*v = vn
		default:
			require.Fail(t, fmt.Sprintf("Unable to create object: %s", reflect.TypeOf(v).String()))
		}
	}
}

type MetaObjectMod[T meta.Object] func(t *testing.T, obj T)

func SetMetaBasedOnType(t *testing.T, object meta.Object) {
	switch v := object.(type) {
	case *core.Secret:
		v.Kind = "Secret"
		v.APIVersion = "v1"
		v.SetSelfLink(fmt.Sprintf("/api/v1/secrets/%s/%s",
			object.GetNamespace(),
			object.GetName()))
	case *api.ArangoDeployment:
		v.Kind = deployment.ArangoDeploymentResourceKind
		v.APIVersion = api.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			api.SchemeGroupVersion.String(),
			deployment.ArangoDeploymentResourcePlural,
			object.GetNamespace(),
			object.GetName()))
	case *api.ArangoClusterSynchronization:
		v.Kind = deployment.ArangoClusterSynchronizationResourceKind
		v.APIVersion = api.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			api.SchemeGroupVersion.String(),
			deployment.ArangoClusterSynchronizationResourcePlural,
			object.GetNamespace(),
			object.GetName()))
	case *backupApi.ArangoBackup:
		v.Kind = backup.ArangoBackupResourceKind
		v.APIVersion = backupApi.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			backupApi.SchemeGroupVersion.String(),
			backup.ArangoBackupResourcePlural,
			object.GetNamespace(),
			object.GetName()))
	case *mlApi.ArangoMLExtension:
		v.Kind = ml.ArangoMLExtensionResourceKind
		v.APIVersion = mlApi.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			mlApi.SchemeGroupVersion.String(),
			ml.ArangoMLExtensionResourcePlural,
			object.GetNamespace(),
			object.GetName()))
	default:
		require.Fail(t, fmt.Sprintf("Unable to create object: %s", reflect.TypeOf(v).String()))
	}
}

func NewMetaObject[T meta.Object](t *testing.T, namespace, name string, mods ...MetaObjectMod[T]) T {
	var obj T

	if objT := reflect.TypeOf(obj); objT.Kind() == reflect.Pointer {
		newObj := reflect.New(objT.Elem())

		reflect.ValueOf(&obj).Elem().Set(newObj)
	}

	obj.SetNamespace(namespace)
	obj.SetName(name)
	obj.SetUID(uuid.NewUUID())

	SetMetaBasedOnType(t, obj)

	for _, mod := range mods {
		mod(t, obj)
	}

	return obj
}

func NewItem(t *testing.T, o operation.Operation, object meta.Object) operation.Item {
	item := operation.Item{
		Operation: o,

		Namespace: object.GetNamespace(),
		Name:      object.GetName(),
	}

	switch v := object.(type) {
	case *core.Secret:
		item.Group = ""
		item.Version = "v1"
		item.Kind = "Secret"
	case *api.ArangoDeployment:
		item.Group = deployment.ArangoDeploymentGroupName
		item.Version = api.ArangoDeploymentVersion
		item.Kind = deployment.ArangoDeploymentResourceKind
	case *api.ArangoClusterSynchronization:
		item.Group = deployment.ArangoDeploymentGroupName
		item.Version = api.ArangoDeploymentVersion
		item.Kind = deployment.ArangoClusterSynchronizationResourceKind
	case *backupApi.ArangoBackup:
		item.Group = backup.ArangoBackupGroupName
		item.Version = backupApi.ArangoBackupVersion
		item.Kind = backup.ArangoBackupResourceKind
	case *mlApi.ArangoMLExtension:
		item.Group = ml.ArangoMLGroupName
		item.Version = mlApi.ArangoMLVersion
		item.Kind = ml.ArangoMLExtensionResourceKind
	default:
		require.Fail(t, fmt.Sprintf("Unable to create object: %s", reflect.TypeOf(v).String()))
	}

	return item
}
