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

package executor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/logging"
)

func Test_Executor(t *testing.T) {
	ctx := context.Background()

	logger := logging.Global().RegisterAndGetLogger("test", logging.Trace)

	require.NoError(t, Run(ctx, logger, 1, func(ctx context.Context, log logging.Logger, th Thread, h Handler) error {
		log.Info("Start main thread")
		defer log.Info("Complete main thread")

		h.RunAsync(ctx, func(ctx context.Context, log logging.Logger, th Thread, h Handler) error {
			log.Info("Start second thread")
			defer log.Info("Complete second thread")

			h.RunAsync(ctx, func(ctx context.Context, log logging.Logger, th Thread, h Handler) error {
				log.Info("Start third thread")
				defer log.Info("Complete third thread")

				return nil
			})

			return nil
		})

		h.WaitForSubThreads(th)

		return nil
	}))
}
