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

package reconcile

import (
	"context"

	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/go-driver"
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconciler"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/conn"
)

type CreateMemberMod func(s *api.DeploymentStatus, g api.ServerGroup, m *api.MemberStatus) error

// Context provides methods to the reconcile package.
type Context interface {
	reconciler.DeploymentStatusUpdate
	reconciler.DeploymentAgencyMaintenance
	reconciler.ArangoMemberContext
	reconciler.DeploymentPodRenderer
	reconciler.DeploymentImageManager
	reconciler.DeploymentModInterfaces
	reconciler.DeploymentCachedStatus
	reconciler.ArangoAgencyGet
	reconciler.ArangoApplier
	reconciler.DeploymentInfoGetter
	reconciler.DeploymentClient
	reconciler.KubernetesEventGenerator
	reconciler.DeploymentSyncClient

	// CreateMember adds a new member to the given group.
	// If ID is non-empty, it will be used, otherwise a new ID is created.
	// Returns ID, error
	CreateMember(ctx context.Context, group api.ServerGroup, id string, mods ...CreateMemberMod) (string, error)
	// GetPod returns pod.
	GetPod(ctx context.Context, podName string) (*v1.Pod, error)
	// DeletePod deletes a pod with given name in the namespace
	// of the deployment. If the pod does not exist, the error is ignored.
	DeletePod(ctx context.Context, podName string, options meta.DeleteOptions) error
	// DeletePvc deletes a persistent volume claim with given name in the namespace
	// of the deployment. If the pvc does not exist, the error is ignored.
	DeletePvc(ctx context.Context, pvcName string) error
	// RemovePodFinalizers removes all the finalizers from the Pod with given name in the namespace
	// of the deployment. If the pod does not exist, the error is ignored.
	RemovePodFinalizers(ctx context.Context, podName string) error
	// UpdatePvc update PVC with given name in the namespace
	// of the deployment.
	UpdatePvc(ctx context.Context, pvc *v1.PersistentVolumeClaim) error
	// GetPvc gets a PVC by the given name, in the samespace of the deployment.
	GetPvc(ctx context.Context, pvcName string) (*v1.PersistentVolumeClaim, error)
	// GetTLSKeyfile returns the keyfile encoded TLS certificate+key for
	// the given member.
	GetTLSKeyfile(group api.ServerGroup, member api.MemberStatus) (string, error)
	// DeleteTLSKeyfile removes the Secret containing the TLS keyfile for the given member.
	// If the secret does not exist, the error is ignored.
	DeleteTLSKeyfile(ctx context.Context, group api.ServerGroup, member api.MemberStatus) error
	// DeleteSecret removes the Secret with given name.
	// If the secret does not exist, the error is ignored.
	DeleteSecret(secretName string) error
	// GetDeploymentHealth returns a copy of the latest known state of cluster health
	GetDeploymentHealth() (driver.ClusterHealth, error)
	// GetShardSyncStatus returns true if all shards are in sync
	GetShardSyncStatus() bool
	// InvalidateSyncStatus resets the sync state to false and triggers an inspection
	InvalidateSyncStatus()
	// DisableScalingCluster disables scaling DBservers and coordinators
	DisableScalingCluster(ctx context.Context) error
	// EnableScalingCluster enables scaling DBservers and coordinators
	EnableScalingCluster(ctx context.Context) error
	// GetBackup receives information about a backup resource
	GetBackup(ctx context.Context, backup string) (*backupApi.ArangoBackup, error)
	// GetAuthentication return authentication for members
	GetAuthentication() conn.Auth
}
