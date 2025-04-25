//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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
	goStrings "strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/crd"
	"github.com/arangodb/kube-arangodb/pkg/crd/crds"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

const (
	AllSchemasValue = "all"
)

var (
	cmdCRD = &cobra.Command{
		Use:   "crd",
		RunE:  cli.Usage,
		Short: "CRD operations",
	}
	cmdCRDInstall = &cobra.Command{
		Use:   "install",
		Run:   cmdCRDInstallRun,
		Short: "Install and update all required CRDs",
	}
	cmdCRDGenerate = &cobra.Command{
		Use:   "generate",
		Run:   cmdCRDGenerateRun,
		Short: "Generates YAML of all required CRDs",
	}
)

var (
	crdInstallOptions struct {
		validationSchema      []string
		preserveUnknownFields []string
		skip                  []string
		force                 bool
	}
)

var (
	defaultValidationSchemaEnabled []string
)

func init() {
	cmdMain.AddCommand(cmdCRD)
	cmdOps.AddCommand(cmdCRD)

	f := cmdCRD.PersistentFlags()
	f.StringArrayVar(&crdInstallOptions.validationSchema, "crd.validation-schema", defaultValidationSchemaEnabled, "Controls which CRD should have validation schema <crd-name>=<true/false>.")
	f.StringArrayVar(&crdInstallOptions.preserveUnknownFields, "crd.preserve-unknown-fields", nil, "Controls which CRD should have enabled preserve unknown fields in validation schema <crd-name>=<true/false>.")
	f.StringArrayVar(&crdInstallOptions.skip, "crd.skip", nil, "Controls which CRD should be skipped.")
	f.BoolVar(&crdInstallOptions.force, "crd.force-update", false, "Enforce CRD Schema update")

	cmdCRD.AddCommand(cmdCRDInstall)
	cmdCRD.AddCommand(cmdCRDGenerate)
}

func prepareCRDOptions(schemaEnabledArgs []string, preserveUnknownFieldsArgs []string) (map[string]crds.CRDOptions, error) {
	defaultOptions := crd.GetDefaultCRDOptions()
	var err error

	schemaEnabled := map[string]bool{}
	preserveUnknownFields := map[string]bool{}

	for _, arg := range schemaEnabledArgs {
		parts := goStrings.SplitN(arg, "=", 2)

		var enabled bool

		if len(parts) == 2 {
			enabled, err = strconv.ParseBool(parts[1])
			if err != nil {
				return nil, errors.Wrapf(err, "not a bool value: %s", parts[1])
			}

		}

		schemaEnabled[parts[0]] = enabled
	}

	for _, arg := range preserveUnknownFieldsArgs {
		parts := goStrings.SplitN(arg, "=", 2)

		var enabled bool

		if len(parts) == 2 {
			enabled, err = strconv.ParseBool(parts[1])
			if err != nil {
				return nil, errors.Wrapf(err, "not a bool value: %s", parts[1])
			}

		}

		preserveUnknownFields[parts[0]] = enabled
	}

	for k := range schemaEnabled {
		if k == AllSchemasValue {
			continue
		}
		if _, ok := defaultOptions[k]; !ok {
			return nil, fmt.Errorf("unknown CRD %s", k)
		}
	}

	for k := range preserveUnknownFields {
		if k == AllSchemasValue {
			continue
		}
		if _, ok := defaultOptions[k]; !ok {
			return nil, fmt.Errorf("unknown CRD %s", k)
		}
	}

	// Override the defaults
	if v, ok := schemaEnabled[AllSchemasValue]; ok {
		delete(preserveUnknownFields, AllSchemasValue)
		for k := range defaultOptions {
			z := defaultOptions[k]
			z.WithSchema = v
			defaultOptions[k] = z
		}
	}
	if v, ok := preserveUnknownFields[AllSchemasValue]; ok {
		delete(preserveUnknownFields, AllSchemasValue)
		for k := range defaultOptions {
			z := defaultOptions[k]
			z.WithPreserve = v
			defaultOptions[k] = z
		}
	}

	// Set explicit words
	for k, v := range schemaEnabled {
		z := defaultOptions[k]
		z.WithSchema = v
		defaultOptions[k] = z
	}
	for k, v := range preserveUnknownFields {
		z := defaultOptions[k]
		z.WithPreserve = v
		defaultOptions[k] = z
	}

	return defaultOptions, nil
}

func cmdCRDInstallRun(cmd *cobra.Command, args []string) {
	crdOpts, err := prepareCRDOptions(crdInstallOptions.validationSchema, crdInstallOptions.preserveUnknownFields)
	if err != nil {
		logger.Fatal("Invalid --crd.validation-schema args: %s", err)
		return
	}

	client, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		logger.Fatal("Failed to get client")
		return
	}

	ctx, cancel := context.WithTimeout(shutdown.Context(), time.Minute)
	defer cancel()

	err = crd.EnsureCRDWithOptions(ctx, client, crd.EnsureCRDOptions{IgnoreErrors: false, CRDOptions: crdOpts, ForceUpdate: crdInstallOptions.force, Skip: crdInstallOptions.skip})
	if err != nil {
		os.Exit(1)
	}
}

func cmdCRDGenerateRun(cmd *cobra.Command, args []string) {
	crdOpts, err := prepareCRDOptions(crdInstallOptions.validationSchema, crdInstallOptions.preserveUnknownFields)
	if err != nil {
		logger.Fatal("Invalid --crd.validation-schema args: %s", err)
		return
	}

	err = crd.GenerateCRDYAMLWithOptions(crd.EnsureCRDOptions{IgnoreErrors: false, CRDOptions: crdOpts, ForceUpdate: crdInstallOptions.force, Skip: crdInstallOptions.skip}, cmd.OutOrStdout())
	if err != nil {
		os.Exit(1)
	}
}
