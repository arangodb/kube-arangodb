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

package gateway

import (
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type TypedFilterConfigGen interface {
	RenderTypedFilterConfig() (util.KV[string, *anypb.Any], error)
}

func NewTypedFilterConfig(gens ...TypedFilterConfigGen) (map[string]*anypb.Any, error) {
	generated := map[string]*anypb.Any{}

	for _, g := range gens {
		if k, err := g.RenderTypedFilterConfig(); err != nil {
			return nil, err
		} else {
			if _, ok := generated[k.K]; ok {
				return nil, errors.Errorf("Duplicated key: %s", k.K)
			}

			if k.V == nil {
				continue
			}

			generated[k.K] = k.V
		}
	}

	if len(generated) == 0 {
		return nil, nil
	}

	return generated, nil
}
