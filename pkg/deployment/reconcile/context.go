//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
// Author Tomasz Mielech
//

package reconcile

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"

	"github.com/arangodb/arangosync-client/client"
	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/agency"
	v1 "k8s.io/api/core/v1"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/conn"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// Context provides methods to the reconcile package.
type Context interface {
	resources.DeploymentStatusUpdate
	resources.DeploymentAgencyMaintenance
	resources.ArangoMemberContext
	resources.DeploymentPodRenderer

	// GetAPIObject returns the deployment as k8s object.
	GetAPIObject() k8sutil.APIObject
	// GetSpec returns the current specification of the deployment
	GetSpec() api.DeploymentSpec
	// GetStatus returns the current status of the deployment
	GetStatus() (api.DeploymentStatus, int32)
	// UpdateStatus replaces the status of the deployment with the given status and
	// updates the resources in k8s.
	UpdateStatus(ctx context.Context, status api.DeploymentStatus, lastVersion int32, force ...bool) error
	// UpdateMember updates the deployment status wrt the given member.
	UpdateMember(ctx context.Context, member api.MemberStatus) error
	// GetDatabaseClient returns a cached client for the entire database (cluster coordinators or single server),
	// creating one if needed.
	GetDatabaseClient(ctx context.Context) (driver.Client, error)
	// GetServerClient returns a cached client for a specific server.
	GetServerClient(ctx context.Context, group api.ServerGroup, id string) (driver.Client, error)
	// GetAgencyClients returns a client connection for every agency member.
	// If the given predicate is not nil, only agents are included where the given predicate returns true.
	GetAgencyClients(ctx context.Context, predicate func(id string) bool) ([]driver.Connection, error)
	// GetAgency returns a connection to the entire agency.
	GetAgency(ctx context.Context) (agency.Agency, error)
	// GetSyncServerClient returns a cached client for a specific arangosync server.
	GetSyncServerClient(ctx context.Context, group api.ServerGroup, id string) (client.API, error)
	// CreateEvent creates a given event.
	// On error, the error is logged.
	CreateEvent(evt *k8sutil.Event)
	// CreateMember adds a new member to the given group.
	// If ID is non-empty, it will be used, otherwise a new ID is created.
	// Returns ID, error
	CreateMember(ctx context.Context, group api.ServerGroup, id string) (string, error)
	// GetPod returns pod.
	GetPod(ctx context.Context, podName string) (*v1.Pod, error)
	// DeletePod deletes a pod with given name in the namespace
	// of the deployment. If the pod does not exist, the error is ignored.
	DeletePod(ctx context.Context, podName string) error
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
	// GetAgencyData object for key path
	GetAgencyData(ctx context.Context, i interface{}, keyParts ...string) error
	// SelectImage select currently used image by pod
	SelectImage(spec api.DeploymentSpec, status api.DeploymentStatus) (api.ImageInfo, bool)
	// SecretsInterface return secret interface
	SecretsInterface() k8sutil.SecretInterface
	// GetBackup receives information about a backup resource
	GetBackup(ctx context.Context, backup string) (*backupApi.ArangoBackup, error)
	// GetName receives deployment name
	GetName() string
	// GetAuthentication return authentication for members
	GetAuthentication() conn.Auth
}
