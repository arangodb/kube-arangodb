//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package reconciler

import (
	"context"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/arangosync-client/client"
	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/acs/sutil"
	agencyCache "github.com/arangodb/kube-arangodb/pkg/deployment/agency"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// ServerGroupIterator provides a helper to callback on every server
// group of the deployment.
type ServerGroupIterator interface {
	// ForeachServerGroupAccepted calls the given callback for all accepted server groups.
	// If the callback returns an error, this error is returned and no other server
	// groups are processed.
	// Groups are processed in this order: agents, single, dbservers, coordinators, syncmasters, syncworkers
	ForeachServerGroupAccepted(cb api.ServerGroupFunc, status *api.DeploymentStatus) error
}

type DeploymentStatusUpdateErrFunc func(s *api.DeploymentStatus) (bool, error)
type DeploymentStatusUpdateFunc func(s *api.DeploymentStatus) bool
type DeploymentMemberStatusUpdateErrFunc func(s *api.MemberStatus) (bool, error)
type DeploymentMemberStatusUpdateFunc func(s *api.MemberStatus) bool

type DeploymentStatusUpdate interface {
	// WithStatusUpdateErr update status of ArangoDeployment with defined modifier. If action returns True action is taken
	WithStatusUpdateErr(ctx context.Context, action DeploymentStatusUpdateErrFunc) error
	// WithStatusUpdate update status of ArangoDeployment with defined modifier. If action returns True action is taken
	WithStatusUpdate(ctx context.Context, action DeploymentStatusUpdateFunc) error

	// WithMemberStatusUpdateErr update status of ArangoDeployment Member with defined modifier. If action returns True action is taken
	WithMemberStatusUpdateErr(ctx context.Context, id string, group api.ServerGroup, action DeploymentMemberStatusUpdateErrFunc) error
	// WithMemberStatusUpdate update status of ArangoDeployment Member with defined modifier. If action returns True action is taken
	WithMemberStatusUpdate(ctx context.Context, id string, group api.ServerGroup, action DeploymentMemberStatusUpdateFunc) error

	// UpdateStatus replaces the status of the deployment with the given status and
	// updates the resources in k8s.
	UpdateStatus(ctx context.Context, status api.DeploymentStatus) error
	// UpdateMember updates the deployment status wrt the given member.
	UpdateMember(ctx context.Context, member api.MemberStatus) error
}

type DeploymentAgencyMaintenance interface {
	// SetAgencyMaintenanceMode set maintenance mode info
	SetAgencyMaintenanceMode(ctx context.Context, enabled bool) error
}

type DeploymentPodRenderer interface {
	// RenderPodForMember Renders Pod definition for member
	RenderPodForMember(ctx context.Context, acs sutil.ACS, spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*core.Pod, error)
	// RenderPodTemplateForMember Renders PodTemplate definition for member
	RenderPodTemplateForMember(ctx context.Context, acs sutil.ACS, spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*core.PodTemplateSpec, error)

	DeploymentEndpoints
}

type DeploymentEndpoints interface {
	// GenerateMemberEndpoint generates endpoint for a member
	GenerateMemberEndpoint(group api.ServerGroup, member api.MemberStatus) (string, error)
}

type DeploymentImageManager interface {
	// SelectImage select currently used image by pod
	SelectImage(spec api.DeploymentSpec, status api.DeploymentStatus) (api.ImageInfo, bool)
	// SelectImageForMember select currently used image by pod in member
	SelectImageForMember(spec api.DeploymentSpec, status api.DeploymentStatus, member api.MemberStatus) (api.ImageInfo, bool)
}

type ArangoAgencyGet interface {
	GetAgencyCache() (state.State, bool)
	GetAgencyArangoDBCache() (state.DB, bool)
	GetAgencyHealth() (agencyCache.Health, bool)
	ShardsInSyncMap() (state.ShardsSyncStatus, bool)
}

type ArangoAgency interface {
	ArangoAgencyGet

	RefreshAgencyCache(ctx context.Context) (uint64, error)
}

type DeploymentInfoGetter interface {
	// GetAPIObject returns the deployment as k8s object.
	GetAPIObject() k8sutil.APIObject
	// GetSpec returns the current specification of the deployment
	GetSpec() api.DeploymentSpec
	// GetStatus returns the current status of the deployment
	GetStatus() api.DeploymentStatus
	// GetMode the specified mode of deployment
	GetMode() api.DeploymentMode
	// GetName returns the name of the deployment
	GetName() string
	// GetNamespace returns the namespace that contains the deployment
	GetNamespace() string
	// IsSyncEnabled returns information if sync is enabled
	IsSyncEnabled() bool
}

type ArangoApplier interface {
	ApplyPatchOnPod(ctx context.Context, pod *core.Pod, p ...patch.Item) error
	ApplyPatch(ctx context.Context, p ...patch.Item) error
}

type DeploymentDatabaseClient interface {
	// GetDatabaseAsyncClient returns a cached client for the entire database (cluster coordinators or single server),
	// creating one if needed. Only in AsyncMode
	GetDatabaseAsyncClient(ctx context.Context) (driver.Client, error)
	// GetServerAsyncClient returns an async client for a specific server.
	GetServerAsyncClient(id string) (driver.Client, error)
}

type DeploymentMemberClient interface {
	// GetServerClient returns a cached client for a specific server.
	GetServerClient(ctx context.Context, group api.ServerGroup, id string) (driver.Client, error)
}

type DeploymentSyncClient interface {
	// GetSyncServerClient returns a cached client for a specific arangosync server.
	GetSyncServerClient(ctx context.Context, group api.ServerGroup, id string) (client.API, error)
}

type KubernetesEventGenerator interface {
	// CreateEvent creates a given event.
	// On error, the error is logged.
	CreateEvent(evt *k8sutil.Event)

	CreateOperatorEngineOpsAlertEvent(message string, args ...interface{})
}

// DeploymentClient provides functionalities to get deployment's clients.
type DeploymentClient interface {
	DeploymentDatabaseClient
	DeploymentMemberClient
	DeploymentSyncClient
}

// DeploymentGetter provides functionalities to get deployment resources.
type DeploymentGetter interface {
	DeploymentClient
	DeploymentInfoGetter
}
