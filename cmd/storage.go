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
	"net"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner"
	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner/service"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/version"
)

var (
	cmdStorage = &cobra.Command{
		Use: "storage",
		Run: executeUsage,
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
	var err error

	// Log version

	logger.Info("Starting arangodb local storage provisioner (%s), version %s build %s", version.GetVersionV1().Edition.Title(), version.GetVersionV1().Version, version.GetVersionV1().Build)

	// Get environment
	nodeName := os.Getenv(constants.EnvOperatorNodeName)
	if len(nodeName) == 0 {
		logger.Fatal("%s environment variable missing", constants.EnvOperatorNodeName)
	}

	config := newProvisionerConfigAndDeps(nodeName)
	p, err := service.New(config)
	if err != nil {
		logger.Err(err).Fatal("Failed to create provisioner")
	}

	ctx := context.TODO()
	p.Run(ctx)
}

// newProvisionerConfigAndDeps creates storage provisioner config & dependencies.
func newProvisionerConfigAndDeps(nodeName string) service.Config {
	cfg := service.Config{
		Address:  net.JoinHostPort("0.0.0.0", strconv.Itoa(storageProvisioner.port)),
		NodeName: nodeName,
	}

	return cfg
}
