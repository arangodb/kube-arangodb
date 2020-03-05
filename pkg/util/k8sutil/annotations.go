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

package k8sutil

import (
	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"
	policyTyped "k8s.io/client-go/kubernetes/typed/policy/v1beta1"
)

func IsChildResource(kind, name, namespace string, resource meta.Object) bool {
	if resource == nil {
		return false
	}

	if namespace != resource.GetNamespace() {
		return false
	}

	ownerRef := resource.GetOwnerReferences()

	if len(ownerRef) == 0 {
		return false
	}

	for _, owner := range ownerRef {
		if owner.Kind != kind {
			continue
		}

		if owner.Name != name {
			continue
		}

		return true
	}

	return false
}

func GetSecretsForParent(client typedCore.SecretInterface, kind, name, namespace string) ([]*core.Secret, error) {
	secrets, err := client.List(meta.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(secrets.Items) == 0 {
		return []*core.Secret{}, nil
	}

	childSecrets := make([]*core.Secret, 0, len(secrets.Items))

	for _, secret := range secrets.Items {
		if IsChildResource(kind, name, namespace, &secret) {
			childSecrets = append(childSecrets, secret.DeepCopy())
		}
	}

	return childSecrets, nil
}

func GetPDBForParent(client policyTyped.PodDisruptionBudgetInterface, kind, name, namespace string) ([]*policy.PodDisruptionBudget, error) {
	pdbs, err := client.List(meta.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(pdbs.Items) == 0 {
		return []*policy.PodDisruptionBudget{}, nil
	}

	childPdbs := make([]*policy.PodDisruptionBudget, 0, len(pdbs.Items))

	for _, pdb := range pdbs.Items {
		if IsChildResource(kind, name, namespace, &pdb) {
			childPdbs = append(childPdbs, pdb.DeepCopy())
		}
	}

	return childPdbs, nil
}

func GetPVCForParent(client typedCore.PersistentVolumeClaimInterface, kind, name, namespace string) ([]*core.PersistentVolumeClaim, error) {
	pvcs, err := client.List(meta.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(pvcs.Items) == 0 {
		return []*core.PersistentVolumeClaim{}, nil
	}

	childPvcs := make([]*core.PersistentVolumeClaim, 0, len(pvcs.Items))

	for _, pvc := range pvcs.Items {
		if IsChildResource(kind, name, namespace, &pvc) {
			childPvcs = append(childPvcs, pvc.DeepCopy())
		}
	}

	return childPvcs, nil
}

func GetServicesForParent(client typedCore.ServiceInterface, kind, name, namespace string) ([]*core.Service, error) {
	services, err := client.List(meta.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(services.Items) == 0 {
		return []*core.Service{}, nil
	}

	childServices := make([]*core.Service, 0, len(services.Items))

	for _, service := range services.Items {
		if IsChildResource(kind, name, namespace, &service) {
			childServices = append(childServices, service.DeepCopy())
		}
	}

	return childServices, nil
}

func GetServiceAccountsForParent(client typedCore.ServiceAccountInterface, kind, name, namespace string) ([]*core.ServiceAccount, error) {
	serviceAccounts, err := client.List(meta.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(serviceAccounts.Items) == 0 {
		return []*core.ServiceAccount{}, nil
	}

	childServiceAccounts := make([]*core.ServiceAccount, 0, len(serviceAccounts.Items))

	for _, serviceAccount := range serviceAccounts.Items {
		if IsChildResource(kind, name, namespace, &serviceAccount) {
			childServiceAccounts = append(childServiceAccounts, serviceAccount.DeepCopy())
		}
	}

	return childServiceAccounts, nil
}

func GetPodsForParent(client typedCore.PodInterface, kind, name, namespace string) ([]*core.Pod, error) {
	podList, err := client.List(meta.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(podList.Items) == 0 {
		return []*core.Pod{}, nil
	}

	pods := make([]*core.Pod, 0, len(podList.Items))

	for _, pod := range podList.Items {
		if IsChildResource(kind, name, namespace, &pod) {
			pods = append(pods, pod.DeepCopy())
		}
	}

	return pods, nil
}
