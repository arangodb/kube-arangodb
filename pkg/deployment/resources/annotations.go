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
// Author Adam Janikowski
// Author Tomasz Mielech
//

package resources

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/util/collection"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/rs/zerolog/log"
	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PatchFunc func(name string, d []byte) error

func (r *Resources) EnsureAnnotations(ctx context.Context, cachedStatus inspectorInterface.Inspector) error {
	kubecli := r.context.GetKubeCli()
	monitoringcli := r.context.GetMonitoringV1Cli()

	log.Info().Msgf("Ensuring annotations")

	patchSecret := func(name string, d []byte) error {
		ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
		defer cancel()

		_, err := kubecli.CoreV1().Secrets(r.context.GetNamespace()).Patch(ctxChild, name, types.JSONPatchType, d,
			meta.PatchOptions{})
		return err
	}

	if err := ensureSecretsAnnotations(patchSecret,
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec()); err != nil {
		return err
	}

	patchServiceAccount := func(name string, d []byte) error {
		ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
		defer cancel()

		_, err := kubecli.CoreV1().ServiceAccounts(r.context.GetNamespace()).Patch(ctxChild, name, types.JSONPatchType, d,
			meta.PatchOptions{})
		return err
	}

	if err := ensureServiceAccountsAnnotations(patchServiceAccount,
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec()); err != nil {
		return err
	}

	patchService := func(name string, d []byte) error {
		ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
		defer cancel()

		_, err := kubecli.CoreV1().Services(r.context.GetNamespace()).Patch(ctxChild, name, types.JSONPatchType, d,
			meta.PatchOptions{})
		return err
	}

	if err := ensureServicesAnnotations(patchService,
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec()); err != nil {
		return err
	}

	patchPDB := func(name string, d []byte) error {
		ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
		defer cancel()

		_, err := kubecli.PolicyV1beta1().PodDisruptionBudgets(r.context.GetNamespace()).Patch(ctxChild, name,
			types.JSONPatchType, d, meta.PatchOptions{})
		return err
	}

	if err := ensurePdbsAnnotations(patchPDB,
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec()); err != nil {
		return err
	}

	patchPVC := func(name string, d []byte) error {
		ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
		defer cancel()

		_, err := kubecli.CoreV1().PersistentVolumeClaims(r.context.GetNamespace()).Patch(ctxChild, name,
			types.JSONPatchType, d, meta.PatchOptions{})
		return err
	}

	if err := ensurePvcsAnnotations(patchPVC,
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec()); err != nil {
		return err
	}

	patchPod := func(name string, d []byte) error {
		ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
		defer cancel()

		_, err := kubecli.CoreV1().Pods(r.context.GetNamespace()).Patch(ctxChild, name, types.JSONPatchType, d,
			meta.PatchOptions{})
		return err
	}

	if err := ensurePodsAnnotations(patchPod,
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec().Annotations,
		r.context.GetSpec()); err != nil {
		return err
	}

	patchServiceMonitor := func(name string, d []byte) error {
		ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
		defer cancel()

		_, err := monitoringcli.ServiceMonitors(r.context.GetNamespace()).Patch(ctxChild, name, types.JSONPatchType, d,
			meta.PatchOptions{})
		return err
	}

	if err := ensureServiceMonitorsAnnotations(patchServiceMonitor,
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec()); err != nil {
		return err
	}

	return nil
}

func ensureSecretsAnnotations(patch PatchFunc, cachedStatus inspectorInterface.Inspector, kind, name, namespace string, spec api.DeploymentSpec) error {
	if err := cachedStatus.IterateSecrets(func(secret *core.Secret) error {
		ensureAnnotationsMap(secret.Kind, secret, spec, patch)
		return nil
	}, func(secret *core.Secret) bool {
		return k8sutil.IsChildResource(kind, name, namespace, secret)
	}); err != nil {
		return err
	}

	return nil
}

