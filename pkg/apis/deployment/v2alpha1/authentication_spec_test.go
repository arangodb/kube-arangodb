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
// Author Ewout Prangsma
//

package v2alpha1

import (
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticationSpecValidate(t *testing.T) {
	// Valid
	assert.Nil(t, AuthenticationSpec{JWTSecretName: util.NewString("None")}.Validate(false))
	assert.Nil(t, AuthenticationSpec{JWTSecretName: util.NewString("foo")}.Validate(false))
	assert.Nil(t, AuthenticationSpec{JWTSecretName: util.NewString("foo")}.Validate(true))

	// Not valid
	assert.Error(t, AuthenticationSpec{JWTSecretName: util.NewString("Foo")}.Validate(false))
}

func TestAuthenticationSpecIsAuthenticated(t *testing.T) {
	assert.False(t, AuthenticationSpec{JWTSecretName: util.NewString("None")}.IsAuthenticated())
	assert.True(t, AuthenticationSpec{JWTSecretName: util.NewString("foo")}.IsAuthenticated())
	assert.True(t, AuthenticationSpec{JWTSecretName: util.NewString("")}.IsAuthenticated())
}

func TestAuthenticationSpecSetDefaults(t *testing.T) {
	def := func(spec AuthenticationSpec) AuthenticationSpec {
		spec.SetDefaults("test-jwt")
		return spec
	}

	assert.Equal(t, "test-jwt", def(AuthenticationSpec{}).GetJWTSecretName())
	assert.Equal(t, "foo", def(AuthenticationSpec{JWTSecretName: util.NewString("foo")}).GetJWTSecretName())
}

func TestAuthenticationSpecResetImmutableFields(t *testing.T) {
	tests := []struct {
		Original AuthenticationSpec
		Target   AuthenticationSpec
		Expected AuthenticationSpec
		Result   []string
	}{
		// Valid "changes"
		{
			AuthenticationSpec{JWTSecretName: util.NewString("None")},
			AuthenticationSpec{JWTSecretName: util.NewString("None")},
			AuthenticationSpec{JWTSecretName: util.NewString("None")},
			nil,
		},
		{
			AuthenticationSpec{JWTSecretName: util.NewString("foo")},
			AuthenticationSpec{JWTSecretName: util.NewString("foo")},
			AuthenticationSpec{JWTSecretName: util.NewString("foo")},
			nil,
		},
		{
			AuthenticationSpec{JWTSecretName: util.NewString("foo")},
			AuthenticationSpec{JWTSecretName: util.NewString("foo2")},
			AuthenticationSpec{JWTSecretName: util.NewString("foo2")},
			nil,
		},

		// Invalid changes
		{
			AuthenticationSpec{JWTSecretName: util.NewString("foo")},
			AuthenticationSpec{JWTSecretName: util.NewString("None")},
			AuthenticationSpec{JWTSecretName: util.NewString("foo")},
			[]string{"test.jwtSecretName"},
		},
		{
			AuthenticationSpec{JWTSecretName: util.NewString("None")},
			AuthenticationSpec{JWTSecretName: util.NewString("foo")},
			AuthenticationSpec{JWTSecretName: util.NewString("None")},
			[]string{"test.jwtSecretName"},
		},
	}

	for _, test := range tests {
		result := test.Original.ResetImmutableFields("test", &test.Target)
		assert.Equal(t, test.Result, result)
		assert.Equal(t, test.Expected, test.Target)
	}
}
