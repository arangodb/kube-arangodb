//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

type HandleP3Func[P1, P2, P3 any] func(ctx context.Context, p1 P1, p2 P2, p3 P3) (bool, error)

type HandleP3ConditionFunc[P1, P2, P3 any] func(ctx context.Context, p1 P1, p2 P2, p3 P3) (*Condition, bool, error)

type HandleP3ConditionExtract[P1, P2, P3 any] func(ctx context.Context, p1 P1, p2 P2, p3 P3) *api.ConditionList

func HandleP3[P1, P2, P3 any](ctx context.Context, p1 P1, p2 P2, p3 P3, handler ...HandleP3Func[P1, P2, P3]) (bool, error) {
	isChanged := false
	for _, h := range handler {
		changed, err := h(ctx, p1, p2, p3)
		if changed {
			isChanged = true
		}

		if err != nil {
			return isChanged, err
		}
	}

	return isChanged, nil
}

func HandleP3WithStop[P1, P2, P3 any](ctx context.Context, p1 P1, p2 P2, p3 P3, handler ...HandleP3Func[P1, P2, P3]) (bool, error) {
	changed, err := HandleP3[P1, P2, P3](ctx, p1, p2, p3, handler...)
	if IsStop(err) {
		return changed, nil
	}

	return changed, err
}

func HandleP3WithCondition[P1, P2, P3 any](ctx context.Context, conditions *api.ConditionList, condition api.ConditionType, p1 P1, p2 P2, p3 P3, handler ...HandleP3Func[P1, P2, P3]) (bool, error) {
	changed, err := HandleP3[P1, P2, P3](ctx, p1, p2, p3, handler...)
	return WithCondition(conditions, condition, changed, err)
}

func HandleP3Condition[P1, P2, P3 any](extract HandleP3ConditionExtract[P1, P2, P3], condition api.ConditionType, handler HandleP3ConditionFunc[P1, P2, P3]) HandleP3Func[P1, P2, P3] {
	return func(ctx context.Context, p1 P1, p2 P2, p3 P3) (bool, error) {
		c, changed, err := handler(ctx, p1, p2, p3)
		return WithConditionChange(extract(ctx, p1, p2, p3), condition, c, changed, err)
	}
}
