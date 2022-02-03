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
	"sort"
	"sync"
)

type DeploymentStatusMemberElementsSortFunc func(a, b DeploymentStatusMemberElement) bool
type DeploymentStatusMemberElementsCondFunc func(a DeploymentStatusMemberElement) bool

type DeploymentStatusMemberElements []DeploymentStatusMemberElement

func (d DeploymentStatusMemberElements) ForEach(f func(id int)) {
	if f == nil {
		return
	}

	var wg sync.WaitGroup

	wg.Add(len(d))

	for id := range d {
		go func(i int) {
			defer wg.Done()

			f(i)
		}(id)
	}

	wg.Wait()
}

func (d DeploymentStatusMemberElements) Filter(f DeploymentStatusMemberElementsCondFunc) DeploymentStatusMemberElements {
	var l DeploymentStatusMemberElements

	for _, a := range d {
		if !f(a) {
			continue
		}

		z := a.DeepCopy()

		l = append(l, *z)
	}

	return l
}

func (d DeploymentStatusMemberElements) Sort(less DeploymentStatusMemberElementsSortFunc) DeploymentStatusMemberElements {
	n := d.DeepCopy()

	sort.Slice(n, func(i, j int) bool {
		return less(n[i], n[j])
	})

	return n
}

// DeploymentStatusMemberElement holds one specific element with group and member status
type DeploymentStatusMemberElement struct {
	Group  ServerGroup  `json:"group,omitempty"`
	Member MemberStatus `json:"member,omitempty"`
}

func (ds DeploymentStatusMembers) AsList() DeploymentStatusMemberElements {
	return ds.AsListInGroups(AllServerGroups...)
}

func (ds DeploymentStatusMembers) AsListInGroups(groups ...ServerGroup) DeploymentStatusMemberElements {
	var elements []DeploymentStatusMemberElement

	// Always return nil, so no error handling
	for _, g := range groups {
		elements = append(elements, ds.AsListInGroup(g)...)
	}

	return elements
}

func (ds DeploymentStatusMembers) AsListInGroup(group ServerGroup) DeploymentStatusMemberElements {
	var r DeploymentStatusMemberElements

	for _, m := range ds.MembersOfGroup(group) {
		r = append(r, DeploymentStatusMemberElement{
			Group:  group,
			Member: m,
		})
	}

	return r
}
