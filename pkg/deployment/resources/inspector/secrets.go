//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Tomasz Mielech
//

package inspector

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret"
	core "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

func (i *inspector) IterateSecrets(action secret.Action, filters ...secret.Filter) error {
	for _, secret := range i.Secrets() {
		if err := i.iterateSecrets(secret, action, filters...); err != nil {
			return err
		}
	}
	return nil
}

func (i *inspector) iterateSecrets(secret *core.Secret, action secret.Action, filters ...secret.Filter) error {
	for _, filter := range filters {
		if !filter(secret) {
			return nil
		}
	}

	return action(secret)
}

func (i *inspector) Secrets() []*core.Secret {
	i.lock.Lock()
	defer i.lock.Unlock()

	var r []*core.Secret
	for _, secret := range i.secrets {
		r = append(r, secret)
	}

	return r
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

func (i *inspector) SecretReadInterface() secret.ReadInterface {
	return &secretReadInterface{i: i}
}

type secretReadInterface struct {
	i *inspector
}

func (s secretReadInterface) Get(ctx context.Context, name string, opts meta.GetOptions) (*core.Secret, error) {
	if s, ok := s.i.Secret(name); !ok {
		return nil, apiErrors.NewNotFound(schema.GroupResource{
			Group:    core.GroupName,
			Resource: "secrets",
		}, name)
	} else {
		return s, nil
	}
}

func secretsToMap(ctx context.Context, inspector *inspector, k kubernetes.Interface, namespace string) func() error {
	return func() error {
		secrets, err := getSecrets(ctx, k, namespace, "")
		if err != nil {
			return err
		}

		secretMap := map[string]*core.Secret{}

		for _, secret := range secrets {
			_, exists := secretMap[secret.GetName()]
			if exists {
				return errors.Newf("Secret %s already exists in map, error received", secret.GetName())
			}

			secretMap[secret.GetName()] = secretPointer(secret)
		}

		inspector.secrets = secretMap

		return nil
	}
}

func secretPointer(pod core.Secret) *core.Secret {
	return &pod
}

func getSecrets(ctx context.Context, k kubernetes.Interface, namespace, cont string) ([]core.Secret, error) {
	ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
	defer cancel()
	secrets, err := k.CoreV1().Secrets(namespace).List(ctxChild, meta.ListOptions{
		Limit:    128,
		Continue: cont,
	})

	if err != nil {
		return nil, err
	}

	if secrets.Continue != "" {
		nextSecretsLayer, err := getSecrets(ctx, k, namespace, secrets.Continue)
		if err != nil {
			return nil, err
		}

		return append(secrets.Items, nextSecretsLayer...), nil
	}

	return secrets.Items, nil
}
