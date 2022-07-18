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
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/crd"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

var (
	cmdCRD = &cobra.Command{
		Use:   "crd",
		Run:   executeUsage,
		Short: "CRD operations",
	}
	cmdCRDInstall = &cobra.Command{
		Use:   "install",
		Run:   cmdCRDInstallRun,
		Short: "Install and update all required CRDs",
	}
)

func init() {
	cmdMain.AddCommand(cmdCRD)
	cmdOps.AddCommand(cmdCRD)

	cmdCRD.AddCommand(cmdCRDInstall)
}

func cmdCRDInstallRun(cmd *cobra.Command, args []string) {
	client, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		logger.Fatal("Failed to get client")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	err := crd.EnsureCRD(ctx, client, false)
	if err != nil {
		os.Exit(1)
	}
}
