//
// DISCLAIMER
//
// Copyright 2023-2024 ArangoDB GmbH, Cologne, Germany
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

package operator

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	analyticsApi "github.com/arangodb/kube-arangodb/pkg/apis/analytics/v1alpha1"
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	mlApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1alpha1"
	mlApi "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1beta1"
)

func WithArangoBackupUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[backupApi.ArangoBackupStatus, *backupApi.ArangoBackup], obj *backupApi.ArangoBackup, status backupApi.ArangoBackupStatus, opts meta.UpdateOptions) (*backupApi.ArangoBackup, error) {
	return WithUpdateStatusInterfaceRetry[backupApi.ArangoBackupStatus, *backupApi.ArangoBackup](ctx, client, obj, status, opts)
}

func WithArangoExtensionUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[mlApi.ArangoMLExtensionStatus, *mlApi.ArangoMLExtension], obj *mlApi.ArangoMLExtension, status mlApi.ArangoMLExtensionStatus, opts meta.UpdateOptions) (*mlApi.ArangoMLExtension, error) {
	return WithUpdateStatusInterfaceRetry[mlApi.ArangoMLExtensionStatus, *mlApi.ArangoMLExtension](ctx, client, obj, status, opts)
}

func WithArangoCronJobUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[mlApiv1alpha1.ArangoMLCronJobStatus, *mlApiv1alpha1.ArangoMLCronJob], obj *mlApiv1alpha1.ArangoMLCronJob, status mlApiv1alpha1.ArangoMLCronJobStatus, opts meta.UpdateOptions) (*mlApiv1alpha1.ArangoMLCronJob, error) {
	return WithUpdateStatusInterfaceRetry[mlApiv1alpha1.ArangoMLCronJobStatus, *mlApiv1alpha1.ArangoMLCronJob](ctx, client, obj, status, opts)
}

func WithArangoBatchJobUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[mlApiv1alpha1.ArangoMLBatchJobStatus, *mlApiv1alpha1.ArangoMLBatchJob], obj *mlApiv1alpha1.ArangoMLBatchJob, status mlApiv1alpha1.ArangoMLBatchJobStatus, opts meta.UpdateOptions) (*mlApiv1alpha1.ArangoMLBatchJob, error) {
	return WithUpdateStatusInterfaceRetry[mlApiv1alpha1.ArangoMLBatchJobStatus, *mlApiv1alpha1.ArangoMLBatchJob](ctx, client, obj, status, opts)
}

func WithAnalyticsGAEUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[analyticsApi.GraphAnalyticsEngineStatus, *analyticsApi.GraphAnalyticsEngine], obj *analyticsApi.GraphAnalyticsEngine, status analyticsApi.GraphAnalyticsEngineStatus, opts meta.UpdateOptions) (*analyticsApi.GraphAnalyticsEngine, error) {
	return WithUpdateStatusInterfaceRetry[analyticsApi.GraphAnalyticsEngineStatus, *analyticsApi.GraphAnalyticsEngine](ctx, client, obj, status, opts)
}

func WithArangoStorageUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[mlApi.ArangoMLStorageStatus, *mlApi.ArangoMLStorage], obj *mlApi.ArangoMLStorage, status mlApi.ArangoMLStorageStatus, opts meta.UpdateOptions) (*mlApi.ArangoMLStorage, error) {
	return WithUpdateStatusInterfaceRetry[mlApi.ArangoMLStorageStatus, *mlApi.ArangoMLStorage](ctx, client, obj, status, opts)
}
