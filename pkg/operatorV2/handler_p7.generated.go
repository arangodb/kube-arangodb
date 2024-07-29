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

type HandleP7Func[P1, P2, P3, P4, P5, P6, P7 any] func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6, p7 P7) (bool, error)

type HandleP7ConditionFunc[P1, P2, P3, P4, P5, P6, P7 any] func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6, p7 P7) (*Condition, bool, error)

type HandleP7ConditionExtract[P1, P2, P3, P4, P5, P6, P7 any] func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6, p7 P7) *api.ConditionList

func HandleP7[P1, P2, P3, P4, P5, P6, P7 any](ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6, p7 P7, handler ...HandleP7Func[P1, P2, P3, P4, P5, P6, P7]) (bool, error) {
	isChanged := false
	for _, h := range handler {
		changed, err := h(ctx, p1, p2, p3, p4, p5, p6, p7)
		if changed {
			isChanged = true
		}

		if err != nil {
			return isChanged, err
		}
	}

	return isChanged, nil
}

func HandleP7WithStop[P1, P2, P3, P4, P5, P6, P7 any](ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6, p7 P7, handler ...HandleP7Func[P1, P2, P3, P4, P5, P6, P7]) (bool, error) {
	changed, err := HandleP7[P1, P2, P3, P4, P5, P6, P7](ctx, p1, p2, p3, p4, p5, p6, p7, handler...)
	if IsStop(err) {
		return changed, nil
	}

	return changed, err
}

func HandleP7WithCondition[P1, P2, P3, P4, P5, P6, P7 any](ctx context.Context, conditions *api.ConditionList, condition api.ConditionType, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6, p7 P7, handler ...HandleP7Func[P1, P2, P3, P4, P5, P6, P7]) (bool, error) {
	changed, err := HandleP7[P1, P2, P3, P4, P5, P6, P7](ctx, p1, p2, p3, p4, p5, p6, p7, handler...)
	return WithCondition(conditions, condition, changed, err)
}

func HandleP7Condition[P1, P2, P3, P4, P5, P6, P7 any](extract HandleP7ConditionExtract[P1, P2, P3, P4, P5, P6, P7], condition api.ConditionType, handler HandleP7ConditionFunc[P1, P2, P3, P4, P5, P6, P7]) HandleP7Func[P1, P2, P3, P4, P5, P6, P7] {
	return func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6, p7 P7) (bool, error) {
		c, changed, err := handler(ctx, p1, p2, p3, p4, p5, p6, p7)
		return WithConditionChange(extract(ctx, p1, p2, p3, p4, p5, p6, p7), condition, c, changed, err)
	}
}
