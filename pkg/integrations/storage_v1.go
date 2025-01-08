//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package integrations

import (
	"context"

	"github.com/spf13/cobra"

	pbImplStorageV1 "github.com/arangodb/kube-arangodb/integrations/storage/v1"
	pbStorageV1 "github.com/arangodb/kube-arangodb/integrations/storage/v1/definition"
	awsHelper "github.com/arangodb/kube-arangodb/pkg/util/aws"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func init() {
	registerer.Register(pbStorageV1.Name, func() Integration {
		return &storageV1{}
	})
}

type storageV1 struct {
	Configuration pbImplStorageV1.Configuration
}

func (b *storageV1) Name() string {
	return pbStorageV1.Name
}

func (b *storageV1) Description() string {
	return "StorageBucket Integration"
}

func (b *storageV1) Register(cmd *cobra.Command, fs FlagEnvHandler) error {
	return errors.Errors(
		fs.StringVar((*string)(&b.Configuration.Type), "type", string(pbImplStorageV1.ConfigurationTypeS3), "Type of the Storage Integration"),
		fs.StringVar(&b.Configuration.S3.Client.Endpoint, "s3.endpoint", "", "Endpoint of S3 API implementation"),
		fs.StringSliceVar(&b.Configuration.S3.Client.TLS.CAFiles, "s3.ca-crt", nil, "Path to file containing CA certificate to validate endpoint connection"),
		fs.BoolVar(&b.Configuration.S3.Client.TLS.Insecure, "s3.allow-insecure", false, "If set to true, the Endpoint certificates won't be checked"),
		fs.BoolVar(&b.Configuration.S3.Client.DisableSSL, "s3.disable-ssl", false, "If set to true, the SSL won't be used when connecting to Endpoint"),
		fs.StringVar(&b.Configuration.S3.Client.Region, "s3.region", "", "Region"),
		fs.StringVar(&b.Configuration.S3.BucketName, "s3.bucket", "", "Bucket name"),
		fs.StringVar((*string)(&b.Configuration.S3.Client.Provider.Type), "s3.provider.type", string(awsHelper.ProviderTypeFile), "S3 Credentials Provider type"),
		fs.StringVar(&b.Configuration.S3.Client.Provider.File.AccessKeyIDFile, "s3.access-key", "", "Path to file containing S3 AccessKey"),
		fs.StringVar(&b.Configuration.S3.Client.Provider.File.SecretAccessKeyFile, "s3.secret-key", "", "Path to file containing S3 SecretKey"),
	)
}

func (b *storageV1) Handler(ctx context.Context, cmd *cobra.Command) (svc.Handler, error) {
	return pbImplStorageV1.New(b.Configuration)
}

func (*storageV1) Init(ctx context.Context, cmd *cobra.Command) error {
	return nil
}

func (*storageV1) Visible() bool {
	return false
}
