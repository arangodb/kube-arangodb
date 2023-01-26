//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/mods"
	servicev1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service/v1"
)

func (i *inspectorState) ServicesModInterface() mods.ServicesMods {
	return servicesMod{
		i: i,
	}
}

type servicesMod struct {
	i *inspectorState
}

func (p servicesMod) V1() servicev1.ModInterface {
	return wrapMod[*core.Service](definitions.Service, p.i.GetThrottles, generic.WithModStatusGetter[*core.Service](constants.ServiceGKv1(), p.clientv1))
}

func (p servicesMod) clientv1() generic.ModClient[*core.Service] {
	return p.i.Client().Kubernetes().CoreV1().Services(p.i.Namespace())
}
