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

package v2alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func TestSyncSpecValidate(t *testing.T) {
	// Valid
	auth := SyncAuthenticationSpec{JWTSecretName: util.NewString("foo"), ClientCASecretName: util.NewString("foo-client")}
	tls := TLSSpec{CASecretName: util.NewString("None")}
	assert.Nil(t, SyncSpec{Authentication: auth}.Validate(DeploymentModeSingle))
	assert.Nil(t, SyncSpec{Authentication: auth}.Validate(DeploymentModeActiveFailover))
	assert.Nil(t, SyncSpec{Authentication: auth}.Validate(DeploymentModeCluster))
	assert.Nil(t, SyncSpec{Authentication: auth, TLS: tls, Enabled: util.NewBool(true)}.Validate(DeploymentModeCluster))

	// Not valid
	assert.Error(t, SyncSpec{Authentication: auth, TLS: tls, Enabled: util.NewBool(true)}.Validate(DeploymentModeSingle))
	assert.Error(t, SyncSpec{Authentication: auth, TLS: tls, Enabled: util.NewBool(true)}.Validate(DeploymentModeActiveFailover))
}

func TestSyncSpecSetDefaults(t *testing.T) {
	def := func(spec SyncSpec) SyncSpec {
		spec.SetDefaults("test-jwt", "test-client-auth-ca", "test-tls-ca", "test-mon")
		return spec
	}

	assert.False(t, def(SyncSpec{}).IsEnabled())
	assert.False(t, def(SyncSpec{Enabled: util.NewBool(false)}).IsEnabled())
	assert.True(t, def(SyncSpec{Enabled: util.NewBool(true)}).IsEnabled())
	assert.Equal(t, "test-jwt", def(SyncSpec{}).Authentication.GetJWTSecretName())
	assert.Equal(t, "test-mon", def(SyncSpec{}).Monitoring.GetTokenSecretName())
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
