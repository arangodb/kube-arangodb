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
	"math"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
)

type TopologyStatus struct {
	ID types.UID `json:"id"`

	Size int `json:"size,omitempty"`

	Zones TopologyStatusZones `json:"zones,omitempty"`

	Label string `json:"label,omitempty"`
}

func (t *TopologyStatus) Equal(b *TopologyStatus) bool {
	if t == nil && b == nil {
		return true
	}

	if t == nil || b == nil {
		return false
	}

	return t.ID == b.ID &&
		t.Size == b.Size &&
		t.Label == b.Label &&
		t.Zones.Equal(b.Zones)
}

func (t *TopologyStatus) GetLeastUsedZone(group ServerGroup) int {
	if t == nil {
		return -1
	}

	r, m := -1, math.MaxInt64

	for i, z := range t.Zones {
		if n, ok := z.Members[group.AsRoleAbbreviated()]; ok {
			if v := len(n); v < m {
				r, m = i, v
			}
		} else {
			if v := 0; v < m {
				r, m = i, v
			}
		}
	}

	return r
}

func (t *TopologyStatus) RegisterTopologyLabel(zone int, label string) bool {
	if t == nil {
		return false
	}

	if zone < 0 || zone >= t.Size {
		return false
	}

	if t.Zones[zone].Labels.Contains(label) {
		return false
	}

	t.Zones[zone].Labels = t.Zones[zone].Labels.Add(label).Sort()

	return true
}

func (t *TopologyStatus) RemoveMember(group ServerGroup, id string) bool {
	if t == nil {
		return false
	}

	for _, zone := range t.Zones {
		if zone.RemoveMember(group, id) {
			return true
		}
	}

	return false
}

func (t *TopologyStatus) IsTopologyOwned(m *TopologyMemberStatus) bool {
	if t == nil {
		return false
	}

	if m == nil {
		return false
	}

	return t.ID == m.ID
}

func (t *TopologyStatus) IsTopologyEvenlyDistributed(group ServerGroup) bool {
	if t == nil {
		return true
	}

	max, min := 0, math.MaxInt64

	for _, z := range t.Zones {
		l := len(z.Members[group.AsRoleAbbreviated()])

		if min > l {
			min = l
		}

		if max < l {
			max = l
		}
	}

	return min+1 >= max
}

func (t *TopologyStatus) Enabled() bool {
	return t != nil
}

type TopologyStatusZones []TopologyStatusZone

func (in TopologyStatusZones) Equal(zones TopologyStatusZones) bool {
	if len(in) == 0 && len(zones) == 0 {
		return true
	}

	if len(in) != len(zones) {
		return false
	}

	for id := range in {
		if !in[id].Equal(&zones[id]) {
			return false
		}
	}

	return true
}

type TopologyStatusZoneMembers map[string]List

func (in TopologyStatusZoneMembers) Equal(members TopologyStatusZoneMembers) bool {
	if len(in) == 0 && len(members) == 0 {
		return true
	}
	if len(in) != len(members) {
		return false
	}
	for k, v := range in {
		mv, ok := members[k]
		if !ok {
			return false
		}
		if !v.Equal(mv) {
			return false
		}
	}

	return true
}

type TopologyStatusZone struct {
	ID int `json:"id"`

	Labels List `json:"labels,omitempty"`

	Members TopologyStatusZoneMembers `json:"members,omitempty"`
}

func (t *TopologyStatusZone) AddMember(group ServerGroup, id string) {
	if t.Members == nil {
		t.Members = TopologyStatusZoneMembers{}
	}

	t.Members[group.AsRoleAbbreviated()] = t.Members[group.AsRoleAbbreviated()].Add(id).Sort()
}

func (t *TopologyStatusZone) RemoveMember(group ServerGroup, id string) bool {
	if t == nil {
		return false
	}
	if t.Members == nil {
		return false
	}
	if !t.Members[group.AsRoleAbbreviated()].Contains(id) {
		return false
	}
	t.Members[group.AsRoleAbbreviated()] = t.Members[group.AsRoleAbbreviated()].Remove(id)
	return true
}

func (t *TopologyStatusZone) Get(group ServerGroup) List {
	if t == nil {
		return nil
	}

	if v, ok := t.Members[group.AsRoleAbbreviated()]; ok {
		return v
	} else {
		return nil
	}
}

func (t *TopologyStatusZone) Equal(b *TopologyStatusZone) bool {
	if t == nil && b == nil {
		return true
	}
	if t == nil || b == nil {
		return false
	}
	return t.ID == b.ID &&
		t.Labels.Equal(b.Labels) &&
		t.Members.Equal(b.Members)
}

func NewTopologyStatus(spec *TopologySpec) *TopologyStatus {
	if spec == nil {
		return nil
	}
	zones := make(TopologyStatusZones, spec.Zones)

	for i := 0; i < spec.Zones; i++ {
		zones[i] = TopologyStatusZone{ID: i}
	}

	return &TopologyStatus{
		ID:    uuid.NewUUID(),
		Size:  spec.Zones,
		Zones: zones,
		Label: spec.GetLabel(),
	}
}
