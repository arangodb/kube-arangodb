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

package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type taintsCase struct {
	tolerations []core.Toleration
	taints      []core.Taint

	schedulable bool
}

func newMetaTimeWithDiff(d time.Duration) *meta.Time {
	return newMetaTime(time.Now().Add(d))
}

func newMetaTime(time time.Time) *meta.Time {
	q := meta.NewTime(time)
	return &q
}

func Test_Taints(t *testing.T) {
	cases := map[string]taintsCase{
		"No taints & tolerations": {
			schedulable: true,
		},
		"Tainted node": {
			schedulable: false,

			taints: []core.Taint{
				{
					Key:    "node.kubernetes.io/unschedulable",
					Effect: core.TaintEffectNoSchedule,
				},
			},
		},
		"Custom tainted node": {
			schedulable: false,

			taints: []core.Taint{
				{
					Key:    "arangodb.com/taint",
					Effect: core.TaintEffectNoSchedule,
				},
			},
		},
		"Custom tainted node - tolerate all": {
			schedulable: true,

			tolerations: []core.Toleration{
				{
					Operator: core.TolerationOpExists,
				},
			},

			taints: []core.Taint{
				{
					Key:    "arangodb.com/taint",
					Value:  "test",
					Effect: core.TaintEffectNoSchedule,
				},
			},
		},
		"Custom tainted node - NoSched - tolerate all for 5 minutes - in range": {
			schedulable: true,

			tolerations: []core.Toleration{
				{
					Operator:          core.TolerationOpExists,
					TolerationSeconds: util.NewType[int64](300),
				},
			},

			taints: []core.Taint{
				{
					Key:       "arangodb.com/taint",
					Value:     "test",
					Effect:    core.TaintEffectNoSchedule,
					TimeAdded: newMetaTimeWithDiff(0),
				},
			},
		},
		"Custom tainted node - NoSched - tolerate all for 5 minutes - out of range": {
			schedulable: false,

			tolerations: []core.Toleration{
				{
					Operator:          core.TolerationOpExists,
					TolerationSeconds: util.NewType[int64](300),
				},
			},

			taints: []core.Taint{
				{
					Key:       "arangodb.com/taint",
					Value:     "test",
					Effect:    core.TaintEffectNoSchedule,
					TimeAdded: newMetaTimeWithDiff(-360 * time.Second),
				},
			},
		},
		"Custom tainted node - NoExec - tolerate all for 5 minute": {
			schedulable: false,

			tolerations: []core.Toleration{
				{
					Operator:          core.TolerationOpExists,
					TolerationSeconds: util.NewType[int64](300),
				},
			},

			taints: []core.Taint{
				{
					Key:       "arangodb.com/taint",
					Value:     "test",
					Effect:    core.TaintEffectNoExecute,
					TimeAdded: newMetaTimeWithDiff(0),
				},
			},
		},
		"Custom tainted node - tolerate different": {
			schedulable: false,

			tolerations: []core.Toleration{
				{
					Key:      "arangodb.com/taint2",
					Operator: core.TolerationOpExists,
				},
			},

			taints: []core.Taint{
				{
					Key:    "arangodb.com/taint",
					Value:  "test",
					Effect: core.TaintEffectNoSchedule,
				},
			},
		},
		"Custom tainted node - tolerate key": {
			schedulable: true,

			tolerations: []core.Toleration{
				{
					Key:      "arangodb.com/taint",
					Operator: core.TolerationOpExists,
				},
			},

			taints: []core.Taint{
				{
					Key:    "arangodb.com/taint",
					Value:  "test",
					Effect: core.TaintEffectNoSchedule,
				},
			},
		},
		"Custom tainted node - tolerate key & diff value": {
			schedulable: false,

			tolerations: []core.Toleration{
				{
					Key:      "arangodb.com/taint",
					Value:    "test2",
					Operator: core.TolerationOpEqual,
				},
			},

			taints: []core.Taint{
				{
					Key:    "arangodb.com/taint",
					Value:  "test",
					Effect: core.TaintEffectNoSchedule,
				},
			},
		},
		"Custom tainted node - tolerate key & same value": {
			schedulable: false,

			tolerations: []core.Toleration{
				{
					Key:      "arangodb.com/taint",
					Value:    "test2",
					Operator: core.TolerationOpEqual,
				},
			},

			taints: []core.Taint{
				{
					Key:    "arangodb.com/taint",
					Value:  "test",
					Effect: core.TaintEffectNoSchedule,
				},
			},
		},
	}

	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			schedulable := AreTaintsTolerated(c.tolerations, c.taints)

			if c.schedulable {
				require.True(t, schedulable)
			} else {
				require.False(t, schedulable)
			}
		})
	}
}
