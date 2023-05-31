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

package k8sutil

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ListContinue interface {
	GetContinue() string
}

type ListAPI[T ListContinue] interface {
	List(ctx context.Context, opts meta.ListOptions) (T, error)
}

func APIList[T ListContinue](ctx context.Context, api ListAPI[T], opts meta.ListOptions, parser func(result T, err error) error) error {
	result, err := api.List(ctx, opts)
	for {
		if err := parser(result, err); err != nil {
			return err
		}

		if c := result.GetContinue(); c == "" {
			return nil
		} else {
			result, err = api.List(ctx, meta.ListOptions{
				Continue: result.GetContinue(),
			})
		}
	}
}
