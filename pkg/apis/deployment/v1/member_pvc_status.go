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

package v1

type MemberPersistentVolumeClaimStatus struct {
	Name string `json:"name"`
}

func (m *MemberPersistentVolumeClaimStatus) Equal(other *MemberPersistentVolumeClaimStatus) bool {
	if m == nil && other == nil {
		return true
	}
	if m == nil || other == nil {
		return false
	}
	return m.Name == other.Name
}

func (m *MemberPersistentVolumeClaimStatus) GetName() string {
	if m == nil {
		return ""
	}

	return m.Name
}

func (m *MemberPersistentVolumeClaimStatus) Propagate(s *MemberStatus) {
	if s == nil {
		return
	}

	if m == nil {
		s.PersistentVolumeClaimName = ""
	} else {
		s.PersistentVolumeClaimName = m.Name
	}
}
