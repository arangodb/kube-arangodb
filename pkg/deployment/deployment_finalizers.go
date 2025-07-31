//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

package deployment

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
)

var expectedFinalizers = []string{
	utilConstants.FinalizerDeplRemoveChildFinalizers,
}

// ensureFinalizers adds all required finalizers to the given deployment (in memory).
func ensureFinalizers(depl *api.ArangoDeployment) bool {
	fx := make(map[string]bool, len(expectedFinalizers))

	st := len(depl.Finalizers)

	for _, fn := range expectedFinalizers {
		fx[fn] = false
	}

	for _, f := range depl.GetFinalizers() {
		if _, ok := fx[f]; ok {
			fx[f] = true
		}
	}

	for _, fn := range expectedFinalizers {
		if !fx[fn] {
			depl.Finalizers = append(depl.Finalizers, fn)
		}
	}

	// Set finalizers
	return st != len(depl.Finalizers)
}

// runDeploymentFinalizers goes through the list of ArangoDeployoment finalizers to see if they can be removed.
func (d *Deployment) runDeploymentFinalizers(ctx context.Context, cachedStatus inspectorInterface.Inspector) error {
	var removalList []string

	depls := d.deps.Client.Arango().DatabaseV1().ArangoDeployments(d.GetNamespace())
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	updated, err := depls.Get(ctxChild, d.currentObject.GetName(), meta.GetOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	for _, f := range updated.ObjectMeta.GetFinalizers() {
		switch f {
		case utilConstants.FinalizerDeplRemoveChildFinalizers:
			d.log.Debug("Inspecting 'remove child finalizers' finalizer")
			if retry, err := d.inspectRemoveChildFinalizers(ctx, updated, cachedStatus); err == nil && !retry {
				removalList = append(removalList, f)
			} else if retry {
				d.log.Str("finalizer", f).Debug("Retry on finalizer removal")
			} else {
				d.log.Err(err).Str("finalizer", f).Debug("Cannot remove finalizer yet")
			}
		}
	}
	// Remove finalizers (if needed)
	if len(removalList) > 0 {
		c := d.deps.Client.Arango().DatabaseV1().ArangoDeployments(updated.GetNamespace())
		if _, err := k8sutil.RemoveSelectedFinalizers[*api.ArangoDeployment](ctx, c, c, updated, removalList, false); err != nil {
			d.log.Err(err).Debug("Failed to update ArangoDeployment (to remove finalizers)")
			return errors.WithStack(err)
		}
	}
	return nil
}

// inspectRemoveChildFinalizers checks the finalizer condition for remove-child-finalizers.
// It returns nil if the finalizer can be removed.
func (d *Deployment) inspectRemoveChildFinalizers(ctx context.Context, _ *api.ArangoDeployment, cachedStatus inspectorInterface.Inspector) (bool, error) {
	retry := false

	if found, err := d.removePodFinalizers(ctx, cachedStatus); err != nil {
		return false, errors.WithStack(err)
	} else if found {
		retry = true
	}
	if found, err := d.removePVCFinalizers(ctx, cachedStatus); err != nil {
		return false, errors.WithStack(err)
	} else if found {
		retry = true
	}

	return retry, nil
}
