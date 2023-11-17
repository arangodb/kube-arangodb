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

package compare

type GenP0[T interface{}] func(spec, status *T) Func

type FuncGenP0[T interface{}] func(in GenP0[T]) Func

func NewFuncGenP0[T interface{}](spec, status *T) FuncGenP0[T] {
	return func(in GenP0[T]) Func {
		return in(spec, status)
	}
}

type GenP2[T, P1, P2 interface{}] func(p1 P1, p2 P2, spec, status *T) Func

type FuncGenP2[T, P1, P2 interface{}] func(in GenP2[T, P1, P2]) Func

func NewFuncGenP2[T, P1, P2 interface{}](p1 P1, p2 P2, spec, status *T) FuncGenP2[T, P1, P2] {
	return func(in GenP2[T, P1, P2]) Func {
		return in(p1, p2, spec, status)
	}
}
