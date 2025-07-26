//
// DISCLAIMER
//
// Copyright 2023-2025 ArangoDB GmbH, Cologne, Germany
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

	"github.com/arangodb/kube-arangodb/pkg/apis/analytics"
	analyticsApi "github.com/arangodb/kube-arangodb/pkg/apis/analytics/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/apis/backup"
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/ml"
	mlApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1alpha1"
	mlApi "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/apis/networking"
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/apis/platform"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/apis/scheduler"
	schedulerApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
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

	return errors.Errorf("Max retries reached")
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
		case **core.ConfigMap:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.CoreV1().ConfigMaps(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **core.Service:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.CoreV1().Services(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **core.Endpoints:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.CoreV1().Endpoints(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
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
		case **apps.Deployment:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.AppsV1().Deployments(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
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
		case **backupApi.ArangoBackupPolicy:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.BackupV1().ArangoBackupPolicies(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **mlApi.ArangoMLExtension:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.MlV1beta1().ArangoMLExtensions(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **mlApi.ArangoMLStorage:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.MlV1beta1().ArangoMLStorages(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **mlApiv1alpha1.ArangoMLExtension:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.MlV1alpha1().ArangoMLExtensions(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **mlApiv1alpha1.ArangoMLStorage:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.MlV1alpha1().ArangoMLStorages(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **mlApiv1alpha1.ArangoMLBatchJob:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.MlV1alpha1().ArangoMLBatchJobs(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **mlApiv1alpha1.ArangoMLCronJob:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.MlV1alpha1().ArangoMLCronJobs(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
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
		case **schedulerApiv1alpha1.ArangoProfile:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.SchedulerV1alpha1().ArangoProfiles(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **schedulerApi.ArangoProfile:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.SchedulerV1beta1().ArangoProfiles(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **schedulerApi.ArangoSchedulerPod:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.SchedulerV1beta1().ArangoSchedulerPods(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **schedulerApi.ArangoSchedulerDeployment:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.SchedulerV1beta1().ArangoSchedulerDeployments(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **schedulerApi.ArangoSchedulerBatchJob:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.SchedulerV1beta1().ArangoSchedulerBatchJobs(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **schedulerApi.ArangoSchedulerCronJob:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.SchedulerV1beta1().ArangoSchedulerCronJobs(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **analyticsApi.GraphAnalyticsEngine:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.AnalyticsV1alpha1().GraphAnalyticsEngines(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **networkingApi.ArangoRoute:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.NetworkingV1beta1().ArangoRoutes(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **platformApi.ArangoPlatformStorage:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.PlatformV1beta1().ArangoPlatformStorages(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **platformApi.ArangoPlatformChart:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.PlatformV1beta1().ArangoPlatformCharts(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		case **platformApi.ArangoPlatformService:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.PlatformV1beta1().ArangoPlatformServices(vl.GetNamespace()).Create(context.Background(), vl, meta.CreateOptions{})
			require.NoError(t, err)
		default:
			require.Fail(t, fmt.Sprintf("Unable to create object: %s", reflect.TypeOf(v).String()))
		}
	}

	return func(t *testing.T) {
		RefreshObjects(t, k8s, arango, objects...)
	}
}

func UpdateObjectsC(t *testing.T, client kclient.Client, objects ...interface{}) func(t *testing.T) {
	return UpdateObjects(t, client.Kubernetes(), client.Arango(), objects...)
}

func Update[T meta.Object](t *testing.T, k8s kubernetes.Interface, arango arangoClientSet.Interface, obj *T, mods ...MetaObjectMod[T]) {
	z := *obj
	Apply[T](t, z, mods...)
	UpdateObjects(t, k8s, arango, &z)
	*obj = z
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
		case **core.ConfigMap:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.CoreV1().ConfigMaps(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **core.Service:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.CoreV1().Services(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **core.Endpoints:
			require.NotNil(t, v)

			vl := *v
			_, err := k8s.CoreV1().Endpoints(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
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
		case **apps.Deployment:
			require.NotNil(t, v)
			vl := *v
			_, err := k8s.AppsV1().Deployments(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
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
		case **backupApi.ArangoBackupPolicy:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.BackupV1().ArangoBackupPolicies(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **mlApi.ArangoMLExtension:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.MlV1beta1().ArangoMLExtensions(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **mlApi.ArangoMLStorage:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.MlV1beta1().ArangoMLStorages(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **mlApiv1alpha1.ArangoMLExtension:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.MlV1alpha1().ArangoMLExtensions(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **mlApiv1alpha1.ArangoMLStorage:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.MlV1alpha1().ArangoMLStorages(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **mlApiv1alpha1.ArangoMLBatchJob:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.MlV1alpha1().ArangoMLBatchJobs(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **mlApiv1alpha1.ArangoMLCronJob:
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
		case **schedulerApiv1alpha1.ArangoProfile:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.SchedulerV1alpha1().ArangoProfiles(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **schedulerApi.ArangoProfile:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.SchedulerV1beta1().ArangoProfiles(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **schedulerApi.ArangoSchedulerPod:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.SchedulerV1beta1().ArangoSchedulerPods(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **schedulerApi.ArangoSchedulerDeployment:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.SchedulerV1beta1().ArangoSchedulerDeployments(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **schedulerApi.ArangoSchedulerBatchJob:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.SchedulerV1beta1().ArangoSchedulerBatchJobs(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **schedulerApi.ArangoSchedulerCronJob:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.SchedulerV1beta1().ArangoSchedulerCronJobs(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **analyticsApi.GraphAnalyticsEngine:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.AnalyticsV1alpha1().GraphAnalyticsEngines(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **networkingApi.ArangoRoute:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.NetworkingV1beta1().ArangoRoutes(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **platformApi.ArangoPlatformStorage:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.PlatformV1beta1().ArangoPlatformStorages(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **platformApi.ArangoPlatformChart:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.PlatformV1beta1().ArangoPlatformCharts(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		case **platformApi.ArangoPlatformService:
			require.NotNil(t, v)

			vl := *v
			_, err := arango.PlatformV1beta1().ArangoPlatformServices(vl.GetNamespace()).Update(context.Background(), vl, meta.UpdateOptions{})
			require.NoError(t, err)
		default:
			require.Fail(t, fmt.Sprintf("Unable to update object: %s", reflect.TypeOf(v).String()))
		}
	}

	return func(t *testing.T) {
		RefreshObjects(t, k8s, arango, objects...)
	}
}

func DeleteObjects(t *testing.T, k8s kubernetes.Interface, arango arangoClientSet.Interface, objects ...interface{}) {
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
		case **core.ConfigMap:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, k8s.CoreV1().ConfigMaps(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **core.Service:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, k8s.CoreV1().Services(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **core.Endpoints:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, k8s.CoreV1().Endpoints(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **core.ServiceAccount:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, k8s.CoreV1().ServiceAccounts(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **apps.StatefulSet:
			require.NotNil(t, v)
			vl := *v
			err := k8s.AppsV1().StatefulSets(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{})
			require.NoError(t, err)
		case **apps.Deployment:
			require.NotNil(t, v)
			vl := *v
			err := k8s.AppsV1().Deployments(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{})
			require.NoError(t, err)
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
		case **backupApi.ArangoBackupPolicy:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.BackupV1().ArangoBackupPolicies(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **mlApi.ArangoMLExtension:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.MlV1beta1().ArangoMLExtensions(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **mlApi.ArangoMLStorage:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.MlV1beta1().ArangoMLStorages(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **mlApiv1alpha1.ArangoMLExtension:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.MlV1alpha1().ArangoMLExtensions(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **mlApiv1alpha1.ArangoMLStorage:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.MlV1alpha1().ArangoMLStorages(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **mlApiv1alpha1.ArangoMLBatchJob:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.MlV1alpha1().ArangoMLBatchJobs(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **mlApiv1alpha1.ArangoMLCronJob:
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
		case **schedulerApiv1alpha1.ArangoProfile:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.SchedulerV1alpha1().ArangoProfiles(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **schedulerApi.ArangoProfile:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.SchedulerV1beta1().ArangoProfiles(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **schedulerApi.ArangoSchedulerPod:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.SchedulerV1beta1().ArangoSchedulerPods(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **schedulerApi.ArangoSchedulerDeployment:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.SchedulerV1beta1().ArangoSchedulerDeployments(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **schedulerApi.ArangoSchedulerBatchJob:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.SchedulerV1beta1().ArangoSchedulerBatchJobs(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **schedulerApi.ArangoSchedulerCronJob:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.SchedulerV1beta1().ArangoSchedulerCronJobs(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **analyticsApi.GraphAnalyticsEngine:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.AnalyticsV1alpha1().GraphAnalyticsEngines(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **networkingApi.ArangoRoute:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.NetworkingV1beta1().ArangoRoutes(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **platformApi.ArangoPlatformStorage:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.PlatformV1beta1().ArangoPlatformStorages(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **platformApi.ArangoPlatformChart:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.PlatformV1beta1().ArangoPlatformCharts(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		case **platformApi.ArangoPlatformService:
			require.NotNil(t, v)

			vl := *v
			require.NoError(t, arango.PlatformV1beta1().ArangoPlatformServices(vl.GetNamespace()).Delete(context.Background(), vl.GetName(), meta.DeleteOptions{}))
		default:
			require.Fail(t, fmt.Sprintf("Unable to delete object: %s", reflect.TypeOf(v).String()))
		}
	}
}

func RefreshObjectsC(t *testing.T, client kclient.Client, objects ...interface{}) {
	RefreshObjects(t, client.Kubernetes(), client.Arango(), objects...)
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
		case **core.ConfigMap:
			require.NotNil(t, v)

			vl := *v

			vn, err := k8s.CoreV1().ConfigMaps(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
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
		case **core.Endpoints:
			require.NotNil(t, v)

			vl := *v

			vn, err := k8s.CoreV1().Endpoints(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
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
		case **apps.Deployment:
			require.NotNil(t, v)

			vl := *v
			vn, err := k8s.AppsV1().Deployments(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
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
		case **backupApi.ArangoBackupPolicy:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.BackupV1().ArangoBackupPolicies(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
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

			vn, err := arango.MlV1beta1().ArangoMLExtensions(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
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

			vn, err := arango.MlV1beta1().ArangoMLStorages(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **mlApiv1alpha1.ArangoMLExtension:
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
		case **mlApiv1alpha1.ArangoMLStorage:
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
		case **mlApiv1alpha1.ArangoMLBatchJob:
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
		case **mlApiv1alpha1.ArangoMLCronJob:
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
		case **schedulerApiv1alpha1.ArangoProfile:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.SchedulerV1alpha1().ArangoProfiles(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **schedulerApi.ArangoProfile:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.SchedulerV1beta1().ArangoProfiles(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **schedulerApi.ArangoSchedulerPod:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.SchedulerV1beta1().ArangoSchedulerPods(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **schedulerApi.ArangoSchedulerDeployment:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.SchedulerV1beta1().ArangoSchedulerDeployments(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **schedulerApi.ArangoSchedulerBatchJob:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.SchedulerV1beta1().ArangoSchedulerBatchJobs(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **schedulerApi.ArangoSchedulerCronJob:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.SchedulerV1beta1().ArangoSchedulerCronJobs(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **analyticsApi.GraphAnalyticsEngine:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.AnalyticsV1alpha1().GraphAnalyticsEngines(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **networkingApi.ArangoRoute:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.NetworkingV1beta1().ArangoRoutes(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **platformApi.ArangoPlatformStorage:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.PlatformV1beta1().ArangoPlatformStorages(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **platformApi.ArangoPlatformChart:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.PlatformV1beta1().ArangoPlatformCharts(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					*v = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				*v = vn
			}
		case **platformApi.ArangoPlatformService:
			require.NotNil(t, v)

			vl := *v

			vn, err := arango.PlatformV1beta1().ArangoPlatformServices(vl.GetNamespace()).Get(context.Background(), vl.GetName(), meta.GetOptions{})
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

func Apply[T meta.Object](t *testing.T, obj T, mods ...MetaObjectMod[T]) {
	for _, m := range mods {
		m(t, obj)
	}
}

func SetMetaBasedOnType(t *testing.T, object meta.Object) {
	switch v := object.(type) {
	case *batch.CronJob:
		v.Kind = " ArangoSchedulerCronJob"
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
		v.Kind = " ArangoSchedulerPod"
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
	case *core.ConfigMap:
		v.Kind = "ConfigMap"
		v.APIVersion = "v1"
		v.SetSelfLink(fmt.Sprintf("/api/v1/configmaps/%s/%s",
			object.GetNamespace(),
			object.GetName()))
	case *core.Service:
		v.Kind = "Service"
		v.APIVersion = "v1"
		v.SetSelfLink(fmt.Sprintf("/api/v1/services/%s/%s",
			object.GetNamespace(),
			object.GetName()))
	case *core.Endpoints:
		v.Kind = "Endpoints"
		v.APIVersion = "v1"
		v.SetSelfLink(fmt.Sprintf("/api/v1/endpoints/%s/%s",
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
	case *apps.Deployment:
		v.Kind = "Deployment"
		v.APIVersion = "v1"
		v.SetSelfLink(fmt.Sprintf("/api/apps/v1/deployments/%s/%s",
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
	case *backupApi.ArangoBackupPolicy:
		v.Kind = backup.ArangoBackupPolicyResourceKind
		v.APIVersion = backupApi.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			backupApi.SchemeGroupVersion.String(),
			backup.ArangoBackupPolicyResourcePlural,
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
	case *mlApiv1alpha1.ArangoMLExtension:
		v.Kind = ml.ArangoMLExtensionResourceKind
		v.APIVersion = mlApiv1alpha1.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			mlApiv1alpha1.SchemeGroupVersion.String(),
			ml.ArangoMLExtensionResourcePlural,
			object.GetNamespace(),
			object.GetName()))
	case *mlApiv1alpha1.ArangoMLStorage:
		v.Kind = ml.ArangoMLStorageResourceKind
		v.APIVersion = mlApiv1alpha1.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			mlApiv1alpha1.SchemeGroupVersion.String(),
			ml.ArangoMLStorageResourcePlural,
			object.GetNamespace(),
			object.GetName()))
	case *mlApiv1alpha1.ArangoMLBatchJob:
		v.Kind = ml.ArangoMLBatchJobResourceKind
		v.APIVersion = mlApiv1alpha1.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			mlApiv1alpha1.SchemeGroupVersion.String(),
			ml.ArangoMLBatchJobResourcePlural,
			object.GetNamespace(),
			object.GetName()))
	case *mlApiv1alpha1.ArangoMLCronJob:
		v.Kind = ml.ArangoMLCronJobResourceKind
		v.APIVersion = mlApiv1alpha1.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			mlApiv1alpha1.SchemeGroupVersion.String(),
			ml.ArangoMLCronJobResourcePlural,
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
	case *schedulerApiv1alpha1.ArangoProfile:
		v.Kind = scheduler.ArangoProfileResourceKind
		v.APIVersion = schedulerApiv1alpha1.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			schedulerApiv1alpha1.SchemeGroupVersion.String(),
			scheduler.ArangoProfileResourcePlural,
			object.GetNamespace(),
			object.GetName()))
	case *schedulerApi.ArangoProfile:
		v.Kind = scheduler.ArangoProfileResourceKind
		v.APIVersion = schedulerApi.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			schedulerApi.SchemeGroupVersion.String(),
			scheduler.ArangoProfileResourcePlural,
			object.GetNamespace(),
			object.GetName()))
	case *schedulerApi.ArangoSchedulerPod:
		v.Kind = scheduler.PodResourceKind
		v.APIVersion = schedulerApi.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			schedulerApi.SchemeGroupVersion.String(),
			scheduler.PodResourcePlural,
			object.GetNamespace(),
			object.GetName()))
	case *schedulerApi.ArangoSchedulerDeployment:
		v.Kind = scheduler.DeploymentResourceKind
		v.APIVersion = schedulerApi.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			schedulerApi.SchemeGroupVersion.String(),
			scheduler.DeploymentResourcePlural,
			object.GetNamespace(),
			object.GetName()))
	case *schedulerApi.ArangoSchedulerBatchJob:
		v.Kind = scheduler.BatchJobResourceKind
		v.APIVersion = schedulerApi.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			schedulerApi.SchemeGroupVersion.String(),
			scheduler.BatchJobResourcePlural,
			object.GetNamespace(),
			object.GetName()))
	case *schedulerApi.ArangoSchedulerCronJob:
		v.Kind = scheduler.CronJobResourceKind
		v.APIVersion = schedulerApi.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			schedulerApi.SchemeGroupVersion.String(),
			scheduler.CronJobResourcePlural,
			object.GetNamespace(),
			object.GetName()))
	case *analyticsApi.GraphAnalyticsEngine:
		v.Kind = analytics.GraphAnalyticsEngineResourceKind
		v.APIVersion = analyticsApi.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			analyticsApi.SchemeGroupVersion.String(),
			analytics.GraphAnalyticsEngineResourcePlural,
			object.GetNamespace(),
			object.GetName()))
	case *networkingApi.ArangoRoute:
		v.Kind = networking.ArangoRouteResourceKind
		v.APIVersion = networkingApi.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			networkingApi.SchemeGroupVersion.String(),
			networking.ArangoRouteResourcePlural,
			object.GetNamespace(),
			object.GetName()))
	case *platformApi.ArangoPlatformStorage:
		v.Kind = platform.ArangoPlatformStorageResourceKind
		v.APIVersion = platformApi.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			platformApi.SchemeGroupVersion.String(),
			platform.ArangoPlatformStorageResourcePlural,
			object.GetNamespace(),
			object.GetName()))
	case *platformApi.ArangoPlatformChart:
		v.Kind = platform.ArangoPlatformChartResourceKind
		v.APIVersion = platformApi.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			platformApi.SchemeGroupVersion.String(),
			platform.ArangoPlatformChartResourcePlural,
			object.GetNamespace(),
			object.GetName()))
	case *platformApi.ArangoPlatformService:
		v.Kind = platform.ArangoPlatformServiceResourceKind
		v.APIVersion = platformApi.SchemeGroupVersion.String()
		v.SetSelfLink(fmt.Sprintf("/api/%s/%s/%s/%s",
			platformApi.SchemeGroupVersion.String(),
			platform.ArangoPlatformServiceResourcePlural,
			object.GetNamespace(),
			object.GetName()))
	default:
		require.Fail(t, fmt.Sprintf("Unable to get object meta: %s", reflect.TypeOf(v).String()))
	}
}

func NewMetaObjectInDefaultNamespace[T meta.Object](t *testing.T, name string, mods ...MetaObjectMod[T]) T {
	return NewMetaObject[T](t, FakeNamespace, name, mods...)
}

func NewMetaObject[T meta.Object](t *testing.T, namespace, name string, mods ...MetaObjectMod[T]) T {
	obj, err := util.DeepType[T]()
	require.NoError(t, err)

	if IsNamespaced(obj) {
		obj.SetNamespace(namespace)
	}
	obj.SetName(name)
	obj.SetUID(uuid.NewUUID())
	obj.SetCreationTimestamp(meta.Now())

	SetMetaBasedOnType(t, obj)

	Apply(t, obj, mods...)

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

	gvk, ok := constants.ExtractGVKFromObject(object)
	require.True(t, ok, reflect.TypeOf(object).String())

	item.Group = gvk.Group
	item.Version = gvk.Version
	item.Kind = gvk.Kind

	return item
}
