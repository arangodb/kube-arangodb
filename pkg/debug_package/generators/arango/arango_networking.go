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

	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/list"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func Networking(f shared.FactoryGen) {
	f.AddSection("networking").
		Register("route", true, shared.WithKubernetesItems[*networkingApi.ArangoRoute](arangoNetworkingV1beta1ArangoRouteList, shared.WithDefinitions[*networkingApi.ArangoRoute]))
}

func arangoNetworkingV1beta1ArangoRouteList(ctx context.Context, client kclient.Client, namespace string) ([]*networkingApi.ArangoRoute, error) {
	return list.ListObjects[*networkingApi.ArangoRouteList, *networkingApi.ArangoRoute](ctx, client.Arango().NetworkingV1beta1().ArangoRoutes(namespace), func(result *networkingApi.ArangoRouteList) []*networkingApi.ArangoRoute {
		q := make([]*networkingApi.ArangoRoute, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}
