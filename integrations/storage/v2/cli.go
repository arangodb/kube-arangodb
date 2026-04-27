//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package v2

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/util/cli"
)

func NewCLI(prefix string) CLI {
	return cliImpl{
		prefix: prefix,

		storageType: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.type", prefix),
			Description: "Type of the Storage Integration (s3, gcs, azureBlobStorage)",
			Default:     string(ConfigurationTypeS3),
		},

		s3:    newS3CLI(fmt.Sprintf("%s.s3", prefix)),
		gcs:   newGCSCLI(fmt.Sprintf("%s.gcs", prefix)),
		azure: newAzureCLI(fmt.Sprintf("%s.azure-blob-storage", prefix)),
	}
}

type CLI interface {
	cli.FlagRegisterer

	Configuration(cmd *cobra.Command) (Configuration, error)
}

type cliImpl struct {
	prefix string

	storageType cli.Flag[string]

	s3    s3CLI
	gcs   gcsCLI
	azure azureCLI
}

func (c cliImpl) GetName() string {
	return c.prefix
}

func (c cliImpl) Register(cmd *cobra.Command) error {
	return cli.RegisterFlags(
		cmd,
		c.storageType,
		c.s3,
		c.gcs,
		c.azure,
	)
}

func (c cliImpl) Validate(cmd *cobra.Command) error {
	return cli.ValidateFlags(
		c.storageType,
	)(cmd, nil)
}

func (c cliImpl) Configuration(cmd *cobra.Command) (Configuration, error) {
	storageType, err := c.storageType.Get(cmd)
	if err != nil {
		return Configuration{}, err
	}

	s3Cfg, err := c.s3.Configuration(cmd)
	if err != nil {
		return Configuration{}, err
	}

	gcsCfg, err := c.gcs.Configuration(cmd)
	if err != nil {
		return Configuration{}, err
	}

	azureCfg, err := c.azure.Configuration(cmd)
	if err != nil {
		return Configuration{}, err
	}

	return Configuration{
		Type:             ConfigurationType(storageType),
		S3:               s3Cfg,
		GCS:              gcsCfg,
		AzureBlobStorage: azureCfg,
	}, nil
}
