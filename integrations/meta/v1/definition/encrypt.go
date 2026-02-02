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

package definition

import (
	"encoding/base64"
	"encoding/json"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/crypto"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func IsEncryptedObject(in *anypb.Any) bool {
	return in.MessageIs(&EncryptedObject{})
}

func (o *ObjectSecret) Encrypt(in *anypb.Any) (*anypb.Any, error) {
	if t := o.GetToken(); t != nil {
		return t.Encrypt(in)
	}

	return in, nil
}

func (o *ObjectSecretToken) Encrypt(in *anypb.Any) (*anypb.Any, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	encrypted, err := crypto.EncryptionKey(o.GetToken()).Encrypt(data)
	if err != nil {
		return nil, err
	}

	return anypb.New(&EncryptedObject{
		Object: base64.StdEncoding.EncodeToString(encrypted),
		Hash:   util.SHA256(data),
	})
}

func (o *ObjectSecret) Decrypt(in *anypb.Any) (*anypb.Any, error) {
	if t := o.GetToken(); t != nil {
		return t.Decrypt(in)
	}

	if IsEncryptedObject(in) {
		return nil, errors.Errorf("Object encrypted, but secret is missing")
	}

	return in, nil
}

func (o *ObjectSecretToken) Decrypt(in *anypb.Any) (*anypb.Any, error) {
	if !IsEncryptedObject(in) {
		return nil, errors.Errorf("Object is not encrypted, but secret provided")
	}

	var z EncryptedObject
	if err := in.UnmarshalTo(&z); err != nil {
		return nil, err
	}

	data, err := base64.StdEncoding.DecodeString(z.Object)
	if err != nil {
		return nil, err
	}

	decrypted, err := crypto.EncryptionKey(o.GetToken()).Decrypt(data)
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
