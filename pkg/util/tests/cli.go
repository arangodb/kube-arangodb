//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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

package tests

import (
	"context"
	goStrings "strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func RunCLI(ctx context.Context, cmd *cobra.Command, args ...string) error {
	cmd.SetArgs(args)

	return cmd.ExecuteContext(ctx)
}

func RunWithCLI(t *testing.T, cmd *cobra.Command, args ...string) func(in func(t *testing.T)) {
	return func(in func(t *testing.T)) {
		t.Run("CLI", func(t *testing.T) {
			ctx, c := context.WithCancel(t.Context())
			defer c()

			t.Logf("Executing: %s", goStrings.Join(args, " "))

			done := make(chan struct{})

			go func() {
				defer close(done)

				require.NoError(t, RunCLI(ctx, cmd, args...))
			}()

			defer func() { c(); <-done }()

			in(t)

			c()

			<-done
		})
	}
}
