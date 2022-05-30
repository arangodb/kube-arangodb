//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

// SecretHashes keeps track of the value of secrets
// so we can detect changes.
// For each used secret, a sha256 hash is stored.
type SecretHashes struct {
	// AuthJWT contains the hash of the auth.jwtSecretName secret
	AuthJWT string `json:"auth-jwt,omitempty"`
	// RocksDBEncryptionKey contains the hash of the rocksdb.encryption.keySecretName secret
	RocksDBEncryptionKey string `json:"rocksdb-encryption-key,omitempty"`
	// TLSCA contains the hash of the tls.caSecretName secret
	TLSCA string `json:"tls-ca,omitempty"`
	// SyncTLSCA contains the hash of the sync.tls.caSecretName secret
	SyncTLSCA string `json:"sync-tls-ca,omitempty"`
	// User's map contains hashes for each user
	Users map[string]string `json:"users,omitempty"`
}

// Equal compares two SecretHashes
func (sh *SecretHashes) Equal(other *SecretHashes) bool {
	if sh == nil && other == nil {
		return true
	} else if sh == nil || other == nil {
		return false
	}

	return sh.AuthJWT == other.AuthJWT &&
		sh.RocksDBEncryptionKey == other.RocksDBEncryptionKey &&
		sh.TLSCA == other.TLSCA &&
		sh.SyncTLSCA == other.SyncTLSCA &&
		isStringMapEqual(sh.Users, other.Users)
}

// NewEmptySecretHashes creates new empty structure
func NewEmptySecretHashes() *SecretHashes {
	sh := &SecretHashes{}
	sh.Users = make(map[string]string)
	return sh
}

func isStringMapEqual(first map[string]string, second map[string]string) bool {
	if first == nil && second == nil {
		return true
	}

	if first == nil || second == nil {
		return false
	}

	if len(first) != len(second) {
		return false
	}

	for key, valueF := range first {
		valueS, ok := second[key]
		if !ok {
			return false
		}

		if valueF != valueS {
			return false
		}
	}

	return true
}
