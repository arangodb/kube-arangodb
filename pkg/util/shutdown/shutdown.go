//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

package shutdown

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	ctx, stop = context.WithCancel(context.Background())

	sigChannel := make(chan os.Signal, 2)

	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)

	go func() {
		defer stop()
		<-sigChannel
	}()
}

var (
	ctx  context.Context
	stop context.CancelFunc
)

func Context() context.Context {
	return ctx
}

func Stop() {
	defer stop()
}

func Channel() <-chan struct{} {
	return ctx.Done()
}
