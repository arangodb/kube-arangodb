//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package main

import (
	"context"
	"os"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/rs/zerolog"

	"github.com/arangodb/kube-arangodb/tests/duration/simple"
	t "github.com/arangodb/kube-arangodb/tests/duration/test"
)

var (
	testPeriod = time.Minute * 2
)

// runTestLoop keeps running tests until the given context is canceled.
func runTestLoop(ctx context.Context, client driver.Client, duration time.Duration) {
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	endTime := time.Now().Add(duration)
	reportDir := "."
	tests := []t.TestScript{}
	tests = append(tests, simple.NewSimpleTest(log, reportDir, simple.SimpleConfig{
		MaxDocuments:   500,
		MaxCollections: 50,
	}))

	log.Info().Msg("Starting tests")
	listener := &testListener{
		Log: log,
		FailedCallback: func() {
			log.Fatal().Msg("Too many recent failures. Aborting test")
		},
	}
	for _, tst := range tests {
		if err := tst.Start(client, listener); err != nil {
			log.Fatal().Err(err).Msg("Failed to start test")
		}
	}
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		// Check end time
		if time.Now().After(endTime) {
			log.Info().Msgf("Test has run for %s. We're done", duration)
			return
		}

		// Run tests
		log.Info().Msg("Running tests...")
		select {
		case <-time.After(testPeriod):
			// Continue
		case <-ctx.Done():
			return
		}

		// Pause tests
		log.Info().Msg("Pause tests")
		for _, tst := range tests {
			if err := tst.Pause(); err != nil {
				log.Fatal().Err(err).Msg("Failed to pause test")
			}
		}

		// Wait for tests to really pause
		log.Info().Msg("Waiting for tests to reach pausing state")
		for _, tst := range tests {
			for !tst.Status().Pausing {
				select {
				case <-time.After(time.Second):
					// Continue
				case <-ctx.Done():
					return
				}
			}
		}

		// Resume tests
		log.Info().Msg("Resuming tests")
		for _, tst := range tests {
			if err := tst.Resume(); err != nil {
				log.Fatal().Err(err).Msg("Failed to resume test")
			}
		}
	}
}
