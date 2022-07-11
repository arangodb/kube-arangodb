//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package sutil

import (
	"context"

	"k8s.io/apimachinery/pkg/types"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

type ACSGetter interface {
	ACS() ACS
}

type ACS interface {
	ACSItem

	Inspect(ctx context.Context, deployment *api.ArangoDeployment, client kclient.Client, cachedStatus inspectorInterface.Inspector) error

	Cluster(uid types.UID) (ACSItem, bool)
	CurrentClusterCache() inspectorInterface.Inspector
	ClusterCache(uid types.UID) (inspectorInterface.Inspector, bool)

	ForEachHealthyCluster(f func(item ACSItem) error) error

	RemoteClusters() []types.UID
}

type ACSItem interface {
	UID() types.UID
	Cache() inspectorInterface.Inspector
	Ready() bool
}
