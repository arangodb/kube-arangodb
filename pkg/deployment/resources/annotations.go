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
	"time"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/backup/utils"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/rs/zerolog/log"
	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"
	policyTyped "k8s.io/client-go/kubernetes/typed/policy/v1beta1"
)

func (r *Resources) EnsureAnnotations(cachedStatus inspector.Inspector) error {
	kubecli := r.context.GetKubeCli()

	log.Info().Msgf("Ensuring annotations")

	if err := ensureSecretsAnnotations(kubecli.CoreV1().Secrets(r.context.GetNamespace()),
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec().Annotations); err != nil {
		return err
	}

	if err := ensureServiceAccountsAnnotations(kubecli.CoreV1().ServiceAccounts(r.context.GetNamespace()),
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec().Annotations); err != nil {
		return err
	}

	if err := ensureServicesAnnotations(kubecli.CoreV1().Services(r.context.GetNamespace()),
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec().Annotations); err != nil {
		return err
	}

	if err := ensurePdbsAnnotations(kubecli.PolicyV1beta1().PodDisruptionBudgets(r.context.GetNamespace()),
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec().Annotations); err != nil {
		return err
	}

	if err := ensurePvcsAnnotations(kubecli.CoreV1().PersistentVolumeClaims(r.context.GetNamespace()),
		cachedStatus,
		deployment.ArangoDeploymentResourceKind,
		r.context.GetAPIObject().GetName(),
		r.context.GetAPIObject().GetNamespace(),
		r.context.GetSpec().Annotations); err != nil {
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

	return nil
}

func ensureSecretsAnnotations(client typedCore.SecretInterface, cachedStatus inspector.Inspector, kind, name, namespace string, annotations map[string]string) error {
	if err := cachedStatus.IterateSecrets(func(secret *core.Secret) error {
		if !k8sutil.CompareAnnotations(secret.GetAnnotations(), annotations) {
			log.Info().Msgf("Replacing annotations for Secret %s", secret.Name)
			if err := setSecretAnnotations(client, secret, annotations); err != nil {
				return err
			}
		}

		return nil
	}, func(secret *core.Secret) bool {
		return k8sutil.IsChildResource(kind, name, namespace, secret)
	}); err != nil {
		return err
	}

	return nil
}

func setSecretAnnotations(client typedCore.SecretInterface, secret *core.Secret, annotations map[string]string) error {
	return utils.Retry(5, 200*time.Millisecond, func() error {
		currentSecret, err := client.Get(secret.Name, meta.GetOptions{})
		if err != nil {
			return err
		}

		currentSecret.Annotations = k8sutil.MergeAnnotations(k8sutil.GetSecuredAnnotations(currentSecret.Annotations), annotations)

		_, err = client.Update(currentSecret)
		if err != nil {
			return err
		}

		return nil
	})
}

func ensureServiceAccountsAnnotations(client typedCore.ServiceAccountInterface, kind, name, namespace string, annotations map[string]string) error {
	serviceAccounts, err := k8sutil.GetServiceAccountsForParent(client,
		kind,
		name,
		namespace)
	if err != nil {
		return err
	}

	for _, serviceAccount := range serviceAccounts {
		if !k8sutil.CompareAnnotations(serviceAccount.GetAnnotations(), annotations) {
			log.Info().Msgf("Replacing annotations for ServiceAccount %s", serviceAccount.Name)
			if err = setServiceAccountAnnotations(client, serviceAccount, annotations); err != nil {
				return err
			}
		}
	}

	return nil
}

func setServiceAccountAnnotations(client typedCore.ServiceAccountInterface, serviceAccount *core.ServiceAccount, annotations map[string]string) error {
	return utils.Retry(5, 200*time.Millisecond, func() error {
		currentServiceAccount, err := client.Get(serviceAccount.Name, meta.GetOptions{})
		if err != nil {
			return err
		}

		currentServiceAccount.Annotations = k8sutil.MergeAnnotations(k8sutil.GetSecuredAnnotations(currentServiceAccount.Annotations), annotations)

		_, err = client.Update(currentServiceAccount)
		if err != nil {
			return err
		}

		return nil
	})
}

func ensureServicesAnnotations(client typedCore.ServiceInterface, cachedStatus inspector.Inspector, kind, name, namespace string, annotations map[string]string) error {
	if err := cachedStatus.IterateServices(func(service *core.Service) error {
		if !k8sutil.CompareAnnotations(service.GetAnnotations(), annotations) {
			log.Info().Msgf("Replacing annotations for Service %s", service.Name)
			if err := setServiceAnnotations(client, service, annotations); err != nil {
				return err
			}
		}

		return nil
	}, func(service *core.Service) bool {
		return k8sutil.IsChildResource(kind, name, namespace, service)
	}); err != nil {
		return err
	}

	return nil
}

func setServiceAnnotations(client typedCore.ServiceInterface, service *core.Service, annotations map[string]string) error {
	return utils.Retry(5, 200*time.Millisecond, func() error {
		currentService, err := client.Get(service.Name, meta.GetOptions{})
		if err != nil {
			return err
		}

		currentService.Annotations = k8sutil.MergeAnnotations(k8sutil.GetSecuredAnnotations(currentService.Annotations), annotations)

		_, err = client.Update(currentService)
		if err != nil {
			return err
		}

		return nil
	})
}

func ensurePdbsAnnotations(client policyTyped.PodDisruptionBudgetInterface, kind, name, namespace string, annotations map[string]string) error {
	podDisruptionBudgets, err := k8sutil.GetPDBForParent(client,
		kind,
		name,
		namespace)
	if err != nil {
		return err
	}

	for _, podDisruptionBudget := range podDisruptionBudgets {
		if !k8sutil.CompareAnnotations(podDisruptionBudget.GetAnnotations(), annotations) {
			log.Info().Msgf("Replacing annotations for PDB %s", podDisruptionBudget.Name)
			if err = setPdbAnnotations(client, podDisruptionBudget, annotations); err != nil {
				return err
			}
		}
	}

	return nil
}

func setPdbAnnotations(client policyTyped.PodDisruptionBudgetInterface, podDisruptionBudget *policy.PodDisruptionBudget, annotations map[string]string) error {
	return utils.Retry(5, 200*time.Millisecond, func() error {
		currentPodDistributionBudget, err := client.Get(podDisruptionBudget.Name, meta.GetOptions{})
		if err != nil {
			return err
		}

		currentPodDistributionBudget.Annotations = k8sutil.MergeAnnotations(k8sutil.GetSecuredAnnotations(currentPodDistributionBudget.Annotations), annotations)

		_, err = client.Update(currentPodDistributionBudget)
		if err != nil {
			return err
		}

		return nil
	})
}

func ensurePvcsAnnotations(client typedCore.PersistentVolumeClaimInterface, cachedStatus inspector.Inspector, kind, name, namespace string, annotations map[string]string) error {
	if err := cachedStatus.IteratePersistentVolumeClaims(func(persistentVolumeClaim *core.PersistentVolumeClaim) error {
		if !k8sutil.CompareAnnotations(persistentVolumeClaim.GetAnnotations(), annotations) {
			log.Info().Msgf("Replacing annotations for PVC %s", persistentVolumeClaim.Name)
			if err := setPvcAnnotations(client, persistentVolumeClaim, annotations); err != nil {
				return err
			}
		}

		return nil
	}, func(persistentVolumeClaim *core.PersistentVolumeClaim) bool {
		return k8sutil.IsChildResource(kind, name, namespace, persistentVolumeClaim)
	}); err != nil {
		return err
	}

	return nil
}

func setPvcAnnotations(client typedCore.PersistentVolumeClaimInterface, persistentVolumeClaim *core.PersistentVolumeClaim, annotations map[string]string) error {
	return utils.Retry(5, 200*time.Millisecond, func() error {
		currentVolumeClaim, err := client.Get(persistentVolumeClaim.Name, meta.GetOptions{})
		if err != nil {
			return err
		}

		currentVolumeClaim.Annotations = k8sutil.MergeAnnotations(k8sutil.GetSecuredAnnotations(currentVolumeClaim.Annotations), annotations)

		_, err = client.Update(currentVolumeClaim)
		if err != nil {
			return err
		}

		return nil
	})
}

func getPodGroup(pod *core.Pod) api.ServerGroup {
	if pod.Labels == nil {
		return api.ServerGroupUnknown
	}

	return api.ServerGroupFromRole(pod.Labels[k8sutil.LabelKeyRole])
}

func ensurePodsAnnotations(client typedCore.PodInterface, cachedStatus inspector.Inspector, kind, name, namespace string, annotations map[string]string, spec api.DeploymentSpec) error {
	if err := cachedStatus.IteratePods(func(pod *core.Pod) error {
		group := getPodGroup(pod)
		mergedAnnotations := k8sutil.MergeAnnotations(annotations, spec.GetServerGroupSpec(group).Annotations)

		if !k8sutil.CompareAnnotations(pod.GetAnnotations(), mergedAnnotations) {
			log.Info().Msgf("Replacing annotations for Pod %s", pod.Name)
			if err := setPodAnnotations(client, pod, mergedAnnotations); err != nil {
				return err
			}
		}

		return nil
	}, func(pod *core.Pod) bool {
		return k8sutil.IsChildResource(kind, name, namespace, pod)
	}); err != nil {
		return err
	}

	return nil
}

func setPodAnnotations(client typedCore.PodInterface, pod *core.Pod, annotations map[string]string) error {
	return utils.Retry(5, 200*time.Millisecond, func() error {
		currentPod, err := client.Get(pod.Name, meta.GetOptions{})
		if err != nil {
			return err
		}

		currentPod.Annotations = k8sutil.MergeAnnotations(k8sutil.GetSecuredAnnotations(currentPod.Annotations), annotations)

		_, err = client.Update(currentPod)
		if err != nil {
			return err
		}

		return nil
	})
}
