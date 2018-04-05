//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package v1alpha

import (
	"math/rand"

	"github.com/pkg/errors"
)

// MemberStatusList is a list of MemberStatus entries
type MemberStatusList []MemberStatus

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
		if x.PodName == podName {
			return l[i], true
		}
	}
	return MemberStatus{}, false
}

// Add a member to the list.
// Returns an AlreadyExistsError if the ID of the given member already exists.
func (l *MemberStatusList) Add(m MemberStatus) error {
	src := *l
	for _, x := range src {
		if x.ID == m.ID {
			return maskAny(errors.Wrapf(AlreadyExistsError, "Member '%s' already exists", m.ID))
		}
	}
	*l = append(src, m)
	return nil
}

// Update a member in the list.
// Returns a NotFoundError if the ID of the given member cannot be found.
func (l MemberStatusList) Update(m MemberStatus) error {
	for i, x := range l {
		if x.ID == m.ID {
			l[i] = m
			return nil
		}
	}
	return maskAny(errors.Wrapf(NotFoundError, "Member '%s' is not a member", m.ID))
}

// RemoveByID a member with given ID from the list.
// Returns a NotFoundError if the ID of the given member cannot be found.
func (l *MemberStatusList) RemoveByID(id string) error {
	src := *l
	for i, x := range src {
		if x.ID == id {
			*l = append(src[:i], src[i+1:]...)
			return nil
		}
	}
	return maskAny(errors.Wrapf(NotFoundError, "Member '%s' is not a member", id))
}

// SelectMemberToRemove selects a member from the given list that should
// be removed in a scale down action.
// Returns an error if the list is empty.
func (l MemberStatusList) SelectMemberToRemove() (MemberStatus, error) {
	if len(l) > 0 {
		// Try to find a not ready member
		for _, m := range l {
			if m.Phase == MemberPhaseNone {
				return m, nil
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
	return MemberStatus{}, maskAny(errors.Wrap(NotFoundError, "No member available for removal"))
}
