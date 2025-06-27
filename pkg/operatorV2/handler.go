//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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
	"time"

	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

// Handler define interface for operator actions
type Handler interface {
	Name() string

	Handle(ctx context.Context, item operation.Item) error

	CanBeHandled(item operation.Item) bool
}

// HandlerTimeout extends the handle definition with timeout
type HandlerTimeout interface {
	Handler

	Timeout() time.Duration
}

// WithHandlerTimeout returns the handler with custom timeout
func WithHandlerTimeout(ctx context.Context, h Handler) (context.Context, context.CancelFunc) {
	if t, ok := h.(HandlerTimeout); ok {
		return context.WithTimeout(ctx, t.Timeout())
	}
	return globals.GetGlobals().Timeouts().Reconciliation().WithTimeout(ctx)
}
