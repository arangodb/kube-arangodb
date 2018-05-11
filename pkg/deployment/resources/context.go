//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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

package resources

import (
	"context"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// ServerGroupIterator provides a helper to callback on every server
// group of the deployment.
type ServerGroupIterator interface {
	// ForeachServerGroup calls the given callback for all server groups.
	// If the callback returns an error, this error is returned and no other server
	// groups are processed.
	// Groups are processed in this order: agents, single, dbservers, coordinators, syncmasters, syncworkers
	ForeachServerGroup(cb func(group api.ServerGroup, spec api.ServerGroupSpec, status *api.MemberStatusList) error, status *api.DeploymentStatus) error
}

// Context provides all functions needed by the Resources service
// to perform its service.
type Context interface {
	// GetAPIObject returns the deployment as k8s object.
	GetAPIObject() k8sutil.APIObject
	// GetServerGroupIterator returns the deployment as ServerGroupIterator.
	GetServerGroupIterator() ServerGroupIterator
	// GetSpec returns the current specification of the deployment
	GetSpec() api.DeploymentSpec
	// GetStatus returns the current status of the deployment
	GetStatus() api.DeploymentStatus
	// UpdateStatus replaces the status of the deployment with the given status and
	// updates the resources in k8s.
	UpdateStatus(status api.DeploymentStatus, force ...bool) error
	// GetKubeCli returns the kubernetes client
	GetKubeCli() kubernetes.Interface
	// GetLifecycleImage returns the image name containing the lifecycle helper (== name of operator image)
	GetLifecycleImage() string
	// GetNamespace returns the namespace that contains the deployment
	GetNamespace() string
	// createEvent creates a given event.
	// On error, the error is logged.
	CreateEvent(evt *v1.Event)
	// GetOwnedPods returns a list of all pods owned by the deployment.
	GetOwnedPods() ([]v1.Pod, error)
	// CleanupPod deletes a given pod with force and explicit UID.
	// If the pod does not exist, the error is ignored.
	CleanupPod(p v1.Pod) error
	// GetDatabaseClient returns a cached client for the entire database (cluster coordinators or single server),
	// creating one if needed.
	GetDatabaseClient(ctx context.Context) (driver.Client, error)
}
