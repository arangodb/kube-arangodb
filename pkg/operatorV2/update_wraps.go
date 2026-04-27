//
// DISCLAIMER
//
// Copyright 2023-2026 ArangoDB GmbH, Cologne, Germany
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
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1beta1"
	permissionApi "github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
)

func WithArangoBackupUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[backupApi.ArangoBackupStatus, *backupApi.ArangoBackup], obj *backupApi.ArangoBackup, status backupApi.ArangoBackupStatus, opts meta.UpdateOptions) (*backupApi.ArangoBackup, error) {
	return WithUpdateStatusInterfaceRetry[backupApi.ArangoBackupStatus, *backupApi.ArangoBackup](ctx, client, obj, status, opts)
}

func WithAnalyticsGAEUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[analyticsApi.GraphAnalyticsEngineStatus, *analyticsApi.GraphAnalyticsEngine], obj *analyticsApi.GraphAnalyticsEngine, status analyticsApi.GraphAnalyticsEngineStatus, opts meta.UpdateOptions) (*analyticsApi.GraphAnalyticsEngine, error) {
	return WithUpdateStatusInterfaceRetry[analyticsApi.GraphAnalyticsEngineStatus, *analyticsApi.GraphAnalyticsEngine](ctx, client, obj, status, opts)
}

func WithNetworkingArangoRouteUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[networkingApi.ArangoRouteStatus, *networkingApi.ArangoRoute], obj *networkingApi.ArangoRoute, status networkingApi.ArangoRouteStatus, opts meta.UpdateOptions) (*networkingApi.ArangoRoute, error) {
	return WithUpdateStatusInterfaceRetry[networkingApi.ArangoRouteStatus, *networkingApi.ArangoRoute](ctx, client, obj, status, opts)
}

func WithSchedulerArangoProfileUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[schedulerApi.ProfileStatus, *schedulerApi.ArangoProfile], obj *schedulerApi.ArangoProfile, status schedulerApi.ProfileStatus, opts meta.UpdateOptions) (*schedulerApi.ArangoProfile, error) {
	return WithUpdateStatusInterfaceRetry[schedulerApi.ProfileStatus, *schedulerApi.ArangoProfile](ctx, client, obj, status, opts)
}

func WithSchedulerPodUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[schedulerApi.ArangoSchedulerPodStatus, *schedulerApi.ArangoSchedulerPod], obj *schedulerApi.ArangoSchedulerPod, status schedulerApi.ArangoSchedulerPodStatus, opts meta.UpdateOptions) (*schedulerApi.ArangoSchedulerPod, error) {
	return WithUpdateStatusInterfaceRetry[schedulerApi.ArangoSchedulerPodStatus, *schedulerApi.ArangoSchedulerPod](ctx, client, obj, status, opts)
}

func WithSchedulerDeploymentUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[schedulerApi.ArangoSchedulerDeploymentStatus, *schedulerApi.ArangoSchedulerDeployment], obj *schedulerApi.ArangoSchedulerDeployment, status schedulerApi.ArangoSchedulerDeploymentStatus, opts meta.UpdateOptions) (*schedulerApi.ArangoSchedulerDeployment, error) {
	return WithUpdateStatusInterfaceRetry[schedulerApi.ArangoSchedulerDeploymentStatus, *schedulerApi.ArangoSchedulerDeployment](ctx, client, obj, status, opts)
}

func WithSchedulerBatchJobUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[schedulerApi.ArangoSchedulerBatchJobStatus, *schedulerApi.ArangoSchedulerBatchJob], obj *schedulerApi.ArangoSchedulerBatchJob, status schedulerApi.ArangoSchedulerBatchJobStatus, opts meta.UpdateOptions) (*schedulerApi.ArangoSchedulerBatchJob, error) {
	return WithUpdateStatusInterfaceRetry[schedulerApi.ArangoSchedulerBatchJobStatus, *schedulerApi.ArangoSchedulerBatchJob](ctx, client, obj, status, opts)
}

func WithSchedulerCronJobUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[schedulerApi.ArangoSchedulerCronJobStatus, *schedulerApi.ArangoSchedulerCronJob], obj *schedulerApi.ArangoSchedulerCronJob, status schedulerApi.ArangoSchedulerCronJobStatus, opts meta.UpdateOptions) (*schedulerApi.ArangoSchedulerCronJob, error) {
	return WithUpdateStatusInterfaceRetry[schedulerApi.ArangoSchedulerCronJobStatus, *schedulerApi.ArangoSchedulerCronJob](ctx, client, obj, status, opts)
}

func WithArangoPlatformStorageUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[platformApi.ArangoPlatformStorageStatus, *platformApi.ArangoPlatformStorage], obj *platformApi.ArangoPlatformStorage, status platformApi.ArangoPlatformStorageStatus, opts meta.UpdateOptions) (*platformApi.ArangoPlatformStorage, error) {
	return WithUpdateStatusInterfaceRetry[platformApi.ArangoPlatformStorageStatus, *platformApi.ArangoPlatformStorage](ctx, client, obj, status, opts)
}

func WithArangoPlatformChartUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[platformApi.ArangoPlatformChartStatus, *platformApi.ArangoPlatformChart], obj *platformApi.ArangoPlatformChart, status platformApi.ArangoPlatformChartStatus, opts meta.UpdateOptions) (*platformApi.ArangoPlatformChart, error) {
	return WithUpdateStatusInterfaceRetry[platformApi.ArangoPlatformChartStatus, *platformApi.ArangoPlatformChart](ctx, client, obj, status, opts)
}

func WithArangoPlatformServiceUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[platformApi.ArangoPlatformServiceStatus, *platformApi.ArangoPlatformService], obj *platformApi.ArangoPlatformService, status platformApi.ArangoPlatformServiceStatus, opts meta.UpdateOptions) (*platformApi.ArangoPlatformService, error) {
	return WithUpdateStatusInterfaceRetry[platformApi.ArangoPlatformServiceStatus, *platformApi.ArangoPlatformService](ctx, client, obj, status, opts)
}

func WithArangoPermissionTokenUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[permissionApi.ArangoPermissionTokenStatus, *permissionApi.ArangoPermissionToken], obj *permissionApi.ArangoPermissionToken, status permissionApi.ArangoPermissionTokenStatus, opts meta.UpdateOptions) (*permissionApi.ArangoPermissionToken, error) {
	return WithUpdateStatusInterfaceRetry[permissionApi.ArangoPermissionTokenStatus, *permissionApi.ArangoPermissionToken](ctx, client, obj, status, opts)
}

func WithArangoPermissionRoleUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[permissionApi.ArangoPermissionRoleStatus, *permissionApi.ArangoPermissionRole], obj *permissionApi.ArangoPermissionRole, status permissionApi.ArangoPermissionRoleStatus, opts meta.UpdateOptions) (*permissionApi.ArangoPermissionRole, error) {
	return WithUpdateStatusInterfaceRetry[permissionApi.ArangoPermissionRoleStatus, *permissionApi.ArangoPermissionRole](ctx, client, obj, status, opts)
}

func WithArangoPermissionPolicyUpdateStatusInterfaceRetry(ctx context.Context, client UpdateStatusInterface[permissionApi.ArangoPermissionPolicyStatus, *permissionApi.ArangoPermissionPolicy], obj *permissionApi.ArangoPermissionPolicy, status permissionApi.ArangoPermissionPolicyStatus, opts meta.UpdateOptions) (*permissionApi.ArangoPermissionPolicy, error) {
	return WithUpdateStatusInterfaceRetry[permissionApi.ArangoPermissionPolicyStatus, *permissionApi.ArangoPermissionPolicy](ctx, client, obj, status, opts)
}
