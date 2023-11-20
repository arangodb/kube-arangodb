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

type HandleP0Func func(tx context.Context) error

type HandleP1Func[P1 interface{}] func(tx context.Context, p1 P1) error

type HandleP2Func[P1, P2 interface{}] func(tx context.Context, p1 P1, p2 P2) error

func HandleP0(ctx context.Context, handler ...HandleP0Func) error {
	for _, h := range handler {
		if err := h(ctx); err != nil {
			return err
		}
	}

	return nil
}

func HandleP1[P1 interface{}](ctx context.Context, p1 P1, handler ...HandleP1Func[P1]) error {
	for _, h := range handler {
		if err := h(ctx, p1); err != nil {
			return err
		}
	}

	return nil
}

func HandleP2[P1, P2 interface{}](ctx context.Context, p1 P1, p2 P2, handler ...HandleP2Func[P1, P2]) error {
	for _, h := range handler {
		if err := h(ctx, p1, p2); err != nil {
			return err
		}
	}

	return nil
}
