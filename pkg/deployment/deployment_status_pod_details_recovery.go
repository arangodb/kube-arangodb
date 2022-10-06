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

package deployment

import api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"

func RecoverPodDetails(in *api.DeploymentStatus) (changed bool, _ error) {
	changed = false
	for _, m := range in.Members.AsList() {
		if m.Member.Pod == nil {
			// Pod is nil, recovery might be needed
			if m.Member.PodName != "" {
				m.Member.Pod = &api.MemberPodStatus{
					Name:        m.Member.PodName,
					UID:         m.Member.PodUID,
					SpecVersion: m.Member.PodSpecVersion,
				}

				if err := in.Members.Update(m.Member, m.Group); err != nil {
					return false, err
				}
				changed = true
			}
		}

		if p := m.Member.PersistentVolumeClaim; p == nil {
			// Recovery is nil, recovery might be needed
			if m.Member.PersistentVolumeClaimName != "" {
				m.Member.PersistentVolumeClaim = &api.MemberPersistentVolumeClaimStatus{
					Name: m.Member.PersistentVolumeClaimName,
				}

				if err := in.Members.Update(m.Member, m.Group); err != nil {
					return false, err
				}
				changed = true
			}
		}
	}
	return
}
