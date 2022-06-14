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

package k8sutil

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func encodeLicenseKey(in string) string {
	return base64.StdEncoding.EncodeToString([]byte(in))
}

func createLicenseSecret(t *testing.T, c kclient.Client, key, value string) string {
	s := fmt.Sprintf("secret-%s", uuid.NewUUID())

	q := &core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      s,
			Namespace: tests.FakeNamespace,
		},
		Data: map[string][]byte{
			key: []byte(value),
		},
	}

	_, err := c.Kubernetes().CoreV1().Secrets(tests.FakeNamespace).Create(context.Background(), q, meta.CreateOptions{})
	require.NoError(t, err)

	return s
}

func generateLK(t *testing.T) string {
	z := map[string]string{
		"grant":     "",
		"signature": "",
	}

	i, err := json.Marshal(z)

	require.NoError(t, err)

	return string(i)
}

func Test_GetLicenseFromSecret(t *testing.T) {
	lk := generateLK(t)
	lke := encodeLicenseKey(lk)

	c := kclient.NewFakeClient()
	i := tests.NewInspector(t, c)

	t.Run(constants.SecretKeyV2License, func(t *testing.T) {
		t.Run("Encoded license", func(t *testing.T) {
			n := createLicenseSecret(t, c, constants.SecretKeyV2License, lke)

			require.NoError(t, i.Refresh(context.Background()))

			license, ok := GetLicenseFromSecret(i, n)
			require.True(t, ok)

			require.Empty(t, license.V1)
			require.NotEmpty(t, license.V2)
			require.EqualValues(t, lke, license.V2)
		})

		t.Run("Raw license", func(t *testing.T) {
			n := createLicenseSecret(t, c, constants.SecretKeyV2License, lk)

			require.NoError(t, i.Refresh(context.Background()))

			license, ok := GetLicenseFromSecret(i, n)
			require.True(t, ok)

			require.Empty(t, license.V1)
			require.NotEmpty(t, license.V2)
			require.EqualValues(t, lke, license.V2)
		})
	})

	t.Run(constants.SecretKeyV2Token, func(t *testing.T) {
		t.Run("Encoded license", func(t *testing.T) {
			n := createLicenseSecret(t, c, constants.SecretKeyV2Token, lke)

			require.NoError(t, i.Refresh(context.Background()))

			license, ok := GetLicenseFromSecret(i, n)
			require.True(t, ok)

			require.Empty(t, license.V1)
			require.NotEmpty(t, license.V2)
			require.EqualValues(t, lke, license.V2)
		})

		t.Run("Raw license", func(t *testing.T) {
			n := createLicenseSecret(t, c, constants.SecretKeyV2Token, lk)

			require.NoError(t, i.Refresh(context.Background()))

			license, ok := GetLicenseFromSecret(i, n)
			require.True(t, ok)

			require.Empty(t, license.V1)
			require.NotEmpty(t, license.V2)
			require.EqualValues(t, lke, license.V2)
		})
	})
}
