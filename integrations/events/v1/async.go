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

package v1

import (
	"context"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

func WithAsync[IN proto.Message, H RemoteStore[IN]](in H, size int, timeout time.Duration, delay time.Duration) RemoteStore[IN] {
	return &asyncRemoteWriter[IN, H]{
		upstream: in,
		cache:    make(chan []IN, size),
		timeout:  timeout,
		delay:    delay,
	}
}

type asyncRemoteWriter[IN proto.Message, H RemoteStore[IN]] struct {
	upstream H

	cache chan []IN

	timeout time.Duration
	delay   time.Duration
}

func (a *asyncRemoteWriter[IN, H]) Init(ctx context.Context) error {
	return a.upstream.Init(ctx)
}

func (a *asyncRemoteWriter[IN, H]) Background(ctx context.Context) {
	logger.Info("Async background started")
	defer func() {
		logger.Info("Async background completed")
	}()

	for {
		select {
		case <-ctx.Done():
			close(a.cache)
			for events := range a.cache {
				// Cleanup the queue
				a.emitEvents(events...)
			}
			return
		case events := <-a.cache:
			a.emitEvents(events...)
		}
	}
}

func (a *asyncRemoteWriter[IN, H]) emitEvents(events ...IN) {
	if len(events) == 0 {
		return
	}

	timeoutTimer := time.NewTimer(a.timeout)
	defer timeoutTimer.Stop()

	delayTimer := time.NewTicker(a.delay)
	defer delayTimer.Stop()

	for {
		err := globals.GetGlobals().Timeouts().ArangoD().RunWithTimeout(context.Background(), func(ctxChild context.Context) error {
			return a.upstream.Emit(ctxChild, events...)
		})
		if err != nil {
			logger.Err(err).Warn("Unable to send events batch, retry")
		} else {
			logger.Debug("Batch sent")
			return
		}

		select {
		case <-delayTimer.C:
			continue
		case <-timeoutTimer.C:
			logger.Error("Unable to send events in expected time")
			return
		}
	}
}

func (a *asyncRemoteWriter[IN, H]) Emit(ctx context.Context, events ...IN) error {
	if len(events) == 0 {
		return nil
	}

	timeout := time.NewTimer(time.Second)
	defer timeout.Stop()

	select {
	case a.cache <- events:
		return nil
	case <-timeout.C:
		return errors.Errorf("timeout waiting for events to be scheduled")
	}
}
