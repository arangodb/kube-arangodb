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

package integrations

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

const ShutdownV1ContextCancelFuncName util.ContextKey = "shutdown.v1.context.cancel"

func WithShutdownV1ContextCancelFunc(ctx context.Context, c context.CancelFunc) context.Context {
	return context.WithValue(ctx, ShutdownV1ContextCancelFuncName, c)
}

func extractShutdownFunc(ctx context.Context) context.CancelFunc {
	if z, ok := ctx.Value(ShutdownV1ContextCancelFuncName).(context.CancelFunc); ok {
		return z
	}

	return shutdown.Stop
}
