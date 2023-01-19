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

package anonymous

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
)

func NewAnonymous[S meta.Object](gvk schema.GroupVersionKind, get generic.ReadClient[S], mod generic.ModStatusClient[S]) Interface {
	return anonymous[S]{
		gvk: gvk,
		get: get,
		mod: mod,
	}
}

type anonymous[S meta.Object] struct {
	gvk schema.GroupVersionKind
	get generic.ReadClient[S]
	mod generic.ModStatusClient[S]
}

func (a anonymous[S]) Get(ctx context.Context, name string, opts meta.GetOptions) (meta.Object, error) {
	return a.get.Get(ctx, name, opts)
}

func (a anonymous[S]) Create(ctx context.Context, obj meta.Object, opts meta.CreateOptions) (meta.Object, error) {
	if o, ok := obj.(S); !ok {
		return nil, newInvalidTypeError(a.gvk)
	} else {
		return a.mod.Create(ctx, o, opts)
	}
}

func (a anonymous[S]) Update(ctx context.Context, obj meta.Object, opts meta.UpdateOptions) (meta.Object, error) {
	if o, ok := obj.(S); !ok {
		return nil, newInvalidTypeError(a.gvk)
	} else {
		return a.mod.Update(ctx, o, opts)
	}
}

func (a anonymous[S]) UpdateStatus(ctx context.Context, obj meta.Object, opts meta.UpdateOptions) (meta.Object, error) {
	if o, ok := obj.(S); !ok {
		return nil, newInvalidTypeError(a.gvk)
	} else {
		return a.mod.Update(ctx, o, opts)
	}
}

func (a anonymous[S]) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions, subresources ...string) (result meta.Object, err error) {

	return a.mod.Patch(ctx, name, pt, data, opts, subresources...)
}

func (a anonymous[S]) Delete(ctx context.Context, name string, opts meta.DeleteOptions) error {
	return a.mod.Delete(ctx, name, opts)
}
