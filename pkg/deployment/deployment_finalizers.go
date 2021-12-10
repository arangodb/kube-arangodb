//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Tomasz Mielech
//

package deployment

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/util/globals"

	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// ensureFinalizers adds all required finalizers to the given deployment (in memory).
func ensureFinalizers(depl *api.ArangoDeployment) {
	for _, f := range depl.GetFinalizers() {
		if f == constants.FinalizerDeplRemoveChildFinalizers {
			// Finalizer already set
			return
		}
	}
	// Set finalizers
	depl.SetFinalizers(append(depl.GetFinalizers(), constants.FinalizerDeplRemoveChildFinalizers))
}

// runDeploymentFinalizers goes through the list of ArangoDeployoment finalizers to see if they can be removed.
func (d *Deployment) runDeploymentFinalizers(ctx context.Context, cachedStatus inspectorInterface.Inspector) error {
	log := d.deps.Log
	var removalList []string

	depls := d.deps.DatabaseCRCli.DatabaseV1().ArangoDeployments(d.GetNamespace())
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	updated, err := depls.Get(ctxChild, d.apiObject.GetName(), metav1.GetOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	for _, f := range updated.ObjectMeta.GetFinalizers() {
		switch f {
		case constants.FinalizerDeplRemoveChildFinalizers:
			log.Debug().Msg("Inspecting 'remove child finalizers' finalizer")
			if err := d.inspectRemoveChildFinalizers(ctx, log, updated, cachedStatus); err == nil {
				removalList = append(removalList, f)
			} else {
				log.Debug().Err(err).Str("finalizer", f).Msg("Cannot remove finalizer yet")
			}
		}
	}
	// Remove finalizers (if needed)
	if len(removalList) > 0 {
		if err := removeDeploymentFinalizers(ctx, log, d.deps.DatabaseCRCli, updated, removalList); err != nil {
			log.Debug().Err(err).Msg("Failed to update ArangoDeployment (to remove finalizers)")
			return errors.WithStack(err)
		}
	}
	return nil
}

// inspectRemoveChildFinalizers checks the finalizer condition for remove-child-finalizers.
// It returns nil if the finalizer can be removed.
func (d *Deployment) inspectRemoveChildFinalizers(ctx context.Context, _ zerolog.Logger, _ *api.ArangoDeployment, cachedStatus inspectorInterface.Inspector) error {
	if err := d.removePodFinalizers(ctx, cachedStatus); err != nil {
		return errors.WithStack(err)
	}
	if err := d.removePVCFinalizers(ctx, cachedStatus); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// removeDeploymentFinalizers removes the given finalizers from the given PVC.
func removeDeploymentFinalizers(ctx context.Context, log zerolog.Logger, cli versioned.Interface,
	depl *api.ArangoDeployment, finalizers []string) error {
	depls := cli.DatabaseV1().ArangoDeployments(depl.GetNamespace())
	getFunc := func() (metav1.Object, error) {
		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()

		result, err := depls.Get(ctxChild, depl.GetName(), metav1.GetOptions{})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return result, nil
	}
	updateFunc := func(updated metav1.Object) error {
		updatedDepl := updated.(*api.ArangoDeployment)
		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()

		result, err := depls.Update(ctxChild, updatedDepl, metav1.UpdateOptions{})
		if err != nil {
			return errors.WithStack(err)
		}
		*depl = *result
		return nil
	}
	ignoreNotFound := false
	if err := k8sutil.RemoveFinalizers(log, finalizers, getFunc, updateFunc, ignoreNotFound); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
