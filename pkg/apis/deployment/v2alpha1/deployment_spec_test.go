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

package v2alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/handlers/utils"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func TestDeploymentSpecValidate(t *testing.T) {
	// TODO
}

func TestDeploymentSpecSetDefaults(t *testing.T) {
	def := func(spec DeploymentSpec) DeploymentSpec {
		spec.SetDefaults("test")
		return spec
	}

	assert.Equal(t, "arangodb/arangodb:latest", def(DeploymentSpec{}).GetImage())
}

func TestDeploymentSpecResetImmutableFields(t *testing.T) {
	tests := []struct {
		Original      DeploymentSpec
		Target        DeploymentSpec
		Expected      DeploymentSpec
		ApplyDefaults bool
		Result        []string
	}{
		// Valid "changes"
		{
			DeploymentSpec{Image: util.NewType[string]("foo")},
			DeploymentSpec{Image: util.NewType[string]("foo2")},
			DeploymentSpec{Image: util.NewType[string]("foo2")},
			false,
			nil,
		},
		{
			DeploymentSpec{Image: util.NewType[string]("foo")},
			DeploymentSpec{Image: util.NewType[string]("foo2")},
			DeploymentSpec{Image: util.NewType[string]("foo2")},
			true,
			nil,
		},
		{
			DeploymentSpec{ImagePullPolicy: util.NewType[core.PullPolicy](core.PullAlways)},
			DeploymentSpec{ImagePullPolicy: util.NewType[core.PullPolicy](core.PullNever)},
			DeploymentSpec{ImagePullPolicy: util.NewType[core.PullPolicy](core.PullNever)},
			false,
			nil,
		},
		{
			DeploymentSpec{ImagePullPolicy: util.NewType[core.PullPolicy](core.PullAlways)},
			DeploymentSpec{ImagePullPolicy: util.NewType[core.PullPolicy](core.PullNever)},
			DeploymentSpec{ImagePullPolicy: util.NewType[core.PullPolicy](core.PullNever)},
			true,
			nil,
		},

		// Invalid changes
		{
			DeploymentSpec{Mode: NewMode(DeploymentModeSingle)},
			DeploymentSpec{Mode: NewMode(DeploymentModeCluster)},
			DeploymentSpec{Mode: NewMode(DeploymentModeSingle)},
			false,
			[]string{"mode"},
		},
		{
			DeploymentSpec{Mode: NewMode(DeploymentModeSingle)},
			DeploymentSpec{Mode: NewMode(DeploymentModeCluster)},
			DeploymentSpec{Mode: NewMode(DeploymentModeSingle)},
			true,
			[]string{"mode", "agents.count"},
		},
		{
			DeploymentSpec{DisableIPv6: util.NewType[bool](false)},
			DeploymentSpec{DisableIPv6: util.NewType[bool](true)},
			DeploymentSpec{DisableIPv6: util.NewType[bool](false)},
			false,
			[]string{"disableIPv6"},
		},
	}

	for _, test := range tests {
		if test.ApplyDefaults {
			test.Original.SetDefaults("foo")
			test.Expected.SetDefaults("foo")
			test.Target.SetDefaultsFrom(test.Original)
			test.Target.SetDefaults("foo")
		}
		result := test.Original.ResetImmutableFields(&test.Target)
		if test.ApplyDefaults {
			if len(result) > 0 {
				test.Target.SetDefaults("foo")
			}
		}
		assert.Equal(t, test.Result, result)
		assert.Equal(t, test.Expected, test.Target)
	}
}

func TestDeploymentSpec_GetCoreContainers(t *testing.T) {
	type fields struct {
		Single       ServerGroupSpec
		Agents       ServerGroupSpec
		DBServers    ServerGroupSpec
		Coordinators ServerGroupSpec
		SyncMasters  ServerGroupSpec
		SyncWorkers  ServerGroupSpec
	}

	type args struct {
		group ServerGroup
	}

	tests := map[string]struct {
		fields fields
		args   args
		want   utils.StringList
	}{
		"one sidecar container": {
			fields: fields{
				DBServers: ServerGroupSpec{
					SidecarCoreNames: []string{"other"},
				},
			},
			args: args{
				group: ServerGroupDBServers,
			},
			want: utils.StringList{"server", "other"},
		},
		"one predefined container and one sidecar container": {
			fields: fields{
				DBServers: ServerGroupSpec{
					SidecarCoreNames: []string{"server", "other"},
				},
			},
			args: args{
				group: ServerGroupDBServers,
			},
			want: utils.StringList{"server", "other"},
		},
		"zero core containers": {
			fields: fields{
				DBServers: ServerGroupSpec{
					SidecarCoreNames: nil,
				},
			},
			args: args{
				group: ServerGroupDBServers,
			},
			want: utils.StringList{"server"},
		},
		"two non-core containers": {
			fields: fields{
				DBServers: ServerGroupSpec{
					SidecarCoreNames: []string{"other1", "other2"},
				},
			},
			args: args{
				group: ServerGroupDBServers,
			},
			want: utils.StringList{"server", "other1", "other2"},
		},
	}
	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			s := DeploymentSpec{
				DBServers: test.fields.DBServers,
			}

			got := s.GetCoreContainers(test.args.group)
			assert.Equal(t, test.want, got)

		})
	}
}
