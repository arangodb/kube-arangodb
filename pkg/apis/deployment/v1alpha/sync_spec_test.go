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

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
)

func TestSyncSpecValidate(t *testing.T) {
	// Valid
	auth := SyncAuthenticationSpec{JWTSecretName: util.NewString("foo"), ClientCASecretName: util.NewString("foo-client")}
	tls := TLSSpec{CASecretName: util.NewString("None")}
	assert.Nil(t, SyncSpec{Image: util.NewString("foo"), Authentication: auth}.Validate(DeploymentModeSingle))
	assert.Nil(t, SyncSpec{Image: util.NewString("foo"), Authentication: auth}.Validate(DeploymentModeActiveFailover))
	assert.Nil(t, SyncSpec{Image: util.NewString("foo"), Authentication: auth}.Validate(DeploymentModeCluster))
	assert.Nil(t, SyncSpec{Image: util.NewString("foo"), Authentication: auth, TLS: tls, Enabled: util.NewBool(true)}.Validate(DeploymentModeCluster))

	// Not valid
	assert.Error(t, SyncSpec{Image: util.NewString(""), Authentication: auth}.Validate(DeploymentModeSingle))
	assert.Error(t, SyncSpec{Image: util.NewString(""), Authentication: auth}.Validate(DeploymentModeActiveFailover))
	assert.Error(t, SyncSpec{Image: util.NewString(""), Authentication: auth}.Validate(DeploymentModeCluster))
	assert.Error(t, SyncSpec{Image: util.NewString("foo"), Authentication: auth, TLS: tls, Enabled: util.NewBool(true)}.Validate(DeploymentModeSingle))
	assert.Error(t, SyncSpec{Image: util.NewString("foo"), Authentication: auth, TLS: tls, Enabled: util.NewBool(true)}.Validate(DeploymentModeActiveFailover))
}

func TestSyncSpecSetDefaults(t *testing.T) {
	def := func(spec SyncSpec) SyncSpec {
		spec.SetDefaults("test-image", v1.PullAlways, "test-jwt", "test-client-auth-ca", "test-tls-ca")
		return spec
	}

	assert.False(t, def(SyncSpec{}).IsEnabled())
	assert.False(t, def(SyncSpec{Enabled: util.NewBool(false)}).IsEnabled())
	assert.True(t, def(SyncSpec{Enabled: util.NewBool(true)}).IsEnabled())
	assert.Equal(t, "test-image", def(SyncSpec{}).GetImage())
	assert.Equal(t, "foo", def(SyncSpec{Image: util.NewString("foo")}).GetImage())
	assert.Equal(t, v1.PullAlways, def(SyncSpec{}).GetImagePullPolicy())
	assert.Equal(t, v1.PullNever, def(SyncSpec{ImagePullPolicy: util.NewPullPolicy(v1.PullNever)}).GetImagePullPolicy())
	assert.Equal(t, "test-jwt", def(SyncSpec{}).Authentication.GetJWTSecretName())
	assert.Equal(t, "foo", def(SyncSpec{Authentication: SyncAuthenticationSpec{JWTSecretName: util.NewString("foo")}}).Authentication.GetJWTSecretName())
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
			SyncSpec{Enabled: util.NewBool(false)},
			SyncSpec{Enabled: util.NewBool(true)},
			SyncSpec{Enabled: util.NewBool(true)},
			nil,
		},
		{
			SyncSpec{Enabled: util.NewBool(true)},
			SyncSpec{Enabled: util.NewBool(false)},
			SyncSpec{Enabled: util.NewBool(false)},
			nil,
		},
		{
			SyncSpec{Image: util.NewString("foo")},
			SyncSpec{Image: util.NewString("foo2")},
			SyncSpec{Image: util.NewString("foo2")},
			nil,
		},
		{
			SyncSpec{ImagePullPolicy: util.NewPullPolicy(v1.PullAlways)},
			SyncSpec{ImagePullPolicy: util.NewPullPolicy(v1.PullNever)},
			SyncSpec{ImagePullPolicy: util.NewPullPolicy(v1.PullNever)},
			nil,
		},
		{
			SyncSpec{Authentication: SyncAuthenticationSpec{JWTSecretName: util.NewString("None"), ClientCASecretName: util.NewString("some")}},
			SyncSpec{Authentication: SyncAuthenticationSpec{JWTSecretName: util.NewString("None"), ClientCASecretName: util.NewString("some")}},
			SyncSpec{Authentication: SyncAuthenticationSpec{JWTSecretName: util.NewString("None"), ClientCASecretName: util.NewString("some")}},
			nil,
		},
		{
			SyncSpec{Authentication: SyncAuthenticationSpec{JWTSecretName: util.NewString("foo"), ClientCASecretName: util.NewString("some")}},
			SyncSpec{Authentication: SyncAuthenticationSpec{JWTSecretName: util.NewString("foo"), ClientCASecretName: util.NewString("some")}},
			SyncSpec{Authentication: SyncAuthenticationSpec{JWTSecretName: util.NewString("foo"), ClientCASecretName: util.NewString("some")}},
			nil,
		},
		{
			SyncSpec{Authentication: SyncAuthenticationSpec{JWTSecretName: util.NewString("foo"), ClientCASecretName: util.NewString("some")}},
			SyncSpec{Authentication: SyncAuthenticationSpec{JWTSecretName: util.NewString("foo2"), ClientCASecretName: util.NewString("some")}},
			SyncSpec{Authentication: SyncAuthenticationSpec{JWTSecretName: util.NewString("foo2"), ClientCASecretName: util.NewString("some")}},
			nil,
		},
	}

	for _, test := range tests {
		result := test.Original.ResetImmutableFields("test", &test.Target)
		assert.Equal(t, test.Result, result)
		assert.Equal(t, test.Expected, test.Target)
	}
}
