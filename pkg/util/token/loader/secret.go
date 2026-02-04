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

package loader

import (
	"context"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
)

func SecretCacheSecretAPI(secrets generic.ReadClient[*core.Secret], secretName string, ttl time.Duration) func(ctx context.Context) (utilToken.Secret, time.Duration, error) {
	return func(ctx context.Context) (utilToken.Secret, time.Duration, error) {
		s, err := LoadSecretSetFromSecretAPI(ctx, secrets, secretName)
		if err != nil {
			return nil, 0, err
		}

		return s, ttl, nil
	}
}

func LoadSecretSetFromSecretAPI(ctx context.Context, secrets generic.ReadClient[*core.Secret], secretName string) (utilToken.Secret, error) {
	active, passive, err := LoadSecretsFromSecretAPI(ctx, secrets, secretName)
	if err != nil {
		return nil, err
	}

	return utilToken.NewSecretSet(active, passive), nil
}

func LoadSecretsFromSecretAPI(ctx context.Context, secrets generic.ReadClient[*core.Secret], secretName string) (utilToken.Secret, utilToken.Secrets, error) {
	s, err := secrets.Get(ctx, secretName, meta.GetOptions{})
	if err != nil {
		return nil, nil, err
	}

	return LoadSecretsFromSecret(s)
}

func LoadSecretsFromSecret(secret *core.Secret) (utilToken.Secret, utilToken.Secrets, error) {
	return LoadSecretsFromData(secret.Data)
}
