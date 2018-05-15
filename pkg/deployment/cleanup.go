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

package deployment

import (
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// removePodFinalizers removes all finalizers from all pods owned by us.
func (d *Deployment) removePodFinalizers() error {
	log := d.deps.Log
	kubecli := d.GetKubeCli()
	pods, err := d.GetOwnedPods()
	if err != nil {
		return maskAny(err)
	}
	for _, p := range pods {
		if err := k8sutil.RemovePodFinalizers(log, kubecli, &p, p.GetFinalizers()); err != nil {
			log.Warn().Err(err).Msg("Failed to remove pod finalizers")
		}
	}
	return nil
}

// removePVCFinalizers removes all finalizers from all PVCs owned by us.
func (d *Deployment) removePVCFinalizers() error {
	log := d.deps.Log
	kubecli := d.GetKubeCli()
	pvcs, err := d.GetOwnedPVCs()
	if err != nil {
		return maskAny(err)
	}
	for _, p := range pvcs {
		if err := k8sutil.RemovePVCFinalizers(log, kubecli, &p, p.GetFinalizers()); err != nil {
			log.Warn().Err(err).Msg("Failed to remove PVC finalizers")
		}
	}
	return nil
}
