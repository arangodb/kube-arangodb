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

package v1

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestSecretHashes_Equal(t *testing.T) {
	// Arrange
	sh := SecretHashes{}
	testCases := []struct {
		Name        string
		CompareFrom *SecretHashes
		CompareTo   *SecretHashes
		Expected    bool
	}{
		{
			Name:        "Parameter can not be nil",
			CompareFrom: &SecretHashes{},
			Expected:    false,
		},
		{
			Name:        "The addresses are the same",
			CompareFrom: &sh,
			CompareTo:   &sh,
			Expected:    true,
		},
		{
			Name: "JWT token is different",
			CompareFrom: &SecretHashes{
				AuthJWT: "1",
			},
			CompareTo: &SecretHashes{
				AuthJWT: "2",
			},
			Expected: false,
		},
		{
			Name: "Users are different",
			CompareFrom: &SecretHashes{
				Users: map[string]string{
					"root": "",
				},
			},
			CompareTo: &SecretHashes{},
			Expected:  false,
		},
		{
			Name: "User's table size is different",
			CompareFrom: &SecretHashes{
				Users: map[string]string{
					"root": "",
				},
			},
			CompareTo: &SecretHashes{
				Users: map[string]string{
					"root": "",
					"user": "",
				},
			},
			Expected: false,
		},
		{
			Name: "User's table has got different users",
			CompareFrom: &SecretHashes{
				Users: map[string]string{
					"root": "",
				},
			},
			CompareTo: &SecretHashes{
				Users: map[string]string{
					"user": "",
				},
			},
			Expected: false,
		},
		{
			Name: "User's table has got different hashes for users",
			CompareFrom: &SecretHashes{
				Users: map[string]string{
					"root": "123",
				},
			},
			CompareTo: &SecretHashes{
				Users: map[string]string{
					"root": "1234",
				},
			},
			Expected: false,
		},
		{
			Name: "Secret hashes are the same",
			CompareFrom: &SecretHashes{
				AuthJWT:              "1",
				RocksDBEncryptionKey: "2",
				TLSCA:                "3",
				SyncTLSCA:            "4",
				Users: map[string]string{
					"root": "123",
				},
			},
			CompareTo: &SecretHashes{
				AuthJWT:              "1",
				RocksDBEncryptionKey: "2",
				TLSCA:                "3",
				SyncTLSCA:            "4",
				Users: map[string]string{
					"root": "123",
				},
			},
			Expected: true,
		},
		{
			Name: "Secret hashes are the same without users",
			CompareFrom: &SecretHashes{
				AuthJWT:              "1",
				RocksDBEncryptionKey: "2",
				TLSCA:                "3",
				SyncTLSCA:            "4",
			},
			CompareTo: &SecretHashes{
				AuthJWT:              "1",
				RocksDBEncryptionKey: "2",
				TLSCA:                "3",
				SyncTLSCA:            "4",
			},
			Expected: true,
		},
	}

	for _, testCase := range testCases {
		//nolint:scopelint
		t.Run(testCase.Name, func(t *testing.T) {
			// Act
			expected := testCase.CompareFrom.Equal(testCase.CompareTo)

			// Assert
			assert.Equal(t, testCase.Expected, expected)
		})
	}
}
