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

	pbImplStorageV2SharedGCS "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared/gcs"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	gcsHelper "github.com/arangodb/kube-arangodb/pkg/util/gcs"
)

func newGCSCLI(prefix string) gcsCLI {
	return gcsCLI{
		prefix: prefix,

		projectID: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.project-id", prefix),
			Description: "GCP Project ID",
			Default:     "",
		},
		bucketName: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.bucket.name", prefix),
			Description: "GCS Bucket name",
			Default:     "",
		},
		bucketPrefix: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.bucket.prefix", prefix),
			Description: "GCS Bucket Prefix",
			Default:     "",
		},
		providerType: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.provider.type", prefix),
			Description: "Type of the provided GCS credentials",
			Default:     string(gcsHelper.ProviderTypeServiceAccount),
		},
		saFile: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.provider.sa.file", prefix),
			Description: "Path to the file with GCP ServiceAccount JSON",
			Default:     "",
		},
		saJSON: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.provider.sa.json", prefix),
			Description: "GCP ServiceAccount JSON",
			Default:     "",
		},
	}
}

type gcsCLI struct {
	prefix string

	projectID    cli.Flag[string]
	bucketName   cli.Flag[string]
	bucketPrefix cli.Flag[string]
	providerType cli.Flag[string]
	saFile       cli.Flag[string]
	saJSON       cli.Flag[string]
}

func (g gcsCLI) GetName() string {
	return g.prefix
}

func (g gcsCLI) Register(cmd *cobra.Command) error {
	return cli.RegisterFlags(
		cmd,
		g.projectID,
		g.bucketName,
		g.bucketPrefix,
		g.providerType,
		g.saFile,
		g.saJSON,
	)
}

func (g gcsCLI) Validate(cmd *cobra.Command) error {
	return nil
}

func (g gcsCLI) Configuration(cmd *cobra.Command) (pbImplStorageV2SharedGCS.Configuration, error) {
	projectID, err := g.projectID.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedGCS.Configuration{}, err
	}
	bucketName, err := g.bucketName.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedGCS.Configuration{}, err
	}
	bucketPrefix, err := g.bucketPrefix.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedGCS.Configuration{}, err
	}
	providerType, err := g.providerType.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedGCS.Configuration{}, err
	}
	saFile, err := g.saFile.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedGCS.Configuration{}, err
	}
	saJSON, err := g.saJSON.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedGCS.Configuration{}, err
	}

	return pbImplStorageV2SharedGCS.Configuration{
		BucketName:   bucketName,
		BucketPrefix: bucketPrefix,
		Client: gcsHelper.Config{
			ProjectID: projectID,
			Provider: gcsHelper.Provider{
				Type: gcsHelper.ProviderType(providerType),
				ServiceAccount: gcsHelper.ProviderServiceAccount{
					File: saFile,
					JSON: saJSON,
				},
			},
		},
	}, nil
}
