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
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_EncryptionKey_RoundTrip(t *testing.T) {
	const key EncryptionKey = "test-key"
	data := []byte("arangodb-secret-token")

	encrypted, err := key.Encrypt(data)
	require.NoError(t, err)
	require.NotEqual(t, data, encrypted)

	decrypted, err := key.Decrypt(encrypted)
	require.NoError(t, err)
	require.Equal(t, data, decrypted)
}

func Test_EncryptionKey_WrongKeyFails(t *testing.T) {
	encrypted, err := EncryptionKey("test-key").Encrypt([]byte("arangodb-secret-token"))
	require.NoError(t, err)

	_, err = EncryptionKey("other-key").Decrypt(encrypted)
	require.Error(t, err)
}

func Test_EncryptionKey_TruncatedInputFails(t *testing.T) {
	encrypted, err := EncryptionKey("test-key").Encrypt([]byte("arangodb-secret-token"))
	require.NoError(t, err)

	_, err = EncryptionKey("test-key").Decrypt(encrypted[:len(encrypted)-5])
	require.Error(t, err)

	_, err = EncryptionKey("test-key").Decrypt(nil)
	require.Error(t, err)
}

// Test_EncryptionKey_DecryptLegacyFormat proves backward compatibility: a ciphertext produced by the
// previous implementation (cipher.NewGCM with a manually prepended random nonce) is still decryptable
// after switching to cipher.NewGCMWithRandomNonce, so encrypted data written by older operators is
// readable across the upgrade without any migration. The fixture was generated with the old code for
// EncryptionKey("test-key") over the plaintext below.
func Test_EncryptionKey_DecryptLegacyFormat(t *testing.T) {
	const key EncryptionKey = "test-key"
	legacyHex := "1e25b981a87f5652cfebf10ae7e2ecfb76cbdde28f312958ec5f2770db8e9e63cbeb5865b41f1d7da62ac5c18f04d7f3ef"

	legacy, err := hex.DecodeString(legacyHex)
	require.NoError(t, err)

	decrypted, err := key.Decrypt(legacy)
	require.NoError(t, err)
	require.Equal(t, "arangodb-secret-token", string(decrypted))
}
