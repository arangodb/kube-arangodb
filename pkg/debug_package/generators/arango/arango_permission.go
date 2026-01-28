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

package arango

import (
	"context"

	permissionApi "github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/list"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func Permission(f shared.FactoryGen) {
	f.AddSection("permission").
		Register("token", true, shared.WithKubernetesItems[*permissionApi.ArangoPermissionToken](arangoPermissionV1alpha1ArangoPermissionTokenList, shared.WithDefinitions[*permissionApi.ArangoPermissionToken]))
}

func arangoPermissionV1alpha1ArangoPermissionTokenList(ctx context.Context, client kclient.Client, namespace string) ([]*permissionApi.ArangoPermissionToken, error) {
	return list.ListObjects[*permissionApi.ArangoPermissionTokenList, *permissionApi.ArangoPermissionToken](ctx, client.Arango().PermissionV1alpha1().ArangoPermissionTokens(namespace), func(result *permissionApi.ArangoPermissionTokenList) []*permissionApi.ArangoPermissionToken {
		q := make([]*permissionApi.ArangoPermissionToken, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}
