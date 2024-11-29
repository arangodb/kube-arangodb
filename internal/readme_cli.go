//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package internal

import (
	"path"

	"github.com/arangodb/kube-arangodb/cmd"
	"github.com/arangodb/kube-arangodb/cmd/integration"
	"github.com/arangodb/kube-arangodb/internal/md"
)

func GenerateCLIArangoDBOperatorReadme(root string) error {
	readmeSections := map[string]string{}

	if section, err := GenerateHelpQuoted(cmd.Command()); err != nil {
		return err
	} else {
		readmeSections["arangodb_operator_cmd"] = section
	}

	if err := md.ReplaceSectionsInFile(path.Join(root, "docs", "cli", "arangodb_operator.md"), readmeSections); err != nil {
		return err
	}

	return nil
}

func GenerateCLIArangoDBOperatorOpsReadme(root string) error {
	readmeSections := map[string]string{}

	if section, err := GenerateHelpQuoted(cmd.CommandOps()); err != nil {
		return err
	} else {
		readmeSections["arangodb_operator_ops_cmd"] = section
	}

	if section, err := GenerateHelpQuoted(cmd.CommandOps(), "crd"); err != nil {
		return err
	} else {
		readmeSections["arangodb_operator_ops_cmd_crd"] = section
	}

	if section, err := GenerateHelpQuoted(cmd.CommandOps(), "crd", "install"); err != nil {
		return err
	} else {
		readmeSections["arangodb_operator_ops_cmd_crd_install"] = section
	}

	if section, err := GenerateHelpQuoted(cmd.CommandOps(), "crd", "generate"); err != nil {
		return err
	} else {
		readmeSections["arangodb_operator_ops_cmd_crd_generate"] = section
	}

	if section, err := GenerateHelpQuoted(cmd.CommandOps(), "debug-package"); err != nil {
		return err
	} else {
		readmeSections["arangodb_operator_ops_cmd_debug_package"] = section
	}

	if err := md.ReplaceSectionsInFile(path.Join(root, "docs", "cli", "arangodb_operator_ops.md"), readmeSections); err != nil {
		return err
	}

	return nil
}

func GenerateCLIArangoDBOperatorIntegrationReadme(root string) error {
	readmeSections := map[string]string{}

	if section, err := GenerateHelpQuoted(integration.Command()); err != nil {
		return err
	} else {
		readmeSections["arangodb_operator_integration_cmd"] = section
	}

	if section, err := GenerateHelpQuoted(integration.Command(), "client"); err != nil {
		return err
	} else {
		readmeSections["arangodb_operator_integration_cmd_client"] = section
	}

	if err := md.ReplaceSectionsInFile(path.Join(root, "docs", "cli", "arangodb_operator_integration.md"), readmeSections); err != nil {
		return err
	}

	return nil
}
