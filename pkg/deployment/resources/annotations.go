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
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/util/collection"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	monitoring "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringTypedClient "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/rs/zerolog/log"
	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"
	policyTyped "k8s.io/client-go/kubernetes/typed/policy/v1beta1"
)

func (r *Resources) EnsureAnnotations(cachedStatus inspectorInterface.Inspector) error {
	kubecli := r.context.GetKubeCli()
	monitoringcli := r.context.GetMonitoringV1Cli()

	log.Info().Msgf("Ensuring annotations")

	if err := ensureSecretsAnnotations(kubecli.CoreV1().Secrets(r.context.GetNamespace()),
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec()); err != nil {
		return err
	}

	if err := ensureServiceAccountsAnnotations(kubecli.CoreV1().ServiceAccounts(r.context.GetNamespace()),
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec()); err != nil {
		return err
	}

	if err := ensureServicesAnnotations(kubecli.CoreV1().Services(r.context.GetNamespace()),
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec()); err != nil {
		return err
	}

	if err := ensurePdbsAnnotations(kubecli.PolicyV1beta1().PodDisruptionBudgets(r.context.GetNamespace()),
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec()); err != nil {
		return err
	}

	if err := ensurePvcsAnnotations(kubecli.CoreV1().PersistentVolumeClaims(r.context.GetNamespace()),
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec()); err != nil {
		return err
	}

	if err := ensurePodsAnnotations(kubecli.CoreV1().Pods(r.context.GetNamespace()),
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec().Annotations,
		r.context.GetSpec()); err != nil {
		return err
	}

	if err := ensureServiceMonitorsAnnotations(monitoringcli.ServiceMonitors(r.context.GetNamespace()),
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec()); err != nil {
		return err
	}

	return nil
}

func ensureSecretsAnnotations(client typedCore.SecretInterface, cachedStatus inspectorInterface.Inspector, kind, name, namespace string, spec api.DeploymentSpec) error {
	if err := cachedStatus.IterateSecrets(func(secret *core.Secret) error {
		ensureAnnotationsMap(secret.Kind, secret, spec, func(name string, d []byte) error {
			_, err := client.Patch(name, types.JSONPatchType, d)
			return err
		})
		return nil
	}, func(secret *core.Secret) bool {
		return k8sutil.IsChildResource(kind, name, namespace, secret)
	}); err != nil {
		return err
	}

	return nil
}

func ensureServiceAccountsAnnotations(client typedCore.ServiceAccountInterface, cachedStatus inspectorInterface.Inspector, kind, name, namespace string, spec api.DeploymentSpec) error {
	if err := cachedStatus.IterateServiceAccounts(func(serviceAccount *core.ServiceAccount) error {
		ensureAnnotationsMap(serviceAccount.Kind, serviceAccount, spec, func(name string, d []byte) error {
			_, err := client.Patch(name, types.JSONPatchType, d)
			return err
		})
		return nil
	}, func(serviceAccount *core.ServiceAccount) bool {
		return k8sutil.IsChildResource(kind, name, namespace, serviceAccount)
	}); err != nil {
		return err
	}

	return nil
}

func ensureServicesAnnotations(client typedCore.ServiceInterface, cachedStatus inspectorInterface.Inspector, kind, name, namespace string, spec api.DeploymentSpec) error {
	if err := cachedStatus.IterateServices(func(service *core.Service) error {
		ensureAnnotationsMap(service.Kind, service, spec, func(name string, d []byte) error {
			_, err := client.Patch(name, types.JSONPatchType, d)
			return err
		})
		return nil
	}, func(service *core.Service) bool {
		return k8sutil.IsChildResource(kind, name, namespace, service)
	}); err != nil {
		return err
	}

	return nil
}

func ensurePdbsAnnotations(client policyTyped.PodDisruptionBudgetInterface, cachedStatus inspectorInterface.Inspector, kind, name, namespace string, spec api.DeploymentSpec) error {
	if err := cachedStatus.IteratePodDisruptionBudgets(func(podDisruptionBudget *policy.PodDisruptionBudget) error {
		ensureAnnotationsMap(podDisruptionBudget.Kind, podDisruptionBudget, spec, func(name string, d []byte) error {
			_, err := client.Patch(name, types.JSONPatchType, d)
			return err
		})
		return nil
	}, func(podDisruptionBudget *policy.PodDisruptionBudget) bool {
		return k8sutil.IsChildResource(kind, name, namespace, podDisruptionBudget)
	}); err != nil {
		return err
	}

	return nil
}

func ensurePvcsAnnotations(client typedCore.PersistentVolumeClaimInterface, cachedStatus inspectorInterface.Inspector, kind, name, namespace string, spec api.DeploymentSpec) error {
	if err := cachedStatus.IteratePersistentVolumeClaims(func(persistentVolumeClaim *core.PersistentVolumeClaim) error {
		ensureGroupAnnotationsMap(persistentVolumeClaim.Kind, persistentVolumeClaim, spec, func(name string, d []byte) error {
			_, err := client.Patch(name, types.JSONPatchType, d)
			return err
		})
		return nil
	}, func(persistentVolumeClaim *core.PersistentVolumeClaim) bool {
		return k8sutil.IsChildResource(kind, name, namespace, persistentVolumeClaim)
	}); err != nil {
		return err
	}

	return nil
}

