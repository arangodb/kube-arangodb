//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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

package patcher

import (
	"context"
	"reflect"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
)

type Patch[T meta.Object] func(in T) []patch.Item

func NewPatchList[T meta.Object](items ...Patch[T]) PatchList[T] {
	return items
}

type PatchList[T meta.Object] []Patch[T]

func (p PatchList[T]) Append(items ...Patch[T]) PatchList[T] {
	z := make([]Patch[T], 0, len(p)+len(items))
	copy(z, p)
	copy(z[len(p):], items)
	return z
}

type Client[T meta.Object] interface {
	generic.PatchInterface[T]
}

func Patcher[T meta.Object](ctx context.Context, client Client[T], in T, opts meta.PatchOptions, functions ...Patch[T]) (T, bool, error) {
	if v := reflect.ValueOf(in); !v.IsValid() || v.IsZero() {
		return util.Default[T](), false, nil
	}

	if in.GetName() == "" {
		return util.Default[T](), false, nil
	}

	items, err := Generate(in, functions...)
	if err != nil {
		return util.Default[T](), false, err
	}

	if len(items) == 0 {
		return in, false, nil
	}

	data, err := patch.NewPatch(items...).Marshal()
	if err != nil {
		return util.Default[T](), false, err
	}

	nctx, c := globals.GetGlobals().Timeouts().Kubernetes().WithTimeout(ctx)
	defer c()

	if obj, err := client.Patch(nctx, in.GetName(), types.JSONPatchType, data, opts); err != nil {
		return util.Default[T](), false, err
	} else {
		return obj, true, nil
	}
}

func Generate[T meta.Object](in T, functions ...Patch[T]) (items patch.Patch, err error) {
	defer func() {
		if r := recover(); r != nil {
			if v, ok := r.(error); ok {
				err = v
				return
			}

			err = errors.Errorf("Recovered from panic: %v", r)
		}
	}()

	for id := range functions {
		if f := functions[id]; f != nil {
			items = append(items, f(in)...)
		}
	}

	return
}

func Optional[T meta.Object](p Patch[T], enabled bool) Patch[T] {
	return func(in T) []patch.Item {
		if !enabled {
			return nil
		}

		if p != nil {
			return p(in)
		}

		return nil
	}
}
