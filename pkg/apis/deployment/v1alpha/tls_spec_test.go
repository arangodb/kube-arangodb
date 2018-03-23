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

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestTLSSpecValidate(t *testing.T) {
	// Valid
	assert.Nil(t, TLSSpec{XCASecretName: util.NewString("foo")}.Validate())
	assert.Nil(t, TLSSpec{XCASecretName: util.NewString("None")}.Validate())
	assert.Nil(t, TLSSpec{XCASecretName: util.NewString("None"), AltNames: []string{}}.Validate())
	assert.Nil(t, TLSSpec{XCASecretName: util.NewString("None"), AltNames: []string{"foo"}}.Validate())
	assert.Nil(t, TLSSpec{XCASecretName: util.NewString("None"), AltNames: []string{"email@example.com", "127.0.0.1"}}.Validate())

	// Not valid
	assert.Error(t, TLSSpec{XCASecretName: nil}.Validate())
	assert.Error(t, TLSSpec{XCASecretName: util.NewString("")}.Validate())
	assert.Error(t, TLSSpec{XCASecretName: util.NewString("Foo")}.Validate())
	assert.Error(t, TLSSpec{XCASecretName: util.NewString("foo"), AltNames: []string{"@@"}}.Validate())
}

func TestTLSSpecIsSecure(t *testing.T) {
	assert.True(t, TLSSpec{XCASecretName: util.NewString("")}.IsSecure())
	assert.True(t, TLSSpec{XCASecretName: util.NewString("foo")}.IsSecure())
	assert.False(t, TLSSpec{XCASecretName: util.NewString("None")}.IsSecure())
}

func TestTLSSpecSetDefaults(t *testing.T) {
	def := func(spec TLSSpec) TLSSpec {
		spec.SetDefaults("")
		return spec
	}

	assert.Equal(t, "", def(TLSSpec{}).GetCASecretName())
	assert.Equal(t, "foo", def(TLSSpec{XCASecretName: util.NewString("foo")}).GetCASecretName())
	assert.Len(t, def(TLSSpec{}).GetAltNames(), 0)
	assert.Len(t, def(TLSSpec{AltNames: []string{"foo.local"}}).GetAltNames(), 1)
	assert.Equal(t, defaultTLSTTL, def(TLSSpec{}).GetTTL())
	assert.Equal(t, time.Hour, def(TLSSpec{XTTL: util.NewDuration(time.Hour)}).GetTTL())
}
