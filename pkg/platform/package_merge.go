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
	"sigs.k8s.io/yaml"

	"github.com/arangodb/kube-arangodb/pkg/util/cli"
)

func packageMerge() (*cobra.Command, error) {
	var cmd cobra.Command

	cmd.Use = "merge ... packages"
	cmd.Short = "Merges definitions into single file"

	if err := cli.RegisterFlags(&cmd); err != nil {
		return nil, err
	}

	cmd.RunE = getRunner().With(packageMergeRun).Run

	return &cmd, nil
}

func packageMergeRun(cmd *cobra.Command, args []string) error {
	p, err := getHelmPackages(args...)
	if err != nil {
		return err
	}

	d, err := yaml.Marshal(p)
	if err != nil {
		return err
	}

	return render(cmd, "---\n\n%s", string(d))
}
