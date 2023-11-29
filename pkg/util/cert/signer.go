//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package cert

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
)

type Signer struct {
	privateKey *rsa.PrivateKey
}

// NewSigner creates a new Signer with a generated private key.
func NewSigner() (*Signer, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return &Signer{privateKey: privateKey}, nil
}

// Sign signs the content with the private key and returns:
// base64 encoded signature, base64 encoded content and error.
func (s *Signer) Sign(content string) (string, string, error) {
	hash := sha256.New()
	hash.Write([]byte(content))
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hash.Sum(nil))
	if err != nil {
		return "", "", err
	}
	return base64.StdEncoding.EncodeToString(signature), base64.StdEncoding.EncodeToString([]byte(content)), nil
}

// PublicKey returns the public key in PKIX format.
func (s *Signer) PublicKey() (string, error) {
	publicKey := &s.privateKey.PublicKey
	publicKeyDer, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", err
	}

	publicKeyBlock := pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyDer,
	}
	return string(pem.EncodeToMemory(&publicKeyBlock)), nil
}
