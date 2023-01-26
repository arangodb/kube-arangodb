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
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	arangotaskv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangotask/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/mods"
)

func (i *inspectorState) ArangoTaskModInterface() mods.ArangoTaskMods {
	return arangoTaskMod{
		i: i,
	}
}

type arangoTaskMod struct {
	i *inspectorState
}

func (p arangoTaskMod) V1() arangotaskv1.ModInterface {
	return wrapMod[*api.ArangoTask](definitions.ArangoTask, p.i.GetThrottles, p.clientv1)
}

func (p arangoTaskMod) clientv1() generic.ModStatusClient[*api.ArangoTask] {
	return p.i.Client().Arango().DatabaseV1().ArangoTasks(p.i.Namespace())
}
