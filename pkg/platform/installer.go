//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package platform

import (
	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

func NewInstaller() (*cobra.Command, error) {
	return installer()
}

func installer() (*cobra.Command, error) {
	var cmd cobra.Command

	cmd.Use = "arangodb_operator_platform"

	cmd.SetContext(shutdown.Context())

	if err := cli.RegisterFlags(&cmd, flagNamespace, flagKubeconfig); err != nil {
		return nil, err
	}

	if err := withRegisterCommand(&cmd,
		pkg,
	); err != nil {
		return nil, err
	}

	return &cmd, nil
}

func withRegisterCommand(parent *cobra.Command, calls ...func() (*cobra.Command, error)) error {
	for _, call := range calls {
		if c, err := call(); err != nil {
			return err
		} else {
			parent.AddCommand(c)
		}
	}

	return nil
}
