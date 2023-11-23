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

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func TestServerGroupSpecSecurityContext_NewPodSecurityContext(t *testing.T) {
	testCases := map[string]struct {
		sc      *ServerGroupSpecSecurityContext
		secured bool
		want    *core.PodSecurityContext
	}{
		"default unsecured pod security": {
			sc:   nil,
			want: nil,
		},
		"default secured pod security": {
			sc:      nil,
			secured: true,
			want: &core.PodSecurityContext{
				FSGroup: util.NewType[int64](shared.DefaultFSGroup),
			},
		},
		"user secured pod security takes precedence": {
			sc: &ServerGroupSpecSecurityContext{
				FSGroup: util.NewType[int64](3001),
			},
			secured: true,
			want: &core.PodSecurityContext{
				FSGroup: util.NewType[int64](3001),
			},
		},
		"user secured pod security with FSGroup==nil": {
			sc: &ServerGroupSpecSecurityContext{
				SupplementalGroups: []int64{1},
			},
			secured: true,
			want: &core.PodSecurityContext{
				FSGroup:            util.NewType[int64](shared.DefaultFSGroup),
				SupplementalGroups: []int64{1},
			},
		},
		"user unsecured pod security": {
			sc: &ServerGroupSpecSecurityContext{
				FSGroup:            util.NewType[int64](3001),
				SupplementalGroups: []int64{1},
			},
			secured: false,
			want: &core.PodSecurityContext{
				FSGroup:            util.NewType[int64](3001),
				SupplementalGroups: []int64{1},
			},
		},
		"pass sysctl opts": {
			sc: &ServerGroupSpecSecurityContext{
				Sysctls: map[string]intstr.IntOrString{
					"opt.1": intstr.FromInt(1),
					"opt.2": intstr.FromString("2"),
				},
			},
			secured: false,
			want: &core.PodSecurityContext{
				Sysctls: []core.Sysctl{
					{
						Name:  "opt.1",
						Value: "1",
					},
					{
						Name:  "opt.2",
						Value: "2",
					},
				},
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			actual := testCase.sc.NewPodSecurityContext(testCase.secured)
			assert.Equalf(t, testCase.want, actual, "NewPodSecurityContext(%v)", testCase.secured)
		})
	}
}

func TestServerGroupSpecSecurityContext_NewPodSecurityContextFromJSON(t *testing.T) {
	testCases := map[string]struct {
		sc      string
		secured bool
		want    *core.PodSecurityContext
	}{
		"pass sysctl opts": {
			sc:      `{"sysctls":{"opt.1":1, "opt.2":"2"}}`,
			secured: false,
			want: &core.PodSecurityContext{
				Sysctls: []core.Sysctl{
					{
						Name:  "opt.1",
						Value: "1",
					},
					{
						Name:  "opt.2",
						Value: "2",
					},
				},
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			var p ServerGroupSpecSecurityContext
			require.NoError(t, json.Unmarshal([]byte(testCase.sc), &p))

			actual := p.NewPodSecurityContext(testCase.secured)
			assert.Equalf(t, testCase.want, actual, "NewPodSecurityContext(%v)", testCase.secured)
		})
	}
}

func TestServerGroupSpecSecurityContext_NewSecurityContext(t *testing.T) {
	tests := map[string]struct {
		sc      *ServerGroupSpecSecurityContext
		secured bool
		want    *core.SecurityContext
	}{
		"default unsecured context security": {
			sc:      nil,
			secured: false,
			want: &core.SecurityContext{
				Capabilities: &core.Capabilities{
					Drop: []core.Capability{"ALL"},
				},
			},
		},
		"default secured context security": {
			sc:      nil,
			secured: true,
			want: &core.SecurityContext{
				Capabilities: &core.Capabilities{
					Drop: []core.Capability{"ALL"},
				},
				ReadOnlyRootFilesystem: util.NewType(true),
				RunAsGroup:             util.NewType[int64](shared.DefaultRunAsGroup),
				RunAsNonRoot:           util.NewType(true),
				RunAsUser:              util.NewType[int64](shared.DefaultRunAsUser),
			},
		},
		"user unsecured context security": {
			sc: &ServerGroupSpecSecurityContext{
				RunAsUser: util.NewType[int64](3001),
			},
			secured: false,
			want: &core.SecurityContext{
				Capabilities: &core.Capabilities{
					Drop: []core.Capability{"ALL"},
				},
				RunAsUser: util.NewType[int64](3001),
			},
		},
		"secured user setting RunAsUser takes precedence": {
			sc: &ServerGroupSpecSecurityContext{
				RunAsUser: util.NewType[int64](3001),
			},
			secured: true,
			want: &core.SecurityContext{
				Capabilities: &core.Capabilities{
					Drop: []core.Capability{"ALL"},
				},
				ReadOnlyRootFilesystem: util.NewType(true),
				RunAsGroup:             util.NewType[int64](shared.DefaultRunAsGroup),
				RunAsNonRoot:           util.NewType(true),
				RunAsUser:              util.NewType[int64](3001),
			},
		},
		"secured mixed users' settings takes precedence": {
			sc: &ServerGroupSpecSecurityContext{
				AddCapabilities:          []core.Capability{"1"},
				AllowPrivilegeEscalation: util.NewType(true),
				DropAllCapabilities:      util.NewType(false), // secured will turn it on
				Privileged:               util.NewType(false),
				RunAsNonRoot:             util.NewType(false),
				RunAsUser:                util.NewType[int64](3001),
			},
			secured: true,
			want: &core.SecurityContext{

				AllowPrivilegeEscalation: util.NewType(true),
				Capabilities: &core.Capabilities{
					Add:  []core.Capability{"1"},
					Drop: []core.Capability{"ALL"},
				},
				Privileged:             util.NewType(false),
				ReadOnlyRootFilesystem: util.NewType(true),
				RunAsGroup:             util.NewType[int64](shared.DefaultRunAsGroup),
				RunAsNonRoot:           util.NewType(false),
				RunAsUser:              util.NewType[int64](3001),
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			actual := testCase.sc.NewSecurityContext(testCase.secured)
			assert.Equalf(t, testCase.want, actual, "NewSecurityContext(%v)", testCase.secured)
		})
	}
}
