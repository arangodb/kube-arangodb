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

package generic

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/gvk"
)

type Inspector[S meta.Object] interface {
	gvk.GVK

	ListSimple() []S
	GetSimple(name string) (S, bool)
	Filter(filters ...Filter[S]) []S
	Iterate(action Action[S], filters ...Filter[S]) error
	Read() ReadClient[S]
}

type Filter[S meta.Object] func(obj S) bool
type Action[S meta.Object] func(obj S) error

func FilterObject[S meta.Object](obj S, filters ...Filter[S]) bool {
	for _, f := range filters {
		if f == nil {
			continue
		}

		if !f(obj) {
			return false
		}
	}

	return true
}

func FilterByLabels[S meta.Object](labels map[string]string) Filter[S] {
	return func(obj S) bool {
		objLabels := obj.GetLabels()
		for key, value := range labels {
			v, ok := objLabels[key]
			if !ok {
				return false
			}

			if v != value {
				return false
			}
		}

		return true
	}
}