func ensureServiceMonitorsAnnotations(client monitoringTypedClient.ServiceMonitorInterface, cachedStatus inspectorInterface.Inspector, kind, name, namespace string, spec api.DeploymentSpec) error {
	if err := cachedStatus.IterateServiceMonitors(func(serviceMonitor *monitoring.ServiceMonitor) error {
		ensureAnnotationsMap(serviceMonitor.Kind, serviceMonitor, spec, func(name string, d []byte) error {
			_, err := client.Patch(name, types.JSONPatchType, d)
			return err
		})
		return nil
	}, func(serviceMonitor *monitoring.ServiceMonitor) bool {
		return k8sutil.IsChildResource(kind, name, namespace, serviceMonitor)
	}); err != nil {
		return err
	}

	return nil
}

func getObjectGroup(obj meta.Object) api.ServerGroup {
	l := obj.GetLabels()
	if len(l) == 0 {
		return api.ServerGroupUnknown
	}

	group, ok := l[k8sutil.LabelKeyRole]
	if !ok {
		return api.ServerGroupUnknown
	}

	return api.ServerGroupFromRole(group)
}

func ensurePodsAnnotations(client typedCore.PodInterface, cachedStatus inspectorInterface.Inspector, kind, name, namespace string, annotations map[string]string, spec api.DeploymentSpec) error {
	if err := cachedStatus.IteratePods(func(pod *core.Pod) error {
		ensureGroupAnnotationsMap(pod.Kind, pod, spec, func(name string, d []byte) error {
			_, err := client.Patch(name, types.JSONPatchType, d)
			return err
		})
		return nil
	}, func(pod *core.Pod) bool {
		return k8sutil.IsChildResource(kind, name, namespace, pod)
	}); err != nil {
		return err
	}

	return nil
}

func (r *Resources) isChildResource(obj meta.Object) bool {
	return k8sutil.IsChildResource(deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		obj)
}

func getDefaultMode(annotations map[string]string) api.LabelsMode {
	if len(annotations) == 0 {
		return api.LabelsDisabledMode
	}
	return api.LabelsReplaceMode
}

func ensureGroupLabelsMap(kind string, obj meta.Object, spec api.DeploymentSpec,
	patchCmd func(name string, d []byte) error) bool {
	group := getObjectGroup(obj)
	groupSpec := spec.GetServerGroupSpec(group)
	expected := collection.MergeAnnotations(spec.Labels, groupSpec.Labels)

	ignoredList := append(spec.LabelsIgnoreList, groupSpec.LabelsIgnoreList...)

	mode := groupSpec.LabelsMode.Get(spec.LabelsMode.Get(getDefaultMode(expected)))

	return ensureObjectMap(kind, obj, mode, expected, obj.GetLabels(), collection.LabelsPatch, patchCmd, ignoredList...)
}

func ensureLabelsMap(kind string, obj meta.Object, spec api.DeploymentSpec,
	patchCmd func(name string, d []byte) error) bool {
	expected := spec.Labels
	ignored := spec.AnnotationsIgnoreList

	mode := spec.LabelsMode.Get(getDefaultMode(expected))

	return ensureObjectMap(kind, obj, mode, expected, obj.GetLabels(), collection.LabelsPatch, patchCmd, ignored...)
}

func ensureGroupAnnotationsMap(kind string, obj meta.Object, spec api.DeploymentSpec,
	patchCmd func(name string, d []byte) error) bool {
	group := getObjectGroup(obj)
	groupSpec := spec.GetServerGroupSpec(group)
	expected := collection.MergeAnnotations(spec.Annotations, groupSpec.Annotations)

	ignoredList := append(spec.AnnotationsIgnoreList, groupSpec.AnnotationsIgnoreList...)

	mode := groupSpec.AnnotationsMode.Get(spec.AnnotationsMode.Get(getDefaultMode(expected)))

	return ensureObjectMap(kind, obj, mode, expected, obj.GetAnnotations(), collection.AnnotationsPatch, patchCmd, ignoredList...)
}

func ensureAnnotationsMap(kind string, obj meta.Object, spec api.DeploymentSpec,
	patchCmd func(name string, d []byte) error) bool {
	expected := spec.Annotations
	ignored := spec.AnnotationsIgnoreList

	mode := spec.AnnotationsMode.Get(getDefaultMode(expected))

	return ensureObjectMap(kind, obj, mode, expected, obj.GetAnnotations(), collection.AnnotationsPatch, patchCmd, ignored...)
}

func ensureObjectMap(kind string, obj meta.Object, mode api.LabelsMode,
	expected, actual map[string]string,
	patchGetter func(mode api.LabelsMode, expected map[string]string, actual map[string]string, ignored ...string) patch.Patch,
	patchCmd func(name string, d []byte) error,
	ignored ...string) bool {
	p := patchGetter(mode, expected, actual, ignored...)

	if len(p) == 0 {
		return false
	}

	log.Info().Msgf("Replacing annotations for %s %s", kind, obj.GetName())

	d, err := p.Marshal()
	if err != nil {
		log.Warn().Err(err).Msgf("Unable to marshal kubernetes patch instruction")
		return false
	}

	if err := patchCmd(obj.GetName(), d); err != nil {
		log.Warn().Err(err).Msgf("Unable to patch Pod")
		return false
	}

	return true
}
