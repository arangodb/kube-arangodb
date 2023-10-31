//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/crd"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
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

var (
	crdInstallOptions struct {
		validationSchema []string
	}
)

var (
	defaultValidationSchemaEnabled []string
)

func init() {
	cmdMain.AddCommand(cmdCRD)
	cmdOps.AddCommand(cmdCRD)

	f := cmdCRDInstall.Flags()
	f.StringArrayVar(&crdInstallOptions.validationSchema, "crd.validation-schema", defaultValidationSchemaEnabled, "Controls which CRD should have validation schema <crd-name>=<true/false>.")
	cmdCRD.AddCommand(cmdCRDInstall)
}

func parseCRDValidationSchemaArgs(args []string) (map[string]bool, error) {
	availableCRDs := crd.ListAvailableNames()
	result := make(map[string]bool)
	for _, arg := range args {
		parts := strings.Split(arg, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("arg value '%s' can't be parsed", arg)
		}

		n := parts[0]
		if availableCRDs.IndexOf(n) < 0 {
			return nil, fmt.Errorf("unknown CRD %s", n)
		}

		enabled, err := strconv.ParseBool(parts[1])
		if err != nil {
			return nil, errors.Wrapf(err, "not a bool value: %s", parts[1])
		}
		result[n] = enabled
	}
	return result, nil
}

func cmdCRDInstallRun(cmd *cobra.Command, args []string) {
	crdValidation, err := parseCRDValidationSchemaArgs(crdInstallOptions.validationSchema)
	if err != nil {
		logger.Fatal("Invalid --crd.validation-schema args: %s", err)
		return
	}

	client, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		logger.Fatal("Failed to get client")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	err = crd.EnsureCRD(ctx, client, false, crdValidation)
	if err != nil {
		os.Exit(1)
	}
}
