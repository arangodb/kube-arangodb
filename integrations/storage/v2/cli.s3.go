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

	pbImplStorageV2SharedS3 "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared/s3"
	awsHelper "github.com/arangodb/kube-arangodb/pkg/util/aws"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
)

func newS3CLI(prefix string) s3CLI {
	return s3CLI{
		prefix: prefix,

		endpoint: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.endpoint", prefix),
			Description: "Endpoint of S3 API implementation",
			Default:     "",
		},
		ca: cli.Flag[[]string]{
			Name:        fmt.Sprintf("%s.ca", prefix),
			Description: "Path to file containing CA certificate to validate endpoint connection",
			Default:     nil,
		},
		allowInsecure: cli.Flag[bool]{
			Name:        fmt.Sprintf("%s.allow-insecure", prefix),
			Description: "If set to true, the Endpoint certificates won't be checked",
			Default:     false,
		},
		disableSSL: cli.Flag[bool]{
			Name:        fmt.Sprintf("%s.disable-ssl", prefix),
			Description: "If set to true, the SSL won't be used when connecting to Endpoint",
			Default:     false,
		},
		region: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.region", prefix),
			Description: "S3 Region",
			Default:     "",
		},
		bucketName: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.bucket.name", prefix),
			Description: "S3 Bucket name",
			Default:     "",
		},
		bucketPrefix: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.bucket.prefix", prefix),
			Description: "S3 Bucket Prefix",
			Default:     "",
		},
		providerType: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.provider.type", prefix),
			Description: "S3 Credentials Provider type",
			Default:     string(awsHelper.ProviderTypeFile),
		},
		accessKeyFile: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.provider.file.access-key", prefix),
			Description: "Path to file containing S3 AccessKey",
			Default:     "",
		},
		secretKeyFile: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.provider.file.secret-key", prefix),
			Description: "Path to file containing S3 SecretKey",
			Default:     "",
		},
	}
}

type s3CLI struct {
	prefix string

	endpoint      cli.Flag[string]
	ca            cli.Flag[[]string]
	allowInsecure cli.Flag[bool]
	disableSSL    cli.Flag[bool]
	region        cli.Flag[string]
	bucketName    cli.Flag[string]
	bucketPrefix  cli.Flag[string]
	providerType  cli.Flag[string]
	accessKeyFile cli.Flag[string]
	secretKeyFile cli.Flag[string]
}

func (s s3CLI) GetName() string {
	return s.prefix
}

func (s s3CLI) Register(cmd *cobra.Command) error {
	return cli.RegisterFlags(
		cmd,
		s.endpoint,
		s.ca,
		s.allowInsecure,
		s.disableSSL,
		s.region,
		s.bucketName,
		s.bucketPrefix,
		s.providerType,
		s.accessKeyFile,
		s.secretKeyFile,
	)
}

func (s s3CLI) Validate(cmd *cobra.Command) error {
	return nil
}

func (s s3CLI) Configuration(cmd *cobra.Command) (pbImplStorageV2SharedS3.Configuration, error) {
	endpoint, err := s.endpoint.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedS3.Configuration{}, err
	}
	caFiles, err := s.ca.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedS3.Configuration{}, err
	}
	allowInsecure, err := s.allowInsecure.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedS3.Configuration{}, err
	}
	disableSSL, err := s.disableSSL.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedS3.Configuration{}, err
	}
	region, err := s.region.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedS3.Configuration{}, err
	}
	bucketName, err := s.bucketName.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedS3.Configuration{}, err
	}
	bucketPrefix, err := s.bucketPrefix.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedS3.Configuration{}, err
	}
	providerType, err := s.providerType.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedS3.Configuration{}, err
	}
	accessKeyFile, err := s.accessKeyFile.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedS3.Configuration{}, err
	}
	secretKeyFile, err := s.secretKeyFile.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedS3.Configuration{}, err
	}

	return pbImplStorageV2SharedS3.Configuration{
		BucketName:   bucketName,
		BucketPrefix: bucketPrefix,
		Client: awsHelper.Config{
			Endpoint:   endpoint,
			Region:     region,
			DisableSSL: disableSSL,
			Provider: awsHelper.Provider{
				Type: awsHelper.ProviderType(providerType),
				File: awsHelper.ProviderConfigFile{
					AccessKeyIDFile:     accessKeyFile,
					SecretAccessKeyFile: secretKeyFile,
				},
			},
			TLS: awsHelper.TLS{
				Insecure: allowInsecure,
				CAFiles:  caFiles,
			},
		},
	}, nil
}
