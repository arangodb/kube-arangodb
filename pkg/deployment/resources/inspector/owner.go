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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/anonymous"
)

func (i *inspectorState) IsOwnerOf(ctx context.Context, owner inspector.Object, obj meta.Object) bool {
	return i.isOwnerOf(ctx, 8, i.AnonymousObjects(), owner.GroupVersionKind(), owner, obj)
}

func (i *inspectorState) isOwnerOf(ctx context.Context, jumps int, a []anonymous.Impl, gvk schema.GroupVersionKind, owner meta.Object, obj meta.Object) bool {
	if jumps <= 0 {
		return false
	}

	for _, o := range obj.GetOwnerReferences() {
		ogvk := extractGVKFromOwnerReference(o)

		if ogvk.Kind == gvk.Kind && ogvk.Group == gvk.Group && o.Name == owner.GetName() && o.UID == owner.GetUID() {
			return true
		}

		for _, q := range a {
			if a == nil {
				continue
			}
			if c, ok := q.Anonymous(ogvk); ok {
				if nobj, err := c.Get(ctx, o.Name, meta.GetOptions{}); err == nil {
					if i.isOwnerOf(ctx, jumps-1, a, gvk, owner, nobj) {
						return true
					}
				}
			}
		}
	}

	return false
}
