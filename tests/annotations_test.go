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

package tests

import (
	"github.com/arangodb/kube-arangodb/pkg/util/collection"
	"testing"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/dchest/uniuri"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/client"
)

func addAnnotation(t *testing.T, kubeClient kubernetes.Interface, arangoClient versioned.Interface, depl *api.ArangoDeployment, annotations map[string]string) {
	object, err := arangoClient.DatabaseV1().ArangoDeployments(depl.GetNamespace()).Get(depl.GetName(), meta.GetOptions{})
	require.NoError(t, err)

	object.Spec.Annotations = annotations
	object.Spec.Coordinators.Annotations = depl.Spec.Coordinators.Annotations

	_, err = arangoClient.DatabaseV1().ArangoDeployments(depl.GetNamespace()).Update(object)
	require.NoError(t, err)

	ensureAnnotations(t, kubeClient, object)
}

func ensureAnnotationsTimeout(t *testing.T, client kubernetes.Interface, depl *api.ArangoDeployment) func() error {
	return func() error {
		if err := ensureSecretAnnotations(t, client, depl); err == nil || !isInterrupt(err) {
			return err
		}

		if err := ensurePDBAnnotation(t, client, depl); err == nil || !isInterrupt(err) {
			return err
		}

		if err := ensurePVCAnnotation(t, client, depl); err == nil || !isInterrupt(err) {
			return err
		}

		if err := ensureServiceAnnotation(t, client, depl); err == nil || !isInterrupt(err) {
			return err
		}

		if err := ensureServiceAccountAnnotation(t, client, depl); err == nil || !isInterrupt(err) {
			return err
		}

		if err := ensurePodAnnotations(t, client, depl); err == nil || !isInterrupt(err) {
			return err
		}

		return interrupt{}
	}
}

func ensureSecretAnnotations(t *testing.T, client kubernetes.Interface, depl *api.ArangoDeployment) error {
	secrets, err := k8sutil.GetSecretsForParent(client.CoreV1().Secrets(depl.Namespace), deployment.ArangoDeploymentResourceKind, depl.Name, depl.Namespace)
	require.NoError(t, err)
	require.True(t, len(secrets) > 0)
	for _, secret := range secrets {
		if !collection.Compare(secret.GetAnnotations(), depl.Spec.Annotations) {
			log.Info().Msgf("Annotations for Secret does not match on %s", secret.Name)
			return nil
		}
	}

	return interrupt{}
}

func getPodGroup(pod *core.Pod) api.ServerGroup {
	if pod.Labels == nil {
		return api.ServerGroupUnknown
	}

	return api.ServerGroupFromRole(pod.Labels[k8sutil.LabelKeyRole])
}

func ensurePodAnnotations(t *testing.T, client kubernetes.Interface, depl *api.ArangoDeployment) error {
	pods, err := k8sutil.GetPodsForParent(client.CoreV1().Pods(depl.Namespace), deployment.ArangoDeploymentResourceKind, depl.Name, depl.Namespace)
	require.NoError(t, err)
	require.True(t, len(pods) > 0)
	for _, pod := range pods {
		group := getPodGroup(pod)
		combinedAnnotations := collection.MergeAnnotations(depl.Spec.Annotations, depl.Spec.GetServerGroupSpec(group).Annotations)
		if !collection.Compare(pod.GetAnnotations(), combinedAnnotations) {
			log.Info().Msgf("Annotations for Pod does not match on %s", pod.Name)
			return nil
		}
	}

	return interrupt{}
}

func ensurePDBAnnotation(t *testing.T, client kubernetes.Interface, depl *api.ArangoDeployment) error {
	podDisruptionBudgets, err := k8sutil.GetPDBForParent(client.PolicyV1beta1().PodDisruptionBudgets(depl.Namespace), deployment.ArangoDeploymentResourceKind, depl.Name, depl.Namespace)
	require.NoError(t, err)
	require.True(t, len(podDisruptionBudgets) > 0)
	for _, podDisruptionBudget := range podDisruptionBudgets {
		if !collection.Compare(podDisruptionBudget.GetAnnotations(), depl.Spec.Annotations) {
			log.Info().Msgf("Annotations for PDB does not match on %s", podDisruptionBudget.Name)
			return nil
		}
	}

	return interrupt{}
}

func ensurePVCAnnotation(t *testing.T, client kubernetes.Interface, depl *api.ArangoDeployment) error {
	persistentVolumeClaims, err := k8sutil.GetPVCForParent(client.CoreV1().PersistentVolumeClaims(depl.Namespace), deployment.ArangoDeploymentResourceKind, depl.Name, depl.Namespace)
	require.NoError(t, err)
	require.True(t, len(persistentVolumeClaims) > 0)
	for _, persistentVolumeClaim := range persistentVolumeClaims {
		if !collection.Compare(persistentVolumeClaim.GetAnnotations(), depl.Spec.Annotations) {
			log.Info().Msgf("Annotations for PVC does not match on %s", persistentVolumeClaim.Name)
			return nil
		}
	}

	return interrupt{}
}

func ensureServiceAnnotation(t *testing.T, client kubernetes.Interface, depl *api.ArangoDeployment) error {
	services, err := k8sutil.GetServicesForParent(client.CoreV1().Services(depl.Namespace), deployment.ArangoDeploymentResourceKind, depl.Name, depl.Namespace)
	require.NoError(t, err)
	require.True(t, len(services) > 0)
	for _, service := range services {
		if !collection.Compare(service.GetAnnotations(), depl.Spec.Annotations) {
			log.Info().Msgf("Annotations for Service does not match on %s", service.Name)
			return nil
		}
	}

	return interrupt{}
}

