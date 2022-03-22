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

package deployment

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/util/globals"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	pvcv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolumeclaim/v1"
	podv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/pod/v1"
)

// removePodFinalizers removes all finalizers from all pods owned by us.
func (d *Deployment) removePodFinalizers(ctx context.Context, cachedStatus inspectorInterface.Inspector) (bool, error) {
	log := d.deps.Log

	found := false

	if err := cachedStatus.Pod().V1().Iterate(func(pod *core.Pod) error {
		log.Info().Str("pod", pod.GetName()).Msgf("Removing Pod Finalizer")
		if count, err := k8sutil.RemovePodFinalizers(ctx, cachedStatus, log, d.PodsModInterface(), pod, constants.ManagedFinalizers(), true); err != nil {
			log.Warn().Err(err).Msg("Failed to remove pod finalizers")
			return err
		} else if count > 0 {
			found = true
		}

		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()

		if err := d.PodsModInterface().Delete(ctxChild, pod.GetName(), meta.DeleteOptions{
			GracePeriodSeconds: util.NewInt64(0),
		}); err != nil {
			if !k8sutil.IsNotFound(err) {
				log.Warn().Err(err).Msg("Failed to remove pod")
				return err
			}
		}
		return nil
	}, podv1.FilterPodsByLabels(k8sutil.LabelsForDeployment(d.GetName(), ""))); err != nil {
		return false, err
	}

	return found, nil
}

// removePVCFinalizers removes all finalizers from all PVCs owned by us.
func (d *Deployment) removePVCFinalizers(ctx context.Context, cachedStatus inspectorInterface.Inspector) (bool, error) {
	log := d.deps.Log

	found := false

	if err := cachedStatus.PersistentVolumeClaim().V1().Iterate(func(pvc *core.PersistentVolumeClaim) error {
		log.Info().Str("pvc", pvc.GetName()).Msgf("Removing PVC Finalizer")
		if count, err := k8sutil.RemovePVCFinalizers(ctx, cachedStatus, log, d.PersistentVolumeClaimsModInterface(), pvc, constants.ManagedFinalizers(), true); err != nil {
			log.Warn().Err(err).Msg("Failed to remove PVC finalizers")
			return err
		} else if count > 0 {
			found = true
		}
		return nil
	}, pvcv1.FilterPersistentVolumeClaimsByLabels(k8sutil.LabelsForDeployment(d.GetName(), ""))); err != nil {
		return false, err
	}

	return found, nil
}
