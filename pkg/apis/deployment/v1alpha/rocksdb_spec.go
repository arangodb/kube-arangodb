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
	"github.com/arangodb/k8s-operator/pkg/util/k8sutil"
)

// RocksDBEncryptionSpec holds rocksdb encryption at rest specific configuration settings
type RocksDBEncryptionSpec struct {
	KeySecretName string `json:"keySecretName,omitempty"`
}

// RocksDBSpec holds rocksdb specific configuration settings
type RocksDBSpec struct {
	Encryption RocksDBEncryptionSpec `json:"encryption"`
}

// IsEncrypted returns true when an encryption key secret name is provided,
// false otherwise.
func (s RocksDBSpec) IsEncrypted() bool {
	return s.Encryption.KeySecretName != ""
}

// Validate the given spec
func (s RocksDBSpec) Validate() error {
	if err := k8sutil.ValidateOptionalResourceName(s.Encryption.KeySecretName); err != nil {
		return maskAny(err)
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *RocksDBSpec) SetDefaults() {
	// Nothing needed
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to given field prefix.
func (s RocksDBSpec) ResetImmutableFields(fieldPrefix string, target *RocksDBSpec) []string {
	var resetFields []string
	if s.IsEncrypted() != target.IsEncrypted() {
		// Note: You can change the name, but not from empty to non-empty (or reverse).
		target.Encryption.KeySecretName = s.Encryption.KeySecretName
		resetFields = append(resetFields, fieldPrefix+".encryption.keySecretName")
	}
	return resetFields
}
