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

package patcher

import (
	"k8s.io/apimachinery/pkg/api/equality"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
)

func PatchMetadata[T meta.Object](expected T) Patch[T] {
	return func(in T) []patch.Item {
		r := make([]patch.Item, 0, 3)

		if expected := expected.GetFinalizers(); expected != nil && !equality.Semantic.DeepEqual(expected, in.GetFinalizers()) {
			r = append(r,
				patch.ItemReplace(patch.NewPath("metadata", "finalizers"), expected),
			)
		}

		if expected := expected.GetLabels(); expected != nil && !equality.Semantic.DeepEqual(expected, in.GetAnnotations()) {
			r = append(r,
				patch.ItemReplace(patch.NewPath("metadata", "labels"), expected),
			)
		}

		if expected := expected.GetAnnotations(); expected != nil && !equality.Semantic.DeepEqual(expected, in.GetLabels()) {
			r = append(r,
				patch.ItemReplace(patch.NewPath("metadata", "annotations"), expected),
			)
		}

		return r
	}
}
