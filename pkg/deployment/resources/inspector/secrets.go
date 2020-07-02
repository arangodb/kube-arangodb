//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package inspector

import (
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type SecretFilter func(pod *core.Secret) bool
type SecretAction func(pod *core.Secret) error

func (i *inspector) IterateSecrets(action SecretAction, filters ...SecretFilter) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	for _, secret := range i.secrets {
		if err := i.iterateSecrets(secret, action, filters...); err != nil {
			return err
		}
	}
	return nil
}

func (i *inspector) iterateSecrets(secret *core.Secret, action SecretAction, filters ...SecretFilter) error {
	for _, filter := range filters {
		if !filter(secret) {
			return nil
		}
	}

	return action(secret)
}

func (i *inspector) Secret(name string) (*core.Secret, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()

	secret, ok := i.secrets[name]
	if !ok {
		return nil, false
	}

	return secret, true
}

func secretsToMap(k kubernetes.Interface, namespace string) (map[string]*core.Secret, error) {
	secrets, err := getSecrets(k, namespace, "")
	if err != nil {
		return nil, err
	}

	secretMap := map[string]*core.Secret{}

	for _, secret := range secrets {
		_, exists := secretMap[secret.GetName()]
		if exists {
			return nil, errors.Errorf("Secret %s already exists in map, error received", secret.GetName())
		}

		secretMap[secret.GetName()] = secretPointer(secret)
	}

	return secretMap, nil
}

func secretPointer(pod core.Secret) *core.Secret {
	return &pod
}

func getSecrets(k kubernetes.Interface, namespace, cont string) ([]core.Secret, error) {
	secrets, err := k.CoreV1().Secrets(namespace).List(meta.ListOptions{
		Limit:    128,
		Continue: cont,
	})

	if err != nil {
		return nil, err
	}

	if secrets.Continue != "" {
		nextSecretsLayer, err := getSecrets(k, namespace, secrets.Continue)
		if err != nil {
			return nil, err
		}

		return append(secrets.Items, nextSecretsLayer...), nil
	}

	return secrets.Items, nil
}
