//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package resources

import (
	"context"

	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/util/collection"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tools"
)

type PatchFunc func(name string, d []byte) error

func (r *Resources) EnsureAnnotations(ctx context.Context, cachedStatus inspectorInterface.Inspector) error {
	log := r.log.Str("section", "annotations")

	log.Trace("Ensuring annotations")

	patchSecret := func(name string, d []byte) error {
		return globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			_, err := cachedStatus.SecretsModInterface().V1().Patch(ctxChild, name, types.JSONPatchType, d,
				meta.PatchOptions{})
			return err
		})
	}

	if err := r.ensureSecretsAnnotations(patchSecret,
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec()); err != nil {
		return err
	}

	patchServiceAccount := func(name string, d []byte) error {
		return globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			_, err := cachedStatus.ServiceAccountsModInterface().V1().Patch(ctxChild, name,
				types.JSONPatchType, d, meta.PatchOptions{})
			return err
		})
	}

	if err := r.ensureServiceAccountsAnnotations(patchServiceAccount,
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec()); err != nil {
		return err
	}

	patchService := func(name string, d []byte) error {
		return globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			_, err := cachedStatus.ServicesModInterface().V1().Patch(ctxChild, name, types.JSONPatchType, d,
				meta.PatchOptions{})
			return err
		})
	}

	if err := r.ensureServicesAnnotations(patchService,
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec()); err != nil {
		return err
	}

	patchPDB := func(name string, d []byte) error {
		return globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			if _, err := cachedStatus.PodDisruptionBudget().V1(); err == nil {
				_, err = cachedStatus.PodDisruptionBudgetsModInterface().V1().Patch(ctxChild, name,
					types.JSONPatchType, d, meta.PatchOptions{})
				return err
			}

			return nil
		})
	}

	if err := r.ensurePdbsAnnotations(patchPDB,
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec()); err != nil {
		return err
	}

	patchPVC := func(name string, d []byte) error {
		return globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			_, err := cachedStatus.PersistentVolumeClaimsModInterface().V1().Patch(ctxChild, name,
				types.JSONPatchType, d, meta.PatchOptions{})
			return err
		})
	}

	if err := r.ensurePvcsAnnotations(patchPVC,
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec()); err != nil {
		return err
	}

	patchPod := func(name string, d []byte) error {
		return globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			_, err := cachedStatus.PodsModInterface().V1().Patch(ctxChild, name, types.JSONPatchType, d,
				meta.PatchOptions{})
			return err
		})
	}

	if err := r.ensurePodsAnnotations(patchPod,
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec()); err != nil {
		return err
	}

	patchServiceMonitor := func(name string, d []byte) error {
		return globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			_, err := cachedStatus.ServiceMonitorsModInterface().V1().Patch(ctxChild, name, types.JSONPatchType, d,
				meta.PatchOptions{})
			return err
		})
	}

	if err := r.ensureServiceMonitorsAnnotations(patchServiceMonitor,
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec()); err != nil {
		return err
	}

	return nil
}

func (r *Resources) ensureSecretsAnnotations(patch PatchFunc, cachedStatus inspectorInterface.Inspector, kind, name, namespace string, spec api.DeploymentSpec) error {
	if err := cachedStatus.Secret().V1().Iterate(func(secret *core.Secret) error {
		r.ensureAnnotationsMap(secret.Kind, secret, spec, patch)
		return nil
	}, func(secret *core.Secret) bool {
		return tools.IsChildResource(kind, name, namespace, secret)
	}); err != nil {
		return err
	}

	return nil
}

func (r *Resources) ensureServiceAccountsAnnotations(patch PatchFunc, cachedStatus inspectorInterface.Inspector, kind, name, namespace string, spec api.DeploymentSpec) error {
	if err := cachedStatus.ServiceAccount().V1().Iterate(func(serviceAccount *core.ServiceAccount) error {
		r.ensureAnnotationsMap(serviceAccount.Kind, serviceAccount, spec, patch)
		return nil
	}, func(serviceAccount *core.ServiceAccount) bool {
		return tools.IsChildResource(kind, name, namespace, serviceAccount)
	}); err != nil {
		return err
	}

	return nil
}

func (r *Resources) ensureServicesAnnotations(patch PatchFunc, cachedStatus inspectorInterface.Inspector, kind, name, namespace string, spec api.DeploymentSpec) error {
	if err := cachedStatus.Service().V1().Iterate(func(service *core.Service) error {
		r.ensureAnnotationsMap(service.Kind, service, spec, patch)
		return nil
	}, func(service *core.Service) bool {
		return tools.IsChildResource(kind, name, namespace, service)
	}); err != nil {
		return err
	}

	return nil
}

func (r *Resources) ensurePdbsAnnotations(patch PatchFunc, cachedStatus inspectorInterface.Inspector, kind, name, namespace string,
	spec api.DeploymentSpec) error {
	if inspector, err := cachedStatus.PodDisruptionBudget().V1(); err == nil {
		if err := inspector.Iterate(func(podDisruptionBudget *policy.PodDisruptionBudget) error {
			r.ensureAnnotationsMap(podDisruptionBudget.Kind, podDisruptionBudget, spec, patch)
			return nil
		}, func(podDisruptionBudget *policy.PodDisruptionBudget) bool {
			return tools.IsChildResource(kind, name, namespace, podDisruptionBudget)
		}); err != nil {
			return err
		}

		return nil
	}

	return nil
}

