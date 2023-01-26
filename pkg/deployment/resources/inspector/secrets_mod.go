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
	secretv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret/v1"
)

func (i *inspectorState) SecretsModInterface() mods.SecretsMods {
	return secretsMod{
		i: i,
	}
}

type secretsMod struct {
	i *inspectorState
}

func (p secretsMod) V1() secretv1.ModInterface {
	return wrapMod[*core.Secret](definitions.Secret, p.i.GetThrottles, generic.WithModStatusGetter[*core.Secret](constants.SecretGKv1(), p.clientv1))
}

func (p secretsMod) clientv1() generic.ModClient[*core.Secret] {
	return p.i.Client().Kubernetes().CoreV1().Secrets(p.i.Namespace())
}
