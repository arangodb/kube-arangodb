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

type HandleP0Func func(ctx context.Context) (bool, error)

type HandleP0ConditionFunc func(ctx context.Context) (*Condition, bool, error)

type HandleP0ConditionExtract func(ctx context.Context) *api.ConditionList

func HandleP0(ctx context.Context, handler ...HandleP0Func) (bool, error) {
	isChanged := false
	for _, h := range handler {
		changed, err := h(ctx)
		if changed {
			isChanged = true
		}

		if err != nil {
			return isChanged, err
		}
	}

	return isChanged, nil
}

func HandleP0WithStop(ctx context.Context, handler ...HandleP0Func) (bool, error) {
	changed, err := HandleP0(ctx, handler...)
	if IsStop(err) {
		return changed, nil
	}

	return changed, err
}

func HandleP0WithCondition(ctx context.Context, conditions *api.ConditionList, condition api.ConditionType, handler ...HandleP0Func) (bool, error) {
	changed, err := HandleP0(ctx, handler...)
	return WithCondition(conditions, condition, changed, err)
}

func HandleP0Condition(extract HandleP0ConditionExtract, condition api.ConditionType, handler HandleP0ConditionFunc) HandleP0Func {
	return func(ctx context.Context) (bool, error) {
		c, changed, err := handler(ctx)
		return WithConditionChange(extract(ctx), condition, c, changed, err)
	}
}

// New

type HandleSharedP0Func func(ctx context.Context) (bool, error)

type HandleSharedP0ConditionFunc func(ctx context.Context) (*Condition, bool, error)

type HandleSharedP0ConditionExtract func(ctx context.Context) *sharedApi.ConditionList

func HandleSharedP0(ctx context.Context, handler ...HandleSharedP0Func) (bool, error) {
	isChanged := false
	for _, h := range handler {
		changed, err := h(ctx)
		if changed {
			isChanged = true
		}

		if err != nil {
			return isChanged, err
		}
	}

	return isChanged, nil
}

func HandleSharedP0WithStop(ctx context.Context, handler ...HandleSharedP0Func) (bool, error) {
	changed, err := HandleSharedP0(ctx, handler...)
	if IsStop(err) {
		return changed, nil
	}

	return changed, err
}

func HandleSharedP0WithCondition(ctx context.Context, conditions *sharedApi.ConditionList, condition sharedApi.ConditionType, handler ...HandleSharedP0Func) (bool, error) {
	changed, err := HandleSharedP0(ctx, handler...)
	return WithSharedCondition(conditions, condition, changed, err)
}

func HandleSharedP0Condition(extract HandleSharedP0ConditionExtract, condition sharedApi.ConditionType, handler HandleSharedP0ConditionFunc) HandleSharedP0Func {
	return func(ctx context.Context) (bool, error) {
		c, changed, err := handler(ctx)
		return WithSharedConditionChange(extract(ctx), condition, c, changed, err)
	}
}
