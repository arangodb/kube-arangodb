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

	pbImplStorageV2 "github.com/arangodb/kube-arangodb/integrations/storage/v2"
	pbStorageV2 "github.com/arangodb/kube-arangodb/integrations/storage/v2/definition"
	awsHelper "github.com/arangodb/kube-arangodb/pkg/util/aws"
	"github.com/arangodb/kube-arangodb/pkg/util/azure"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/gcs"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func init() {
	registerer.Register(pbStorageV2.Name, func() Integration {
		return &storageV2{}
	})
}

type storageV2 struct {
	Configuration pbImplStorageV2.Configuration
}

func (b *storageV2) Name() string {
	return pbStorageV2.Name
}

func (b *storageV2) Description() string {
	return "StorageBucket V2 Integration"
}

func (b *storageV2) Register(cmd *cobra.Command, fs FlagEnvHandler) error {
	return errors.Errors(
		fs.StringVar((*string)(&b.Configuration.Type), "type", string(pbImplStorageV2.ConfigurationTypeS3), "Type of the Storage Integration"),

		fs.StringVar(&b.Configuration.S3.Client.Endpoint, "s3.endpoint", "", "Endpoint of S3 API implementation"),
		fs.StringSliceVar(&b.Configuration.S3.Client.TLS.CAFiles, "s3.ca", nil, "Path to file containing CA certificate to validate endpoint connection"),
		fs.BoolVar(&b.Configuration.S3.Client.TLS.Insecure, "s3.allow-insecure", false, "If set to true, the Endpoint certificates won't be checked"),
		fs.BoolVar(&b.Configuration.S3.Client.DisableSSL, "s3.disable-ssl", false, "If set to true, the SSL won't be used when connecting to Endpoint"),
		fs.StringVar(&b.Configuration.S3.Client.Region, "s3.region", "", "Region"),
		fs.StringVar(&b.Configuration.S3.BucketName, "s3.bucket.name", "", "Bucket name"),
		fs.StringVar(&b.Configuration.S3.BucketPrefix, "s3.bucket.prefix", "", "Bucket Prefix"),
		fs.StringVar((*string)(&b.Configuration.S3.Client.Provider.Type), "s3.provider.type", string(awsHelper.ProviderTypeFile), "S3 Credentials Provider type"),
		fs.StringVar(&b.Configuration.S3.Client.Provider.File.AccessKeyIDFile, "s3.provider.file.access-key", "", "Path to file containing S3 AccessKey"),
		fs.StringVar(&b.Configuration.S3.Client.Provider.File.SecretAccessKeyFile, "s3.provider.file.secret-key", "", "Path to file containing S3 SecretKey"),

		fs.StringVar(&b.Configuration.GCS.Client.ProjectID, "gcs.project-id", "", "GCP Project ID"),
		fs.StringVar(&b.Configuration.GCS.BucketName, "gcs.bucket.name", "", "Bucket name"),
		fs.StringVar(&b.Configuration.GCS.BucketPrefix, "gcs.bucket.prefix", "", "Bucket Prefix"),
		fs.StringVar((*string)(&b.Configuration.GCS.Client.Provider.Type), "gcs.provider.type", string(gcs.ProviderTypeServiceAccount), "Type of the provided credentials"),
		fs.StringVar(&b.Configuration.GCS.Client.Provider.ServiceAccount.File, "gcs.provider.sa.file", "", "Path to the file with ServiceAccount JSON"),
		fs.StringVar(&b.Configuration.GCS.Client.Provider.ServiceAccount.JSON, "gcs.provider.sa.json", "", "ServiceAccount JSON"),

		fs.StringVar(&b.Configuration.AzureBlobStorage.Client.Provider.TenantID, "azure-blob-storage.client.tenant-id", "", "Azure Client Tenant ID"),
		fs.StringVar(&b.Configuration.AzureBlobStorage.Client.AccountName, "azure-blob-storage.account-name", "", "AzureBlobStorage Account ID"),
		fs.StringVar(&b.Configuration.AzureBlobStorage.Client.Endpoint, "azure-blob-storage.endpoint", "", "AzureBlobStorage Endpoint"),
		fs.StringVar(&b.Configuration.AzureBlobStorage.BucketName, "azure-blob-storage.bucket.name", "", "Bucket name"),
		fs.StringVar(&b.Configuration.AzureBlobStorage.BucketPrefix, "azure-blob-storage.bucket.prefix", "", "Bucket Prefix"),
		fs.StringVar((*string)(&b.Configuration.AzureBlobStorage.Client.Provider.Type), "azure-blob-storage.client.type", string(azure.ProviderTypeSecret), "Azure Client Provider"),
		fs.StringVar(&b.Configuration.AzureBlobStorage.Client.Provider.Secret.ClientID, "azure-blob-storage.client.secret.client-id", "", "Azure ClientID"),
		fs.StringVar(&b.Configuration.AzureBlobStorage.Client.Provider.Secret.ClientIDFile, "azure-blob-storage.client.secret.client-id-file", "", "Azure ClientID File"),
		fs.StringVar(&b.Configuration.AzureBlobStorage.Client.Provider.Secret.ClientSecret, "azure-blob-storage.client.secret.client-secret", "", "Azure ClientSecret"),
		fs.StringVar(&b.Configuration.AzureBlobStorage.Client.Provider.Secret.ClientSecretFile, "azure-blob-storage.client.secret.client-secret-file", "", "Azure ClientSecret File"),
	)
}

func (b *storageV2) Handler(ctx context.Context, cmd *cobra.Command) (svc.Handler, error) {
	return pbImplStorageV2.New(b.Configuration)
}

func (*storageV2) Init(ctx context.Context, cmd *cobra.Command) error {
	return nil
}
