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

package svc

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type Background interface {
	Background(ctx context.Context)
}

func RunBackgroundSync(ctx context.Context, in any) {
	if h, ok := in.(Background); ok {
		h.Background(ctx)
	}
}

func RunBackground(in any) context.CancelFunc {
	if h, ok := in.(Background); ok {
		return util.RunContextAsync(context.Background(), h.Background)
	}

	return func() {

	}
}
