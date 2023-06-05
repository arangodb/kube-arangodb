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
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// RocksDBEncryptionSpec holds rocksdb encryption at rest specific configuration settings
type RocksDBEncryptionSpec struct {
	KeySecretName *string `json:"keySecretName,omitempty"`
}

// GetKeySecretName returns the value of keySecretName.
func (s RocksDBEncryptionSpec) GetKeySecretName() string {
	return util.TypeOrDefault[string](s.KeySecretName)
}

// IsEncrypted returns true when an encryption key secret name is provided,
// false otherwise.
func (s RocksDBEncryptionSpec) IsEncrypted() bool {
	return s.GetKeySecretName() != ""
}

// RocksDBSpec holds rocksdb specific configuration settings
type RocksDBSpec struct {
	Encryption RocksDBEncryptionSpec `json:"encryption"`
}

// IsEncrypted returns true when an encryption key secret name is provided,
// false otherwise.
func (s RocksDBSpec) IsEncrypted() bool {
	return s.Encryption.IsEncrypted()
}

// Validate the given spec
func (s RocksDBSpec) Validate() error {
	if err := shared.ValidateOptionalResourceName(s.Encryption.GetKeySecretName()); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *RocksDBSpec) SetDefaults() {
	// Nothing needed
}

// SetDefaultsFrom fills unspecified fields with a value from given source spec.
func (s *RocksDBSpec) SetDefaultsFrom(source RocksDBSpec) {
	if s.Encryption.KeySecretName == nil {
		s.Encryption.KeySecretName = util.NewTypeOrNil[string](source.Encryption.KeySecretName)
	}
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to given field prefix.
func (s RocksDBSpec) ResetImmutableFields(fieldPrefix string, target *RocksDBSpec) []string {
	var resetFields []string
	if s.IsEncrypted() != target.IsEncrypted() {
		// Note: You can change the name, but not from empty to non-empty (or reverse).
		target.Encryption.KeySecretName = util.NewTypeOrNil[string](s.Encryption.KeySecretName)
		resetFields = append(resetFields, fieldPrefix+".encryption.keySecretName")
	}
	return resetFields
}
