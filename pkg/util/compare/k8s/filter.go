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

package k8s

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FilterFunc[T any] func(in T) T

func FilterP[T any](in *T, filters ...FilterFunc[*T]) *T {
	if in == nil {
		return nil
	}

	return Filter(in, filters...)
}

func Filter[T any](in T, filters ...FilterFunc[T]) T {
	for _, f := range filters {
		in = f(in)
	}

	return in
}

func ObjectMetaFilter(in meta.ObjectMeta) meta.ObjectMeta {
	return meta.ObjectMeta{
		Labels:          in.Labels,
		Annotations:     in.Annotations,
		OwnerReferences: in.OwnerReferences,
		Finalizers:      in.Finalizers,
	}
}
