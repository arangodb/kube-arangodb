//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package member

import (
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

const (
	recentTerminationsKeepPeriod = time.Minute * 30
)

type phaseMapFunc func(action api.Action, m *api.MemberStatus)
type phaseMapTo map[api.MemberPhase]phaseMapFunc
type phaseMap map[api.MemberPhase]phaseMapTo

type PhaseExecutor interface {
	Execute(m *api.MemberStatus, action api.Action, to api.MemberPhase) bool
}

func GetPhaseExecutor() PhaseExecutor {
	return phase
}

var phase = phaseMap{
	api.MemberPhaseNone: {
		api.MemberPhasePending: func(action api.Action, m *api.MemberStatus) {
			// Change member RID
			m.RID = uuid.NewUUID()

			// Clean Pod details
			m.PodUID = ""
		},
	},
	api.MemberPhasePending: {
		api.MemberPhaseCreated: func(action api.Action, m *api.MemberStatus) {
			// Clean conditions
			removeMemberConditionsMapFunc(m)
		},
		api.MemberPhaseUpgrading: func(action api.Action, m *api.MemberStatus) {
			removeMemberConditionsMapFunc(m)
		},
	},
}

func removeMemberConditionsMapFunc(m *api.MemberStatus) {
	// Clean conditions
	m.Conditions.Remove(api.ConditionTypeReady)
	m.Conditions.Remove(api.ConditionTypeTerminated)
	m.Conditions.Remove(api.ConditionTypeTerminating)
	m.Conditions.Remove(api.ConditionTypeAgentRecoveryNeeded)
	m.Conditions.Remove(api.ConditionTypeAutoUpgrade)
	m.Conditions.Remove(api.ConditionTypeUpgradeFailed)
	m.Conditions.Remove(api.ConditionTypePendingTLSRotation)
	m.Conditions.Remove(api.ConditionTypePendingRestart)
	m.Conditions.Remove(api.ConditionTypeRestart)
	m.Conditions.Remove(api.ConditionTypePendingUpdate)
	m.Conditions.Remove(api.ConditionTypeUpdating)
	m.Conditions.Remove(api.ConditionTypeUpdateFailed)
	m.Conditions.Remove(api.ConditionTypeCleanedOut)
	m.Conditions.Remove(api.ConditionTypeTopologyAware)
	m.Conditions.Remove(api.MemberReplacementRequired)

	m.RemoveTerminationsBefore(time.Now().Add(-1 * recentTerminationsKeepPeriod))

	m.Upgrade = false
}

func (p phaseMap) empty(action api.Action, m *api.MemberStatus) {

}

func (p phaseMap) getFunc(from, to api.MemberPhase) phaseMapFunc {
	if f, ok := p[from]; ok {
		if t, ok := f[to]; ok {
			return t
		}
	}

	return p.empty
}

func (p phaseMap) Execute(m *api.MemberStatus, action api.Action, to api.MemberPhase) bool {
	from := m.Phase

	if from == to {
		return false
	}

	f := p.getFunc(from, to)

	m.Phase = to

	f(action, m)

	return true
}
