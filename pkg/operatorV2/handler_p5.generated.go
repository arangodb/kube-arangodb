package operator

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

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