func ensureServiceAccountsAnnotations(patch PatchFunc, cachedStatus inspectorInterface.Inspector, kind, name, namespace string, spec api.DeploymentSpec) error {
	if err := cachedStatus.IterateServiceAccounts(func(serviceAccount *core.ServiceAccount) error {
		ensureAnnotationsMap(serviceAccount.Kind, serviceAccount, spec, patch)
		return nil
	}, func(serviceAccount *core.ServiceAccount) bool {
		return k8sutil.IsChildResource(kind, name, namespace, serviceAccount)
	}); err != nil {
		return err
	}

	return nil
}

func ensureServicesAnnotations(patch PatchFunc, cachedStatus inspectorInterface.Inspector, kind, name, namespace string, spec api.DeploymentSpec) error {
	if err := cachedStatus.IterateServices(func(service *core.Service) error {
		ensureAnnotationsMap(service.Kind, service, spec, patch)
		return nil
	}, func(service *core.Service) bool {
		return k8sutil.IsChildResource(kind, name, namespace, service)
	}); err != nil {
		return err
	}

	return nil
}

func ensurePdbsAnnotations(patch PatchFunc, cachedStatus inspectorInterface.Inspector, kind, name, namespace string, spec api.DeploymentSpec) error {
	if err := cachedStatus.IteratePodDisruptionBudgets(func(podDisruptionBudget *policy.PodDisruptionBudget) error {
		ensureAnnotationsMap(podDisruptionBudget.Kind, podDisruptionBudget, spec, patch)
		return nil
	}, func(podDisruptionBudget *policy.PodDisruptionBudget) bool {
		return k8sutil.IsChildResource(kind, name, namespace, podDisruptionBudget)
	}); err != nil {
		return err
	}

	return nil
}

func ensurePvcsAnnotations(patch PatchFunc, cachedStatus inspectorInterface.Inspector, kind, name, namespace string, spec api.DeploymentSpec) error {
	if err := cachedStatus.IteratePersistentVolumeClaims(func(persistentVolumeClaim *core.PersistentVolumeClaim) error {
		ensureGroupAnnotationsMap(persistentVolumeClaim.Kind, persistentVolumeClaim, spec, patch)
		return nil
	}, func(persistentVolumeClaim *core.PersistentVolumeClaim) bool {
		return k8sutil.IsChildResource(kind, name, namespace, persistentVolumeClaim)
	}); err != nil {
		return err
	}

	return nil
}

func ensureServiceMonitorsAnnotations(patch PatchFunc, cachedStatus inspectorInterface.Inspector, kind, name, namespace string, spec api.DeploymentSpec) error {
	if err := cachedStatus.IterateServiceMonitors(func(serviceMonitor *monitoring.ServiceMonitor) error {
		ensureAnnotationsMap(serviceMonitor.Kind, serviceMonitor, spec, patch)
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

func ensurePodsAnnotations(patch PatchFunc, cachedStatus inspectorInterface.Inspector, kind, name, namespace string, annotations map[string]string, spec api.DeploymentSpec) error {
	if err := cachedStatus.IteratePods(func(pod *core.Pod) error {
		ensureGroupAnnotationsMap(pod.Kind, pod, spec, patch)
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

func ensureAnnotationsMap(kind string, obj meta.Object, spec api.DeploymentSpec, patchCmd PatchFunc) bool {
	expected := spec.Annotations
	ignored := spec.AnnotationsIgnoreList

	mode := spec.AnnotationsMode.Get(getDefaultMode(expected))

	return ensureObjectMap(kind, obj, mode, expected, obj.GetAnnotations(), collection.AnnotationsPatch, patchCmd, ignored...)
}

func ensureObjectMap(kind string, obj meta.Object, mode api.LabelsMode,
	expected, actual map[string]string,
	patchGetter func(mode api.LabelsMode, expected map[string]string, actual map[string]string, ignored ...string) patch.Patch,
	patchCmd PatchFunc,
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
