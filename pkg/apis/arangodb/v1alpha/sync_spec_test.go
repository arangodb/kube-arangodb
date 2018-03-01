//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package v1alpha

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
)

func TestSyncSpecValidate(t *testing.T) {
	// Valid
	auth := AuthenticationSpec{JWTSecretName: "foo"}
	assert.Nil(t, SyncSpec{Image: "foo", Authentication: auth}.Validate(DeploymentModeSingle))
	assert.Nil(t, SyncSpec{Image: "foo", Authentication: auth}.Validate(DeploymentModeResilientSingle))
	assert.Nil(t, SyncSpec{Image: "foo", Authentication: auth}.Validate(DeploymentModeCluster))
	assert.Nil(t, SyncSpec{Image: "foo", Authentication: auth, Enabled: true}.Validate(DeploymentModeCluster))

	// Not valid
	assert.Error(t, SyncSpec{Image: "", Authentication: auth}.Validate(DeploymentModeSingle))
	assert.Error(t, SyncSpec{Image: "", Authentication: auth}.Validate(DeploymentModeResilientSingle))
	assert.Error(t, SyncSpec{Image: "", Authentication: auth}.Validate(DeploymentModeCluster))
	assert.Error(t, SyncSpec{Image: "foo", Authentication: auth, Enabled: true}.Validate(DeploymentModeSingle))
	assert.Error(t, SyncSpec{Image: "foo", Authentication: auth, Enabled: true}.Validate(DeploymentModeResilientSingle))
}

func TestSyncSpecSetDefaults(t *testing.T) {
	def := func(spec SyncSpec) SyncSpec {
		spec.SetDefaults("test-image", v1.PullAlways, "test-jwt")
		return spec
	}

	assert.False(t, def(SyncSpec{}).Enabled)
	assert.False(t, def(SyncSpec{Enabled: false}).Enabled)
	assert.True(t, def(SyncSpec{Enabled: true}).Enabled)
	assert.Equal(t, "test-image", def(SyncSpec{}).Image)
	assert.Equal(t, "foo", def(SyncSpec{Image: "foo"}).Image)
	assert.Equal(t, v1.PullAlways, def(SyncSpec{}).ImagePullPolicy)
	assert.Equal(t, v1.PullNever, def(SyncSpec{ImagePullPolicy: v1.PullNever}).ImagePullPolicy)
	assert.Equal(t, "test-jwt", def(SyncSpec{}).Authentication.JWTSecretName)
	assert.Equal(t, "foo", def(SyncSpec{Authentication: AuthenticationSpec{JWTSecretName: "foo"}}).Authentication.JWTSecretName)
}

func TestSyncSpecResetImmutableFields(t *testing.T) {
	tests := []struct {
		Original SyncSpec
		Target   SyncSpec
		Expected SyncSpec
		Result   []string
	}{
		// Valid "changes"
		{
			SyncSpec{Enabled: false},
			SyncSpec{Enabled: true},
			SyncSpec{Enabled: true},
			nil,
		},
		{
			SyncSpec{Enabled: true},
			SyncSpec{Enabled: false},
			SyncSpec{Enabled: false},
			nil,
		},
		{
			SyncSpec{Image: "foo"},
			SyncSpec{Image: "foo2"},
			SyncSpec{Image: "foo2"},
			nil,
		},
		{
			SyncSpec{ImagePullPolicy: v1.PullAlways},
			SyncSpec{ImagePullPolicy: v1.PullNever},
			SyncSpec{ImagePullPolicy: v1.PullNever},
			nil,
		},
		{
			SyncSpec{Authentication: AuthenticationSpec{JWTSecretName: "None"}},
			SyncSpec{Authentication: AuthenticationSpec{JWTSecretName: "None"}},
			SyncSpec{Authentication: AuthenticationSpec{JWTSecretName: "None"}},
			nil,
		},
		{
			SyncSpec{Authentication: AuthenticationSpec{JWTSecretName: "foo"}},
			SyncSpec{Authentication: AuthenticationSpec{JWTSecretName: "foo"}},
			SyncSpec{Authentication: AuthenticationSpec{JWTSecretName: "foo"}},
			nil,
		},
		{
			SyncSpec{Authentication: AuthenticationSpec{JWTSecretName: "foo"}},
			SyncSpec{Authentication: AuthenticationSpec{JWTSecretName: "foo2"}},
			SyncSpec{Authentication: AuthenticationSpec{JWTSecretName: "foo2"}},
			nil,
		},

		// Invalid changes
		{
			SyncSpec{Authentication: AuthenticationSpec{JWTSecretName: "foo"}},
			SyncSpec{Authentication: AuthenticationSpec{JWTSecretName: "None"}},
			SyncSpec{Authentication: AuthenticationSpec{JWTSecretName: "foo"}},
			[]string{"test.auth.jwtSecretName"},
		},
		{
			SyncSpec{Authentication: AuthenticationSpec{JWTSecretName: "None"}},
			SyncSpec{Authentication: AuthenticationSpec{JWTSecretName: "foo"}},
			SyncSpec{Authentication: AuthenticationSpec{JWTSecretName: "None"}},
			[]string{"test.auth.jwtSecretName"},
		},
	}

	for _, test := range tests {
		result := test.Original.ResetImmutableFields("test", &test.Target)
		assert.Equal(t, test.Result, result)
		assert.Equal(t, test.Expected, test.Target)
	}
}
