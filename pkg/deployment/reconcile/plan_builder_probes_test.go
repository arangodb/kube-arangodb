//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package reconcile

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	"testing"
)

func TestPlanBuilderProbes(t *testing.T) {
	type testCase struct {
		probe pod.Probe
		groupProbeDisabled *bool
		groupProbeSpec *api.ServerGroupProbeSpec
		containerProbe *core.Probe

		result bool
	}

	testCases := map[string]testCase{
		"Defaults": {

		},
		"Disable created probe": {
			containerProbe: &core.Probe{},

			result: true,
		},
		"Create probe": {
			probe: pod.Probe{
				EnabledByDefault: true,
				CanBeEnabled:     true,
			},

			result: true,
		},
		"Compare probe - without overrides": {
			probe: pod.Probe{
				EnabledByDefault: true,
				CanBeEnabled:     true,
			},
			containerProbe: &core.Probe{},
		},
		"Compare probe - with disabled": {
			probe: pod.Probe{
				EnabledByDefault: true,
				CanBeEnabled:     true,
			},
			groupProbeDisabled: util.NewBool(true),
			containerProbe:     &core.Probe{},

			result: true,
		},
		"Compare probe - with enabled": {
			probe: pod.Probe{
				EnabledByDefault: true,
				CanBeEnabled:     true,
			},
			groupProbeDisabled: util.NewBool(false),

			result: true,
		},
		"Compare probe - override defaults": {
			probe: pod.Probe{
				EnabledByDefault: true,
				CanBeEnabled:     true,
			},

			groupProbeSpec: &api.ServerGroupProbeSpec{
				TimeoutSeconds: util.NewInt32(10),
			},

			containerProbe: &core.Probe{},

			result: true,
		},
		"Compare probe - override defaults - do not change": {
			probe: pod.Probe{
				EnabledByDefault: true,
				CanBeEnabled:     true,
			},

			groupProbeSpec: &api.ServerGroupProbeSpec{
				TimeoutSeconds: util.NewInt32(10),
			},

			containerProbe: &core.Probe{
					TimeoutSeconds: 10,
			},
		},
	}

	for name, c := range testCases {
		t.Run(name, func(t *testing.T) {
			res, out := compareProbes(c.probe, c.groupProbeDisabled, c.groupProbeSpec, c.containerProbe)

			require.Equal(t, c.result, res, out)
		})
	}
}
