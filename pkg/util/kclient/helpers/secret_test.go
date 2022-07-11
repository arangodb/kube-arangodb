//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package helpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func Test_SecretConfigGetter(t *testing.T) {
	t.Run("Missing secret", func(t *testing.T) {
		c := kclient.NewFakeClient()

		i := inspector.NewInspector(throttle.NewAlwaysThrottleComponents(), c, "default", "default")
		require.NoError(t, i.Refresh(context.Background()))

		_, _, err := SecretConfigGetter(i, "secret", "key")()
		require.EqualError(t, err, "Secret secret not found")
	})

	t.Run("Missing key", func(t *testing.T) {
		c := kclient.NewFakeClient()

		s := core.Secret{
			ObjectMeta: meta.ObjectMeta{
				Name:      "secret",
				Namespace: "default",
			},
		}

		_, err := c.Kubernetes().CoreV1().Secrets("default").Create(context.Background(), &s, meta.CreateOptions{})
		require.NoError(t, err)

		i := inspector.NewInspector(throttle.NewAlwaysThrottleComponents(), c, "default", "default")
		require.NoError(t, i.Refresh(context.Background()))

		_, _, err = SecretConfigGetter(i, "secret", "key")()
		require.EqualError(t, err, "Key secret/key not found")
	})

	t.Run("Invalid data", func(t *testing.T) {
		c := kclient.NewFakeClient()

		s := core.Secret{
			ObjectMeta: meta.ObjectMeta{
				Name:      "secret",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"key": []byte(`
random data
`),
			},
		}

		_, err := c.Kubernetes().CoreV1().Secrets("default").Create(context.Background(), &s, meta.CreateOptions{})
		require.NoError(t, err)

		i := inspector.NewInspector(throttle.NewAlwaysThrottleComponents(), c, "default", "default")
		require.NoError(t, i.Refresh(context.Background()))

		_, _, err = SecretConfigGetter(i, "secret", "key")()
		require.Error(t, err, "Key secret/key not found")
	})

	t.Run("Valid data", func(t *testing.T) {
		c := kclient.NewFakeClient()

		s := core.Secret{
			ObjectMeta: meta.ObjectMeta{
				Name:      "secret",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"key": []byte(`
apiVersion: v1
clusters:
- cluster:
    server: https://localhost
  name: test
contexts:
- context:
    cluster: test
    user: test
    namespace: test
  name: test
current-context: test
kind: Config
preferences: {}
users:
- name: test
  user:
    token: x
`),
			},
		}

		_, err := c.Kubernetes().CoreV1().Secrets("default").Create(context.Background(), &s, meta.CreateOptions{})
		require.NoError(t, err)

		i := inspector.NewInspector(throttle.NewAlwaysThrottleComponents(), c, "default", "default")
		require.NoError(t, i.Refresh(context.Background()))

		_, _, err = SecretConfigGetter(i, "secret", "key")()
		require.NoError(t, err)
	})
}
