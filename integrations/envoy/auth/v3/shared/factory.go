//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package shared

import (
	"context"

	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
)

func NewFactory(gen ...FactoryGen) Factory {
	return append(factories{}, gen...)
}

type Factory interface {
	Render(ctx context.Context, configuration Configuration) AuthHandler
}

type FactoryGen func(ctx context.Context, configuration Configuration) (AuthHandler, bool)

type factories []FactoryGen

func (f factories) Render(ctx context.Context, configuration Configuration) AuthHandler {
	hand := make(handlers, 0, len(f))

	for id := range f {
		if v, ok := f[id](ctx, configuration); ok {
			hand = append(hand, v)
		}
	}

	return hand
}

type handlers []AuthHandler

func (h handlers) Handle(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest, current *Response) error {
	for _, handler := range h {
		if err := handler.Handle(ctx, request, current); err != nil {
			return err
		}
	}

	return nil
}
