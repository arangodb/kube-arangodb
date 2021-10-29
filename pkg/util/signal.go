//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Jakub Wierzbowski
//

package util

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// CreateSignalContext creates and returns the context which is closed when one of the provided signal occurs.
// If the provided list of signals is empty, then SIGINT and SIGTERM is used by default.
func CreateSignalContext(ctx context.Context, signals ...os.Signal) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	ctxSignal, cancelSignal := context.WithCancel(ctx)
	sigChannel := make(chan os.Signal)

	if len(signals) > 0 {
		signal.Notify(sigChannel, signals...)
	} else {
		signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	}

	go func() {
		// Wait until signal occurs.
		<-sigChannel
		// Close the context which is used by the caller.
		cancelSignal()
	}()

	return ctxSignal
}
