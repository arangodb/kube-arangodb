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
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
)

// Legacy

type HandleP5Func[P1, P2, P3, P4, P5 any] func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5) (bool, error)

type HandleP5ConditionFunc[P1, P2, P3, P4, P5 any] func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5) (*Condition, bool, error)

type HandleP5ConditionExtract[P1, P2, P3, P4, P5 any] func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5) *api.ConditionList

func HandleP5[P1, P2, P3, P4, P5 any](ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, handler ...HandleP5Func[P1, P2, P3, P4, P5]) (bool, error) {
	isChanged := false
	for _, h := range handler {
		changed, err := h(ctx, p1, p2, p3, p4, p5)
		if changed {
			isChanged = true
		}

		if err != nil {
			return isChanged, err
		}
	}

	return isChanged, nil
}

func HandleP5WithStop[P1, P2, P3, P4, P5 any](ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, handler ...HandleP5Func[P1, P2, P3, P4, P5]) (bool, error) {
	changed, err := HandleP5[P1, P2, P3, P4, P5](ctx, p1, p2, p3, p4, p5, handler...)
	if IsStop(err) {
		return changed, nil
	}

	return changed, err
}

func HandleP5WithCondition[P1, P2, P3, P4, P5 any](ctx context.Context, conditions *api.ConditionList, condition api.ConditionType, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, handler ...HandleP5Func[P1, P2, P3, P4, P5]) (bool, error) {
	changed, err := HandleP5[P1, P2, P3, P4, P5](ctx, p1, p2, p3, p4, p5, handler...)
	return WithCondition(conditions, condition, changed, err)
}

func HandleP5Condition[P1, P2, P3, P4, P5 any](extract HandleP5ConditionExtract[P1, P2, P3, P4, P5], condition api.ConditionType, handler HandleP5ConditionFunc[P1, P2, P3, P4, P5]) HandleP5Func[P1, P2, P3, P4, P5] {
	return func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5) (bool, error) {
		c, changed, err := handler(ctx, p1, p2, p3, p4, p5)
		return WithConditionChange(extract(ctx, p1, p2, p3, p4, p5), condition, c, changed, err)
	}
}

// New

type HandleSharedP5Func[P1, P2, P3, P4, P5 any] func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5) (bool, error)

type HandleSharedP5ConditionFunc[P1, P2, P3, P4, P5 any] func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5) (*Condition, bool, error)

type HandleSharedP5ConditionExtract[P1, P2, P3, P4, P5 any] func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5) *sharedApi.ConditionList

func HandleSharedP5[P1, P2, P3, P4, P5 any](ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, handler ...HandleSharedP5Func[P1, P2, P3, P4, P5]) (bool, error) {
	isChanged := false
	for _, h := range handler {
		changed, err := h(ctx, p1, p2, p3, p4, p5)
		if changed {
			isChanged = true
		}

		if err != nil {
			return isChanged, err
		}
	}

	return isChanged, nil
}

func HandleSharedP5WithStop[P1, P2, P3, P4, P5 any](ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, handler ...HandleSharedP5Func[P1, P2, P3, P4, P5]) (bool, error) {
	changed, err := HandleSharedP5[P1, P2, P3, P4, P5](ctx, p1, p2, p3, p4, p5, handler...)
	if IsStop(err) {
		return changed, nil
	}

	return changed, err
}

func HandleSharedP5WithCondition[P1, P2, P3, P4, P5 any](ctx context.Context, conditions *sharedApi.ConditionList, condition sharedApi.ConditionType, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, handler ...HandleSharedP5Func[P1, P2, P3, P4, P5]) (bool, error) {
	changed, err := HandleSharedP5[P1, P2, P3, P4, P5](ctx, p1, p2, p3, p4, p5, handler...)
	return WithSharedCondition(conditions, condition, changed, err)
}

func HandleSharedP5Condition[P1, P2, P3, P4, P5 any](extract HandleSharedP5ConditionExtract[P1, P2, P3, P4, P5], condition sharedApi.ConditionType, handler HandleSharedP5ConditionFunc[P1, P2, P3, P4, P5]) HandleSharedP5Func[P1, P2, P3, P4, P5] {
	return func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5) (bool, error) {
		c, changed, err := handler(ctx, p1, p2, p3, p4, p5)
		return WithSharedConditionChange(extract(ctx, p1, p2, p3, p4, p5), condition, c, changed, err)
	}
}
