//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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

func TestRocksDBSpecValidate(t *testing.T) {
	// Valid
	assert.Nil(t, RocksDBSpec{}.Validate())
	assert.Nil(t, RocksDBSpec{Encryption: RocksDBEncryptionSpec{KeySecretName: util.NewString("foo")}}.Validate())

	// Not valid
	assert.Error(t, RocksDBSpec{Encryption: RocksDBEncryptionSpec{KeySecretName: util.NewString("Foo")}}.Validate())
}

func TestRocksDBSpecIsEncrypted(t *testing.T) {
	assert.False(t, RocksDBSpec{}.IsEncrypted())
	assert.False(t, RocksDBSpec{Encryption: RocksDBEncryptionSpec{KeySecretName: util.NewString("")}}.IsEncrypted())
	assert.True(t, RocksDBSpec{Encryption: RocksDBEncryptionSpec{KeySecretName: util.NewString("foo")}}.IsEncrypted())
}

func TestRocksDBSpecSetDefaults(t *testing.T) {
	def := func(spec RocksDBSpec) RocksDBSpec {
		spec.SetDefaults()
		return spec
	}

	assert.Equal(t, "", def(RocksDBSpec{}).Encryption.GetKeySecretName())
}

func TestRocksDBSpecResetImmutableFields(t *testing.T) {
	tests := []struct {
		Original RocksDBSpec
		Target   RocksDBSpec
		Expected RocksDBSpec
		Result   []string
	}{
		// Valid "changes"
		{
			RocksDBSpec{},
			RocksDBSpec{},
			RocksDBSpec{},
			nil,
		},
		{
			RocksDBSpec{Encryption: RocksDBEncryptionSpec{KeySecretName: util.NewString("foo")}},
			RocksDBSpec{Encryption: RocksDBEncryptionSpec{KeySecretName: util.NewString("foo")}},
			RocksDBSpec{Encryption: RocksDBEncryptionSpec{KeySecretName: util.NewString("foo")}},
			nil,
		},
		{
			RocksDBSpec{Encryption: RocksDBEncryptionSpec{KeySecretName: util.NewString("foo")}},
			RocksDBSpec{Encryption: RocksDBEncryptionSpec{KeySecretName: util.NewString("foo2")}},
			RocksDBSpec{Encryption: RocksDBEncryptionSpec{KeySecretName: util.NewString("foo2")}},
			nil,
		},

		// Invalid changes
		{
			RocksDBSpec{Encryption: RocksDBEncryptionSpec{KeySecretName: util.NewString("foo")}},
			RocksDBSpec{Encryption: RocksDBEncryptionSpec{KeySecretName: util.NewString("")}},
			RocksDBSpec{Encryption: RocksDBEncryptionSpec{KeySecretName: util.NewString("foo")}},
			[]string{"test.encryption.keySecretName"},
		},
	}

	for _, test := range tests {
		result := test.Original.ResetImmutableFields("test", &test.Target)
		assert.Equal(t, test.Result, result)
		assert.Equal(t, test.Expected, test.Target)
	}
}
