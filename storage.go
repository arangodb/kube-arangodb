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
	"net"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner"
	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner/service"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
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
		port int
	}
)

func init() {
	cmdMain.AddCommand(cmdStorage)
	cmdStorage.AddCommand(cmdStorageProvisioner)

	f := cmdStorageProvisioner.Flags()
	f.IntVar(&storageProvisioner.port, "port", provisioner.DefaultPort, "Port to listen on")
}

// Run the provisioner
func cmdStorageProvisionerRun(cmd *cobra.Command, args []string) {
	//goflag.CommandLine.Parse([]string{"-logtostderr"})
	var err error
	logService, err = logging.NewService(logLevel)
	if err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to initialize log service")
	}

	// Log version
	cliLog.Info().Msgf("Starting arangodb local storage provisioner, version %s build %s", projectVersion, projectBuild)

	// Get environment
	nodeName := os.Getenv(constants.EnvOperatorNodeName)
	if len(nodeName) == 0 {
		cliLog.Fatal().Msgf("%s environment variable missing", constants.EnvOperatorNodeName)
	}

	config, deps, err := newProvisionerConfigAndDeps(nodeName)
	if err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to create provisioner config & dependencies")
	}
	p, err := service.New(config, deps)
	if err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to create provisioner")
	}

	ctx := context.TODO()
	p.Run(ctx)
}

// newProvisionerConfigAndDeps creates storage provisioner config & dependencies.
func newProvisionerConfigAndDeps(nodeName string) (service.Config, service.Dependencies, error) {
	cfg := service.Config{
		Address:  net.JoinHostPort("0.0.0.0", strconv.Itoa(storageProvisioner.port)),
		NodeName: nodeName,
	}
	deps := service.Dependencies{
		Log: logService.MustGetLogger("provisioner"),
	}

	return cfg, deps, nil
}
