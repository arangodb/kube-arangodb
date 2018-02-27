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
	"k8s.io/client-go/kubernetes"

	"github.com/arangodb/k8s-operator/pkg/util/constants"
)

// ValidateEncryptionKeySecret checks that a secret with given name in given namespace
// exists and it contains a 'key' data field of exactly 32 bytes.
func ValidateEncryptionKeySecret(kubecli kubernetes.Interface, secretName, namespace string) error {
	s, err := kubecli.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
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
func CreateEncryptionKeySecret(kubecli kubernetes.Interface, secretName, namespace string, key []byte) error {
	// Create secret
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			constants.SecretEncryptionKey: key,
		},
	}
	if _, err := kubecli.CoreV1().Secrets(namespace).Create(secret); err != nil {
		// Failed to create secret
		return maskAny(err)
	}
	return nil
}

// GetJWTSecret loads the JWT secret from a Secret with given name.
func GetJWTSecret(kubecli kubernetes.Interface, secretName, namespace string) (string, error) {
	s, err := kubecli.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		return "", maskAny(err)
	}
	// Take the first data
	for _, v := range s.Data {
		return string(v), nil
	}
	return "", maskAny(fmt.Errorf("No data found in secret '%s'", secretName))
}
