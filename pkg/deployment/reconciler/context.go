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

package reconciler

import (
	"context"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/arangosync-client/client"
	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/agency"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/acs/sutil"
	agencyCache "github.com/arangodb/kube-arangodb/pkg/deployment/agency"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	persistentvolumeclaimv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolumeclaim/v1"
	podv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/pod/v1"
	poddisruptionbudgetv1beta1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/poddisruptionbudget/v1beta1"
	secretv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret/v1"
	servicev1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service/v1"
	serviceaccountv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/serviceaccount/v1"
	servicemonitorv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/servicemonitor/v1"
)

// ServerGroupIterator provides a helper to callback on every server
// group of the deployment.
type ServerGroupIterator interface {
	// ForeachServerGroup calls the given callback for all server groups.
	// If the callback returns an error, this error is returned and no other server
	// groups are processed.
	// Groups are processed in this order: agents, single, dbservers, coordinators, syncmasters, syncworkers
	ForeachServerGroup(cb api.ServerGroupFunc, status *api.DeploymentStatus) error
}

type DeploymentStatusUpdateErrFunc func(s *api.DeploymentStatus) (bool, error)
type DeploymentStatusUpdateFunc func(s *api.DeploymentStatus) bool

type DeploymentStatusUpdate interface {
	// WithStatusUpdateErr update status of ArangoDeployment with defined modifier. If action returns True action is taken
	WithStatusUpdateErr(ctx context.Context, action DeploymentStatusUpdateErrFunc, force ...bool) error
	// WithStatusUpdate update status of ArangoDeployment with defined modifier. If action returns True action is taken
	WithStatusUpdate(ctx context.Context, action DeploymentStatusUpdateFunc, force ...bool) error

	// UpdateStatus replaces the status of the deployment with the given status and
	// updates the resources in k8s.
	UpdateStatus(ctx context.Context, status api.DeploymentStatus, lastVersion int32, force ...bool) error
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
	// RenderPodForMemberFromCurrent Renders PodTemplate definition for member from current state
	RenderPodForMemberFromCurrent(ctx context.Context, acs sutil.ACS, memberID string) (*core.Pod, error)
	// RenderPodTemplateForMemberFromCurrent Renders PodTemplate definition for member
	RenderPodTemplateForMemberFromCurrent(ctx context.Context, acs sutil.ACS, memberID string) (*core.PodTemplateSpec, error)

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

type DeploymentModInterfaces interface {
	// SecretsModInterface define secret modification interface
	SecretsModInterface() secretv1.ModInterface
	// PodsModInterface define pod modification interface
	PodsModInterface() podv1.ModInterface
	// ServiceAccountsModInterface define serviceaccounts modification interface
	ServiceAccountsModInterface() serviceaccountv1.ModInterface
	// ServicesModInterface define services modification interface
	ServicesModInterface() servicev1.ModInterface
	// PersistentVolumeClaimsModInterface define persistentvolumeclaims modification interface
	PersistentVolumeClaimsModInterface() persistentvolumeclaimv1.ModInterface
	// PodDisruptionBudgetsModInterface define poddisruptionbudgets modification interface
	PodDisruptionBudgetsModInterface() poddisruptionbudgetv1beta1.ModInterface

	// ServiceMonitorsModInterface define servicemonitor modification interface
	ServiceMonitorsModInterface() servicemonitorv1.ModInterface
}

type DeploymentCachedStatus interface {
	// GetCachedStatus current cached state of deployment
	GetCachedStatus() inspectorInterface.Inspector
}

type ArangoAgencyGet interface {
	GetAgencyCache() (agencyCache.State, bool)
	GetAgencyLeaderID() string
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
	GetStatus() (api.DeploymentStatus, int32)
	// GetStatusSnapshot returns the current status of the deployment without revision
	GetStatusSnapshot() api.DeploymentStatus
	// GetMode the specified mode of deployment
	GetMode() api.DeploymentMode
	// GetName returns the name of the deployment
	GetName() string
	// GetNamespace returns the namespace that contains the deployment
	GetNamespace() string
}

type ArangoApplier interface {
	ApplyPatchOnPod(ctx context.Context, pod *core.Pod, p ...patch.Item) error
	ApplyPatch(ctx context.Context, p ...patch.Item) error
}

type DeploymentAgencyClient interface {
	// GetAgencyClients returns a client connection for every agency member.
	GetAgencyClients(ctx context.Context) ([]driver.Connection, error)
	// GetAgencyClientsWithPredicate returns a client connection for every agency member which match condition.
	GetAgencyClientsWithPredicate(ctx context.Context, predicate func(id string) bool) ([]driver.Connection, error)
	// GetAgency returns a connection to the entire agency.
	GetAgency(ctx context.Context, agencyIDs ...string) (agency.Agency, error)
}

type DeploymentDatabaseClient interface {
	// GetDatabaseClient returns a cached client for the entire database (cluster coordinators or single server),
	// creating one if needed.
	GetDatabaseClient(ctx context.Context) (driver.Client, error)

	// GetDatabaseAsyncClient returns a cached client for the entire database (cluster coordinators or single server),
	// creating one if needed. Only in AsyncMode
	GetDatabaseAsyncClient(ctx context.Context) (driver.Client, error)
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
}

type DeploymentClient interface {
	DeploymentAgencyClient
	DeploymentDatabaseClient
	DeploymentMemberClient
	DeploymentSyncClient
}
