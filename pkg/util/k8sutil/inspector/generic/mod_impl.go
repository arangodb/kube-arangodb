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

package generic

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func WithModStatusGetter[S meta.Object](gvk schema.GroupVersionKind, in func() ModClient[S]) func() ModStatusClient[S] {
	return func() ModStatusClient[S] {
		return WithModStatus[S](gvk, in())
	}
}

func WithModStatus[S meta.Object](gvk schema.GroupVersionKind, in ModClient[S]) ModStatusClient[S] {
	return modStatus[S]{
		in,
		gvk,
	}
}

type modStatus[S meta.Object] struct {
	ModClient[S]
	gvk schema.GroupVersionKind
}

func (m modStatus[S]) UpdateStatus(ctx context.Context, obj S, opts meta.UpdateOptions) (S, error) {
	return *new(S), newNotImplementedError(m.gvk)
}

func WithEmptyMod[S meta.Object](gvk schema.GroupVersionKind) ModClient[S] {
	return emptyMod[S]{
		gvk,
	}
}

type emptyMod[S meta.Object] struct {
	gvk schema.GroupVersionKind
}

func (e emptyMod[S]) Create(ctx context.Context, obj S, opts meta.CreateOptions) (S, error) {
	return *new(S), newNotImplementedError(e.gvk)
}

func (e emptyMod[S]) Update(ctx context.Context, obj S, opts meta.UpdateOptions) (S, error) {
	return *new(S), newNotImplementedError(e.gvk)
}

func (e emptyMod[S]) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions, subresources ...string) (result S, err error) {
	return *new(S), newNotImplementedError(e.gvk)
}

func (e emptyMod[S]) Delete(ctx context.Context, name string, opts meta.DeleteOptions) error {
	return newNotImplementedError(e.gvk)
}