func ensureServiceAccountAnnotation(t *testing.T, client kubernetes.Interface, depl *api.ArangoDeployment) error {
	serviceAccounts, err := k8sutil.GetServiceAccountsForParent(client.CoreV1().ServiceAccounts(depl.Namespace), deployment.ArangoDeploymentResourceKind, depl.Name, depl.Namespace)
	require.NoError(t, err)
	for _, serviceAccount := range serviceAccounts {
		if !collection.Compare(serviceAccount.GetAnnotations(), depl.Spec.Annotations) {
			log.Info().Msgf("Annotations for Service Account does not match on %s", serviceAccount.Name)
			return nil
		}
	}

	return interrupt{}
}

func ensureAnnotations(t *testing.T, client kubernetes.Interface, depl *api.ArangoDeployment) {
	if err := timeout(2*time.Second, 5*time.Minute, ensureAnnotationsTimeout(t, client, depl)); err != nil {
		require.NoError(t, err)
	}
}

func TestAnnotations(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewClient()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-annotations-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.Environment = api.NewEnvironment(api.EnvironmentProduction)
	depl.Spec.SetDefaults(depl.GetName())

	// Create deployment
	depl, err := c.DatabaseV1().ArangoDeployments(ns).Create(depl)
	require.NoError(t, err)

	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	if _, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady()); err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	t.Run("Add annotation", func(t *testing.T) {
		annotations := map[string]string{
			"annotation": uniuri.NewLen(8),
		}

		addAnnotation(t, kubecli, c, depl, annotations)

		addAnnotation(t, kubecli, c, depl, nil)
	})

	t.Run("Add kubernetes annotation", func(t *testing.T) {
		key := "kubernetes.io/test-only-annotation"

		annotations := map[string]string{
			key:          uniuri.NewLen(8),
			"annotation": uniuri.NewLen(8),
		}

		addAnnotation(t, kubecli, c, depl, annotations)

		addAnnotation(t, kubecli, c, depl, nil)

		secrets, err := k8sutil.GetSecretsForParent(kubecli.CoreV1().Secrets(depl.Namespace),
			deployment.ArangoDeploymentResourceKind,
			depl.Name,
			depl.Namespace)
		require.NoError(t, err)
		require.True(t, len(secrets) > 0)

		for _, secret := range secrets {
			require.NotNil(t, secret.Annotations)

			_, ok := secret.Annotations[key]

			require.True(t, ok)
		}
	})

	t.Run("Add arangodb annotation", func(t *testing.T) {
		key := "arangodb.com/test-only-annotation"

		annotations := map[string]string{
			key:          uniuri.NewLen(8),
			"annotation": uniuri.NewLen(8),
		}

		addAnnotation(t, kubecli, c, depl, annotations)

		addAnnotation(t, kubecli, c, depl, nil)

		secrets, err := k8sutil.GetSecretsForParent(kubecli.CoreV1().Secrets(depl.Namespace),
			deployment.ArangoDeploymentResourceKind,
			depl.Name,
			depl.Namespace)
		require.NoError(t, err)
		require.True(t, len(secrets) > 0)

		for _, secret := range secrets {
			require.NotNil(t, secret.Annotations)

			_, ok := secret.Annotations[key]

			require.True(t, ok)
		}
	})

	t.Run("Replace annotation", func(t *testing.T) {
		annotations := map[string]string{
			"annotation": uniuri.NewLen(8),
		}

		addAnnotation(t, kubecli, c, depl, annotations)

		annotations["annotation"] = uniuri.NewLen(16)

		addAnnotation(t, kubecli, c, depl, annotations)

		addAnnotation(t, kubecli, c, depl, nil)
	})

	t.Run("Add annotations", func(t *testing.T) {
		annotations := map[string]string{
			"annotation":  uniuri.NewLen(8),
			"annotation2": uniuri.NewLen(16),
		}

		addAnnotation(t, kubecli, c, depl, annotations)

		addAnnotation(t, kubecli, c, depl, nil)
	})

	t.Run("Add annotations for group", func(t *testing.T) {
		annotations := map[string]string{
			"annotation":  uniuri.NewLen(8),
			"annotation2": uniuri.NewLen(16),
		}

		depl.Spec.Coordinators.Annotations = map[string]string{
			"coordinator-only": uniuri.NewLen(32),
			"annotation":       uniuri.NewLen(8),
		}

		addAnnotation(t, kubecli, c, depl, annotations)

		pods, err := k8sutil.GetPodsForParent(kubecli.CoreV1().Pods(depl.Namespace),
			deployment.ArangoDeploymentResourceKind,
			depl.Name,
			depl.Namespace)
		require.NoError(t, err)
		require.True(t, len(pods) > 0)

		for _, pod := range pods {
			require.NotNil(t, pod.Annotations)

			value, ok := pod.Annotations["annotation"]
			_, coordOnly := pod.Annotations["coordinator-only"]

			require.True(t, ok)

			if getPodGroup(pod) == api.ServerGroupCoordinators {
				require.Equal(t, depl.Spec.Coordinators.Annotations["annotation"], value)
				require.True(t, coordOnly)
			} else {
				require.Equal(t, annotations["annotation"], value)
				require.False(t, coordOnly)
			}
		}

		depl.Spec.Coordinators.Annotations = nil

		addAnnotation(t, kubecli, c, depl, nil)
	})
}
