package operator

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

type HandleP6Func[P1, P2, P3, P4, P5, P6 any] func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6) (bool, error)

type HandleP6ConditionFunc[P1, P2, P3, P4, P5, P6 any] func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6) (*Condition, bool, error)

type HandleP6ConditionExtract[P1, P2, P3, P4, P5, P6 any] func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6) *api.ConditionList

func HandleP6[P1, P2, P3, P4, P5, P6 any](ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6, handler ...HandleP6Func[P1, P2, P3, P4, P5, P6]) (bool, error) {
	isChanged := false
	for _, h := range handler {
		changed, err := h(ctx, p1, p2, p3, p4, p5, p6)
		if changed {
			isChanged = true
		}

		if err != nil {
			return isChanged, err
		}
	}

	return isChanged, nil
}

func HandleP6WithStop[P1, P2, P3, P4, P5, P6 any](ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6, handler ...HandleP6Func[P1, P2, P3, P4, P5, P6]) (bool, error) {
	changed, err := HandleP6[P1, P2, P3, P4, P5, P6](ctx, p1, p2, p3, p4, p5, p6, handler...)
	if IsStop(err) {
		return changed, nil
	}

	return changed, err
}

func HandleP6WithCondition[P1, P2, P3, P4, P5, P6 any](ctx context.Context, conditions *api.ConditionList, condition api.ConditionType, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6, handler ...HandleP6Func[P1, P2, P3, P4, P5, P6]) (bool, error) {
	changed, err := HandleP6[P1, P2, P3, P4, P5, P6](ctx, p1, p2, p3, p4, p5, p6, handler...)
	return WithCondition(conditions, condition, changed, err)
}

func HandleP6Condition[P1, P2, P3, P4, P5, P6 any](extract HandleP6ConditionExtract[P1, P2, P3, P4, P5, P6], condition api.ConditionType, handler HandleP6ConditionFunc[P1, P2, P3, P4, P5, P6]) HandleP6Func[P1, P2, P3, P4, P5, P6] {
	return func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6) (bool, error) {
		c, changed, err := handler(ctx, p1, p2, p3, p4, p5, p6)
		return WithConditionChange(extract(ctx, p1, p2, p3, p4, p5, p6), condition, c, changed, err)
	}
}
