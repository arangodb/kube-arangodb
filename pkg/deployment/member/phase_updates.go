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

package member

import (
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

const (
	recentTerminationsKeepPeriod = time.Minute * 30
)

type phaseMapFunc func(obj meta.Object, spec api.DeploymentSpec, group api.ServerGroup, action api.Action, m *api.MemberStatus)
type phaseMapTo map[api.MemberPhase]phaseMapFunc
type phaseMap map[api.MemberPhase]phaseMapTo

type PhaseExecutor interface {
	Execute(obj meta.Object, spec api.DeploymentSpec, group api.ServerGroup, m *api.MemberStatus, action api.Action, to api.MemberPhase) bool
}

func GetPhaseExecutor() PhaseExecutor {
	return phase
}

var phase = phaseMap{
	api.MemberPhaseNone: {
		api.MemberPhasePending: func(obj meta.Object, spec api.DeploymentSpec, group api.ServerGroup, action api.Action, m *api.MemberStatus) {
			// Change member RID
			m.RID = uuid.NewUUID()

			// Clean Pod details
			if m.Pod != nil {
				m.Pod.UID = ""
			}
			m.Pod.Propagate(m)

			// Add ClusterID
			if m.ClusterID == "" {
				m.ClusterID = obj.GetUID()
			}

			if m.Architecture == nil {
				d := spec.Architecture.GetDefault()
				m.Architecture = &d
			}
		},
	},
	api.MemberPhasePending: {
		api.MemberPhaseCreated: func(obj meta.Object, spec api.DeploymentSpec, group api.ServerGroup, action api.Action, m *api.MemberStatus) {
			// Clean conditions
			removeMemberConditionsMapFunc(m)
		},
		api.MemberPhaseUpgrading: func(obj meta.Object, spec api.DeploymentSpec, group api.ServerGroup, action api.Action, m *api.MemberStatus) {
			removeMemberConditionsMapFunc(m)
		},
	},
}

func removeMemberConditionsMapFunc(m *api.MemberStatus) {
	// Clean conditions
	m.Conditions.Remove(api.ConditionTypeReady)
	m.Conditions.Remove(api.ConditionTypeActive)
	m.Conditions.Remove(api.ConditionTypeStarted)
	m.Conditions.Remove(api.ConditionTypeScheduled)
	m.Conditions.Remove(api.ConditionTypeReachable)
	m.Conditions.Remove(api.ConditionTypeServing)
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
	m.Conditions.Remove(api.ConditionTypePVCResizePending)
	m.Conditions.Remove(api.ConditionTypeArchitectureMismatch)
	m.Conditions.Remove(api.ConditionTypeArchitectureChangeCannotBeApplied)
	m.Conditions.Remove(api.ConditionTypeMemberVolumeUnschedulable)

	m.RemoveTerminationsBefore(time.Now().Add(-1 * recentTerminationsKeepPeriod))

	m.Upgrade = false
}

func (p phaseMap) empty(obj meta.Object, spec api.DeploymentSpec, group api.ServerGroup, action api.Action, m *api.MemberStatus) {

}

func (p phaseMap) getFunc(from, to api.MemberPhase) phaseMapFunc {
	if f, ok := p[from]; ok {
		if t, ok := f[to]; ok {
			return t
		}
	}

	return p.empty
}

func (p phaseMap) Execute(obj meta.Object, spec api.DeploymentSpec, group api.ServerGroup, m *api.MemberStatus, action api.Action, to api.MemberPhase) bool {
	from := m.Phase

	if from == to {
		return false
	}

	f := p.getFunc(from, to)

	m.Phase = to

	f(obj, spec, group, action, m)

	return true
}
