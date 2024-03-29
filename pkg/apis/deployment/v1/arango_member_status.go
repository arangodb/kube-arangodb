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

package v1

import meta "k8s.io/apimachinery/pkg/apis/meta/v1"

type ArangoMemberStatus struct {
	Conditions ConditionList `json:"conditions,omitempty"`

	Template *ArangoMemberPodTemplate `json:"template,omitempty"`

	// Message keeps the information about time when ArangoMember Status was modified last time
	LastUpdateTime meta.Time `json:"lastUpdateTime,omitempty"`

	// Message keeps the information about ArangoMember Message in the String format
	Message string `json:"message,omitempty"`
}

func (a ArangoMemberStatus) InSync(status MemberStatus) bool {
	return status.Conditions.Equal(a.Conditions)
}

func (a *ArangoMemberStatus) Propagate(status MemberStatus) (changed bool) {
	if !status.Conditions.Equal(a.Conditions) {
		changed = true
		a.Conditions = status.Conditions.DeepCopy()
	}

	return
}
