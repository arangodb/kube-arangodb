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

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
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

func Test_GetLicensePredefined(t *testing.T) {
	lk := "{\"grant\":\"eyJmZWF0dXJlcyI6eyJleHBpcmVzIjoxNjU5MDc3OTk5fSwibm90QmVmb3JlIjoiMjAyMi0wNy0yN1QxMDowOTozNFoiLCJub3RBZnRlciI6IjIwMjItMDctMjlUMDY6NTk6NTlaIiwic3ViamVjdCI6eyJlbWFpbCI6ImdvdGVhbUBhcmFuZ29kYi5jb20ifSwiaXNzdWVyIjp7ImNvbXBhbnkiOiJBcmFuZ29EQiBHbWJIIiwibmFtZSI6IlFBIChBZGFtKSIsImVtYWlsIjoiYWRhbUBhcmFuZ29kYi5jb20ifSwidmVyc2lvbiI6MX0=\",\"signature\":\"F1jGVQ6nuqQJidGzAeNpzbLmTF6J9cv/sPcEUiR0uxY9202PDC7jfqGSgyrAvEXXX5HaLAsRsf2w0/wgkO5KroaJGlONqVKyFkCL2gYSHL1cZl1O9Zt128qAvUL1L5xGeDXEaS9XC7H3DvLecdAYF9URUumytFfUvHUH9jAPo1k8U2WF7sG3c1z0TlPOJbHo5Z0uy2rlcAEiuCKdrZgWXvBcYYnAaBNQIC8HJwTlqMnUI9NY8qm8St8L6lPHTv6cRl0OBW+Z/+3DmZoOW/j0jZNAE57b8DmwUHsdVDtFBhsl4u62jCD1qcRxuYeQ1QTG8+mFzsB0I5MMj4obSeSYOg==\"}"
	lke := "eyJncmFudCI6ImV5Sm1aV0YwZFhKbGN5STZleUpsZUhCcGNtVnpJam94TmpVNU1EYzNPVGs1ZlN3aWJtOTBRbVZtYjNKbElqb2lNakF5TWkwd055MHlOMVF4TURvd09Ub3pORm9pTENKdWIzUkJablJsY2lJNklqSXdNakl0TURjdE1qbFVNRFk2TlRrNk5UbGFJaXdpYzNWaWFtVmpkQ0k2ZXlKbGJXRnBiQ0k2SW1kdmRHVmhiVUJoY21GdVoyOWtZaTVqYjIwaWZTd2lhWE56ZFdWeUlqcDdJbU52YlhCaGJua2lPaUpCY21GdVoyOUVRaUJIYldKSUlpd2libUZ0WlNJNklsRkJJQ2hCWkdGdEtTSXNJbVZ0WVdsc0lqb2lZV1JoYlVCaGNtRnVaMjlrWWk1amIyMGlmU3dpZG1WeWMybHZiaUk2TVgwPSIsInNpZ25hdHVyZSI6IkYxakdWUTZudXFRSmlkR3pBZU5wemJMbVRGNko5Y3Yvc1BjRVVpUjB1eFk5MjAyUERDN2pmcUdTZ3lyQXZFWFhYNUhhTEFzUnNmMncwL3dna081S3JvYUpHbE9OcVZLeUZrQ0wyZ1lTSEwxY1psMU85WnQxMjhxQXZVTDFMNXhHZURYRWFTOVhDN0gzRHZMZWNkQVlGOVVSVXVteXRGZlV2SFVIOWpBUG8xazhVMldGN3NHM2MxejBUbFBPSmJIbzVaMHV5MnJsY0FFaXVDS2RyWmdXWHZCY1lZbkFhQk5RSUM4SEp3VGxxTW5VSTlOWThxbThTdDhMNmxQSFR2NmNSbDBPQlcrWi8rM0RtWm9PVy9qMGpaTkFFNTdiOERtd1VIc2RWRHRGQmhzbDR1NjJqQ0QxcWNSeHVZZVExUVRHOCttRnpzQjBJNU1NajRvYlNlU1lPZz09In0="

	getLicenseFromSecret(t, lk, lke)
}

func Test_GetLicenseFromSecret(t *testing.T) {
	lk := generateLK(t)
	lke := encodeLicenseKey(lk)

	getLicenseFromSecret(t, lk, lke)
}

func getLicenseFromSecret(t *testing.T, raw, encoded string) {

	c := kclient.NewFakeClient()
	i := tests.NewInspector(t, c)

	t.Run(constants.SecretKeyV2License, func(t *testing.T) {
		t.Run("Encoded license", func(t *testing.T) {
			n := createLicenseSecret(t, c, constants.SecretKeyV2License, encoded)

			require.NoError(t, i.Refresh(context.Background()))

			license, err := GetLicenseFromSecret(i, n)
			require.NoError(t, err)

			require.Empty(t, license.V1)
			require.NotEmpty(t, license.V2)
			require.EqualValues(t, encoded, license.V2)
		})

		t.Run("Raw license", func(t *testing.T) {
			n := createLicenseSecret(t, c, constants.SecretKeyV2License, raw)

			require.NoError(t, i.Refresh(context.Background()))

			license, err := GetLicenseFromSecret(i, n)
			require.NoError(t, err)

			require.Empty(t, license.V1)
			require.NotEmpty(t, license.V2)
			require.EqualValues(t, encoded, license.V2)
		})
	})

	t.Run(constants.SecretKeyV2Token, func(t *testing.T) {
		t.Run("Encoded license", func(t *testing.T) {
			n := createLicenseSecret(t, c, constants.SecretKeyV2Token, encoded)

			require.NoError(t, i.Refresh(context.Background()))

			license, err := GetLicenseFromSecret(i, n)
			require.NoError(t, err)

			require.Empty(t, license.V1)
			require.NotEmpty(t, license.V2)
			require.EqualValues(t, encoded, license.V2)
		})

		t.Run("Raw license", func(t *testing.T) {
			n := createLicenseSecret(t, c, constants.SecretKeyV2Token, raw)

			require.NoError(t, i.Refresh(context.Background()))

			license, err := GetLicenseFromSecret(i, n)
			require.NoError(t, err)

			require.Empty(t, license.V1)
			require.NotEmpty(t, license.V2)
			require.EqualValues(t, encoded, license.V2)
		})

		t.Run("Non existing Secret license", func(t *testing.T) {
			require.NoError(t, i.Refresh(context.Background()))

			_, err := GetLicenseFromSecret(i, "non-existing-secret")
			require.Error(t, err)
		})
		t.Run("Non existing license secret key", func(t *testing.T) {
			n := createLicenseSecret(t, c, "wrong-key", raw)

			require.NoError(t, i.Refresh(context.Background()))

			_, err := GetLicenseFromSecret(i, n)
			require.Error(t, err)
		})
	})
}
