//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
	goflag "flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/arangodb/k8s-operator/pkg/logging"
	"github.com/arangodb/k8s-operator/pkg/storage/provisioner"
	"github.com/arangodb/k8s-operator/pkg/util/constants"
	"github.com/arangodb/k8s-operator/pkg/util/k8sutil"
)

var (
	cmdStorage = &cobra.Command{
		Use: "storage",
		Run: cmdUsage,
	}

	cmdStorageProvisioner = &cobra.Command{
		Use: "provisioner",
		Run: cmdStorageProvisionerRun,
	}

	storageProvisioner struct {
		localPath        []string
		storageClassName string
	}
)

func init() {
	cmdMain.AddCommand(cmdStorage)
	cmdStorage.AddCommand(cmdStorageProvisioner)

	f := cmdStorageProvisioner.Flags()
	f.StringSliceVar(&storageProvisioner.localPath, "local-path", nil, "Local directory to provision volumes into")
	f.StringVar(&storageProvisioner.storageClassName, "storage-class-name", "", "StorageClassName set in provisioned volumes")
}

// Run the provisioner
func cmdStorageProvisionerRun(cmd *cobra.Command, args []string) {
	goflag.CommandLine.Parse([]string{"-logtostderr"})
	var err error
	logService, err = logging.NewService(logLevel)
	if err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to initialize log service")
	}

	// Log version
	cliLog.Info().Msgf("Starting arangodb local storage provisioner, version %s build %s", projectVersion, projectBuild)

	// Get environment
	namespace := os.Getenv(constants.EnvOperatorPodNamespace)
	if len(namespace) == 0 {
		cliLog.Fatal().Msgf("%s environment variable missing", constants.EnvOperatorPodNamespace)
	}
	name := os.Getenv(constants.EnvOperatorPodName)
	if len(name) == 0 {
		cliLog.Fatal().Msgf("%s environment variable missing", constants.EnvOperatorPodName)
	}
	nodeName := os.Getenv(constants.EnvOperatorNodeName)
	if len(name) == 0 {
		cliLog.Fatal().Msgf("%s environment variable missing", constants.EnvOperatorNodeName)
	}

	config, deps, err := newProvisionerConfigAndDeps(nodeName, namespace, name)
	if err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to create provisioner config & dependencies")
	}
	p, err := provisioner.New(config, deps)
	if err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to create provisioner")
	}

	ctx := context.TODO()
	p.Run(ctx)
}

// newProvisionerConfigAndDeps creates storage provisioner config & dependencies.
func newProvisionerConfigAndDeps(nodeName, namespace, name string) (provisioner.Config, provisioner.Dependencies, error) {
	kubecli, err := k8sutil.NewKubeClient()
	if err != nil {
		return provisioner.Config{}, provisioner.Dependencies{}, maskAny(err)
	}

	serviceAccount, err := getMyPodServiceAccount(kubecli, namespace, name)
	if err != nil {
		return provisioner.Config{}, provisioner.Dependencies{}, maskAny(fmt.Errorf("Failed to get my pod's service account: %s", err))
	}

	cfg := provisioner.Config{
		LocalPath:      storageProvisioner.localPath,
		NodeName:       nodeName,
		Namespace:      namespace,
		ServiceAccount: serviceAccount,
	}
	deps := provisioner.Dependencies{
		Log:     logService.MustGetLogger("provisioner"),
		KubeCli: kubecli,
	}

	return cfg, deps, nil
}
