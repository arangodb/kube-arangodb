//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

package reconcile

import (
	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/rotation"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func newUpdateUpgradeDecisionItemFromElement(element api.DeploymentStatusMemberElement) updateUpgradeDecisionItem {
	return newUpdateUpgradeDecisionItem(element.Group, element.Member.ID)
}

func newUpdateUpgradeDecisionItem(group api.ServerGroup, id string) updateUpgradeDecisionItem {
	return updateUpgradeDecisionItem{
		ID:    id,
		Group: group,
	}
}

type updateUpgradeDecisionItem struct {
	ID    string
	Group api.ServerGroup
}

type updateUpgradeDecisionMap map[updateUpgradeDecisionItem]updateUpgradeDecision

func (u updateUpgradeDecisionMap) GetFromToVersion() (from driver.Version, to driver.Version) {
	for _, el := range u {
		if !el.upgrade || !el.upgradeDecision.UpgradeNeeded {
			continue
		}

		if from == "" || from.CompareTo(el.upgradeDecision.FromVersion) > 0 {
			from = el.upgradeDecision.FromVersion
		}

		if to == "" || to.CompareTo(el.upgradeDecision.ToVersion) < 0 {
			to = el.upgradeDecision.ToVersion
		}
	}

	return
}

func (u updateUpgradeDecisionMap) AreGroupsPendingUpgrade(groups ...api.ServerGroup) bool {
	for _, group := range groups {
		if u.IsGroupPendingUpgrade(group) {
			return true
		}
	}
	return false
}

func (u updateUpgradeDecisionMap) IsGroupPendingUpgrade(group api.ServerGroup) bool {
	for i, k := range u {
		if i.Group != group {
			continue
		}
		if k.upgrade {
			return true
		}
	}

	return false
}

func (u updateUpgradeDecisionMap) IsUpgrade() bool {
	for _, k := range u {
		if k.upgrade {
			return true
		}
	}

	return false
}

func (u updateUpgradeDecisionMap) AreGroupsPendingUpdate(groups ...api.ServerGroup) bool {
	for _, group := range groups {
		if u.IsGroupPendingUpdate(group) {
			return true
		}
	}
	return false
}

func (u updateUpgradeDecisionMap) IsGroupPendingUpdate(group api.ServerGroup) bool {
	for i, k := range u {
		if i.Group != group {
			continue
		}
		if k.upgrade {
			return true
		}
	}

	return false
}

func (u updateUpgradeDecisionMap) IsUpdate() bool {
	for _, k := range u {
		if k.update {
			return true
		}
	}

	return false
}

type updateUpgradeDecision struct {
	upgrade         bool
	upgradeDecision upgradeDecision

	unsafeUpdateAllowed bool
	updateAllowed       bool
	updateMessage       string
	update              bool
}

func (r *Reconciler) createRotateOrUpgradeDecision(spec api.DeploymentSpec, status api.DeploymentStatus, context PlanBuilderContext) updateUpgradeDecisionMap {
	d := updateUpgradeDecisionMap{}

	// Init phase
	for _, m := range status.Members.AsListInGroups(api.AllServerGroups...) {
		d[newUpdateUpgradeDecisionItem(m.Group, m.Member.ID)] = r.createRotateOrUpgradeDecisionMember(spec.Mode.Get(), spec, status, context, m)
	}

	return d
}

func (r *Reconciler) createRotateOrUpgradeDecisionMember(mode api.DeploymentMode, spec api.DeploymentSpec, status api.DeploymentStatus, context PlanBuilderContext, element api.DeploymentStatusMemberElement) (d updateUpgradeDecision) {
	if element.Member.Phase == api.MemberPhaseCreated && element.Member.Pod.GetName() != "" {
		// Only upgrade when phase is created

		// Got pod, compare it with what it should be
		decision := r.podNeedsUpgrading(mode, element.Member, spec, status.Images)

		if decision.UpgradeNeeded || decision.Hold {
			d.upgrade = true
			d.upgradeDecision = decision
		}
	}

	d.updateAllowed, d.updateMessage = groupReadyForRestart(context, status, element.Member, element.Group)
	d.unsafeUpdateAllowed = util.TypeOrDefault[bool](spec.AllowUnsafeUpgrade, false)

	if rotation.CheckPossible(element.Member) {
		if element.Member.Conditions.IsTrue(api.ConditionTypeRestart) {
			d.update = true
		} else if element.Member.Conditions.IsTrue(api.ConditionTypePendingUpdate) {
			if !element.Member.Conditions.IsTrue(api.ConditionTypeUpdating) && !element.Member.Conditions.IsTrue(api.ConditionTypeUpdateFailed) {
				d.update = true
			}
		}
	}
	return
}
