//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package arango

import (
	"context"

	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/list"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func Platform(f shared.FactoryGen) {
	f.AddSection("platform").
		Register("storage", true, shared.WithKubernetesItems[*platformApi.ArangoPlatformStorage](arangoPlatformV1Alpha1ArangoPlatformStorageList, shared.WithDefinitions[*platformApi.ArangoPlatformStorage], arangoPlatformV1Alpha1ArangoPlatformStorageDebug)).
		Register("chart", true, shared.WithKubernetesItems[*platformApi.ArangoPlatformChart](arangoPlatformV1Alpha1ArangoPlatformChartList, shared.WithDefinitions[*platformApi.ArangoPlatformChart], arangoPlatformV1Alpha1ArangoPlatformChartExtract)).
		Register("service", true, shared.WithKubernetesItems[*platformApi.ArangoPlatformService](arangoPlatformV1Alpha1ArangoPlatformServiceList, shared.WithDefinitions[*platformApi.ArangoPlatformService]))
}

func arangoPlatformV1Alpha1ArangoPlatformStorageList(ctx context.Context, client kclient.Client, namespace string) ([]*platformApi.ArangoPlatformStorage, error) {
	return list.ListObjects[*platformApi.ArangoPlatformStorageList, *platformApi.ArangoPlatformStorage](ctx, client.Arango().PlatformV1alpha1().ArangoPlatformStorages(namespace), func(result *platformApi.ArangoPlatformStorageList) []*platformApi.ArangoPlatformStorage {
		q := make([]*platformApi.ArangoPlatformStorage, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}

func arangoPlatformV1Alpha1ArangoPlatformChartList(ctx context.Context, client kclient.Client, namespace string) ([]*platformApi.ArangoPlatformChart, error) {
	return list.ListObjects[*platformApi.ArangoPlatformChartList, *platformApi.ArangoPlatformChart](ctx, client.Arango().PlatformV1alpha1().ArangoPlatformCharts(namespace), func(result *platformApi.ArangoPlatformChartList) []*platformApi.ArangoPlatformChart {
		q := make([]*platformApi.ArangoPlatformChart, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}

func arangoPlatformV1Alpha1ArangoPlatformServiceList(ctx context.Context, client kclient.Client, namespace string) ([]*platformApi.ArangoPlatformService, error) {
	return list.ListObjects[*platformApi.ArangoPlatformServiceList, *platformApi.ArangoPlatformService](ctx, client.Arango().PlatformV1alpha1().ArangoPlatformServices(namespace), func(result *platformApi.ArangoPlatformServiceList) []*platformApi.ArangoPlatformService {
		q := make([]*platformApi.ArangoPlatformService, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}
