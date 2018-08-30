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

package k8sutil

import (
	"fmt"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
)

// SecretInterface has methods to work with Secret resources.
type SecretInterface interface {
	Create(*v1.Secret) (*v1.Secret, error)
	Get(name string, options metav1.GetOptions) (*v1.Secret, error)
}

// ValidateEncryptionKeySecret checks that a secret with given name in given namespace
// exists and it contains a 'key' data field of exactly 32 bytes.
func ValidateEncryptionKeySecret(secrets SecretInterface, secretName string) error {
	s, err := secrets.Get(secretName, metav1.GetOptions{})
	if err != nil {
		return maskAny(err)
	}
	// Check `key` field
	keyData, found := s.Data[constants.SecretEncryptionKey]
	if !found {
		return maskAny(fmt.Errorf("No '%s' found in secret '%s'", constants.SecretEncryptionKey, secretName))
	}
	if len(keyData) != 32 {
		return maskAny(fmt.Errorf("'%s' in secret '%s' is expected to be 32 bytes long, found %d", constants.SecretEncryptionKey, secretName, len(keyData)))
	}
	return nil
}

// CreateEncryptionKeySecret creates a secret used to store a RocksDB encryption key.
func CreateEncryptionKeySecret(secrets SecretInterface, secretName string, key []byte) error {
	if len(key) != 32 {
		return maskAny(fmt.Errorf("Key in secret '%s' is expected to be 32 bytes long, got %d", secretName, len(key)))
	}
	// Create secret
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			constants.SecretEncryptionKey: key,
		},
	}
	if _, err := secrets.Create(secret); err != nil {
		// Failed to create secret
		return maskAny(err)
	}
	return nil
}

// ValidateCACertificateSecret checks that a secret with given name in given namespace
// exists and it contains a 'ca.crt' data field.
func ValidateCACertificateSecret(secrets SecretInterface, secretName string) error {
	s, err := secrets.Get(secretName, metav1.GetOptions{})
	if err != nil {
		return maskAny(err)
	}
	// Check `ca.crt` field
	_, found := s.Data[constants.SecretCACertificate]
	if !found {
		return maskAny(fmt.Errorf("No '%s' found in secret '%s'", constants.SecretCACertificate, secretName))
	}
	return nil
}

// GetCACertficateSecret loads a secret with given name in the given namespace
// and extracts the `ca.crt` field.
// If the secret does not exists the field is missing,
// an error is returned.
// Returns: certificate, error
func GetCACertficateSecret(secrets SecretInterface, secretName string) (string, error) {
	s, err := secrets.Get(secretName, metav1.GetOptions{})
	if err != nil {
		return "", maskAny(err)
	}
	// Load `ca.crt` field
	cert, found := s.Data[constants.SecretCACertificate]
	if !found {
		return "", maskAny(fmt.Errorf("No '%s' found in secret '%s'", constants.SecretCACertificate, secretName))
	}
	return string(cert), nil
}

// GetCASecret loads a secret with given name in the given namespace
// and extracts the `ca.crt` & `ca.key` field.
// If the secret does not exists or one of the fields is missing,
// an error is returned.
// Returns: certificate, private-key, isOwnedByDeployment, error
func GetCASecret(secrets SecretInterface, secretName string, ownerRef *metav1.OwnerReference) (string, string, bool, error) {
	s, err := secrets.Get(secretName, metav1.GetOptions{})
	if err != nil {
		return "", "", false, maskAny(err)
	}
	isOwned := false
	if ownerRef != nil {
		for _, x := range s.GetOwnerReferences() {
			if x.UID == ownerRef.UID {
				isOwned = true
				break
			}
		}
	}
	// Load `ca.crt` field
	cert, found := s.Data[constants.SecretCACertificate]
	if !found {
		return "", "", isOwned, maskAny(fmt.Errorf("No '%s' found in secret '%s'", constants.SecretCACertificate, secretName))
	}
	priv, found := s.Data[constants.SecretCAKey]
	if !found {
		return "", "", isOwned, maskAny(fmt.Errorf("No '%s' found in secret '%s'", constants.SecretCAKey, secretName))
	}
	return string(cert), string(priv), isOwned, nil
}

