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

package resources

import (
	"context"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/member"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconciler"
	"github.com/arangodb/kube-arangodb/pkg/operator/scope"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// Context provides all functions needed by the Resources service
// to perform its service.
type Context interface {
	reconciler.DeploymentStatusUpdate
	reconciler.DeploymentAgencyMaintenance
	reconciler.ArangoMemberContext
	reconciler.DeploymentImageManager
	reconciler.DeploymentModInterfaces
	reconciler.DeploymentCachedStatus
	reconciler.ArangoAgency
	reconciler.ArangoApplier
	reconciler.DeploymentInfoGetter
	reconciler.DeploymentClient
	reconciler.DeploymentSyncClient
	reconciler.KubernetesEventGenerator

	member.StateInspectorGetter

	// GetServerGroupIterator returns the deployment as ServerGroupIterator.
	GetServerGroupIterator() reconciler.ServerGroupIterator
	// UpdateStatus replaces the status of the deployment with the given status and
	// updates the resources in k8s.
	UpdateStatus(ctx context.Context, status api.DeploymentStatus, lastVersion int32, force ...bool) error
	// GetOperatorImage returns the image name of operator image
	GetOperatorImage() string
	// GetArangoImage returns the image name containing the default arango image
	GetArangoImage() string
	// CreateEvent creates a given event.
	// On error, the error is logged.
	CreateEvent(evt *k8sutil.Event)
	// GetOwnedPVCs returns a list of all PVCs owned by the deployment.
	GetOwnedPVCs() ([]core.PersistentVolumeClaim, error)
	// CleanupPod deletes a given pod with force and explicit UID.
	// If the pod does not exist, the error is ignored.
	CleanupPod(ctx context.Context, p *core.Pod) error
	// DeletePod deletes a pod with given name in the namespace
	// of the deployment. If the pod does not exist, the error is ignored.
	DeletePod(ctx context.Context, podName string, options meta.DeleteOptions) error
	// DeletePvc deletes a persistent volume claim with given name in the namespace
	// of the deployment. If the pvc does not exist, the error is ignored.
	DeletePvc(ctx context.Context, pvcName string) error
	// GetBackup receives information about a backup resource
	GetBackup(ctx context.Context, backup string) (*backupApi.ArangoBackup, error)
	GetScope() scope.Scope
}
