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

package reconcile

import (
	"context"
	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	agencyData "github.com/arangodb/kube-arangodb/pkg/deployment/agency"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PlanBuilderContext contains context methods provided to plan builders.
type PlanBuilderContext interface {
	// GetTLSKeyfile returns the keyfile encoded TLS certificate+key for
	// the given member.
	GetTLSKeyfile(group api.ServerGroup, member api.MemberStatus) (string, error)
	// GetTLSCA returns the TLS CA certificate in the secret with given name.
	// Returns: publicKey, privateKey, ownerByDeployment, error
	GetTLSCA(secretName string) (string, string, bool, error)
	// CreateEvent creates a given event.
	// On error, the error is logged.
	CreateEvent(evt *k8sutil.Event)
	// GetPvc gets a PVC by the given name, in the samespace of the deployment.
	GetPvc(pvcName string) (*v1.PersistentVolumeClaim, error)
	// GetExpectedPodArguments creates command line arguments for a server in the given group with given ID.
	GetExpectedPodArguments(apiObject metav1.Object, deplSpec api.DeploymentSpec, group api.ServerGroup,
		agents api.MemberStatusList, id string, version driver.Version) []string
	// GetShardSyncStatus returns true if all shards are in sync
	GetShardSyncStatus() bool
	// InvalidateSyncStatus resets the sync state to false and triggers an inspection
	InvalidateSyncStatus()
	// GetStatus returns the current status of the deployment
	GetStatus() (api.DeploymentStatus, int32)
	// GetAgencyData returns fetched keys for agency data
	GetAgencyData(ctx context.Context, keys ... string) (*agencyData.Agency, error)
}

// newPlanBuilderContext creates a PlanBuilderContext from the given context
func newPlanBuilderContext(ctx Context) PlanBuilderContext {
	return ctx
}
