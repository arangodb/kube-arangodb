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
	serviceaccountv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/serviceaccount/v1"
)

func (i *inspectorState) ServiceAccountsModInterface() mods.ServiceAccountsMods {
	return serviceAccountsMod{
		i: i,
	}
}

type serviceAccountsMod struct {
	i *inspectorState
}

func (p serviceAccountsMod) V1() serviceaccountv1.ModInterface {
	return wrapMod[*core.ServiceAccount](definitions.ServiceAccount, p.i.GetThrottles, generic.WithModStatusGetter[*core.ServiceAccount](constants.ServiceAccountGKv1(), p.clientv1))
}

func (p serviceAccountsMod) clientv1() generic.ModClient[*core.ServiceAccount] {
	return p.i.Client().Kubernetes().CoreV1().ServiceAccounts(p.i.Namespace())
}
