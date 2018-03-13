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
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTLSSpecValidate(t *testing.T) {
	// Valid
	assert.Nil(t, TLSSpec{CASecretName: ""}.Validate())
	assert.Nil(t, TLSSpec{CASecretName: "foo"}.Validate())
	assert.Nil(t, TLSSpec{AltNames: []string{}}.Validate())
	assert.Nil(t, TLSSpec{AltNames: []string{"foo"}}.Validate())
	assert.Nil(t, TLSSpec{AltNames: []string{"email@example.com", "127.0.0.1"}}.Validate())

	// Not valid
	assert.Error(t, TLSSpec{CASecretName: "Foo"}.Validate())
	assert.Error(t, TLSSpec{AltNames: []string{"@@"}}.Validate())
}

func TestTLSSpecIsSecure(t *testing.T) {
	assert.False(t, TLSSpec{CASecretName: ""}.IsSecure())
	assert.True(t, TLSSpec{CASecretName: "foo"}.IsSecure())
}

func TestTLSSpecSetDefaults(t *testing.T) {
	def := func(spec TLSSpec) TLSSpec {
		spec.SetDefaults("")
		return spec
	}

	assert.Equal(t, "", def(TLSSpec{}).CASecretName)
	assert.Equal(t, "foo", def(TLSSpec{CASecretName: "foo"}).CASecretName)
	assert.Len(t, def(TLSSpec{}).AltNames, 0)
	assert.Len(t, def(TLSSpec{AltNames: []string{"foo.local"}}).AltNames, 1)
	assert.Equal(t, defaultTLSTTL, def(TLSSpec{}).TTL)
	assert.Equal(t, time.Hour, def(TLSSpec{TTL: time.Hour}).TTL)
}
