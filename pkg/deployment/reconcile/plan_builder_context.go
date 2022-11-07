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

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/acs/sutil"
	"github.com/arangodb/kube-arangodb/pkg/deployment/member"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconciler"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/conn"
)

// PlanBuilderContext contains context methods provided to plan builders.
type PlanBuilderContext interface {
	reconciler.DeploymentStatusUpdate

	reconciler.DeploymentInfoGetter
	reconciler.DeploymentAgencyMaintenance
	reconciler.DeploymentPodRenderer
	reconciler.DeploymentImageManager
	reconciler.ArangoAgencyGet
	reconciler.DeploymentDatabaseClient
	reconciler.KubernetesEventGenerator

	member.StateInspectorGetter

	sutil.ACSGetter
	// GetAuthentication return authentication for members
	GetAuthentication() conn.Auth
	// GetBackup receives information about a backup resource
	GetBackup(ctx context.Context, backup string) (*backupApi.ArangoBackup, error)
}

// newPlanBuilderContext creates a PlanBuilderContext from the given context
func newPlanBuilderContext(ctx Context) PlanBuilderContext {
	return ctx
}
