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

package kubernetes

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

type ObjectList[T meta.Object] map[types.UID]T

func (d ObjectList[T]) AsList() []T {
	list := make([]T, 0, len(d))
	for _, p := range d {
		list = append(list, p)
	}

	return list
}

func MapObjects[L k8sutil.ListContinue, T meta.Object](ctx context.Context, k k8sutil.ListAPI[L], extract func(result L) []T) (ObjectList[T], error) {
	objects := ObjectList[T]{}

	if err := k8sutil.APIList[L](ctx, k, meta.ListOptions{}, func(result L, err error) error {
		if err != nil {
			return err
		}
		for _, obj := range extract(result) {
			obj.SetManagedFields(nil)

			objects[obj.GetUID()] = obj
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return objects, nil
}

func ListObjects[L k8sutil.ListContinue, T meta.Object](ctx context.Context, k k8sutil.ListAPI[L], extract func(result L) []T) ([]T, error) {
	objects, err := MapObjects[L, T](ctx, k, extract)
	if err != nil {
		return nil, err
	}

	return objects.AsList(), nil
}
