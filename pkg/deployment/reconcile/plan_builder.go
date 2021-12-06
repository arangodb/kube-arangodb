//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package reconcile

import (
	"context"

	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
)

const (
	// Component name for reconciliation of this package
	reconciliationComponent = "deployment_reconciliation"
)

// CreatePlan considers the current specification & status of the deployment creates a plan to
// get the status in line with the specification.
// If a plan already exists, nothing is done.
func (d *Reconciler) CreatePlan(ctx context.Context, cachedStatus inspectorInterface.Inspector) (error, bool) {
	var updated bool

	if err, u := d.CreateHighPlan(ctx, cachedStatus); err != nil {
		return err, false
	} else if u {
		updated = true
	}

	if err, u := d.CreateNormalPlan(ctx, cachedStatus); err != nil {
		return err, false
	} else if u {
		updated = true
	}

	return nil, updated
}
