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

package list

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
)

func APIMap[L generic.ListContinue, S meta.Object](ctx context.Context, i generic.ListInterface[L], opts meta.ListOptions, call generic.ExtractorList[L, S]) (map[string]S, error) {
	res, err := APIList(ctx, i, opts, call)
	if err != nil {
		return nil, err
	}

	result := make(map[string]S, len(res))

	for _, el := range res {
		if _, ok := result[el.GetName()]; ok {
			return nil, errors.Errorf("Key %s already exists", el.GetName())
		}

		result[el.GetName()] = el
	}

	return result, nil
}

func APIList[L generic.ListContinue, S meta.Object](ctx context.Context, i generic.ListInterface[L], opts meta.ListOptions, call generic.ExtractorList[L, S]) ([]S, error) {
	var results []S

	var cont string

	for {
		opts.Continue = cont
		if v := globals.GetGlobals().Kubernetes().RequestBatchSize().Get(); opts.Limit <= 0 || opts.Limit > v {
			opts.Limit = v
		}
		res, err := i.List(ctx, opts)
		if err != nil {
			return nil, err
		}

		objs := call(res)

		results = append(results, objs...)

		if res.GetContinue() == "" {
			break
		}

		cont = res.GetContinue()
	}

	return results, nil
}
