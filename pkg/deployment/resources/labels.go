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
// Author Adam Janikowski
//

package resources

import (
	"encoding/json"

	monitoring "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"

	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	reservedLabels = RestrictedList{
		k8sutil.LabelKeyArangoDeployment,
		k8sutil.LabelKeyArangoLocalStorage,
		k8sutil.LabelKeyApp,
		k8sutil.LabelKeyRole,
		k8sutil.LabelKeyArangoExporter,
	}
)

func (r *Resources) EnsureLabels(cachedStatus inspector.Inspector) error {
	r.log.Info().Msgf("Ensuring labels")

	if err := r.EnsureSecretLabels(cachedStatus); err != nil {
		return err
	}

	if err := r.EnsureServiceAccountsLabels(cachedStatus); err != nil {
		return err
	}

	if err := r.EnsureServicesLabels(cachedStatus); err != nil {
		return err
	}

	if err := r.EnsureServiceMonitorsLabels(cachedStatus); err != nil {
		return err
	}

	if err := r.EnsurePodsLabels(cachedStatus); err != nil {
		return err
	}

	if err := r.EnsurePersistentVolumeClaimsLabels(cachedStatus); err != nil {
		return err
	}

	if err := r.EnsurePodDisruptionBudgetsLabels(cachedStatus); err != nil {
		return err
	}

	return nil
}

func (r *Resources) EnsureSecretLabels(cachedStatus inspector.Inspector) error {
	changed := false
	if err := cachedStatus.IterateSecrets(func(secret *core.Secret) error {
		if p := ensureLabelsFromMaps(secret, r.context.GetSpec().Labels, r.context.GetSpec().GetServerGroupSpec(getObjectGroup(secret)).Labels); len(p) != 0 {
			patch, err := json.Marshal(p)
			if err != nil {
				return err
			}
			r.log.Info().Int("changes", len(p)).Msgf("Updating labels for secret %s", secret.GetName())

			if _, err = r.context.GetKubeCli().CoreV1().Secrets(r.context.GetAPIObject().GetNamespace()).Patch(secret.GetName(), types.JSONPatchType, patch); err != nil {
				return err
			}

			changed = true
			return nil
		}

		return nil
	}, func(secret *core.Secret) bool {
		return r.isChildResource(secret)
	}); err != nil {
		return err
	}

	if changed {
		return errors.Reconcile()
	}

	return nil
}

func (r *Resources) EnsureServiceAccountsLabels(cachedStatus inspector.Inspector) error {
	changed := false
	if err := cachedStatus.IterateServiceAccounts(func(serviceAccount *core.ServiceAccount) error {
		if p := ensureLabelsFromMaps(serviceAccount, r.context.GetSpec().Labels, r.context.GetSpec().GetServerGroupSpec(getObjectGroup(serviceAccount)).Labels); len(p) != 0 {
			patch, err := json.Marshal(p)
			if err != nil {
				return err
			}
			r.log.Info().Int("changes", len(p)).Msgf("Updating labels for ServiceAccount %s", serviceAccount.GetName())

			if _, err = r.context.GetKubeCli().CoreV1().ServiceAccounts(r.context.GetAPIObject().GetNamespace()).Patch(serviceAccount.GetName(), types.JSONPatchType, patch); err != nil {
				return err
			}

			changed = true
			return nil
		}

		return nil
	}, func(serviceAccount *core.ServiceAccount) bool {
		return r.isChildResource(serviceAccount)
	}); err != nil {
		return err
	}

	if changed {
		return errors.Reconcile()
	}

	return nil
}

func (r *Resources) EnsureServicesLabels(cachedStatus inspector.Inspector) error {
	changed := false
	if err := cachedStatus.IterateServices(func(service *core.Service) error {
		if p := ensureLabelsFromMaps(service, r.context.GetSpec().Labels, r.context.GetSpec().GetServerGroupSpec(getObjectGroup(service)).Labels); len(p) != 0 {
			patch, err := json.Marshal(p)
			if err != nil {
				return err
			}
			r.log.Info().Int("changes", len(p)).Msgf("Updating labels for Service %s", service.GetName())

			if _, err = r.context.GetKubeCli().CoreV1().Services(r.context.GetAPIObject().GetNamespace()).Patch(service.GetName(), types.JSONPatchType, patch); err != nil {
				return err
			}

			changed = true
			return nil
		}

		return nil
	}, func(service *core.Service) bool {
		return r.isChildResource(service)
	}); err != nil {
		return err
	}

	if changed {
		return errors.Reconcile()
	}

	return nil
}

func (r *Resources) EnsureServiceMonitorsLabels(cachedStatus inspector.Inspector) error {
	changed := false
	if err := cachedStatus.IterateServiceMonitors(func(serviceMonitor *monitoring.ServiceMonitor) error {
		if p := ensureLabelsFromMaps(serviceMonitor, r.context.GetSpec().Labels, r.context.GetSpec().GetServerGroupSpec(getObjectGroup(serviceMonitor)).Labels); len(p) != 0 {
			patch, err := json.Marshal(p)
			if err != nil {
				return err
			}
			r.log.Info().Int("changes", len(p)).Msgf("Updating labels for ServiceMonitor %s", serviceMonitor.GetName())

			if _, err = r.context.GetMonitoringV1Cli().ServiceMonitors(r.context.GetAPIObject().GetNamespace()).Patch(serviceMonitor.GetName(), types.JSONPatchType, patch); err != nil {
				return err
			}

			changed = true
			return nil
		}

		return nil
	}, func(serviceMonitor *monitoring.ServiceMonitor) bool {
		return r.isChildResource(serviceMonitor)
	}); err != nil {
		return err
	}

	if changed {
		return errors.Reconcile()
	}

	return nil
}

