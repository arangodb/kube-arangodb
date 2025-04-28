//
// DISCLAIMER
//
// Copyright 2023-2025 ArangoDB GmbH, Cologne, Germany
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

	analyticsApi "github.com/arangodb/kube-arangodb/pkg/apis/analytics/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/list"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func Analytics(f shared.FactoryGen) {
	f.AddSection("analytics").
		Register("gae", true, shared.WithKubernetesItems[*analyticsApi.GraphAnalyticsEngine](arangoAnalyticsV1Alpha1GraphAnalyticsEngineList, shared.WithDefinitions[*analyticsApi.GraphAnalyticsEngine]))
}

func arangoAnalyticsV1Alpha1GraphAnalyticsEngineList(ctx context.Context, client kclient.Client, namespace string) ([]*analyticsApi.GraphAnalyticsEngine, error) {
	return list.ListObjects[*analyticsApi.GraphAnalyticsEngineList, *analyticsApi.GraphAnalyticsEngine](ctx, client.Arango().AnalyticsV1alpha1().GraphAnalyticsEngines(namespace), func(result *analyticsApi.GraphAnalyticsEngineList) []*analyticsApi.GraphAnalyticsEngine {
		q := make([]*analyticsApi.GraphAnalyticsEngine, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}
