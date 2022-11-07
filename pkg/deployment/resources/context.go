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

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/acs/sutil"
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
	reconciler.DeploymentImageManager
	reconciler.ArangoAgency
	reconciler.ArangoApplier
	reconciler.DeploymentGetter
	reconciler.KubernetesEventGenerator

	member.StateInspectorGetter

	sutil.ACSGetter

	// GetServerGroupIterator returns the deployment as ServerGroupIterator.
	GetServerGroupIterator() reconciler.ServerGroupIterator
	// GetOperatorImage returns the image name of operator image
	GetOperatorImage() string
	// GetArangoImage returns the image name containing the default arango image
	GetArangoImage() string
	// CreateEvent creates a given event.
	// On error, the error is logged.
	CreateEvent(evt *k8sutil.Event)
	// GetOwnedPVCs returns a list of all PVCs owned by the deployment.
	GetOwnedPVCs() ([]core.PersistentVolumeClaim, error)
	// GetBackup receives information about a backup resource
	GetBackup(ctx context.Context, backup string) (*backupApi.ArangoBackup, error)
	GetScope() scope.Scope
}
