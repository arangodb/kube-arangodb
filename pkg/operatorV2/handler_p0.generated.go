package operator

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

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
