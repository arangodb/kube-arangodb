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

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func TestAuthenticationSpecValidate(t *testing.T) {
	// Valid
	assert.Nil(t, AuthenticationSpec{JWTSecretName: util.NewType[string]("None")}.Validate(false))
	assert.Nil(t, AuthenticationSpec{JWTSecretName: util.NewType[string]("foo")}.Validate(false))
	assert.Nil(t, AuthenticationSpec{JWTSecretName: util.NewType[string]("foo")}.Validate(true))

	// Not valid
	assert.Error(t, AuthenticationSpec{JWTSecretName: util.NewType[string]("Foo")}.Validate(false))
}

func TestAuthenticationSpecIsAuthenticated(t *testing.T) {
	assert.False(t, AuthenticationSpec{JWTSecretName: util.NewType[string]("None")}.IsAuthenticated())
	assert.True(t, AuthenticationSpec{JWTSecretName: util.NewType[string]("foo")}.IsAuthenticated())
	assert.True(t, AuthenticationSpec{JWTSecretName: util.NewType[string]("")}.IsAuthenticated())
}

func TestAuthenticationSpecSetDefaults(t *testing.T) {
	def := func(spec AuthenticationSpec) AuthenticationSpec {
		spec.SetDefaults("test-jwt")
		return spec
	}

	assert.Equal(t, "test-jwt", def(AuthenticationSpec{}).GetJWTSecretName())
	assert.Equal(t, "foo", def(AuthenticationSpec{JWTSecretName: util.NewType[string]("foo")}).GetJWTSecretName())
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
			AuthenticationSpec{JWTSecretName: util.NewType[string]("None")},
			AuthenticationSpec{JWTSecretName: util.NewType[string]("None")},
			AuthenticationSpec{JWTSecretName: util.NewType[string]("None")},
			nil,
		},
		{
			AuthenticationSpec{JWTSecretName: util.NewType[string]("foo")},
			AuthenticationSpec{JWTSecretName: util.NewType[string]("foo")},
			AuthenticationSpec{JWTSecretName: util.NewType[string]("foo")},
			nil,
		},
		{
			AuthenticationSpec{JWTSecretName: util.NewType[string]("foo")},
			AuthenticationSpec{JWTSecretName: util.NewType[string]("foo2")},
			AuthenticationSpec{JWTSecretName: util.NewType[string]("foo2")},
			nil,
		},

		// Invalid changes
		{
			AuthenticationSpec{JWTSecretName: util.NewType[string]("foo")},
			AuthenticationSpec{JWTSecretName: util.NewType[string]("None")},
			AuthenticationSpec{JWTSecretName: util.NewType[string]("foo")},
			[]string{"test.jwtSecretName"},
		},
		{
			AuthenticationSpec{JWTSecretName: util.NewType[string]("None")},
			AuthenticationSpec{JWTSecretName: util.NewType[string]("foo")},
			AuthenticationSpec{JWTSecretName: util.NewType[string]("None")},
			[]string{"test.jwtSecretName"},
		},
	}

	for _, test := range tests {
		result := test.Original.ResetImmutableFields("test", &test.Target)
		assert.Equal(t, test.Result, result)
		assert.Equal(t, test.Expected, test.Target)
	}
}
