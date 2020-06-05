//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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

package deployment

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"

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
func (d *Deployment) runDeploymentFinalizers(ctx context.Context, cachedStatus inspector.Inspector) error {
	log := d.deps.Log
	var removalList []string

	depls := d.deps.DatabaseCRCli.DatabaseV1().ArangoDeployments(d.GetNamespace())
	updated, err := depls.Get(d.apiObject.GetName(), metav1.GetOptions{})
	if err != nil {
		return maskAny(err)
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
		if err := removeDeploymentFinalizers(log, d.deps.DatabaseCRCli, updated, removalList); err != nil {
			log.Debug().Err(err).Msg("Failed to update ArangoDeployment (to remove finalizers)")
			return maskAny(err)
		}
	}
	return nil
}

// inspectRemoveChildFinalizers checks the finalizer condition for remove-child-finalizers.
// It returns nil if the finalizer can be removed.
func (d *Deployment) inspectRemoveChildFinalizers(ctx context.Context, log zerolog.Logger, depl *api.ArangoDeployment, cachedStatus inspector.Inspector) error {
	if err := d.removePodFinalizers(cachedStatus); err != nil {
		return maskAny(err)
	}
	if err := d.removePVCFinalizers(); err != nil {
		return maskAny(err)
	}

	return nil
}

// removeDeploymentFinalizers removes the given finalizers from the given PVC.
func removeDeploymentFinalizers(log zerolog.Logger, cli versioned.Interface, depl *api.ArangoDeployment, finalizers []string) error {
	depls := cli.DatabaseV1().ArangoDeployments(depl.GetNamespace())
	getFunc := func() (metav1.Object, error) {
		result, err := depls.Get(depl.GetName(), metav1.GetOptions{})
		if err != nil {
			return nil, maskAny(err)
		}
		return result, nil
	}
	updateFunc := func(updated metav1.Object) error {
		updatedDepl := updated.(*api.ArangoDeployment)
		result, err := depls.Update(updatedDepl)
		if err != nil {
			return maskAny(err)
		}
		*depl = *result
		return nil
	}
	ignoreNotFound := false
	if err := k8sutil.RemoveFinalizers(log, finalizers, getFunc, updateFunc, ignoreNotFound); err != nil {
		return maskAny(err)
	}
	return nil
}
