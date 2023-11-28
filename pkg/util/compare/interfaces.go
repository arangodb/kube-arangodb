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

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

type Template[T interface{}] interface {
	GetTemplate() *T
	SetTemplate(*T)

	GetTemplateChecksum() string
	SetTemplateChecksum(string)

	GetChecksum() string
	SetChecksum(string)
}

type Checksum[T interface{}] func(in *T) (string, error)

type FuncGen[T interface{}] func(spec, status *T) Func

type Func func(builder api.ActionBuilder) (mode Mode, plan api.Plan, err error)

func Merge(f ...Func) Func {
	return func(builder api.ActionBuilder) (mode Mode, plan api.Plan, err error) {
		for _, q := range f {
			if m, p, err := q(builder); err != nil {
				return 0, nil, err
			} else {
				mode = mode.And(m)
				plan = append(plan, p...)
			}
		}

		return
	}
}

func Evaluate(builder api.ActionBuilder, f ...Func) (mode Mode, plan api.Plan, err error) {
	for _, q := range f {
		if m, p, err := q(builder); err != nil {
			return 0, nil, err
		} else {
			mode = mode.And(m)
			plan = append(plan, p...)
		}
	}

	return
}
