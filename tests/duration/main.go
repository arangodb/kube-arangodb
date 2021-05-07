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
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/pkg/errors"

	"github.com/arangodb/kube-arangodb/pkg/util/retry"
)

const (
	defaultTestDuration = time.Hour * 24 * 7 // 7 days
)

var (
	maskAny          = errors.WithStack
	userName         string
	password         string
	clusterEndpoints string
	testDuration     time.Duration
)

func init() {
	flag.StringVar(&userName, "username", "", "Authenticating username")
	flag.StringVar(&password, "password", "", "Authenticating password")
	flag.StringVar(&clusterEndpoints, "cluster", "", "Endpoints for database cluster")
	flag.DurationVar(&testDuration, "duration", defaultTestDuration, "Duration of the test")
}

func main() {
	flag.Parse()

	// Create clients & wait for cluster available
	client, err := createClusterClient(clusterEndpoints, userName, password)
	if err != nil {
		log.Fatalf("Failed to create cluster client: %#v\n", err)
	}
	if err := waitUntilClusterUp(client); err != nil {
		log.Fatalf("Failed to reach cluster: %#v\n", err)
	}

	// Start running tests
	ctx, cancel := context.WithCancel(context.Background())
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	go handleSignal(sigChannel, cancel)
	runTestLoop(ctx, client, testDuration)
}

// createClusterClient creates a configuration, connection and client for
// one of the two ArangoDB clusters in the test. It uses the go-driver.
// It needs a list of endpoints.
func createClusterClient(endpoints string, user string, password string) (driver.Client, error) {
	// This will always use HTTP, and user and password authentication
	config := http.ConnectionConfig{
		Endpoints: strings.Split(endpoints, ","),
		TLSConfig: &tls.Config{InsecureSkipVerify: true},
	}
	connection, err := http.NewConnection(config)
	if err != nil {
		return nil, maskAny(err)
	}
	clientCfg := driver.ClientConfig{
		Connection:     connection,
		Authentication: driver.BasicAuthentication(user, password),
	}
	client, err := driver.NewClient(clientCfg)
	if err != nil {
		return nil, maskAny(err)
	}
	return client, nil
}

func waitUntilClusterUp(c driver.Client) error {
	op := func() error {
		ctx := context.Background()
		if _, err := c.Version(ctx); err != nil {
			return maskAny(err)
		}
		return nil
	}
	if err := retry.Retry(op, time.Minute); err != nil {
		return maskAny(err)
	}
	return nil
}

// handleSignal listens for termination signals and stops this process on termination.
func handleSignal(sigChannel chan os.Signal, cancel context.CancelFunc) {
	signalCount := 0
	for s := range sigChannel {
		signalCount++
		fmt.Println("Received signal:", s)
		if signalCount > 1 {
			os.Exit(1)
		}
		cancel()
	}
}
