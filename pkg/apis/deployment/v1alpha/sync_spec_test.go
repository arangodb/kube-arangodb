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
	auth := AuthenticationSpec{XJWTSecretName: util.NewString("foo")}
	tls := TLSSpec{XCASecretName: util.NewString("None")}
	assert.Nil(t, SyncSpec{XImage: util.NewString("foo"), Authentication: auth}.Validate(DeploymentModeSingle))
	assert.Nil(t, SyncSpec{XImage: util.NewString("foo"), Authentication: auth}.Validate(DeploymentModeResilientSingle))
	assert.Nil(t, SyncSpec{XImage: util.NewString("foo"), Authentication: auth}.Validate(DeploymentModeCluster))
	assert.Nil(t, SyncSpec{XImage: util.NewString("foo"), Authentication: auth, TLS: tls, XEnabled: util.NewBool(true)}.Validate(DeploymentModeCluster))

	// Not valid
	assert.Error(t, SyncSpec{XImage: util.NewString(""), Authentication: auth}.Validate(DeploymentModeSingle))
	assert.Error(t, SyncSpec{XImage: util.NewString(""), Authentication: auth}.Validate(DeploymentModeResilientSingle))
	assert.Error(t, SyncSpec{XImage: util.NewString(""), Authentication: auth}.Validate(DeploymentModeCluster))
	assert.Error(t, SyncSpec{XImage: util.NewString("foo"), Authentication: auth, TLS: tls, XEnabled: util.NewBool(true)}.Validate(DeploymentModeSingle))
	assert.Error(t, SyncSpec{XImage: util.NewString("foo"), Authentication: auth, TLS: tls, XEnabled: util.NewBool(true)}.Validate(DeploymentModeResilientSingle))
}

func TestSyncSpecSetDefaults(t *testing.T) {
	def := func(spec SyncSpec) SyncSpec {
		spec.SetDefaults("test-image", v1.PullAlways, "test-jwt", "test-ca")
		return spec
	}

	assert.False(t, def(SyncSpec{}).IsEnabled())
	assert.False(t, def(SyncSpec{XEnabled: util.NewBool(false)}).IsEnabled())
	assert.True(t, def(SyncSpec{XEnabled: util.NewBool(true)}).IsEnabled())
	assert.Equal(t, "test-image", def(SyncSpec{}).GetImage())
	assert.Equal(t, "foo", def(SyncSpec{XImage: util.NewString("foo")}).GetImage())
	assert.Equal(t, v1.PullAlways, def(SyncSpec{}).GetImagePullPolicy())
	assert.Equal(t, v1.PullNever, def(SyncSpec{XImagePullPolicy: util.NewPullPolicy(v1.PullNever)}).GetImagePullPolicy())
	assert.Equal(t, "test-jwt", def(SyncSpec{}).Authentication.GetJWTSecretName())
	assert.Equal(t, "foo", def(SyncSpec{Authentication: AuthenticationSpec{XJWTSecretName: util.NewString("foo")}}).Authentication.GetJWTSecretName())
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
			SyncSpec{XEnabled: util.NewBool(false)},
			SyncSpec{XEnabled: util.NewBool(true)},
			SyncSpec{XEnabled: util.NewBool(true)},
			nil,
		},
		{
			SyncSpec{XEnabled: util.NewBool(true)},
			SyncSpec{XEnabled: util.NewBool(false)},
			SyncSpec{XEnabled: util.NewBool(false)},
			nil,
		},
		{
			SyncSpec{XImage: util.NewString("foo")},
			SyncSpec{XImage: util.NewString("foo2")},
			SyncSpec{XImage: util.NewString("foo2")},
			nil,
		},
		{
			SyncSpec{XImagePullPolicy: util.NewPullPolicy(v1.PullAlways)},
			SyncSpec{XImagePullPolicy: util.NewPullPolicy(v1.PullNever)},
			SyncSpec{XImagePullPolicy: util.NewPullPolicy(v1.PullNever)},
			nil,
		},
		{
			SyncSpec{Authentication: AuthenticationSpec{XJWTSecretName: util.NewString("None")}},
			SyncSpec{Authentication: AuthenticationSpec{XJWTSecretName: util.NewString("None")}},
			SyncSpec{Authentication: AuthenticationSpec{XJWTSecretName: util.NewString("None")}},
			nil,
		},
		{
			SyncSpec{Authentication: AuthenticationSpec{XJWTSecretName: util.NewString("foo")}},
			SyncSpec{Authentication: AuthenticationSpec{XJWTSecretName: util.NewString("foo")}},
			SyncSpec{Authentication: AuthenticationSpec{XJWTSecretName: util.NewString("foo")}},
			nil,
		},
		{
			SyncSpec{Authentication: AuthenticationSpec{XJWTSecretName: util.NewString("foo")}},
			SyncSpec{Authentication: AuthenticationSpec{XJWTSecretName: util.NewString("foo2")}},
			SyncSpec{Authentication: AuthenticationSpec{XJWTSecretName: util.NewString("foo2")}},
			nil,
		},

		// Invalid changes
		{
			SyncSpec{Authentication: AuthenticationSpec{XJWTSecretName: util.NewString("foo")}},
			SyncSpec{Authentication: AuthenticationSpec{XJWTSecretName: util.NewString("None")}},
			SyncSpec{Authentication: AuthenticationSpec{XJWTSecretName: util.NewString("foo")}},
			[]string{"test.auth.jwtSecretName"},
		},
		{
			SyncSpec{Authentication: AuthenticationSpec{XJWTSecretName: util.NewString("None")}},
			SyncSpec{Authentication: AuthenticationSpec{XJWTSecretName: util.NewString("foo")}},
			SyncSpec{Authentication: AuthenticationSpec{XJWTSecretName: util.NewString("None")}},
			[]string{"test.auth.jwtSecretName"},
		},
	}

	for _, test := range tests {
		result := test.Original.ResetImmutableFields("test", &test.Target)
		assert.Equal(t, test.Result, result)
		assert.Equal(t, test.Expected, test.Target)
	}
}
