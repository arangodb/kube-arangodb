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

	apps "k8s.io/api/apps/v1"

	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/list"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func KubernetesApps(f shared.FactoryGen) {
	f.AddSection("apps").
		Register("replicaset", true, shared.WithKubernetesItems[*apps.ReplicaSet](kubernetesAppsReplicaSetsList, shared.WithDefinitions[*apps.ReplicaSet])).
		Register("deployment", true, shared.WithKubernetesItems[*apps.Deployment](kubernetesAppsDeploymentsList, shared.WithDefinitions[*apps.Deployment])).
		Register("statefulset", true, shared.WithKubernetesItems[*apps.StatefulSet](kubernetesAppsStatefulSetsList, shared.WithDefinitions[*apps.StatefulSet]))
}

func kubernetesAppsReplicaSetsList(ctx context.Context, client kclient.Client, namespace string) ([]*apps.ReplicaSet, error) {
	return list.ListObjects[*apps.ReplicaSetList, *apps.ReplicaSet](ctx, client.Kubernetes().AppsV1().ReplicaSets(namespace), func(result *apps.ReplicaSetList) []*apps.ReplicaSet {
		q := make([]*apps.ReplicaSet, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}

func kubernetesAppsDeploymentsList(ctx context.Context, client kclient.Client, namespace string) ([]*apps.Deployment, error) {
	return list.ListObjects[*apps.DeploymentList, *apps.Deployment](ctx, client.Kubernetes().AppsV1().Deployments(namespace), func(result *apps.DeploymentList) []*apps.Deployment {
		q := make([]*apps.Deployment, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}

func kubernetesAppsStatefulSetsList(ctx context.Context, client kclient.Client, namespace string) ([]*apps.StatefulSet, error) {
	return list.ListObjects[*apps.StatefulSetList, *apps.StatefulSet](ctx, client.Kubernetes().AppsV1().StatefulSets(namespace), func(result *apps.StatefulSetList) []*apps.StatefulSet {
		q := make([]*apps.StatefulSet, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}
