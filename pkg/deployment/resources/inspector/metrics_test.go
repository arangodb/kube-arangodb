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

package inspector

import (
	"context"
	"testing"

	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func extractMetric(t *testing.T, definition definitions.Component, verb definitions.Verb) clientMetricsFields {
	defer clientMetricsInstance.reset()
	if m := clientMetricsInstance.metrics; m != nil {
		if a, ok := m[definition]; ok {
			if b, ok := a[verb]; ok {
				return b
			}
		}
	}

	require.Fail(t, "Metric not found")
	return clientMetricsFields{}
}

func testModClientMetrics[S meta.Object](t *testing.T, definition definitions.Component, in generic.ModClient[S], generator func(name string) S) {
	t.Run("Create with collision", func(t *testing.T) {
		obj := generator("test")

		t.Run("Create", func(t *testing.T) {
			_, err := in.Create(context.Background(), obj, meta.CreateOptions{})
			require.NoError(t, err)

			f := extractMetric(t, definition, definitions.Create)
			require.Equal(t, 1, f.calls)
			require.Equal(t, 0, f.errors)
		})

		t.Run("Recreate", func(t *testing.T) {
			_, err := in.Create(context.Background(), obj, meta.CreateOptions{})
			require.Error(t, err)

			f := extractMetric(t, definition, definitions.Create)
			require.Equal(t, 1, f.calls)
			require.Equal(t, 1, f.errors)
		})

		t.Run("Update", func(t *testing.T) {
			_, err := in.Update(context.Background(), obj, meta.UpdateOptions{})
			require.NoError(t, err)

			f := extractMetric(t, definition, definitions.Update)
			require.Equal(t, 1, f.calls)
			require.Equal(t, 0, f.errors)
		})

		t.Run("Update Missing", func(t *testing.T) {
			obj := generator("test2")

			_, err := in.Update(context.Background(), obj, meta.UpdateOptions{})
			require.Error(t, err)

			f := extractMetric(t, definition, definitions.Update)
			require.Equal(t, 1, f.calls)
			require.Equal(t, 1, f.errors)
		})

		if us, ok := in.(generic.ModStatusClient[S]); ok {
			_, err := us.UpdateStatus(context.Background(), obj, meta.UpdateOptions{})
			extractMetric(t, definition, definitions.UpdateStatus)
			if err == nil {
				t.Run("UpdateStatus", func(t *testing.T) {
					_, err := us.UpdateStatus(context.Background(), obj, meta.UpdateOptions{})
					require.NoError(t, err)

					f := extractMetric(t, definition, definitions.UpdateStatus)
					require.Equal(t, 1, f.calls)
					require.Equal(t, 0, f.errors)
				})

				t.Run("UpdateStatus Missing", func(t *testing.T) {
					obj := generator("test2")

					_, err := us.UpdateStatus(context.Background(), obj, meta.UpdateOptions{})
					require.Error(t, err)

					f := extractMetric(t, definition, definitions.UpdateStatus)
					require.Equal(t, 1, f.calls)
					require.Equal(t, 1, f.errors)
				})
			}
		}

		t.Run("Patch", func(t *testing.T) {
			p, err := patch.NewPatch(patch.ItemReplace(patch.NewPath("metadata", "labels"), map[string]string{})).Marshal()
			require.NoError(t, err)

			_, err = in.Patch(context.Background(), obj.GetName(), types.JSONPatchType, p, meta.PatchOptions{})
			require.NoError(t, err)

			f := extractMetric(t, definition, definitions.Patch)
			require.Equal(t, 1, f.calls)
			require.Equal(t, 0, f.errors)
		})

		t.Run("Patch Missing", func(t *testing.T) {
			obj := generator("test2")
			p, err := patch.NewPatch(patch.ItemReplace(patch.NewPath("metadata", "labels"), map[string]string{})).Marshal()
			require.NoError(t, err)

			_, err = in.Patch(context.Background(), obj.GetName(), types.JSONPatchType, p, meta.PatchOptions{})
			require.Error(t, err)

			f := extractMetric(t, definition, definitions.Patch)
			require.Equal(t, 1, f.calls)
			require.Equal(t, 1, f.errors)
		})

		t.Run("Delete", func(t *testing.T) {
			err := in.Delete(context.Background(), obj.GetName(), meta.DeleteOptions{})
			require.NoError(t, err)

			f := extractMetric(t, definition, definitions.Delete)
			require.Equal(t, 1, f.calls)
			require.Equal(t, 0, f.errors)
		})

		t.Run("Delete - missing", func(t *testing.T) {
			err := in.Delete(context.Background(), obj.GetName(), meta.DeleteOptions{})
			require.Error(t, err)

			f := extractMetric(t, definition, definitions.Delete)
			require.Equal(t, 1, f.calls)
			require.Equal(t, 1, f.errors)
		})

		t.Run("Create for deletion", func(t *testing.T) {
			_, err := in.Create(context.Background(), obj, meta.CreateOptions{})
			require.NoError(t, err)

			f := extractMetric(t, definition, definitions.Create)
			require.Equal(t, 1, f.calls)
			require.Equal(t, 0, f.errors)
		})

		t.Run("ForceDelete", func(t *testing.T) {
			err := in.Delete(context.Background(), obj.GetName(), meta.DeleteOptions{
				GracePeriodSeconds: util.NewType[int64](0),
			})
			require.NoError(t, err)

			f := extractMetric(t, definition, definitions.ForceDelete)
			require.Equal(t, 1, f.calls)
			require.Equal(t, 0, f.errors)
		})

		t.Run("ForceDelete - missing", func(t *testing.T) {
			err := in.Delete(context.Background(), obj.GetName(), meta.DeleteOptions{
				GracePeriodSeconds: util.NewType[int64](0),
			})
			require.Error(t, err)

			f := extractMetric(t, definition, definitions.ForceDelete)
			require.Equal(t, 1, f.calls)
			require.Equal(t, 1, f.errors)
		})
	})
}

func Test_Metrics(t *testing.T) {
	c := kclient.NewFakeClient()
	q := NewInspector(throttle.NewAlwaysThrottleComponents(), c, "test", "test")

	t.Run(string(definitions.ArangoMember), func(t *testing.T) {
		testModClientMetrics[*api.ArangoMember](t, definitions.ArangoMember, q.ArangoMemberModInterface().V1(), func(name string) *api.ArangoMember {
			return &api.ArangoMember{
				ObjectMeta: meta.ObjectMeta{
					Name:      name,
					Namespace: "test",
				},
			}
		})
	})

	t.Run(string(definitions.ArangoTask), func(t *testing.T) {
		testModClientMetrics[*api.ArangoTask](t, definitions.ArangoTask, q.ArangoTaskModInterface().V1(), func(name string) *api.ArangoTask {
			return &api.ArangoTask{
				ObjectMeta: meta.ObjectMeta{
					Name:      name,
					Namespace: "test",
				},
			}
		})
	})

	t.Run(string(definitions.ArangoClusterSynchronization), func(t *testing.T) {
		testModClientMetrics[*api.ArangoClusterSynchronization](t, definitions.ArangoClusterSynchronization, q.ArangoClusterSynchronizationModInterface().V1(), func(name string) *api.ArangoClusterSynchronization {
			return &api.ArangoClusterSynchronization{
				ObjectMeta: meta.ObjectMeta{
					Name:      name,
					Namespace: "test",
				},
			}
		})
	})

	t.Run(string(definitions.Pod), func(t *testing.T) {
		testModClientMetrics[*core.Pod](t, definitions.Pod, q.PodsModInterface().V1(), func(name string) *core.Pod {
			return &core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Name:      name,
					Namespace: "test",
				},
			}
		})
	})

	t.Run(string(definitions.PersistentVolumeClaim), func(t *testing.T) {
		testModClientMetrics[*core.PersistentVolumeClaim](t, definitions.PersistentVolumeClaim, q.PersistentVolumeClaimsModInterface().V1(), func(name string) *core.PersistentVolumeClaim {
			return &core.PersistentVolumeClaim{
				ObjectMeta: meta.ObjectMeta{
					Name:      name,
					Namespace: "test",
				},
			}
		})
	})

	t.Run(string(definitions.Secret), func(t *testing.T) {
		testModClientMetrics[*core.Secret](t, definitions.Secret, q.SecretsModInterface().V1(), func(name string) *core.Secret {
			return &core.Secret{
				ObjectMeta: meta.ObjectMeta{
					Name:      name,
					Namespace: "test",
				},
			}
		})
	})

	t.Run(string(definitions.Service), func(t *testing.T) {
		testModClientMetrics[*core.Service](t, definitions.Service, q.ServicesModInterface().V1(), func(name string) *core.Service {
			return &core.Service{
				ObjectMeta: meta.ObjectMeta{
					Name:      name,
					Namespace: "test",
				},
			}
		})
	})

	t.Run(string(definitions.ServiceAccount), func(t *testing.T) {
		testModClientMetrics[*core.ServiceAccount](t, definitions.ServiceAccount, q.ServiceAccountsModInterface().V1(), func(name string) *core.ServiceAccount {
			return &core.ServiceAccount{
				ObjectMeta: meta.ObjectMeta{
					Name:      name,
					Namespace: "test",
				},
			}
		})
	})

	t.Run(string(definitions.Endpoints), func(t *testing.T) {
		testModClientMetrics[*core.Endpoints](t, definitions.Endpoints, q.EndpointsModInterface().V1(), func(name string) *core.Endpoints {
			return &core.Endpoints{
				ObjectMeta: meta.ObjectMeta{
					Name:      name,
					Namespace: "test",
				},
			}
		})
	})

	t.Run(string(definitions.ServiceMonitor), func(t *testing.T) {
		testModClientMetrics[*monitoring.ServiceMonitor](t, definitions.ServiceMonitor, q.ServiceMonitorsModInterface().V1(), func(name string) *monitoring.ServiceMonitor {
			return &monitoring.ServiceMonitor{
				ObjectMeta: meta.ObjectMeta{
					Name:      name,
					Namespace: "test",
				},
			}
		})
	})
}
