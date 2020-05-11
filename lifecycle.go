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
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

var (
	cmdLifecycle = &cobra.Command{
		Use:    "lifecycle",
		Run:    cmdUsage,
		Hidden: true,
	}

	cmdLifecyclePreStop = &cobra.Command{
		Use:    "preStop",
		Run:    cmdLifecyclePreStopRun,
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
	cmdLifecycle.AddCommand(cmdLifecyclePreStop)
	cmdLifecycle.AddCommand(cmdLifecycleCopy)

	cmdLifecycleCopy.Flags().StringVar(&lifecycleCopyOptions.TargetDir, "target", "", "Target directory to copy the executable to")
}

// Wait until all finalizers of the current pod have been removed.
func cmdLifecyclePreStopRun(cmd *cobra.Command, args []string) {
	cliLog.Info().Msgf("Starting arangodb-operator, lifecycle preStop, version %s build %s", projectVersion, projectBuild)

	// Get environment
	namespace := os.Getenv(constants.EnvOperatorPodNamespace)
	if len(namespace) == 0 {
		cliLog.Fatal().Msgf("%s environment variable missing", constants.EnvOperatorPodNamespace)
	}
	name := os.Getenv(constants.EnvOperatorPodName)
	if len(name) == 0 {
		cliLog.Fatal().Msgf("%s environment variable missing", constants.EnvOperatorPodName)
	}

	// Create kubernetes client
	kubecli, err := k8sutil.NewKubeClient()
	if err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}

	pods := kubecli.CoreV1().Pods(namespace)
	recentErrors := 0
	for {
		p, err := pods.Get(name, metav1.GetOptions{})
		if k8sutil.IsNotFound(err) {
			cliLog.Warn().Msg("Pod not found")
			return
		} else if err != nil {
			recentErrors++
			cliLog.Error().Err(err).Msg("Failed to get pod")
			if recentErrors > 20 {
				cliLog.Fatal().Err(err).Msg("Too many recent errors")
				return
			}
		} else {
			// We got our pod
			finalizerCount := len(p.GetFinalizers())
			if finalizerCount == 0 {
				// No more finalizers, we're done
				cliLog.Info().Msg("All finalizers gone, we can stop now")
				return
			}
			cliLog.Info().Msgf("Waiting for %d more finalizers to be removed", finalizerCount)
		}
		// Wait a bit
		time.Sleep(time.Second)
	}
}

// Copy the executable to a given place.
func cmdLifecycleCopyRun(cmd *cobra.Command, args []string) {
	cliLog.Info().Msgf("Starting arangodb-operator, lifecycle copy, version %s build %s", projectVersion, projectBuild)

	exePath, err := os.Executable()
	if err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to get executable uuidPath")
	}

	// Open source
	rd, err := os.Open(exePath)
	if err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to open executable file")
	}
	defer rd.Close()

	// Open target
	targetPath := filepath.Join(lifecycleCopyOptions.TargetDir, filepath.Base(exePath))
	wr, err := os.Create(targetPath)
	if err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to create target file")
	}
	defer wr.Close()

	if _, err := io.Copy(wr, rd); err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to copy")
	}

	// Set file mode
	if err := os.Chmod(targetPath, 0755); err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to chmod")
	}

	cliLog.Info().Msgf("Executable copied to %s", targetPath)
}
