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

package license

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	errors2 "github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func NewSecretLoader(factory kclient.Factory, namespace, name, key string) Loader {
	return secretLoader{
		factory:   factory,
		namespace: namespace,
		name:      name,
		key:       key,
	}
}

type secretLoader struct {
	factory              kclient.Factory
	namespace, name, key string
}

func (s secretLoader) Refresh(ctx context.Context) (string, bool, error) {
	client, ok := s.factory.Client()
	if !ok {
		return "", false, errors2.Newf("Client is not yet ready")
	}

	secret, err := client.Kubernetes().CoreV1().Secrets(s.namespace).Get(ctx, s.name, meta.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return "", false, nil
		}

		return "", false, err
	}

	if len(secret.Data) == 0 {
		return "", false, nil
	}

	license, ok := secret.Data[s.key]
	if !ok {
		return "", false, nil
	}

	return string(license), true, nil
}
