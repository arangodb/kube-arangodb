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
	"time"

	"github.com/google/uuid"

	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

var (
	flagKubeconfig = cli.Flag[string]{
		Name:        "kubeconfig",
		Description: "Kubernetes Config File",
		Default:     "",
		Persistent:  true,
	}

	flagNamespace = cli.Flag[string]{
		Name:        "namespace",
		Short:       "n",
		Description: "Kubernetes Namespace",
		Default:     "default",
		Persistent:  true,
	}

	flagSecret = cli.Flag[string]{
		Name:        "secret",
		Description: "Kubernetes Secret Name",
		Default:     "",
		Check: func(in string) error {
			if in == "" {
				return nil
			}

			if err := sharedApi.IsValidName(in); err != nil {
				return errors.Errorf("Invalid deployment name: %s", err.Error())
			}

			return nil
		},
		Persistent: true,
	}

	flagPlatformName = cli.Flag[string]{
		Name:        "platform.name",
		Description: "Kubernetes Platform Name (name of the ArangoDeployment)",
		Default:     "",
		Persistent:  true,
		Check: func(in string) error {
			if err := sharedApi.IsValidName(in); err != nil {
				return errors.Errorf("Invalid deployment name: %s", err.Error())
			}

			return nil
		},
	}

	flagLicenseManagerEndpoint = cli.Flag[string]{
		Name:        "license.endpoint",
		Description: "LicenseManager Endpoint",
		Default:     "license.arangodb.com",
		Persistent:  false,
		Check: func(in string) error {
			return nil
		},
	}

	flagDeploymentID = cli.Flag[string]{
		Name:        "deployment.id",
		Description: "Deployment ID",
		Default:     "",
		Persistent:  false,
		Check: func(in string) error {
			return nil
		},
	}

	flagLicenseManagerClientID = cli.Flag[string]{
		Name:        "license.client.id",
		Description: "LicenseManager Client ID",
		Default:     "",
		Persistent:  false,
		Check: func(in string) error {
			if in == "" {
				return errors.New("Platform Client ID is required")
			}

			return nil
		},
	}

	flagLicenseManagerStages = cli.Flag[[]string]{
		Name:        "license.client.stage",
		Description: "LicenseManager Stages",
		Default:     []string{"prd"},
		Persistent:  false,
		Check: func(in []string) error {
			if len(in) == 0 {
				return errors.New("At least one stage needs to be defined")
			}

			return nil
		},
	}

	flagLicenseManagerClientSecret = cli.Flag[string]{
		Name:        "license.client.secret",
		Description: "LicenseManager Client Secret",
		Default:     "",
		Persistent:  false,
		Check: func(in string) error {
			if _, err := uuid.Parse(in); err != nil {
				return err
			}

			return nil
		},
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

	flagDeployment = cli.NewDeployment("arango")

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

	flagRegistryUseCredentials = cli.Flag[bool]{
		Name:        "registry.docker.credentials",
		Description: "Use Docker Credentials",
		Default:     false,
		Check: func(in bool) error {
			return nil
		},
	}

	flagRegistryInsecure = cli.Flag[[]string]{
		Name:        "registry.docker.insecure",
		Description: "List of insecure registries",
		Default:     nil,
	}

	flagRegistryList = cli.Flag[[]string]{
		Name:        "registry.docker.endpoint",
		Description: "List of boosted registries",
		Default:     nil,
	}

	flagActivateInterval = cli.Flag[time.Duration]{
		Name:        "license.interval",
		Description: "Interval of the license synchronization",
		Default:     0,
		Persistent:  false,
		Check: func(in time.Duration) error {
			if in < 0 {
				return errors.New("License Generation Interval cannot be negative")
			}
			return nil
		},
	}
)
