//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package kubernetes

import (
	"context"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/list"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func KubernetesCore(f shared.FactoryGen) {
	f.AddSection("core").
		Register("configmap", true, shared.WithKubernetesItems[*core.ConfigMap](kubernetesCoreConfigMapsList, shared.WithDefinitions[*core.ConfigMap])).
		Register("event", true, shared.WithKubernetesItems[*core.Event](kubernetesCoreEventsList)).
		Register("pod", true, shared.WithKubernetesItems[*core.Pod](kubernetesCorePodsList, shared.WithDefinitions[*core.Pod], kubernetesCorePodLogs)).
		Register("secret", true, shared.WithKubernetesItems[*core.Secret](kubernetesCoreSecretList, shared.WithModification[*core.Secret](shared.WithDefinitions[*core.Secret], kubernetesCoreSecretModHideSensitiveData))).
		Register("service", true, shared.WithKubernetesItems[*core.Service](kubernetesCoreServiceList, shared.WithDefinitions[*core.Service], kubernetesCoreServiceEndpoints))
}

func kubernetesCoreConfigMapsList(ctx context.Context, client kclient.Client, namespace string) ([]*core.ConfigMap, error) {
	return list.ListObjects[*core.ConfigMapList, *core.ConfigMap](ctx, client.Kubernetes().CoreV1().ConfigMaps(namespace), func(result *core.ConfigMapList) []*core.ConfigMap {
		q := make([]*core.ConfigMap, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}

func kubernetesCoreEventsList(ctx context.Context, client kclient.Client, namespace string) ([]*core.Event, error) {
	return list.ListObjects[*core.EventList, *core.Event](ctx, client.Kubernetes().CoreV1().Events(namespace), func(result *core.EventList) []*core.Event {
		q := make([]*core.Event, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}

func kubernetesCorePodsList(ctx context.Context, client kclient.Client, namespace string) ([]*core.Pod, error) {
	return list.ListObjects[*core.PodList, *core.Pod](ctx, client.Kubernetes().CoreV1().Pods(namespace), func(result *core.PodList) []*core.Pod {
		q := make([]*core.Pod, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}

func kubernetesCoreSecretList(ctx context.Context, client kclient.Client, namespace string) ([]*core.Secret, error) {
	return list.ListObjects[*core.SecretList, *core.Secret](ctx, client.Kubernetes().CoreV1().Secrets(namespace), func(result *core.SecretList) []*core.Secret {
		q := make([]*core.Secret, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}

func kubernetesCoreServiceList(ctx context.Context, client kclient.Client, namespace string) ([]*core.Service, error) {
	return list.ListObjects[*core.ServiceList, *core.Service](ctx, client.Kubernetes().CoreV1().Services(namespace), func(result *core.ServiceList) []*core.Service {
		q := make([]*core.Service, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}
