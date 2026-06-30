//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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

// Package collect implements the collector run by the arangod/gateway postStart lifecycle hook.
//
// The collector is a simple application started by the postStart hook. It runs in the foreground,
// collects the events for the current boot on a fixed interval and stops as soon as a cycle
// succeeds. It is not a daemon - there is no background process to keep alive - it simply runs until
// it is done. A timeout bounds the whole run so it cannot retry indefinitely if it never succeeds.
//
// Collected events are printed to stdout.
//
// Each boot is identified by a unique boot id (see shutdown.BootID), which is attached to every
// event the collector emits so that all activity from a single pod boot can be correlated.
package collect

import (
	"context"
	"fmt"
	"os"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
	"github.com/arangodb/kube-arangodb/pkg/version"
)

const (
	// DefaultInterval is the default retry interval between collector cycles.
	DefaultInterval = 30 * time.Second

	// DefaultTimeout is the default upper bound for the collector, after which it gives up.
	DefaultTimeout = 30 * time.Minute
)

// Options configures a collector run.
type Options struct {
	// Interval is the retry interval between collector cycles. Defaults to DefaultInterval.
	Interval time.Duration

	// Timeout bounds the whole run. When the timeout elapses before a cycle succeeds the collector
	// gives up and returns. A non-positive value falls back to DefaultTimeout.
	Timeout time.Duration
}

// interval returns the configured retry interval, falling back to DefaultInterval.
func (o Options) interval() time.Duration {
	if o.Interval <= 0 {
		return DefaultInterval
	}
	return o.Interval
}

// timeout returns the configured run timeout, falling back to DefaultTimeout.
func (o Options) timeout() time.Duration {
	if o.Timeout <= 0 {
		return DefaultTimeout
	}
	return o.Timeout
}

// PostStart is the entrypoint for the postStart collector lifecycle hook.
//
// It runs the collector loop in the foreground and returns once a cycle succeeds, the context is
// cancelled, or the timeout elapses. It never spawns a background process.
func PostStart(ctx context.Context, opts Options) error {
	ctx, cancel := context.WithTimeout(ctx, opts.timeout())
	defer cancel()

	return run(ctx, opts)
}

// run executes the collector loop. A single failing cycle never aborts the run, it just waits for
// the next interval and retries, until a cycle succeeds or the context is closed (cancelled or timed
// out).
func run(ctx context.Context, opts Options) error {
	// bootID is stable for the lifetime of this process and identifies the current boot.
	bootID := shutdown.BootID()

	// created is the process boot time, injected into every collected event so all events share a
	// single, stable timestamp regardless of which cycle produced them. It is trimmed to seconds to
	// keep cross-platform compatibility with the events integration.
	created := shutdown.BootTime().Truncate(time.Second)

	logger.Str("bootID", bootID).Info("Starting arangodb-operator collector (%s), version %s build %s",
		version.GetVersionV1().Edition.Title(), version.GetVersionV1().Version, version.GetVersionV1().Build)

	t := time.NewTicker(opts.interval())
	defer t.Stop()

	for {
		if err := collect(bootID, created); err != nil {
			logger.Err(err).Str("bootID", bootID).Warn("Collector cycle failed, will retry")
		} else {
			logger.Str("bootID", bootID).Info("Collector finished")
			return nil
		}

		select {
		case <-ctx.Done():
			logger.Err(ctx.Err()).Str("bootID", bootID).Warn("Collector stopped before completion")
			return ctx.Err()
		case <-t.C:
		}
	}
}

// collect performs a single collection cycle for the given boot: it runs every registered collector,
// each pushing its events into a shared collector, waits until all of them have completed and prints
// the aggregated events to stdout. Every collected event is stamped with the boot id and the start
// timestamp so all activity from a single pod boot can be correlated.
func collect(bootID string, created time.Time) error {
	events, err := GetCollector().Collect()
	if err != nil {
		return err
	}

	stamp(events, bootID, created)

	if err := print(events); err != nil {
		return err
	}

	logger.Str("bootID", bootID).Int("events", len(events)).Debug("Collected events")
	return nil
}

// stamp injects the shared boot id and start timestamp into every collected event, so events from a
// single boot can be correlated regardless of which collector produced them.
func stamp(events []*Event, bootID string, created time.Time) {
	ts := timestamppb.New(created)

	for _, e := range events {
		e.Created = ts

		if e.Dimensions == nil {
			e.Dimensions = map[string]string{}
		}
		e.Dimensions[dimensionBootID] = bootID
	}
}

// print writes every event to stdout as a single JSON line.
func print(events []*Event) error {
	for _, e := range events {
		data, err := protojson.Marshal(e)
		if err != nil {
			return errors.Wrapf(err, "unable to marshal event")
		}

		if _, err := fmt.Fprintln(os.Stdout, string(data)); err != nil {
			return errors.Wrapf(err, "unable to write event")
		}
	}

	return nil
}
