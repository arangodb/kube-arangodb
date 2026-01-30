//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package tokens

import (
	"context"
	"fmt"
	"path"

	admission "k8s.io/api/admission/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	permissionApi "github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/webhook"
)

func Handler(client kclient.Client) webhook.Handler[*core.Pod] {
	return handler{
		client: client,
	}
}

var _ webhook.MutationHandler[*core.Pod] = handler{}

type handler struct {
	client kclient.Client
}

func (h handler) CanHandle(ctx context.Context, log logging.Logger, t webhook.AdmissionRequestType, request *admission.AdmissionRequest, old, new *core.Pod) bool {
	if request == nil {
		return false
	}

	if request.Operation != admission.Create {
		return false
	}

	if new == nil {
		return false
	}

	if _, ok := new.GetLabels()[utilConstants.ProfilesDeployment]; !ok {
		return false
	}

	return true
}

func (h handler) Mutate(ctx context.Context, log logging.Logger, t webhook.AdmissionRequestType, request *admission.AdmissionRequest, old, new *core.Pod) (webhook.MutationResponse, error) {
	if !h.CanHandle(ctx, log, t, request, old, new) {
		return webhook.MutationResponse{}, errors.Errorf("Object cannot be handled")
	}

	deploymentName, ok := new.GetLabels()[utilConstants.ProfilesDeployment]
	if !ok {
		logger.Warn("No deployment label found")
		return webhook.MutationResponse{
			ValidationResponse: webhook.ValidationResponse{Allowed: true},
		}, nil
	}

	if tokenName, ok := new.GetLabels()[utilConstants.TokenAttachment]; ok {
		return h.generateMutationResponse(ctx, new, deploymentName, new.GetNamespace(), tokenName)
	}

	if saName := new.Spec.ServiceAccountName; saName != "" {
		sa, err := h.client.Kubernetes().CoreV1().ServiceAccounts(new.GetNamespace()).Get(ctx, saName, meta.GetOptions{})
		if err != nil {
			logger.Err(err).Warn("Failed to get service account")
			return webhook.MutationResponse{}, errors.Wrapf(err, "Failed to get service account")
		}

		if tokenName, ok := sa.GetLabels()[utilConstants.TokenAttachment]; ok {
			return h.generateMutationResponse(ctx, new, deploymentName, new.GetNamespace(), tokenName)
		}
	}

	logger.Str("namespace", new.GetNamespace()).Str("name", new.GetName()).Warn("Unable to apply the token")

	return webhook.MutationResponse{
		ValidationResponse: webhook.ValidationResponse{Allowed: true},
	}, nil
}

func (h handler) generateMutationResponse(ctx context.Context, pod *core.Pod, deploymentName, namespace, name string) (webhook.MutationResponse, error) {
	if v, ok := pod.GetLabels()[utilConstants.TokenAttached]; ok {
		if v != name {
			return webhook.MutationResponse{}, errors.Errorf("Unable to reattach Token")
		}

		return webhook.MutationResponse{
			ValidationResponse: webhook.ValidationResponse{Allowed: true},
		}, nil
	}

	token, err := h.client.Arango().PermissionV1alpha1().ArangoPermissionTokens(namespace).Get(ctx, name, meta.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return webhook.MutationResponse{}, errors.Errorf("Token '%s' not found", name)
		}

		return webhook.MutationResponse{}, errors.Wrapf(err, "Unable to get token '%s'", name)
	}

	if !token.Status.Conditions.IsTrue(permissionApi.ReadyCondition) {
		return webhook.MutationResponse{}, errors.Errorf("Token '%s' is not ready", name)
	}

	if token.Status.Deployment.GetName() != deploymentName {
		return webhook.MutationResponse{}, errors.Errorf("Token '%s' is not deployed to '%s'", name, deploymentName)
	}

	volumeName := token.Status.Secret.GetName()

	var items []patch.Item

	l := pod.GetLabels()
	if l == nil {
		l = make(map[string]string)
	}
	l[utilConstants.TokenAttached] = name

	items = append(items, patch.ItemReplace(patch.NewPath("metadata", "labels"), l))

	volume := core.Volume{
		Name: volumeName,
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName: token.Status.Secret.GetName(),
			},
		},
	}

	volumeMount := core.VolumeMount{
		Name:      volumeName,
		ReadOnly:  true,
		MountPath: utilConstants.TokenMountPath,
	}

	env := core.EnvVar{
		Name:      utilConstants.TokenEnvName,
		Value:     path.Join(utilConstants.TokenMountPath, core.ServiceAccountTokenKey),
		ValueFrom: nil,
	}

	// Volumes
	{
		if len(pod.Spec.Volumes) == 0 {
			items = append(items, patch.ItemReplace(patch.NewPath("spec", "volumes"), []core.Volume{volume}))
		} else {
			items = append(items, patch.ItemAdd(patch.NewPath("spec", "volumes", "-"), volume))
		}
	}

	for id, c := range pod.Spec.Containers {
		if len(c.VolumeMounts) == 0 {
			items = append(items, patch.ItemReplace(patch.NewPath("spec", "containers", fmt.Sprintf("%d", id), "volumeMounts"), []core.VolumeMount{volumeMount}))
		} else {
			items = append(items, patch.ItemAdd(patch.NewPath("spec", "containers", fmt.Sprintf("%d", id), "volumeMounts", "-"), volumeMount))
		}
		if len(c.Env) == 0 {
			items = append(items, patch.ItemReplace(patch.NewPath("spec", "containers", fmt.Sprintf("%d", id), "env"), []core.EnvVar{env}))
		} else {
			items = append(items, patch.ItemAdd(patch.NewPath("spec", "containers", fmt.Sprintf("%d", id), "env", "-"), env))
		}
	}

	for id, c := range pod.Spec.InitContainers {
		if len(c.VolumeMounts) == 0 {
			items = append(items, patch.ItemReplace(patch.NewPath("spec", "initContainers", fmt.Sprintf("%d", id), "volumeMounts"), []core.VolumeMount{volumeMount}))
		} else {
			items = append(items, patch.ItemAdd(patch.NewPath("spec", "initContainers", fmt.Sprintf("%d", id), "volumeMounts", "-"), volumeMount))
		}
		if len(c.Env) == 0 {
			items = append(items, patch.ItemReplace(patch.NewPath("spec", "initContainers", fmt.Sprintf("%d", id), "env"), []core.EnvVar{env}))
		} else {
			items = append(items, patch.ItemAdd(patch.NewPath("spec", "initContainers", fmt.Sprintf("%d", id), "env", "-"), env))
		}
	}

	return webhook.MutationResponse{
		ValidationResponse: webhook.ValidationResponse{Allowed: true},
		Patch:              items,
	}, nil
}
