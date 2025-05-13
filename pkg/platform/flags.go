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
	"os"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

var (
	flagNamespace = cli.Flag[string]{
		Name:        "namespace",
		Short:       "n",
		Description: "Kubernetes Namespace",
		Default:     "default",
		Persistent:  true,
	}

	flagPlatformName = cli.Flag[string]{
		Name:        "platform.name",
		Description: "Kubernetes Platform Name (name of the ArangoDeployment)",
		Default:     "",
		Persistent:  true,
	}

	flagOutput = cli.Flag[string]{
		Name:        "output",
		Short:       "o",
		Description: "Output format. Allowed table, json, yaml",
		Default:     "table",
		Persistent:  true,
		Check: func(in string) error {
			switch in {
			case "table", "json", "yaml":
				return nil
			}

			return errors.Errorf("Unknow output format: %s", in)
		},
	}

	flagPlatformEndpoint = cli.Flag[string]{
		Name:        "platform.endpoint",
		Description: "Platform Repository URL",
		Default:     "https://arangodb-platform-prd-chart-registry.s3.amazonaws.com",
		Persistent:  true,
		Hidden:      true,
	}

	flagUpgradeVersions = cli.Flag[bool]{
		Name:        "upgrade",
		Short:       "u",
		Description: "Enable upgrade procedure",
		Check: func(in bool) error {
			return nil
		},
	}

	flagAll = cli.Flag[bool]{
		Name:        "all",
		Short:       "a",
		Description: "Runs on all items",
		Check: func(in bool) error {
			return nil
		},
	}

	flagValues = cli.Flag[[]string]{
		Name:        "values",
		Short:       "f",
		Description: "Chart values",
		Check: func(in []string) error {
			for _, f := range in {
				data, err := os.ReadFile(f)
				if err != nil {
					return errors.Wrapf(err, "Unable to find file: %s", f)
				}

				_, err = util.JsonOrYamlUnmarshal[map[string]interface{}](data)
				if err != nil {
					return errors.Wrapf(err, "Unable to load file: %s", f)
				}
			}
			return nil
		},
	}
)
