//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package crd

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	authorization "k8s.io/api/authorization/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/crd/crds"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func dropLogMessages(t *testing.T, s tests.LogScanner) map[string]string {
	lines := map[string]string{}

	for i := 0; i < len(crds.AllDefinitions()); i++ {
		var data map[string]string
		require.True(t, s.GetData(t, 500*time.Millisecond, &data))

		p, ok := data["crd"]
		require.True(t, ok)

		m, ok := data["message"]
		require.True(t, ok)

		lines[p] = m
	}

	return lines
}

func Test_Apply(t *testing.T) {
	verifyCRDAccessForTests = &authorization.SubjectAccessReviewStatus{
		Allowed: true,
	}

	tests.WithLogScanner(t, "Run", func(t *testing.T, s tests.LogScanner) {
		t.Run("Create CRDs", func(t *testing.T) {
			logger = s.Factory().RegisterAndGetLogger("crd", logging.Info)

			c := kclient.NewFakeClient()

			t.Run("Ensure", func(t *testing.T) {
				require.NoError(t, EnsureCRD(context.Background(), c, false))

				for k, v := range dropLogMessages(t, s) {
					t.Run(k, func(t *testing.T) {
						require.Equal(t, "CRD Created", v)
					})
				}
			})

			t.Run("Verify", func(t *testing.T) {
				for _, q := range crds.AllDefinitions() {
					_, err := c.KubernetesExtensions().ApiextensionsV1().CustomResourceDefinitions().Get(context.Background(), q.CRD.GetName(), meta.GetOptions{})
					require.NoError(t, err)
				}
			})
		})

		t.Run("Create partially CRDs without version", func(t *testing.T) {
			c := kclient.NewFakeClient()

			t.Run("Create", func(t *testing.T) {
				_, err := c.KubernetesExtensions().ApiextensionsV1().CustomResourceDefinitions().Create(context.Background(), crds.AllDefinitions()[0].CRD, meta.CreateOptions{})
				require.NoError(t, err)
			})

			t.Run("Ensure", func(t *testing.T) {
				require.NoError(t, EnsureCRD(context.Background(), c, false))

				for k, v := range dropLogMessages(t, s) {
					t.Run(k, func(t *testing.T) {
						if k == crds.AllDefinitions()[0].CRD.GetName() {
							require.Equal(t, "CRD Updated", v)
						} else {
							require.Equal(t, "CRD Created", v)
						}
					})
				}
			})
		})

		t.Run("Create partially CRDs with version", func(t *testing.T) {
			c := kclient.NewFakeClient()

			t.Run("Create", func(t *testing.T) {
				d := crds.AllDefinitions()[0]
				q := d.CRD.DeepCopy()
				q.Labels = map[string]string{
					Version: string(d.Version),
				}
				_, err := c.KubernetesExtensions().ApiextensionsV1().CustomResourceDefinitions().Create(context.Background(), q, meta.CreateOptions{})
				require.NoError(t, err)
			})

			t.Run("Ensure", func(t *testing.T) {
				require.NoError(t, EnsureCRD(context.Background(), c, false))

				for k, v := range dropLogMessages(t, s) {
					t.Run(k, func(t *testing.T) {
						if k == crds.AllDefinitions()[0].CRD.GetName() {
							require.Equal(t, "CRD Update not required", v)
						} else {
							require.Equal(t, "CRD Created", v)
						}
					})
				}
			})
		})
	})
}
