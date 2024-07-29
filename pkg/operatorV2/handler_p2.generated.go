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

package operator

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

type HandleP2Func[P1, P2 any] func(ctx context.Context, p1 P1, p2 P2) (bool, error)

type HandleP2ConditionFunc[P1, P2 any] func(ctx context.Context, p1 P1, p2 P2) (*Condition, bool, error)

type HandleP2ConditionExtract[P1, P2 any] func(ctx context.Context, p1 P1, p2 P2) *api.ConditionList

func HandleP2[P1, P2 any](ctx context.Context, p1 P1, p2 P2, handler ...HandleP2Func[P1, P2]) (bool, error) {
	isChanged := false
	for _, h := range handler {
		changed, err := h(ctx, p1, p2)
		if changed {
			isChanged = true
		}

		if err != nil {
			return isChanged, err
		}
	}

	return isChanged, nil
}

func HandleP2WithStop[P1, P2 any](ctx context.Context, p1 P1, p2 P2, handler ...HandleP2Func[P1, P2]) (bool, error) {
	changed, err := HandleP2[P1, P2](ctx, p1, p2, handler...)
	if IsStop(err) {
		return changed, nil
	}

	return changed, err
}

func HandleP2WithCondition[P1, P2 any](ctx context.Context, conditions *api.ConditionList, condition api.ConditionType, p1 P1, p2 P2, handler ...HandleP2Func[P1, P2]) (bool, error) {
	changed, err := HandleP2[P1, P2](ctx, p1, p2, handler...)
	return WithCondition(conditions, condition, changed, err)
}

func HandleP2Condition[P1, P2 any](extract HandleP2ConditionExtract[P1, P2], condition api.ConditionType, handler HandleP2ConditionFunc[P1, P2]) HandleP2Func[P1, P2] {
	return func(ctx context.Context, p1 P1, p2 P2) (bool, error) {
		c, changed, err := handler(ctx, p1, p2)
		return WithConditionChange(extract(ctx, p1, p2), condition, c, changed, err)
	}
}