func (r *Resources) ensurePvcsAnnotations(patch PatchFunc, cachedStatus inspectorInterface.Inspector, kind, name, namespace string, spec api.DeploymentSpec) error {
	if err := cachedStatus.PersistentVolumeClaim().V1().Iterate(func(persistentVolumeClaim *core.PersistentVolumeClaim) error {
		r.ensureGroupAnnotationsMap(persistentVolumeClaim.Kind, persistentVolumeClaim, spec, patch)
		return nil
	}, func(persistentVolumeClaim *core.PersistentVolumeClaim) bool {
		return tools.IsChildResource(kind, name, namespace, persistentVolumeClaim)
	}); err != nil {
		return err
	}

	return nil
}

func (r *Resources) ensureServiceMonitorsAnnotations(patch PatchFunc, cachedStatus inspectorInterface.Inspector, kind, name, namespace string, spec api.DeploymentSpec) error {
	i, err := cachedStatus.ServiceMonitor().V1()
	if err != nil {
		if kerrors.IsForbiddenOrNotFound(err) {
			return nil
		}
		return err
	}
	if err := i.Iterate(func(serviceMonitor *monitoring.ServiceMonitor) error {
		r.ensureAnnotationsMap(serviceMonitor.Kind, serviceMonitor, spec, patch)
		return nil
	}, func(serviceMonitor *monitoring.ServiceMonitor) bool {
		return tools.IsChildResource(kind, name, namespace, serviceMonitor)
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

func (r *Resources) ensurePodsAnnotations(patch PatchFunc, cachedStatus inspectorInterface.Inspector, kind, name, namespace string, spec api.DeploymentSpec) error {

	if err := cachedStatus.Pod().V1().Iterate(func(pod *core.Pod) error {
		r.ensureGroupAnnotationsMap(pod.Kind, pod, spec, patch)
		return nil
	}, func(pod *core.Pod) bool {
		return tools.IsChildResource(kind, name, namespace, pod)
	}); err != nil {
		return err
	}

	return nil
}

func (r *Resources) isChildResource(obj meta.Object) bool {
	return tools.IsChildResource(deployment.ArangoDeploymentResourceKind,
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

func (r *Resources) ensureGroupLabelsMap(kind string, obj meta.Object, spec api.DeploymentSpec,
	patchCmd func(name string, d []byte) error) bool {
	group := getObjectGroup(obj)
	groupSpec := spec.GetServerGroupSpec(group)
	expected := collection.MergeAnnotations(spec.Labels, groupSpec.Labels)

	ignoredList := append(spec.LabelsIgnoreList, groupSpec.LabelsIgnoreList...)

	mode := groupSpec.LabelsMode.Get(spec.LabelsMode.Get(getDefaultMode(expected)))

	return r.ensureObjectMap(kind, obj, mode, expected, obj.GetLabels(), collection.LabelsPatch, patchCmd, ignoredList...)
}

func (r *Resources) ensureLabelsMap(kind string, obj meta.Object, spec api.DeploymentSpec,
	patchCmd func(name string, d []byte) error) bool {
	expected := spec.Labels
	ignored := spec.LabelsIgnoreList

	mode := spec.LabelsMode.Get(getDefaultMode(expected))

	return r.ensureObjectMap(kind, obj, mode, expected, obj.GetLabels(), collection.LabelsPatch, patchCmd, ignored...)
}

func (r *Resources) ensureGroupAnnotationsMap(kind string, obj meta.Object, spec api.DeploymentSpec,
	patchCmd func(name string, d []byte) error) {
	group := getObjectGroup(obj)
	groupSpec := spec.GetServerGroupSpec(group)
	expected := collection.MergeAnnotations(spec.Annotations, groupSpec.Annotations)

	ignoredList := append(spec.AnnotationsIgnoreList, groupSpec.AnnotationsIgnoreList...)

	mode := groupSpec.AnnotationsMode.Get(spec.AnnotationsMode.Get(getDefaultMode(expected)))

	r.ensureObjectMap(kind, obj, mode, expected, obj.GetAnnotations(), collection.AnnotationsPatch, patchCmd, ignoredList...)
}

func (r *Resources) ensureAnnotationsMap(kind string, obj meta.Object, spec api.DeploymentSpec, patchCmd PatchFunc) {
	expected := spec.Annotations
	ignored := spec.AnnotationsIgnoreList

	mode := spec.AnnotationsMode.Get(getDefaultMode(expected))

	r.ensureObjectMap(kind, obj, mode, expected, obj.GetAnnotations(), collection.AnnotationsPatch, patchCmd, ignored...)
}

func (r *Resources) ensureObjectMap(kind string, obj meta.Object, mode api.LabelsMode,
	expected, actual map[string]string,
	patchGetter func(mode api.LabelsMode, expected map[string]string, actual map[string]string, ignored ...string) patch.Patch,
	patchCmd PatchFunc,
	ignored ...string) bool {
	p := patchGetter(mode, expected, actual, ignored...)

	log := r.log.Str("section", "annotations")

	if len(p) == 0 {
		return false
	}

	log.Info("Replacing annotations for %s %s", kind, obj.GetName())

	d, err := p.Marshal()
	if err != nil {
		log.Err(err).Warn("Unable to marshal kubernetes patch instruction")
		return false
	}

	if err := patchCmd(obj.GetName(), d); err != nil {
		log.Err(err).Warn("Unable to patch Pod")
		return false
	}

	return true
}
