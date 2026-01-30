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

package v1

import (
	"encoding/base64"
	"encoding/json"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/crypto"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func NewEncryptObject(in *anypb.Any, secret crypto.EncryptionKey) (*anypb.Any, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	encrypted, err := secret.Encrypt(data)
	if err != nil {
		return nil, err
	}

	return anypb.New(&EncryptedObject{
		Object: base64.StdEncoding.EncodeToString(encrypted),
		Hash:   util.SHA256(data),
	})
}

func NewDecryptObject(in *anypb.Any, secret crypto.EncryptionKey) (*anypb.Any, error) {
	var z EncryptedObject
	if err := in.UnmarshalTo(&z); err != nil {
		return nil, err
	}

	data, err := base64.StdEncoding.DecodeString(z.Object)
	if err != nil {
		return nil, err
	}

	decrypted, err := secret.Decrypt(data)
	if err != nil {
		return nil, err
	}

	if util.SHA256(decrypted) != z.Hash {
		return nil, errors.Errorf("Invalid encrypted object")
	}

	var a anypb.Any
	if err := json.Unmarshal(decrypted, &a); err != nil {
		return nil, err
	}

	return &a, nil
}
