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

package helm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/action"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func cleanup(t *testing.T, c Client) func() {
	t.Run("Cleanup Pre", func(t *testing.T) {
		items, err := c.List(context.Background())
		require.NoError(t, err)

		for _, item := range items {
			t.Run(item.Name, func(t *testing.T) {
				_, err := c.Uninstall(context.Background(), item.Name)
				require.NoError(t, err)
			})
		}
	})

	return func() {
		t.Run("Cleanup Post", func(t *testing.T) {
			items, err := c.List(context.Background())
			require.NoError(t, err)

			for _, item := range items {
				t.Run(item.Name, func(t *testing.T) {
					_, err := c.Uninstall(context.Background(), item.Name)
					require.NoError(t, err)
				})
			}
		})
	}
}

func Test_Connection(t *testing.T) {
	c := newClient(t, tests.FakeNamespace)

	require.NoError(t, c.Alive(context.Background()))

	defer cleanup(t, c)()

	t.Run("NonExisting", func(t *testing.T) {
		resp, err := c.Status(context.Background(), "test-missing")
		require.NoError(t, err)

		require.Nil(t, resp)
	})

	t.Run("Install", func(t *testing.T) {
		resp, err := c.Install(context.Background(), example_1_0_0, nil, func(in *action.Install) {
			in.ReleaseName = "test"
		})
		require.NoError(t, err)

		require.NotNil(t, resp)
	})

	t.Run("Upgrade", func(t *testing.T) {
		resp, err := c.Upgrade(context.Background(), "test", example_1_0_0, nil, func(in *action.Upgrade) {
			in.Install = true
			in.Wait = true
		})
		require.NoError(t, err)

		require.NotNil(t, resp)
	})

	t.Run("Test", func(t *testing.T) {
		resp, err := c.Test(context.Background(), "test")
		require.NoError(t, err)

		require.NotNil(t, resp)
	})

	t.Run("Uninstall", func(t *testing.T) {
		_, err := c.Uninstall(context.Background(), "test")
		require.NoError(t, err)
	})

	t.Run("Reinstall", func(t *testing.T) {
		resp, err := c.Install(context.Background(), example_1_0_0, nil, func(in *action.Install) {
			in.ReleaseName = "test"
			in.Labels = map[string]string{
				"X1": "X1",
			}
		})
		require.NoError(t, err)

		require.Len(t, resp.Labels, 1)
	})

	t.Run("List", func(t *testing.T) {
		defer cleanup(t, c)()

		t.Run("Install", func(t *testing.T) {
			resp, err := c.Install(context.Background(), example_1_0_0, nil, func(in *action.Install) {
				in.ReleaseName = "test-1"
				in.Labels = map[string]string{
					"X1": "X1",
				}
			})
			require.NoError(t, err)

			require.Len(t, resp.Labels, 1)

			resp, err = c.Install(context.Background(), example_1_0_0, nil, func(in *action.Install) {
				in.ReleaseName = "test-2"
				in.Labels = map[string]string{
					"X1": "X2",
				}
			})
			require.NoError(t, err)

			require.Len(t, resp.Labels, 1)

			resp, err = c.Install(context.Background(), example_1_0_0, nil, func(in *action.Install) {
				in.ReleaseName = "test-3"
				in.Labels = map[string]string{
					"X1": "X1",
					"X2": "X2",
				}
			})
			require.NoError(t, err)

			require.Len(t, resp.Labels, 2)

			resp, err = c.Install(context.Background(), example_1_0_0, nil, func(in *action.Install) {
				in.ReleaseName = "test-4"
			})
			require.NoError(t, err)

			require.Len(t, resp.Labels, 0)

			elems, err := c.List(context.Background())
			require.NoError(t, err)
			require.Len(t, elems, 4)
		})

		t.Run("List", func(t *testing.T) {
			t.Run("All", func(t *testing.T) {
				l, err := c.List(context.Background(), func(in *action.List) {
					in.Selector = ""
				})
				require.NoError(t, err)
				require.Len(t, l, 4)
			})
			t.Run("Specified", func(t *testing.T) {
				l, err := c.List(context.Background(), func(in *action.List) {
					in.Selector = "X1==X1"
				})
				require.NoError(t, err)
				require.Len(t, l, 2)
			})
			t.Run("Specified", func(t *testing.T) {
				l, err := c.List(context.Background(), func(in *action.List) {
					in.Selector = "X2==X2"
				})
				require.NoError(t, err)
				require.Len(t, l, 1)
			})
			t.Run("Specified", func(t *testing.T) {
				l, err := c.List(context.Background(), func(in *action.List) {
					in.Selector = "X1==X2"
				})
				require.NoError(t, err)
				require.Len(t, l, 1)
			})
			t.Run("Specified", func(t *testing.T) {
				l, err := c.List(context.Background(), func(in *action.List) {
					in.Selector = "X3==X3"
				})
				require.NoError(t, err)
				require.Len(t, l, 0)
			})
		})
	})

	t.Run("Update", func(t *testing.T) {
		defer cleanup(t, c)()

		t.Run("Install", func(t *testing.T) {
			resp, err := c.Install(context.Background(), example_1_0_0, nil, func(in *action.Install) {
				in.ReleaseName = "test"
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
		})

		t.Run("Verify", func(t *testing.T) {
			cm, err := c.Client().Kubernetes().CoreV1().ConfigMaps(tests.FakeNamespace).Get(context.Background(), "test", meta.GetOptions{})
			require.NoError(t, err)
			require.Len(t, cm.Data, 0)
		})

		t.Run("Update", func(t *testing.T) {
			resp, err := c.Upgrade(context.Background(), "test", example_1_0_0, newValues(t, map[string]any{
				"data": map[string]string{
					"test": "test",
				},
			}))
			require.NoError(t, err)
			require.NotNil(t, resp)
		})

		t.Run("Verify", func(t *testing.T) {
			cm, err := c.Client().Kubernetes().CoreV1().ConfigMaps(tests.FakeNamespace).Get(context.Background(), "test", meta.GetOptions{})
			require.NoError(t, err)
			require.Len(t, cm.Data, 1)
			require.Contains(t, cm.Data, "test")
			require.EqualValues(t, cm.Data["test"], "test")
		})
	})
}
