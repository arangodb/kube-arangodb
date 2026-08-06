//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package cmd

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/collect"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
)

var (
	// cmdLifecyclePostStart is the parent for postStart lifecycle hooks (`lifecycle postStart ...`).
	cmdLifecyclePostStart = &cobra.Command{
		Use:    "postStart",
		RunE:   cli.Usage,
		Hidden: true,
	}

	// cmdLifecyclePostStartCollector runs the collector as a postStart hook (`lifecycle postStart collector`).
	cmdLifecyclePostStartCollector = &cobra.Command{
		Use:    "collector",
		Hidden: true,
	}

	// lifecyclePostStartCollectorOptions holds the flags bound to the collector command.
	lifecyclePostStartCollectorOptions collect.Options

	// lifecyclePostStartCollectorForeground runs the collector loop in the foreground instead of
	// spawning a detached background process. It is set on the re-executed child, and can be set
	// manually to run the collector synchronously (e.g. for debugging).
	lifecyclePostStartCollectorForeground bool
)

func init() {
	cmdLifecyclePostStartCollector.RunE = cmdLifecyclePostStartCollectorRunE

	f := cmdLifecyclePostStartCollector.Flags()
	f.DurationVar(&lifecyclePostStartCollectorOptions.Interval, "interval", collect.DefaultInterval, "Collector retry interval")
	f.DurationVar(&lifecyclePostStartCollectorOptions.Timeout, "timeout", collect.DefaultTimeout, "Collector run timeout")
	f.StringVar(&lifecyclePostStartCollectorOptions.Endpoint, "endpoint", "", "ArangoDB endpoint the startup event is written to; empty prints to stdout")
	f.StringVar(&lifecyclePostStartCollectorOptions.JWTPath, "jwt-path", "", "Folder holding the cluster JWT secret used to authenticate to ArangoDB")
	f.BoolVar(&lifecyclePostStartCollectorForeground, "foreground", false, "Run the collector loop in the foreground instead of detaching a background process")
	_ = f.MarkHidden("foreground")

	cmdLifecyclePostStart.AddCommand(cmdLifecyclePostStartCollector)
	cmdLifecycle.AddCommand(cmdLifecyclePostStart)
}

// cmdLifecyclePostStartCollectorRunE runs the collector.
//
// A postStart hook blocks the container from reaching a Running state until it returns, so the
// collector - which may wait for the local arangod to come up - must never run synchronously here:
// on a cluster member (e.g. a DBServer) that would deadlock startup. Instead the hook spawns a
// detached background copy of itself (--foreground) and returns immediately, letting the container
// start while the collector waits and writes its event in the background. The ArangoDB endpoint and
// JWT folder are provided by the operator through the hook flags; when no endpoint is given the
// collector falls back to printing the event to stdout.
func cmdLifecyclePostStartCollectorRunE(cmd *cobra.Command, _ []string) error {
	if lifecyclePostStartCollectorForeground {
		return collect.PostStart(cmd.Context(), lifecyclePostStartCollectorOptions)
	}

	return spawnDetachedCollector(lifecyclePostStartCollectorOptions)
}

// spawnDetachedCollector re-executes this binary as a detached (own session) background process that
// runs the collector loop in the foreground, then returns immediately without waiting for it.
func spawnDetachedCollector(opts collect.Options) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	args := []string{"lifecycle", "postStart", "collector", "--foreground",
		"--interval", opts.Interval.String(),
		"--timeout", opts.Timeout.String(),
	}
	if opts.Endpoint != "" {
		args = append(args, "--endpoint", opts.Endpoint)
	}
	if opts.JWTPath != "" {
		args = append(args, "--jwt-path", opts.JWTPath)
	}

	c := exec.Command(exe, args...)
	// Detach from the hook process: a new session survives the hook returning and is not killed
	// with the hook's process group.
	c.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	return c.Start()
}
