//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/mods"
)

func (i *inspectorState) ArangoClusterSynchronizationModInterface() mods.ArangoClusterSynchronizationMods {
	return &arangoClusterSynchronizationMod{
		i: i,
	}
}

type arangoClusterSynchronizationMod struct {
	i *inspectorState
}

func (p arangoClusterSynchronizationMod) V1() generic.ModStatusClient[*api.ArangoClusterSynchronization] {
	return wrapMod[*api.ArangoClusterSynchronization](definitions.ArangoClusterSynchronization, p.i.GetThrottles, p.clientv1)
}

func (p arangoClusterSynchronizationMod) clientv1() generic.ModStatusClient[*api.ArangoClusterSynchronization] {
	return p.i.Client().Arango().DatabaseV1().ArangoClusterSynchronizations(p.i.Namespace())
}
