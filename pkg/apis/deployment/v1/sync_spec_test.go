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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func TestSyncSpecValidate(t *testing.T) {
	// Valid
	auth := SyncAuthenticationSpec{JWTSecretName: util.NewType[string]("foo"), ClientCASecretName: util.NewType[string]("foo-client")}
	tls := TLSSpec{CASecretName: util.NewType[string]("None")}
	assert.Nil(t, SyncSpec{Authentication: auth}.Validate(DeploymentModeSingle))
	assert.Nil(t, SyncSpec{Authentication: auth}.Validate(DeploymentModeActiveFailover))
	assert.Nil(t, SyncSpec{Authentication: auth}.Validate(DeploymentModeCluster))
	assert.Nil(t, SyncSpec{Authentication: auth, TLS: tls, Enabled: util.NewType[bool](true)}.Validate(DeploymentModeCluster))

	// Not valid
	assert.Error(t, SyncSpec{Authentication: auth, TLS: tls, Enabled: util.NewType[bool](true)}.Validate(DeploymentModeSingle))
	assert.Error(t, SyncSpec{Authentication: auth, TLS: tls, Enabled: util.NewType[bool](true)}.Validate(DeploymentModeActiveFailover))
}

func TestSyncSpecSetDefaults(t *testing.T) {
	def := func(spec SyncSpec) SyncSpec {
		spec.SetDefaults("test-jwt", "test-client-auth-ca", "test-tls-ca", "test-mon")
		return spec
	}

	assert.False(t, def(SyncSpec{}).IsEnabled())
	assert.False(t, def(SyncSpec{Enabled: util.NewType[bool](false)}).IsEnabled())
	assert.True(t, def(SyncSpec{Enabled: util.NewType[bool](true)}).IsEnabled())
	assert.Equal(t, "test-jwt", def(SyncSpec{}).Authentication.GetJWTSecretName())
	assert.Equal(t, "test-mon", def(SyncSpec{}).Monitoring.GetTokenSecretName())
	assert.Equal(t, "foo", def(SyncSpec{Authentication: SyncAuthenticationSpec{JWTSecretName: util.NewType[string]("foo")}}).Authentication.GetJWTSecretName())
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
			SyncSpec{Enabled: util.NewType[bool](false)},
			SyncSpec{Enabled: util.NewType[bool](true)},
			SyncSpec{Enabled: util.NewType[bool](true)},
			nil,
		},
		{
			SyncSpec{Enabled: util.NewType[bool](true)},
			SyncSpec{Enabled: util.NewType[bool](false)},
			SyncSpec{Enabled: util.NewType[bool](false)},
			nil,
		},
		{
			SyncSpec{Authentication: SyncAuthenticationSpec{JWTSecretName: util.NewType[string]("None"), ClientCASecretName: util.NewType[string]("some")}},
			SyncSpec{Authentication: SyncAuthenticationSpec{JWTSecretName: util.NewType[string]("None"), ClientCASecretName: util.NewType[string]("some")}},
			SyncSpec{Authentication: SyncAuthenticationSpec{JWTSecretName: util.NewType[string]("None"), ClientCASecretName: util.NewType[string]("some")}},
			nil,
		},
		{
			SyncSpec{Authentication: SyncAuthenticationSpec{JWTSecretName: util.NewType[string]("foo"), ClientCASecretName: util.NewType[string]("some")}},
			SyncSpec{Authentication: SyncAuthenticationSpec{JWTSecretName: util.NewType[string]("foo"), ClientCASecretName: util.NewType[string]("some")}},
			SyncSpec{Authentication: SyncAuthenticationSpec{JWTSecretName: util.NewType[string]("foo"), ClientCASecretName: util.NewType[string]("some")}},
			nil,
		},
		{
			SyncSpec{Authentication: SyncAuthenticationSpec{JWTSecretName: util.NewType[string]("foo"), ClientCASecretName: util.NewType[string]("some")}},
			SyncSpec{Authentication: SyncAuthenticationSpec{JWTSecretName: util.NewType[string]("foo2"), ClientCASecretName: util.NewType[string]("some")}},
			SyncSpec{Authentication: SyncAuthenticationSpec{JWTSecretName: util.NewType[string]("foo2"), ClientCASecretName: util.NewType[string]("some")}},
			nil,
		},
	}

	for _, test := range tests {
		result := test.Original.ResetImmutableFields("test", &test.Target)
		assert.Equal(t, test.Result, result)
		assert.Equal(t, test.Expected, test.Target)
	}
}