func (r *Resources) EnsurePodsLabels(cachedStatus inspector.Inspector) error {
	changed := false
	if err := cachedStatus.IteratePods(func(pod *core.Pod) error {
		if p := ensureLabelsFromMaps(pod, r.context.GetSpec().Labels, r.context.GetSpec().GetServerGroupSpec(getObjectGroup(pod)).Labels); len(p) != 0 {
			patch, err := json.Marshal(p)
			if err != nil {
				return err
			}
			r.log.Info().Int("changes", len(p)).Msgf("Updating labels for Pod %s", pod.GetName())

			if _, err = r.context.GetKubeCli().CoreV1().Pods(r.context.GetAPIObject().GetNamespace()).Patch(pod.GetName(), types.JSONPatchType, patch); err != nil {
				return err
			}

			changed = true
			return nil
		}

		return nil
	}, func(pod *core.Pod) bool {
		return r.isChildResource(pod)
	}); err != nil {
		return err
	}

	if changed {
		return errors.Reconcile()
	}

	return nil
}

func (r *Resources) EnsurePersistentVolumeClaimsLabels(cachedStatus inspector.Inspector) error {
	changed := false
	if err := cachedStatus.IteratePersistentVolumeClaims(func(persistentVolumeClaim *core.PersistentVolumeClaim) error {
		if p := ensureLabelsFromMaps(persistentVolumeClaim, r.context.GetSpec().Labels, r.context.GetSpec().GetServerGroupSpec(getObjectGroup(persistentVolumeClaim)).Labels); len(p) != 0 {
			patch, err := json.Marshal(p)
			if err != nil {
				return err
			}
			r.log.Info().Int("changes", len(p)).Msgf("Updating labels for PersistentVolumeClaim %s", persistentVolumeClaim.GetName())

			if _, err = r.context.GetKubeCli().CoreV1().PersistentVolumeClaims(r.context.GetAPIObject().GetNamespace()).Patch(persistentVolumeClaim.GetName(), types.JSONPatchType, patch); err != nil {
				return err
			}

			changed = true
			return nil
		}

		return nil
	}, func(persistentVolumeClaim *core.PersistentVolumeClaim) bool {
		return r.isChildResource(persistentVolumeClaim)
	}); err != nil {
		return err
	}

	if changed {
		return errors.Reconcile()
	}

	return nil
}

func (r *Resources) EnsurePodDisruptionBudgetsLabels(cachedStatus inspector.Inspector) error {
	changed := false
	if err := cachedStatus.IteratePodDisruptionBudgets(func(budget *policy.PodDisruptionBudget) error {
		if p := ensureLabelsFromMaps(budget, r.context.GetSpec().Labels, r.context.GetSpec().GetServerGroupSpec(getObjectGroup(budget)).Labels); len(p) != 0 {
			patch, err := json.Marshal(p)
			if err != nil {
				return err
			}
			r.log.Info().Int("changes", len(p)).Msgf("Updating labels for PodDisruptionBudget %s", budget.GetName())

			if _, err = r.context.GetKubeCli().PolicyV1beta1().PodDisruptionBudgets(r.context.GetAPIObject().GetNamespace()).Patch(budget.GetName(), types.JSONPatchType, patch); err != nil {
				return err
			}

			changed = true
			return nil
		}

		return nil
	}, func(budget *policy.PodDisruptionBudget) bool {
		return r.isChildResource(budget)
	}); err != nil {
		return err
	}

	if changed {
		return errors.Reconcile()
	}

	return nil
}

type RestrictedList []string

func (r RestrictedList) IsRestricted(s string) bool {
	for _, i := range r {
		if i == s {
			return true
		}
	}

	return false
}

func ensureLabelsFromMaps(obj meta.Object, labels ...map[string]string) patch.Patch {
	m := map[string]string{}
	for _, labelMap := range labels {
		if len(labelMap) == 0 {
			continue
		}
		for k, v := range labelMap {
			m[k] = v
		}
	}

	return ensureLabels(obj, m)
}

func ensureLabels(obj meta.Object, labels map[string]string) patch.Patch {
	objLabels := obj.GetLabels()
	if objLabels == nil {
		objLabels = map[string]string{}
	}

	if labels == nil {
		labels = map[string]string{}
	}

	m := ensureMap(objLabels, labels, reservedLabels)
	if len(m) == 0 {
		return m
	}

	// Map not present, we need to fix this
	if obj.GetLabels() == nil {
		p := patch.NewPatch()
		p.ItemAdd(patch.NewPath("metadata", "labels"), map[string]string{})
		p.Add(m...)
		return p
	}

	return m
}

func ensureMap(obj map[string]string, labels map[string]string, restricted RestrictedList) patch.Patch {
	p := patch.Patch{}

	for k := range obj {
		if restricted.IsRestricted(k) {
			continue
		}

		if _, ok := labels[k]; !ok {
			p.ItemRemove(patch.NewPath("metadata", "labels", k))
		}
	}

	for k, v := range labels {
		if restricted.IsRestricted(k) {
			continue
		}

		if objV, ok := obj[k]; !ok {
			p.ItemAdd(patch.NewPath("metadata", "labels", k), v)
			continue
		} else if objV != v {
			p.ItemReplace(patch.NewPath("metadata", "labels", k), v)
			continue
		}
	}

	return p
}
