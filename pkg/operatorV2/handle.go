//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

import "context"

type HandleP0Func func(ctx context.Context) (bool, error)

type HandleP1Func[P1 interface{}] func(ctx context.Context, p1 P1) (bool, error)

type HandleP2Func[P1, P2 interface{}] func(ctx context.Context, p1 P1, p2 P2) (bool, error)

type HandleP3Func[P1, P2, P3 interface{}] func(ctx context.Context, p1 P1, p2 P2, p3 P3) (bool, error)

type HandleP4Func[P1, P2, P3, P4 interface{}] func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4) (bool, error)

type HandleP9Func[P1, P2, P3, P4, P5, P6, P7, P8, P9 interface{}] func(ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6, p7 P7, p8 P8, p9 P9) (bool, error)

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

func HandleP1[P1 interface{}](ctx context.Context, p1 P1, handler ...HandleP1Func[P1]) (bool, error) {
	isChanged := false
	for _, h := range handler {
		changed, err := h(ctx, p1)
		if changed {
			isChanged = true
		}

		if err != nil {
			return isChanged, err
		}
	}

	return isChanged, nil
}

func HandleP2[P1, P2 interface{}](ctx context.Context, p1 P1, p2 P2, handler ...HandleP2Func[P1, P2]) (bool, error) {
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

func HandleP3[P1, P2, P3 interface{}](ctx context.Context, p1 P1, p2 P2, p3 P3, handler ...HandleP3Func[P1, P2, P3]) (bool, error) {
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

func HandleP4[P1, P2, P3, P4 interface{}](ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, handler ...HandleP4Func[P1, P2, P3, P4]) (bool, error) {
	isChanged := false
	for _, h := range handler {
		changed, err := h(ctx, p1, p2, p3, p4)
		if changed {
			isChanged = true
		}

		if err != nil {
			return isChanged, err
		}
	}

	return isChanged, nil
}

func HandleP9[P1, P2, P3, P4, P5, P6, P7, P8, P9 interface{}](ctx context.Context, p1 P1, p2 P2, p3 P3, p4 P4, p5 P5, p6 P6, p7 P7, p8 P8, p9 P9, handler ...HandleP9Func[P1, P2, P3, P4, P5, P6, P7, P8, P9]) (bool, error) {
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
