//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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

package cmd

import (
	"context"
	"time"

	"github.com/dchest/uniuri"
	core "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func newBasicAuthCacheObject(client kclient.Client, namespace, name string) cache.Object[map[string]string] {
	return cache.NewObject(func(ctx context.Context) (map[string]string, time.Duration, error) {
		secret, err := client.Kubernetes().CoreV1().Secrets(namespace).Get(ctx, name, meta.GetOptions{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return nil, 0, err
			} else {
				// Create one
				secret = &core.Secret{
					ObjectMeta: meta.ObjectMeta{
						Name:      apiOptions.basicSecretName,
						Namespace: namespace,
					},
					Data: map[string][]byte{
						"admin": []byte(uniuri.NewLen(12)),
					},
				}

				secret, err = client.Kubernetes().CoreV1().Secrets(namespace).Create(ctx, secret, meta.CreateOptions{})
				if err != nil {
					if !apiErrors.IsAlreadyExists(err) {
						return nil, 0, err
					}

					secret, err = client.Kubernetes().CoreV1().Secrets(namespace).Get(ctx, apiOptions.basicSecretName, meta.GetOptions{})
					if err != nil {
						return nil, 0, err
					}
				}
			}
		}
		return util.FormatMap(secret.Data, func(k string, a []byte) string {
			return string(a)
		}), 15 * time.Second, nil
	})
}
