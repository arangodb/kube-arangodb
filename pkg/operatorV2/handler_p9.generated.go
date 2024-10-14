package operator

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

type HandleP9Func[P1, P2, P3, P4, P5, P6, P7, P8, P9 any] func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6, p7 P7, p8 P8, p9 P9) (bool, error)

type HandleP9ConditionFunc[P1, P2, P3, P4, P5, P6, P7, P8, P9 any] func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6, p7 P7, p8 P8, p9 P9) (*Condition, bool, error)

type HandleP9ConditionExtract[P1, P2, P3, P4, P5, P6, P7, P8, P9 any] func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6, p7 P7, p8 P8, p9 P9) *api.ConditionList

func HandleP9[P1, P2, P3, P4, P5, P6, P7, P8, P9 any](ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6, p7 P7, p8 P8, p9 P9, handler ...HandleP9Func[P1, P2, P3, P4, P5, P6, P7, P8, P9]) (bool, error) {
	isChanged := false
	for _, h := range handler {
		changed, err := h(ctx, p1, p2, p3, p4, p5, p6, p7, p8, p9)
		if changed {
			isChanged = true
		}

		if err != nil {
			return isChanged, err
		}
	}

	return isChanged, nil
}

func HandleP9WithStop[P1, P2, P3, P4, P5, P6, P7, P8, P9 any](ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6, p7 P7, p8 P8, p9 P9, handler ...HandleP9Func[P1, P2, P3, P4, P5, P6, P7, P8, P9]) (bool, error) {
	changed, err := HandleP9[P1, P2, P3, P4, P5, P6, P7, P8, P9](ctx, p1, p2, p3, p4, p5, p6, p7, p8, p9, handler...)
	if IsStop(err) {
		return changed, nil
	}

	return changed, err
}

func HandleP9WithCondition[P1, P2, P3, P4, P5, P6, P7, P8, P9 any](ctx context.Context, conditions *api.ConditionList, condition api.ConditionType, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6, p7 P7, p8 P8, p9 P9, handler ...HandleP9Func[P1, P2, P3, P4, P5, P6, P7, P8, P9]) (bool, error) {
	changed, err := HandleP9[P1, P2, P3, P4, P5, P6, P7, P8, P9](ctx, p1, p2, p3, p4, p5, p6, p7, p8, p9, handler...)
	return WithCondition(conditions, condition, changed, err)
}

func HandleP9Condition[P1, P2, P3, P4, P5, P6, P7, P8, P9 any](extract HandleP9ConditionExtract[P1, P2, P3, P4, P5, P6, P7, P8, P9], condition api.ConditionType, handler HandleP9ConditionFunc[P1, P2, P3, P4, P5, P6, P7, P8, P9]) HandleP9Func[P1, P2, P3, P4, P5, P6, P7, P8, P9] {
	return func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6, p7 P7, p8 P8, p9 P9) (bool, error) {
		c, changed, err := handler(ctx, p1, p2, p3, p4, p5, p6, p7, p8, p9)
		return WithConditionChange(extract(ctx, p1, p2, p3, p4, p5, p6, p7, p8, p9), condition, c, changed, err)
	}
}
