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

package inspector

import (
	discovery "k8s.io/api/discovery/v1"

	inspectorConstants "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/mods"
)

func (i *inspectorState) EndpointSlicesModInterface() mods.EndpointSlicesMods {
	return endpointsMod{
		i: i,
	}
}

type endpointsMod struct {
	i *inspectorState
}

func (p endpointsMod) V1() generic.ModClient[*discovery.EndpointSlice] {
	return wrapMod[*discovery.EndpointSlice](definitions.EndpointSlices, p.i.GetThrottles, generic.WithModStatusGetter[*discovery.EndpointSlice](inspectorConstants.EndpointSlicesGKv1(), p.clientv1))
}

func (p endpointsMod) clientv1() generic.ModClient[*discovery.EndpointSlice] {
	return p.i.Client().Kubernetes().DiscoveryV1().EndpointSlices(p.i.Namespace())
}
