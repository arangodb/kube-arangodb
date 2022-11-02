//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/handlers/utils"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/retry"
	"github.com/arangodb/kube-arangodb/pkg/version"
)

var (
	cmdLifecycle = &cobra.Command{
		Use:    "lifecycle",
		Run:    executeUsage,
		Hidden: true,
	}

	cmdLifecyclePreStop = &cobra.Command{
		Use:    "preStop",
		Hidden: true,
	}
	cmdLifecyclePreStopFinalizers = &cobra.Command{
		Use:    "finalizers",
		Run:    cmdLifecyclePreStopRunFinalizer,
		Hidden: true,
	}
	cmdLifecyclePreStopPort = &cobra.Command{
		Use:    "port",
		Hidden: true,
	}
	cmdLifecycleCopy = &cobra.Command{
		Use:    "copy",
		Run:    cmdLifecycleCopyRun,
		Hidden: true,
	}

	lifecycleCopyOptions struct {
		TargetDir string
	}
)

func init() {
	cmdMain.AddCommand(cmdLifecycle)

	var preStopPort cmdLifecyclePreStopRunPort

	cmdLifecyclePreStopPort.RunE = preStopPort.run

	f := cmdLifecyclePreStopPort.Flags()

	f.DurationVar(&preStopPort.timeout, "timeout", 6*60*time.Minute, "PreStopTimeout")

	cmdLifecyclePreStop.AddCommand(cmdLifecyclePreStopFinalizers, cmdLifecyclePreStopPort)

	cmdLifecycle.AddCommand(cmdLifecyclePreStop)
	cmdLifecycle.AddCommand(cmdLifecycleCopy)
	cmdLifecycle.AddCommand(cmdLifecycleProbe)
	cmdLifecycle.AddCommand(cmdLifecycleWait)
	cmdLifecycle.AddCommand(cmdLifecycleStartup)

	cmdLifecycleCopy.Flags().StringVar(&lifecycleCopyOptions.TargetDir, "target", "", "Target directory to copy the executable to")
}

// Wait until all finalizers of the current pod have been removed.
func cmdLifecyclePreStopRunFinalizer(cmd *cobra.Command, args []string) {
	logger.Info("Starting arangodb-operator (%s), lifecycle preStop, version %s build %s", version.GetVersionV1().Edition.Title(), version.GetVersionV1().Version, version.GetVersionV1().Build)

	// Get environment
	namespace := os.Getenv(constants.EnvOperatorPodNamespace)
	if len(namespace) == 0 {
		logger.Fatal("%s environment variable missing", constants.EnvOperatorPodNamespace)
	}
	name := os.Getenv(constants.EnvOperatorPodName)
	if len(name) == 0 {
		logger.Fatal("%s environment variable missing", constants.EnvOperatorPodName)
	}

	// Create kubernetes client
	client, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		logger.Fatal("Client not initialised")
	}

	pods := client.Kubernetes().CoreV1().Pods(namespace)
	recentErrors := 0
	for {
		p, err := pods.Get(context.Background(), name, meta.GetOptions{})
		if kerrors.IsNotFound(err) {
			logger.Warn("Pod not found")
			return
		} else if err != nil {
			recentErrors++
			logger.Err(err).Error("Failed to get pod")
			if recentErrors > 20 {
				logger.Err(err).Fatal("Too many recent errors")
				return
			}
		} else {
			// We got our pod
			finalizerCount := len(p.GetFinalizers())
			if finalizerCount == 0 {
				// No more finalizers, we're done
				logger.Info("All finalizers gone, we can stop now")
				return
			}
			logger.Info("Waiting for %d more finalizers to be removed", finalizerCount)
		}
		// Wait a bit
		time.Sleep(time.Second)
	}
}

// Copy the executable to a given place.
func cmdLifecycleCopyRun(cmd *cobra.Command, args []string) {
	logger.Info("Starting arangodb-operator (%s), lifecycle copy, version %s build %s", version.GetVersionV1().Edition.Title(), version.GetVersionV1().Version, version.GetVersionV1().Build)

	exePath, err := os.Executable()
	if err != nil {
		logger.Err(err).Fatal("Failed to get executable path")
	}

	// Open source
	rd, err := os.Open(exePath)
	if err != nil {
		logger.Err(err).Fatal("Failed to open executable file")
	}
	defer rd.Close()

	// Open target
	targetPath := filepath.Join(lifecycleCopyOptions.TargetDir, filepath.Base(exePath))
	wr, err := os.Create(targetPath)
	if err != nil {
		logger.Err(err).Fatal("Failed to create target file")
	}
	defer wr.Close()

	if _, err := io.Copy(wr, rd); err != nil {
		logger.Err(err).Fatal("Failed to copy")
	}

	// Set file mode
	if err := os.Chmod(targetPath, 0755); err != nil {
		logger.Err(err).Fatal("Failed to chmod")
	}

	logger.Info("Executable copied to %s", targetPath)
}

type cmdLifecyclePreStopRunPort struct {
	timeout time.Duration
}

// Wait until port 8529 is closed.
func (c *cmdLifecyclePreStopRunPort) run(cmd *cobra.Command, args []string) error {
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(shared.ArangoPort))

	// Get environment
	namespace := os.Getenv(constants.EnvOperatorPodNamespace)
	if len(namespace) == 0 {
		logger.Fatal("%s environment variable missing", constants.EnvOperatorPodNamespace)
	}
	name := os.Getenv(constants.EnvOperatorPodName)
	if len(name) == 0 {
		logger.Fatal("%s environment variable missing", constants.EnvOperatorPodName)
	}

	// Create kubernetes client
	client, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		logger.Fatal("Client not initialised")
	}

	pods := client.Kubernetes().CoreV1().Pods(namespace)

	recentErrors := 0

	return retry.NewTimeout(func() error {
		conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)

		if err != nil {
			return retry.Interrput()
		}

		conn.Close()

		p, err := pods.Get(context.Background(), name, meta.GetOptions{})
		if kerrors.IsNotFound(err) {
			logger.Warn("Pod not found")
			return nil
		} else if err != nil {
			recentErrors++
			logger.Err(err).Error("Failed to get pod")
			if recentErrors > 20 {
				logger.Err(err).Fatal("Too many recent errors")
				return nil
			}
		} else {
			// We got our pod
			finalizers := utils.StringList(p.GetFinalizers())
			if !finalizers.Has(constants.FinalizerPodGracefulShutdown) {
				return retry.Interrput()
			}
		}

		return nil
	}).Timeout(125*time.Millisecond, c.timeout)
}
