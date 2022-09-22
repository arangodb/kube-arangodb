//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package v2alpha1

import (
	"math/rand"
	"sort"
	"time"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// MemberStatusList is a list of MemberStatus entries
type MemberStatusList []MemberStatus

// Equal checks for equality
func (l MemberStatusList) Equal(other MemberStatusList) bool {
	if len(l) != len(other) {
		return false
	}

	for i := 0; i < len(l); i++ {
		o, found := other.ElementByID(l[i].ID)
		if !found {
			return false
		}

		if !l[i].Equal(o) {
			return false
		}
	}

	return true
}

// ContainsID returns true if the given list contains a member with given ID.
func (l MemberStatusList) ContainsID(id string) bool {
	for _, x := range l {
		if x.ID == id {
			return true
		}
	}
	return false
}

// ElementByID returns the element in the given list that has the given ID and true.
// If no such element exists, false is returned.
func (l MemberStatusList) ElementByID(id string) (MemberStatus, bool) {
	for i, x := range l {
		if x.ID == id {
			return l[i], true
		}
	}
	return MemberStatus{}, false
}

// ElementByPodName returns the element in the given list that has the given pod name and true.
// If no such element exists, an empty element and false is returned.
func (l MemberStatusList) ElementByPodName(podName string) (MemberStatus, bool) {
	for i, x := range l {
		if x.Pod.GetName() == podName {
			return l[i], true
		}
	}
	return MemberStatus{}, false
}

// ElementByPVCName returns the element in the given list that has the given PVC name and true.
// If no such element exists, an empty element and false is returned.
func (l MemberStatusList) ElementByPVCName(pvcName string) (MemberStatus, bool) {
	for i, x := range l {
		if x.PersistentVolumeClaimName == pvcName {
			return l[i], true
		}
	}
	return MemberStatus{}, false
}

// Add a member to the list.
// Returns an AlreadyExistsError if the ID of the given member already exists.
func (l *MemberStatusList) add(m MemberStatus) error {
	src := *l
	for _, x := range src {
		if x.ID == m.ID {
			return errors.WithStack(errors.Wrapf(AlreadyExistsError, "Member '%s' already exists", m.ID))
		}
	}
	newList := append(src, m)
	sort.Slice(newList, func(i, j int) bool { return newList[i].ID < newList[j].ID })
	*l = newList
	return nil
}

// Update a member in the list.
// Returns a NotFoundError if the ID of the given member cannot be found.
func (l MemberStatusList) update(m MemberStatus) error {
	for i, x := range l {
		if x.ID == m.ID {
			l[i] = m
			return nil
		}
	}
	return errors.WithStack(errors.Wrapf(NotFoundError, "Member '%s' is not a member", m.ID))
}

// RemoveByID a member with given ID from the list.
// Returns a NotFoundError if the ID of the given member cannot be found.
func (l *MemberStatusList) removeByID(id string) error {
	src := *l
	for i, x := range src {
		if x.ID == id {
			*l = append(src[:i], src[i+1:]...)
			return nil
		}
	}
	return errors.WithStack(errors.Wrapf(NotFoundError, "Member '%s' is not a member", id))
}

type MemberToRemoveSelector func(m MemberStatusList) (string, error)

// SelectMemberToRemove selects a member from the given list that should
// be removed in a scale down action.
// Returns an error if the list is empty.
func (l MemberStatusList) SelectMemberToRemove(selectors ...MemberToRemoveSelector) (MemberStatus, error) {
	if len(l) > 0 {
		// Try to find member with phase to be removed
		for _, m := range l {
			if m.Conditions.IsTrue(ConditionTypeMarkedToRemove) {
				return m, nil
			}
		}
		for _, m := range l {
			if m.Conditions.IsTrue(ConditionTypeScaleDownCandidate) {
				return m, nil
			}
		}
		// Try to find a not ready member
		for _, m := range l {
			if m.Phase.IsPending() {
				return m, nil
			}
		}
		for _, m := range l {
			if !m.Conditions.IsTrue(ConditionTypeReady) {
				return m, nil
			}
		}
		for _, m := range l {
			if m.Conditions.IsTrue(ConditionTypeCleanedOut) {
				return m, nil
			}
		}

		// Run conditional picker
		for _, selector := range selectors {
			if selector == nil {
				continue
			}
			if m, err := selector(l); err != nil {
				return MemberStatus{}, err
			} else if m != "" {
				if member, ok := l.ElementByID(m); ok {
					return member, nil
				} else {
					return MemberStatus{}, errors.Newf("Unable to find member with id %s", m)
				}
			}
		}

		// Pick a random member that is in created state
		perm := rand.Perm(len(l))
		for _, idx := range perm {
			m := l[idx]
			if m.Phase == MemberPhaseCreated {
				return m, nil
			}
		}
	}
	return MemberStatus{}, errors.WithStack(errors.Wrap(NotFoundError, "No member available for removal"))
}

// MembersReady returns the number of members that are in the Ready state.
func (l MemberStatusList) MembersReady() int {
	readyCount := 0
	for _, x := range l {
		if x.Conditions.IsTrue(ConditionTypeReady) {
			readyCount++
		}
	}
	return readyCount
}

// MembersServing returns the number of members that are in the Serving state.
func (l MemberStatusList) MembersServing() int {
	servingCount := 0
	for _, x := range l {
		if x.Conditions.IsTrue(ConditionTypeServing) {
			servingCount++
		}
	}
	return servingCount
}

// AllMembersServing returns the true if all members are in the Serving state.
func (l MemberStatusList) AllMembersServing() bool {
	return len(l) == l.MembersServing()
}

// AllMembersReady returns the true if all members are in the Ready state.
func (l MemberStatusList) AllMembersReady() bool {
	return len(l) == l.MembersReady()
}

// AllConditionTrueSince returns true if all members satisfy the condition since the given period
func (l MemberStatusList) AllConditionTrueSince(cond ConditionType, status core.ConditionStatus, period time.Duration) bool {
	for _, x := range l {
		if c, ok := x.Conditions.Get(cond); ok {
			if c.Status == status && c.LastTransitionTime.Time.Add(period).Before(time.Now()) {
				continue
			}
		}
		return false
	}

	return true
}

// AllFailed returns true if all members are failed
func (l MemberStatusList) AllFailed() bool {
	for _, x := range l {
		if !x.Phase.IsFailed() {
			return false
		}
	}
	return true
}
