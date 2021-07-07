//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
//

package reconcile

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"

	"github.com/arangodb/kube-arangodb/pkg/util/arangod/conn"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"

	"github.com/arangodb/go-driver/agency"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	core "k8s.io/api/core/v1"
)

// PlanBuilderContext contains context methods provided to plan builders.
type PlanBuilderContext interface {
	resources.DeploymentStatusUpdate
	resources.DeploymentAgencyMaintenance
	resources.ArangoMemberContext

	// GetTLSKeyfile returns the keyfile encoded TLS certificate+key for
	// the given member.
	GetTLSKeyfile(group api.ServerGroup, member api.MemberStatus) (string, error)
	// CreateEvent creates a given event.
	// On error, the error is logged.
	CreateEvent(evt *k8sutil.Event)
	// GetPvc gets a PVC by the given name, in the samespace of the deployment.
	GetPvc(ctx context.Context, pvcName string) (*core.PersistentVolumeClaim, error)
	// GetShardSyncStatus returns true if all shards are in sync
	GetShardSyncStatus() bool
	// InvalidateSyncStatus resets the sync state to false and triggers an inspection
	InvalidateSyncStatus()
	// GetStatus returns the current status of the deployment
	GetStatus() (api.DeploymentStatus, int32)
	// GetStatus returns the current spec of the deployment
	GetSpec() api.DeploymentSpec
	// GetAgencyData object for key path
	GetAgencyData(ctx context.Context, i interface{}, keyParts ...string) error
	// RenderPodForMember Renders Pod definition for member
	RenderPodForMember(ctx context.Context, cachedStatus inspectorInterface.Inspector, spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*core.Pod, error)
	// SelectImage select currently used image by pod
	SelectImage(spec api.DeploymentSpec, status api.DeploymentStatus) (api.ImageInfo, bool)
	// GetDatabaseClient returns a cached client for the entire database (cluster coordinators or single server),
	// creating one if needed.
	GetDatabaseClient(ctx context.Context) (driver.Client, error)
	// GetServerClient returns a cached client for a specific server.
	GetServerClient(ctx context.Context, group api.ServerGroup, id string) (driver.Client, error)
	// GetAuthentication return authentication for members
	GetAuthentication() conn.Auth
	// SecretsInterface return secret interface
	SecretsInterface() k8sutil.SecretInterface
	// GetBackup receives information about a backup resource
	GetBackup(ctx context.Context, backup string) (*backupApi.ArangoBackup, error)
	// GetName receives deployment name
	GetName() string
	// GetAgency returns a connection to the entire agency.
	GetAgency(ctx context.Context) (agency.Agency, error)
}

// newPlanBuilderContext creates a PlanBuilderContext from the given context
func newPlanBuilderContext(ctx Context) PlanBuilderContext {
	return ctx
}