// CreateCASecret creates a secret used to store a PEM encoded CA certificate & private key.
func CreateCASecret(secrets SecretInterface, secretName string, certificate, key string, ownerRef *metav1.OwnerReference) error {
	// Create secret
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			constants.SecretCACertificate: []byte(certificate),
			constants.SecretCAKey:         []byte(key),
		},
	}
	// Attach secret to owner
	addOwnerRefToObject(secret, ownerRef)
	if _, err := secrets.Create(secret); err != nil {
		// Failed to create secret
		return maskAny(err)
	}
	return nil
}

// GetTLSKeyfileSecret loads a secret used to store a PEM encoded keyfile
// in the format ArangoDB accepts it for its `--ssl.keyfile` option.
// Returns: keyfile (pem encoded), error
func GetTLSKeyfileSecret(secrets SecretInterface, secretName string) (string, error) {
	s, err := secrets.Get(secretName, metav1.GetOptions{})
	if err != nil {
		return "", maskAny(err)
	}
	// Load `tls.keyfile` field
	keyfile, found := s.Data[constants.SecretTLSKeyfile]
	if !found {
		return "", maskAny(fmt.Errorf("No '%s' found in secret '%s'", constants.SecretTLSKeyfile, secretName))
	}
	return string(keyfile), nil
}

// CreateTLSKeyfileSecret creates a secret used to store a PEM encoded keyfile
// in the format ArangoDB accepts it for its `--ssl.keyfile` option.
func CreateTLSKeyfileSecret(secrets SecretInterface, secretName string, keyfile string, ownerRef *metav1.OwnerReference) error {
	// Create secret
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			constants.SecretTLSKeyfile: []byte(keyfile),
		},
	}
	// Attach secret to owner
	addOwnerRefToObject(secret, ownerRef)
	if _, err := secrets.Create(secret); err != nil {
		// Failed to create secret
		return maskAny(err)
	}
	return nil
}

// ValidateTokenSecret checks that a secret with given name in given namespace
// exists and it contains a 'token' data field.
func ValidateTokenSecret(secrets SecretInterface, secretName string) error {
	s, err := secrets.Get(secretName, metav1.GetOptions{})
	if err != nil {
		return maskAny(err)
	}
	// Check `token` field
	_, found := s.Data[constants.SecretKeyToken]
	if !found {
		return maskAny(fmt.Errorf("No '%s' found in secret '%s'", constants.SecretKeyToken, secretName))
	}
	return nil
}

// GetTokenSecret loads the token secret from a Secret with given name.
func GetTokenSecret(secrets SecretInterface, secretName string) (string, error) {
	s, err := secrets.Get(secretName, metav1.GetOptions{})
	if err != nil {
		return "", maskAny(err)
	}
	// Take the first data from the token key
	data, found := s.Data[constants.SecretKeyToken]
	if !found {
		return "", maskAny(fmt.Errorf("No '%s' data found in secret '%s'", constants.SecretKeyToken, secretName))
	}
	return string(data), nil
}

// CreateTokenSecret creates a secret with given name in given namespace
// with a given token as value.
func CreateTokenSecret(secrets SecretInterface, secretName, token string, ownerRef *metav1.OwnerReference) error {
	// Create secret
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			constants.SecretKeyToken: []byte(token),
		},
	}
	// Attach secret to owner
	addOwnerRefToObject(secret, ownerRef)
	if _, err := secrets.Create(secret); err != nil {
		// Failed to create secret
		return maskAny(err)
	}
	return nil
}

// GetBasicAuthSecret loads a secret with given name in the given namespace
// and extracts the `username` & `password` field.
// If the secret does not exists or one of the fields is missing,
// an error is returned.
// Returns: username, password, error
func GetBasicAuthSecret(secrets SecretInterface, secretName string) (string, string, error) {
	s, err := secrets.Get(secretName, metav1.GetOptions{})
	if err != nil {
		return "", "", maskAny(err)
	}
	// Load `ca.crt` field
	username, found := s.Data[constants.SecretUsername]
	if !found {
		return "", "", maskAny(fmt.Errorf("No '%s' found in secret '%s'", constants.SecretUsername, secretName))
	}
	password, found := s.Data[constants.SecretPassword]
	if !found {
		return "", "", maskAny(fmt.Errorf("No '%s' found in secret '%s'", constants.SecretPassword, secretName))
	}
	return string(username), string(password), nil
}
