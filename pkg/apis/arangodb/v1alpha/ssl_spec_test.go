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
)

func TestSSLSpecValidate(t *testing.T) {
	// Valid
	assert.Nil(t, SSLSpec{KeySecretName: ""}.Validate())
	assert.Nil(t, SSLSpec{KeySecretName: "foo"}.Validate())

	// Not valid
	assert.Error(t, SSLSpec{KeySecretName: "Foo"}.Validate())
}

func TestSSLSpecIsSecure(t *testing.T) {
	assert.False(t, SSLSpec{KeySecretName: ""}.IsSecure())
	assert.True(t, SSLSpec{KeySecretName: "foo"}.IsSecure())
}

func TestSSLSpecSetDefaults(t *testing.T) {
	def := func(spec SSLSpec) SSLSpec {
		spec.SetDefaults()
		return spec
	}

	assert.Equal(t, "", def(SSLSpec{}).KeySecretName)
	assert.Equal(t, "foo", def(SSLSpec{KeySecretName: "foo"}).KeySecretName)
	assert.Equal(t, "ArangoDB", def(SSLSpec{}).OrganizationName)
	assert.Equal(t, "foo", def(SSLSpec{OrganizationName: "foo"}).OrganizationName)
	assert.Equal(t, "", def(SSLSpec{}).ServerName)
	assert.Equal(t, "foo", def(SSLSpec{ServerName: "foo"}).ServerName)
}
