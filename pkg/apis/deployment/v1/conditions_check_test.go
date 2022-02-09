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

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_ConditionCheck(t *testing.T) {
	type testCase struct {
		name string

		conditions ConditionList

		check func(l ConditionList) ConditionCheck

		expected bool
	}

	var ct ConditionType = "test"

	cases := []testCase{
		{
			name: "IsTrue when true & exists",
			conditions: ConditionList{
				{
					Type:   ct,
					Status: core.ConditionTrue,
				},
			},
			check: func(l ConditionList) ConditionCheck {
				return l.Check(ct).IsTrue()
			},
			expected: true,
		},
		{
			name: "IsFalse when false & exists",
			conditions: ConditionList{
				{
					Type:   ct,
					Status: core.ConditionFalse,
				},
			},
			check: func(l ConditionList) ConditionCheck {
				return l.Check(ct).IsFalse()
			},
			expected: true,
		},
		{
			name: "IsTrue when does not exists",
			check: func(l ConditionList) ConditionCheck {
				return l.Check(ct).IsTrue()
			},
			expected: false,
		},
		{
			name: "IsFalse when does not exists",
			check: func(l ConditionList) ConditionCheck {
				return l.Check(ct).IsFalse()
			},
			expected: true,
		},
		{
			name: "explicit IsTrue when true & exists",
			conditions: ConditionList{
				{
					Type:   ct,
					Status: core.ConditionTrue,
				},
			},
			check: func(l ConditionList) ConditionCheck {
				return l.Check(ct).Exists().IsTrue()
			},
			expected: true,
		},
		{
			name: "explicit IsFalse when false & exists",
			conditions: ConditionList{
				{
					Type:   ct,
					Status: core.ConditionFalse,
				},
			},
			check: func(l ConditionList) ConditionCheck {
				return l.Check(ct).Exists().IsFalse()
			},
			expected: true,
		},
		{
			name: "explicit IsTrue when does not exists",
			check: func(l ConditionList) ConditionCheck {
				return l.Check(ct).Exists().IsTrue()
			},
			expected: false,
		},
		{
			name: "explicit IsFalse when does not exists",
			check: func(l ConditionList) ConditionCheck {
				return l.Check(ct).Exists().IsFalse()
			},
			expected: false,
		},
		{
			name: "transitionTime - with current time and duration set to 0",
			conditions: ConditionList{
				{
					Type:               ct,
					LastTransitionTime: meta.NewTime(time.Now()),
				},
			},
			check: func(l ConditionList) ConditionCheck {
				return l.Check(ct).LastTransition(0)
			},
			expected: true,
		},
		{
			name: "transitionTime - with current time and duration set to 1s",
			conditions: ConditionList{
				{
					Type:               ct,
					LastTransitionTime: meta.NewTime(time.Now()),
				},
			},
			check: func(l ConditionList) ConditionCheck {
				return l.Check(ct).LastTransition(time.Second)
			},
			expected: false,
		},
		{
			name: "transitionTime - with current time and duration set to 1s",
			conditions: ConditionList{
				{
					Type:               ct,
					LastTransitionTime: meta.NewTime(time.Now()),
				},
			},
			check: func(l ConditionList) ConditionCheck {
				return l.Check(ct).LastTransition(time.Second)
			},
			expected: false,
		},
		{
			name: "transitionTime - with zero time",
			conditions: ConditionList{
				{
					Type: ct,
				},
			},
			check: func(l ConditionList) ConditionCheck {
				return l.Check(ct).LastTransition(0)
			},
			expected: false,
		},
		{
			name: "transitionTime - with current time and duration set to 15s",
			conditions: ConditionList{
				{
					Type:               ct,
					LastTransitionTime: meta.NewTime(time.Now().Add(-15 * time.Second)),
				},
			},
			check: func(l ConditionList) ConditionCheck {
				return l.Check(ct).LastTransition(15 * time.Second)
			},
			expected: true,
		},
		{
			name: "transitionTime - with current time and duration set to 14.75s",
			conditions: ConditionList{
				{
					Type:               ct,
					LastTransitionTime: meta.NewTime(time.Now().Add(-15*time.Second + 250*time.Millisecond)),
				},
			},
			check: func(l ConditionList) ConditionCheck {
				return l.Check(ct).LastTransition(15 * time.Second)
			},
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := c.check(c.conditions).Evaluate()

			if c.expected {
				require.True(t, r)
			} else {
				require.False(t, r)
			}
		})
	}
}
