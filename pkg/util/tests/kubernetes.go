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
	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
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
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

type handleFunc struct {
	in func(ctx context.Context) (bool, error)
}

func (h handleFunc) Name() string {
	return "mock"
}

func (h handleFunc) Handle(ctx context.Context, item operation.Item) error {
	_, err := h.in(ctx)
	return err
}

func (h handleFunc) CanBeHandled(item operation.Item) bool {
	return true
}

func HandleFunc(in func(ctx context.Context) (bool, error)) error {
	return Handle(handleFunc{in: in}, operation.Item{})
}

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
		case **batch.CronJob:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.BatchV1().CronJobs(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **batch.Job:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.BatchV1().Jobs(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **core.Pod:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.CoreV1().Pods(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **core.Secret:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.CoreV1().Secrets(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **core.Service:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.CoreV1().Services(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **core.ServiceAccount:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.CoreV1().ServiceAccounts(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **apps.StatefulSet:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.AppsV1().StatefulSets(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
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
		case **mlApi.ArangoMLStorage:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.MlV1alpha1().ArangoMLStorages(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **rbac.ClusterRole:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.RbacV1().ClusterRoles().Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **rbac.ClusterRoleBinding:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.RbacV1().ClusterRoleBindings().Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **rbac.Role:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.RbacV1().Roles(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **rbac.RoleBinding:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.RbacV1().RoleBindings(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **mlApi.ArangoMLBatchJob:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.MlV1alpha1().ArangoMLBatchJobs(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **mlApi.ArangoMLCronJob:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.MlV1alpha1().ArangoMLCronJobs(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		default:
			require.Fail(t, fmt.Sprintf("Unable to create object: %s", reflect.TypeOf(v).String()))
		}
	}

	return func(t *testing.T) {
		RefreshObjects(t, k8s, arango, objects...)
	}
}

func UpdateObjects(t *testing.T, k8s kubernetes.Interface, arango arangoClientSet.Interface, objects ...interface{}) func(t *testing.T) {
	for _, object := range objects {
		switch v := object.(type) {
		case **batch.CronJob:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.BatchV1().CronJobs(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **batch.Job:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.BatchV1().Jobs(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **core.Pod:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.CoreV1().Pods(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **core.Secret:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.CoreV1().Secrets(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **core.Service:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.CoreV1().Services(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **core.ServiceAccount:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.CoreV1().ServiceAccounts(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **apps.StatefulSet:
			require.NotNil(t, v)
			vl := *v
			_, err := k8s.AppsV1().StatefulSets(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **api.ArangoDeployment:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.DatabaseV1().ArangoDeployments(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **api.ArangoClusterSynchronization:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.DatabaseV1().ArangoClusterSynchronizations(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **backupApi.ArangoBackup:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.BackupV1().ArangoBackups(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **mlApi.ArangoMLExtension:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.MlV1alpha1().ArangoMLExtensions(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **mlApi.ArangoMLStorage:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.MlV1alpha1().ArangoMLStorages(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **mlApi.ArangoMLBatchJob:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.MlV1alpha1().ArangoMLBatchJobs(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **mlApi.ArangoMLCronJob:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.MlV1alpha1().ArangoMLCronJobs(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **rbac.ClusterRole:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.RbacV1().ClusterRoles().Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **rbac.ClusterRoleBinding:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.RbacV1().ClusterRoleBindings().Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **rbac.Role:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.RbacV1().Roles(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **rbac.RoleBinding:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.RbacV1().RoleBindings(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		default:
			require.Fail(t, fmt.Sprintf("Unable to create object: %s", reflect.TypeOf(v).String()))
		}
	}

	return func(t *testing.T) {
		RefreshObjects(t, k8s, arango, objects...)
	}
}

func DeleteObjects(t *testing.T, k8s kubernetes.Interface, arango arangoClientSet.Interface, objects ...interface{}) func(t *testing.T) {
	for _, object := range objects {
		switch v := object.(type) {
		case **batch.CronJob:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, k8s.BatchV1().CronJobs(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **batch.Job:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, k8s.BatchV1().Jobs(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **core.Pod:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, k8s.CoreV1().Pods(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **core.Secret:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, k8s.CoreV1().Secrets(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **core.ServiceAccount:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, k8s.CoreV1().ServiceAccounts(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **api.ArangoDeployment:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.DatabaseV1().ArangoDeployments(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **api.ArangoClusterSynchronization:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.DatabaseV1().ArangoClusterSynchronizations(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **backupApi.ArangoBackup:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.BackupV1().ArangoBackups(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **mlApi.ArangoMLExtension:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.MlV1alpha1().ArangoMLExtensions(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **mlApi.ArangoMLStorage:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.MlV1alpha1().ArangoMLStorages(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **mlApi.ArangoMLBatchJob:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.MlV1alpha1().ArangoMLBatchJobs(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **mlApi.ArangoMLCronJob:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.MlV1alpha1().ArangoMLCronJobs(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **rbac.ClusterRole:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, k8s.RbacV1().ClusterRoles().Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **rbac.ClusterRoleBinding:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, k8s.RbacV1().ClusterRoleBindings().Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **rbac.Role:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, k8s.RbacV1().Roles(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **rbac.RoleBinding:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, k8s.RbacV1().RoleBindings(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
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
		case **batch.CronJob:
			require.NotNil(t, v)

			vl := *v

			vn, err := k8s.BatchV1().CronJobs(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **batch.Job:
			require.NotNil(t, v)

			vl := *v

			vn, err := k8s.BatchV1().Jobs(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **core.Pod:
			require.NotNil(t, v)

			vl := *v

			vn, err := k8s.CoreV1().Pods(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **core.Secret:
			require.NotNil(t, v)

			vl := *v

			vn, err := k8s.CoreV1().Secrets(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **core.Service:
			require.NotNil(t, v)

			vl := *v

			vn, err := k8s.CoreV1().Services(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **core.ServiceAccount:
			require.NotNil(t, v)

			vl := *v

			vn, err := k8s.CoreV1().ServiceAccounts(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **apps.StatefulSet:
			require.NotNil(t, v)

			vl := *v
			vn, err := k8s.AppsV1().StatefulSets(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **api.ArangoDeployment:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.DatabaseV1().ArangoDeployments(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **api.ArangoClusterSynchronization:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.DatabaseV1().ArangoClusterSynchronizations(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **backupApi.ArangoBackup:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.BackupV1().ArangoBackups(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **mlApi.ArangoMLExtension:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.MlV1alpha1().ArangoMLExtensions(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **mlApi.ArangoMLStorage:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.MlV1alpha1().ArangoMLStorages(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **mlApi.ArangoMLBatchJob:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.MlV1alpha1().ArangoMLBatchJobs(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **mlApi.ArangoMLCronJob:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.MlV1alpha1().ArangoMLCronJobs(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **rbac.ClusterRole:
			require.NotNil(t, v)

			vl := *v

			vn, err := k8s.RbacV1().ClusterRoles().Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **rbac.ClusterRoleBinding:
			require.NotNil(t, v)

			vl := *v

			vn, err := k8s.RbacV1().ClusterRoleBindings().Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **rbac.Role:
			require.NotNil(t, v)

			vl := *v

			vn, err := k8s.RbacV1().Roles(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **rbac.RoleBinding:
			require.NotNil(t, v)

			vl := *v

			vn, err := k8s.RbacV1().RoleBindings(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		default:
			require.Fail(t, fmt.Sprintf("Unable to get object: %s", reflect.TypeOf(v).String()))
		}
	}
}

type MetaObjectMod[T meta.Object] func(t *testing.T, obj T)

func SetMetaBasedOnType(t *testing.T, object meta.Object) {
	switch v := object.(type) {
	case *batch.CronJob:
		v.Kind = "CronJob"
		v.APIVersion = "batch/v1"
		v.SetSelfLink(fmt.Sprintf("/api/batch/v1/cronjobs/%s/%s",
			object.GetNamespace(),
			object.GetName()))
	case *batch.Job:
		v.Kind = "Job"
		v.APIVersion = "batch/v1"
		v.SetSelfLink(fmt.Sprintf("/api/batch/v1/jobs/%s/%s",
			object.GetNamespace(),
			object.GetName()))
	case *core.Pod:
		v.Kind = "Pod"
		v.APIVersion = "v1"
		v.SetSelfLink(fmt.Sprintf("/api/v1/Pods/%s/%s",
			object.GetNamespace(),
			object.GetName()))
	case *core.Secret:
		v.Kind = "Secret"
		v.APIVersion = "v1"
		v.SetSelfLink(fmt.Sprintf("/api/v1/secrets/%s/%s",
			object.GetNamespace(),
			object.GetName()))
	case *core.Service:
		v.Kind = "Service"
		v.APIVersion = "v1"
		v.SetSelfLink(fmt.Sprintf("/api/v1/services/%s/%s",
			object.GetNamespace(),
			object.GetName()))
	case *core.ServiceAccount:
		v.Kind = "ServiceAccount"
		v.APIVersion = "v1"
		v.SetSelfLink(fmt.Sprintf("/api/v1/serviceaccounts/%s/%s",
			object.GetNamespace(),
			object.GetName()))
	case *apps.StatefulSet:
		v.Kind = "StatefulSet"
		v.APIVersion = "v1"
		v.SetSelfLink(fmt.Sprintf("/api/apps/v1/statefulsets/%s/%s",
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
	case *mlApi.ArangoMLStorage:
		v.Kind = ml.ArangoMLStorageResourceKind
		v.APIVersion = mlApi.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			mlApi.SchemeGroupVersion.String(),
			ml.ArangoMLStorageResourcePlural,
			object.GetNamespace(),
			object.GetName()))
	case *rbac.ClusterRole:
		v.Kind = "ClusterRole"
		v.APIVersion = "rbac.authorization.k8s.io/v1"
		v.SetSelfLink(fmt.Sprintf("/api/rbac.authorization.k8s.io/v1/clusterroles/%s/%s",
			object.GetNamespace(),
			object.GetName()))
	case *rbac.ClusterRoleBinding:
		v.Kind = "ClusterRoleBinding"
		v.APIVersion = "rbac.authorization.k8s.io/v1"
		v.SetSelfLink(fmt.Sprintf("/api/rbac.authorization.k8s.io/v1/clusterrolebingings/%s/%s",
			object.GetNamespace(),
			object.GetName()))
	case *rbac.Role:
		v.Kind = "Role"
		v.APIVersion = "rbac.authorization.k8s.io/v1"
		v.SetSelfLink(fmt.Sprintf("/api/rbac.authorization.k8s.io/v1/roles/%s/%s",
			object.GetNamespace(),
			object.GetName()))
	case *rbac.RoleBinding:
		v.Kind = "RoleBinding"
		v.APIVersion = "rbac.authorization.k8s.io/v1"
		v.SetSelfLink(fmt.Sprintf("/api/rbac.authorization.k8s.io/v1/rolebingings/%s/%s",
			object.GetNamespace(),
			object.GetName()))
	case *mlApi.ArangoMLBatchJob:
		v.Kind = ml.ArangoMLBatchJobResourceKind
		v.APIVersion = mlApi.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			mlApi.SchemeGroupVersion.String(),
			ml.ArangoMLBatchJobResourcePlural,
			object.GetNamespace(),
			object.GetName()))
	case *mlApi.ArangoMLCronJob:
		v.Kind = ml.ArangoMLCronJobResourceKind
		v.APIVersion = mlApi.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			mlApi.SchemeGroupVersion.String(),
			ml.ArangoMLCronJobResourcePlural,
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

	if IsNamespaced(obj) {
		obj.SetNamespace(namespace)
	}
	obj.SetName(name)
	obj.SetUID(uuid.NewUUID())

	SetMetaBasedOnType(t, obj)

	for _, mod := range mods {
		mod(t, obj)
	}

	return obj
}

func IsNamespaced(in meta.Object) bool {
	switch in.(type) {
	case *rbac.ClusterRole, *rbac.ClusterRoleBinding:
		return false
	default:
		return true
	}
}

func NewItem(t *testing.T, o operation.Operation, object meta.Object) operation.Item {
	item := operation.Item{
		Operation: o,

		Namespace: object.GetNamespace(),
		Name:      object.GetName(),
	}

	switch v := object.(type) {
	case *batch.CronJob:
		item.Group = "batch"
		item.Version = "v1"
		item.Kind = "CronJob"
	case *batch.Job:
		item.Group = "batch"
		item.Version = "v1"
		item.Kind = "Job"
	case *core.Pod:
		item.Group = ""
		item.Version = "v1"
		item.Kind = "Pod"
	case *core.Secret:
		item.Group = ""
		item.Version = "v1"
		item.Kind = "Secret"
	case *core.Service:
		item.Group = ""
		item.Version = "v1"
		item.Kind = "Service"
	case *core.ServiceAccount:
		item.Group = ""
		item.Version = "v1"
		item.Kind = "ServiceAccount"
	case *apps.StatefulSet:
		item.Group = "apps"
		item.Version = "v1"
		item.Kind = "StatefulSet"
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
	case *mlApi.ArangoMLStorage:
		item.Group = ml.ArangoMLGroupName
		item.Version = mlApi.ArangoMLVersion
		item.Kind = ml.ArangoMLStorageResourceKind
	case *rbac.ClusterRole:
		item.Group = "rbac.authorization.k8s.io"
		item.Version = "v1"
		item.Kind = "ClusterRole"
	case *rbac.ClusterRoleBinding:
		item.Group = "rbac.authorization.k8s.io"
		item.Version = "v1"
		item.Kind = "ClusterRoleBinding"
	case *rbac.Role:
		item.Group = "rbac.authorization.k8s.io"
		item.Version = "v1"
		item.Kind = "Role"
	case *rbac.RoleBinding:
		item.Group = "rbac.authorization.k8s.io"
		item.Version = "v1"
		item.Kind = "RoleBinding"
	case *mlApi.ArangoMLBatchJob:
		item.Group = ml.ArangoMLGroupName
		item.Version = mlApi.ArangoMLVersion
		item.Kind = ml.ArangoMLBatchJobResourceKind
	case *mlApi.ArangoMLCronJob:
		item.Group = ml.ArangoMLGroupName
		item.Version = mlApi.ArangoMLVersion
		item.Kind = ml.ArangoMLCronJobResourceKind
	default:
		require.Fail(t, fmt.Sprintf("Unable to create object: %s", reflect.TypeOf(v).String()))
	}

	return item
}
