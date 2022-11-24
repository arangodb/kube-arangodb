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

package inspector

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/uuid"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func Test_OwnerRef(t *testing.T) {
	t.Run("Missing owner", func(t *testing.T) {
		obj1 := &core.Service{
			ObjectMeta: meta.ObjectMeta{
				Name:      "obj1",
				Namespace: "test",
				UID:       uuid.NewUUID(),
			},
		}
		obj2 := &api.ArangoDeployment{
			ObjectMeta: meta.ObjectMeta{
				Name:      "obj2",
				Namespace: "test",
				UID:       uuid.NewUUID(),
			},
		}

		c := clientWithGVK(t, obj1, obj2)

		i := NewInspector(nil, c, "test", "test")

		require.NoError(t, i.Refresh(context.Background()))

		require.False(t, i.IsOwnerOf(context.Background(), obj2, obj1))
	})
	t.Run("Owner 1 hop", func(t *testing.T) {
		obj2 := &api.ArangoDeployment{
			ObjectMeta: meta.ObjectMeta{
				Name:      "obj2",
				Namespace: "test",
				UID:       uuid.NewUUID(),
			},
		}
		obj1 := &core.Service{
			ObjectMeta: meta.ObjectMeta{
				Name:      "obj1",
				Namespace: "test",
				UID:       uuid.NewUUID(),
				OwnerReferences: []meta.OwnerReference{
					obj2.AsOwner(),
				},
			},
		}

		c := clientWithGVK(t, obj1, obj2)

		i := NewInspector(nil, c, "test", "test")

		require.NoError(t, i.Refresh(context.Background()))

		require.True(t, i.IsOwnerOf(context.Background(), obj2, obj1))
	})
	t.Run("Owner 2 hops", func(t *testing.T) {
		obj3 := &api.ArangoDeployment{
			ObjectMeta: meta.ObjectMeta{
				Name:      "obj2",
				Namespace: "test",
				UID:       uuid.NewUUID(),
			},
		}
		obj2 := &core.Service{
			ObjectMeta: meta.ObjectMeta{
				Name:      "obj2",
				Namespace: "test",
				UID:       uuid.NewUUID(),
				OwnerReferences: []meta.OwnerReference{
					obj3.AsOwner(),
				},
			},
		}
		sapi, skind := constants.ServiceGKv1().ToAPIVersionAndKind()
		obj1 := &core.Service{
			ObjectMeta: meta.ObjectMeta{
				Name:      "obj1",
				Namespace: "test",
				UID:       uuid.NewUUID(),
				OwnerReferences: []meta.OwnerReference{
					{
						APIVersion: sapi,
						Kind:       skind,
						Name:       obj2.Name,
						UID:        obj2.UID,
					},
				},
			},
		}

		c := clientWithGVK(t, obj1, obj2, obj3)

		i := NewInspector(nil, c, "test", "test")

		require.NoError(t, i.Refresh(context.Background()))

		require.True(t, i.IsOwnerOf(context.Background(), obj3, obj1))
	})
	t.Run("Owner - infinite loop", func(t *testing.T) {
		s2uid := uuid.NewUUID()
		sapi, skind := constants.ServiceGKv1().ToAPIVersionAndKind()
		obj4 := &api.ArangoDeployment{
			ObjectMeta: meta.ObjectMeta{
				Name:      "obj4",
				Namespace: "test",
				UID:       uuid.NewUUID(),
			},
		}
		obj3 := &core.Service{
			ObjectMeta: meta.ObjectMeta{
				Name:      "obj3",
				Namespace: "test",
				UID:       s2uid,
				OwnerReferences: []meta.OwnerReference{
					{
						APIVersion: sapi,
						Kind:       skind,
						Name:       "obj2",
						UID:        s2uid,
					},
				},
			},
		}
		obj2 := &core.Service{
			ObjectMeta: meta.ObjectMeta{
				Name:      "obj2",
				Namespace: "test",
				UID:       s2uid,
				OwnerReferences: []meta.OwnerReference{
					{
						APIVersion: sapi,
						Kind:       skind,
						Name:       obj3.Name,
						UID:        obj3.UID,
					},
				},
			},
		}
		obj1 := &core.Service{
			ObjectMeta: meta.ObjectMeta{
				Name:      "obj1",
				Namespace: "test",
				UID:       uuid.NewUUID(),
				OwnerReferences: []meta.OwnerReference{
					{
						APIVersion: sapi,
						Kind:       skind,
						Name:       obj2.Name,
						UID:        obj2.UID,
					},
				},
			},
		}

		c := clientWithGVK(t, obj1, obj2, obj3, obj4)

		i := NewInspector(nil, c, "test", "test")

		require.NoError(t, i.Refresh(context.Background()))

		require.False(t, i.IsOwnerOf(context.Background(), obj4, obj1))
	})
	t.Run("Owner - above limit", func(t *testing.T) {
		objs := make([]GVKEnsurer, 10)

		depl := &api.ArangoDeployment{
			ObjectMeta: meta.ObjectMeta{
				Name:      "depl",
				Namespace: "test",
				UID:       uuid.NewUUID(),
			},
		}
		sapi, skind := constants.ServiceGKv1().ToAPIVersionAndKind()

		last := &core.Service{
			ObjectMeta: meta.ObjectMeta{
				Name:      "obj8",
				Namespace: "test",
				UID:       uuid.NewUUID(),
				OwnerReferences: []meta.OwnerReference{
					depl.AsOwner(),
				},
			},
		}

		objs[len(objs)-1] = depl
		objs[len(objs)-2] = last

		for i := len(objs) - 3; i >= 0; i-- {
			n := &core.Service{
				ObjectMeta: meta.ObjectMeta{
					Name:      fmt.Sprintf("obj%d", i),
					Namespace: "test",
					UID:       uuid.NewUUID(),
					OwnerReferences: []meta.OwnerReference{
						{
							APIVersion: sapi,
							Kind:       skind,
							Name:       last.Name,
							UID:        last.UID,
						},
					},
				},
			}
			objs[i] = n
			last = n
		}

		c := clientWithGVK(t, objs...)

		i := NewInspector(nil, c, "test", "test")

		require.NoError(t, i.Refresh(context.Background()))

		require.False(t, i.IsOwnerOf(context.Background(), depl, last))
	})
	t.Run("Owner - on limit", func(t *testing.T) {
		objs := make([]GVKEnsurer, 9)

		depl := &api.ArangoDeployment{
			ObjectMeta: meta.ObjectMeta{
				Name:      "depl",
				Namespace: "test",
				UID:       uuid.NewUUID(),
			},
		}
		sapi, skind := constants.ServiceGKv1().ToAPIVersionAndKind()

		last := &core.Service{
			ObjectMeta: meta.ObjectMeta{
				Name:      "obj8",
				Namespace: "test",
				UID:       uuid.NewUUID(),
				OwnerReferences: []meta.OwnerReference{
					depl.AsOwner(),
				},
			},
		}

		objs[len(objs)-1] = depl
		objs[len(objs)-2] = last

		for i := len(objs) - 3; i >= 0; i-- {
			n := &core.Service{
				ObjectMeta: meta.ObjectMeta{
					Name:      fmt.Sprintf("obj%d", i),
					Namespace: "test",
					UID:       uuid.NewUUID(),
					OwnerReferences: []meta.OwnerReference{
						{
							APIVersion: sapi,
							Kind:       skind,
							Name:       last.Name,
							UID:        last.UID,
						},
					},
				},
			}
			objs[i] = n
			last = n
		}

		c := clientWithGVK(t, objs...)

		i := NewInspector(nil, c, "test", "test")

		require.NoError(t, i.Refresh(context.Background()))

		require.True(t, i.IsOwnerOf(context.Background(), depl, last))
	})
}

type GVKEnsurer interface {
	runtime.Object
	SetGroupVersionKind(gvk schema.GroupVersionKind)
}

func clientWithGVK(t *testing.T, obj ...GVKEnsurer) kclient.Client {
	ensureGVK(t, obj...)

	f := kclient.NewFakeClientBuilder()

	for _, o := range obj {
		f = f.Add(o)
	}

	return f.Client()
}

func ensureGVK(t *testing.T, objs ...GVKEnsurer) {
	for _, o := range objs {
		gvk, ok := constants.ExtractGVKFromObject(o)
		require.True(t, ok)

		o.SetGroupVersionKind(gvk)
	}
}
