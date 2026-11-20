//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
)

type EncryptionKey string

func (ek EncryptionKey) Key() []byte {
	v := sha256.Sum256([]byte(ek))
	return v[:]
}

func (ek EncryptionKey) Encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(ek.Key())
	if err != nil {
		return nil, err
	}

	// NewGCMWithRandomNonce generates the nonce inside the crypto module and prepends it to the
	// ciphertext, which is required under GODEBUG=fips140=only. The on-the-wire layout
	// (12-byte nonce || ciphertext || 16-byte tag) is identical to the previous NewGCM usage with a
	// manually prepended random nonce, so data encrypted by older operators stays decryptable.
	gcm, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nil, nil, data, nil), nil
}

func (ek EncryptionKey) Decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(ek.Key())
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nil, data, nil)
}
