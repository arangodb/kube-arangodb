//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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
	"fmt"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes/fake"

	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func init() {
	logging.Global().SetRoot(zerolog.New(os.Stdout).With().Timestamp().Logger())
}

func runUpdate[T Object](t *testing.T, iterations int, client Client[T], ref **sharedApi.Object, generator Generate[T], decisions ...Decision[T]) {
	logger := logging.Global().Get("test")

	i := 0

	for i = 1; i < 1024; i++ {
		var changed bool
		t.Run(fmt.Sprintf("Iteration %d", i), func(t *testing.T) {
			ok, err := Update[T](context.Background(), logger, client, ref, generator, decisions...)
			require.NoError(t, err)
			changed = ok
		})

		if !changed {
			break
		}
	}

	require.EqualValues(t, iterations, i, fmt.Sprintf("Expected %d iterations, got %d", iterations, i))
}

func get[T Object](t *testing.T, client Client[T], in T) (T, bool) {
	obj, err := client.Get(context.Background(), in.GetName(), meta.GetOptions{})
	if err != nil {
		if kerrors.Is(err, kerrors.NotFound) {
			return util.Default[T](), false
		}

		require.NoError(t, err)
	}

	return obj, true
}

func Test_Updator(t *testing.T) {
	logging.Global().RegisterLogger("test", logging.Trace)

	client := fake.NewSimpleClientset().CoreV1().Secrets(tests.FakeNamespace)

	var secret = core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      "secret",
			Namespace: tests.FakeNamespace,
			UID:       uuid.NewUUID(),
		},
	}
	var checksum = util.SHA256FromString("")

	var ref *sharedApi.Object

	var retSecret Generate[*core.Secret] = func(_ context.Context, _ *sharedApi.Object) (*core.Secret, bool, string, error) {
		return secret.DeepCopy(), false, checksum, nil
	}

	t.Run("Ensure default is handled", func(t *testing.T) {
		runUpdate[*core.Secret](t, 2, client, &ref, retSecret)

		_, ok := get[*core.Secret](t, client, &secret)
		require.True(t, ok)
	})

	t.Run("Ensure rerun is handled", func(t *testing.T) {
		runUpdate[*core.Secret](t, 1, client, &ref, retSecret)

		_, ok := get[*core.Secret](t, client, &secret)
		require.True(t, ok)
	})

	t.Run("Ensure delete is not handled when skip is requested", func(t *testing.T) {
		runUpdate[*core.Secret](t, 1, client, &ref, func(ctx context.Context, _ *sharedApi.Object) (*core.Secret, bool, string, error) {
			return nil, true, "", nil
		})

		_, ok := get[*core.Secret](t, client, &secret)
		require.True(t, ok)
	})

	t.Run("Ensure delete is handled", func(t *testing.T) {
		runUpdate[*core.Secret](t, 3, client, &ref, func(ctx context.Context, _ *sharedApi.Object) (*core.Secret, bool, string, error) {
			return nil, false, "", nil
		})

		_, ok := get[*core.Secret](t, client, &secret)
		require.False(t, ok)
	})

	t.Run("Recreate", func(t *testing.T) {
		runUpdate[*core.Secret](t, 2, client, &ref, retSecret)

		_, ok := get[*core.Secret](t, client, &secret)
		require.True(t, ok)
	})

	t.Run("Change checksum without handler", func(t *testing.T) {
		checksum = util.SHA256FromString("NEW")

		runUpdate[*core.Secret](t, 1, client, &ref, retSecret)

		require.NotEqual(t, checksum, ref.GetChecksum())

		_, ok := get[*core.Secret](t, client, &secret)
		require.True(t, ok)
	})

	t.Run("Change checksum with recreate handler", func(t *testing.T) {
		runUpdate[*core.Secret](t, 4, client, &ref, retSecret, ReplaceChecksum[*core.Secret])

		require.Equal(t, checksum, ref.GetChecksum())

		_, ok := get[*core.Secret](t, client, &secret)
		require.True(t, ok)
	})

	t.Run("UUID Changed", func(t *testing.T) {
		ref.UID = util.NewType(uuid.NewUUID())

		runUpdate[*core.Secret](t, 4, client, &ref, retSecret)

		s, ok := get[*core.Secret](t, client, &secret)
		require.True(t, ok)
		require.Equal(t, ref.GetUID(), s.GetUID())
	})

	t.Run("Owner Added Without Handler", func(t *testing.T) {
		secret.SetOwnerReferences([]meta.OwnerReference{
			{
				UID: uuid.NewUUID(),
			},
		})

		runUpdate[*core.Secret](t, 1, client, &ref, retSecret)

		s, ok := get[*core.Secret](t, client, &secret)
		require.True(t, ok)
		require.NotEqual(t, secret.GetOwnerReferences(), s.GetOwnerReferences())
	})

	t.Run("Owner Added With Handler", func(t *testing.T) {
		runUpdate[*core.Secret](t, 2, client, &ref, retSecret, UpdateOwnerReference[*core.Secret])

		s, ok := get[*core.Secret](t, client, &secret)
		require.True(t, ok)
		require.Equal(t, secret.GetOwnerReferences(), s.GetOwnerReferences())
	})
}
