//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

type waitFunc func() error

func (w waitFunc) Require(t *testing.T) {
	require.NoError(t, w())
}

func executeSync(t *testing.T, ctx context.Context, args ...string) error {
	var c configuration

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		defer cancel()
		<-shutdown.Channel()
	}()

	c.test = &configurationTest{
		ctx:    ctx,
		cancel: cancel,
	}

	cmd := &cobra.Command{}

	tCmd := &cobra.Command{
		Use: "test",
	}

	require.NoError(t, c.Register(tCmd))

	cmd.AddCommand(tCmd)

	cmd.SetOut(os.Stdout)

	cmd.SetArgs(append([]string{"test"}, args...))
	logger.Info("Command: %s", strings.Join(args, " "))

	return cmd.Execute()
}

func executeAsync(t *testing.T, ctx context.Context, args ...string) waitFunc {
	ctx, cancel := context.WithCancel(ctx)

	var err error
	done := make(chan struct{})

	go func() {
		defer close(done)

		err = executeSync(t, ctx, args...)
	}()

	return func() error {
		cancel()
		<-done
		return err
	}
}

func startService(t *testing.T, args ...string) (waitFunc, int, int, int) {
	_, health := tests.ResolveAddress(t, "127.0.0.1:0")
	_, internal := tests.ResolveAddress(t, "127.0.0.1:0")
	_, external := tests.ResolveAddress(t, "127.0.0.1:0")

	cancel := executeAsync(t, shutdown.Context(), append([]string{
		fmt.Sprintf("--health.address=127.0.0.1:%d", health),
		fmt.Sprintf("--services.address=127.0.0.1:%d", internal),
		fmt.Sprintf("--services.external.address=127.0.0.1:%d", external),
		"--services.external.enabled",
	}, args...)...)

	tests.WaitForAddress(t, "127.0.0.1", health)
	tests.WaitForAddress(t, "127.0.0.1", internal)
	tests.WaitForAddress(t, "127.0.0.1", external)

	return cancel, health, internal, external
}
