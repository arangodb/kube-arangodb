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
	arangomemberv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangomember/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/mods"
)

func (i *inspectorState) ArangoMemberModInterface() mods.ArangoMemberMods {
	return arangoMemberMod{
		i: i,
	}
}

type arangoMemberMod struct {
	i *inspectorState
}

func (p arangoMemberMod) V1() arangomemberv1.ModInterface {
	return wrapMod[*api.ArangoMember](definitions.ArangoMember, p.i.GetThrottles, p.clientv1)
}

func (p arangoMemberMod) clientv1() generic.ModStatusClient[*api.ArangoMember] {
	return p.i.Client().Arango().DatabaseV1().ArangoMembers(p.i.Namespace())
}
