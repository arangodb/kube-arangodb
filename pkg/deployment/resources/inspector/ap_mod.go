//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/mods"
)

func (i *inspectorState) ArangoProfileModInterface() mods.ArangoProfileMods {
	return arangoProfileMod{
		i: i,
	}
}

type arangoProfileMod struct {
	i *inspectorState
}

func (p arangoProfileMod) V1Beta1() generic.ModStatusClient[*schedulerApi.ArangoProfile] {
	return wrapMod[*schedulerApi.ArangoProfile](definitions.ArangoProfile, p.i.GetThrottles, p.clientv1beta1)
}

func (p arangoProfileMod) clientv1beta1() generic.ModStatusClient[*schedulerApi.ArangoProfile] {
	return p.i.Client().Arango().SchedulerV1beta1().ArangoProfiles(p.i.Namespace())
}
