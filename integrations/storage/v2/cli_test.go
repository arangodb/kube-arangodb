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
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	awsHelper "github.com/arangodb/kube-arangodb/pkg/util/aws"
	azureHelper "github.com/arangodb/kube-arangodb/pkg/util/azure"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	gcsHelper "github.com/arangodb/kube-arangodb/pkg/util/gcs"
)

func newCmdWithCLI(t *testing.T) (*cobra.Command, CLI) {
	c := NewCLI("storage.v2")
	cmd := &cobra.Command{Use: "test"}
	require.NoError(t, cli.RegisterFlags(cmd, c))
	return cmd, c
}

func TestCLI_Defaults(t *testing.T) {
	cmd, c := newCmdWithCLI(t)

	require.NoError(t, cmd.ParseFlags([]string{}))

	cfg, err := c.Configuration(cmd)
	require.NoError(t, err)

	require.Equal(t, ConfigurationTypeS3, cfg.Type)
	require.Equal(t, awsHelper.ProviderTypeFile, cfg.S3.Client.Provider.Type)
	require.Equal(t, gcsHelper.ProviderTypeServiceAccount, cfg.GCS.Client.Provider.Type)
	require.Equal(t, azureHelper.ProviderTypeSecret, cfg.AzureBlobStorage.Client.Provider.Type)

	require.Empty(t, cfg.S3.BucketName)
	require.Empty(t, cfg.GCS.BucketName)
	require.Empty(t, cfg.AzureBlobStorage.BucketName)
}

func TestCLI_S3_FromFlags(t *testing.T) {
	cmd, c := newCmdWithCLI(t)

	require.NoError(t, cmd.ParseFlags([]string{
		"--storage.v2.type=s3",
		"--storage.v2.s3.endpoint=https://s3.example.com",
		"--storage.v2.s3.region=eu-central-1",
		"--storage.v2.s3.bucket.name=my-bucket",
		"--storage.v2.s3.bucket.prefix=app/",
		"--storage.v2.s3.allow-insecure=true",
		"--storage.v2.s3.disable-ssl=true",
		"--storage.v2.s3.ca=/etc/ca.pem,/etc/ca2.pem",
		"--storage.v2.s3.provider.type=file",
		"--storage.v2.s3.provider.file.access-key=/var/run/secrets/access",
		"--storage.v2.s3.provider.file.secret-key=/var/run/secrets/secret",
	}))

	cfg, err := c.Configuration(cmd)
	require.NoError(t, err)

	require.Equal(t, ConfigurationTypeS3, cfg.Type)
	require.Equal(t, "my-bucket", cfg.S3.BucketName)
	require.Equal(t, "app/", cfg.S3.BucketPrefix)
	require.Equal(t, "https://s3.example.com", cfg.S3.Client.Endpoint)
	require.Equal(t, "eu-central-1", cfg.S3.Client.Region)
	require.True(t, cfg.S3.Client.DisableSSL)
	require.True(t, cfg.S3.Client.TLS.Insecure)
	require.Equal(t, []string{"/etc/ca.pem", "/etc/ca2.pem"}, cfg.S3.Client.TLS.CAFiles)
	require.Equal(t, awsHelper.ProviderTypeFile, cfg.S3.Client.Provider.Type)
	require.Equal(t, "/var/run/secrets/access", cfg.S3.Client.Provider.File.AccessKeyIDFile)
	require.Equal(t, "/var/run/secrets/secret", cfg.S3.Client.Provider.File.SecretAccessKeyFile)
}

func TestCLI_GCS_FromFlags(t *testing.T) {
	cmd, c := newCmdWithCLI(t)

	require.NoError(t, cmd.ParseFlags([]string{
		"--storage.v2.type=gcs",
		"--storage.v2.gcs.project-id=my-project",
		"--storage.v2.gcs.bucket.name=gcs-bucket",
		"--storage.v2.gcs.bucket.prefix=data/",
		"--storage.v2.gcs.provider.type=serviceAccount",
		"--storage.v2.gcs.provider.sa.file=/etc/sa.json",
		"--storage.v2.gcs.provider.sa.json={}",
	}))

	cfg, err := c.Configuration(cmd)
	require.NoError(t, err)

	require.Equal(t, ConfigurationTypeGCS, cfg.Type)
	require.Equal(t, "my-project", cfg.GCS.Client.ProjectID)
	require.Equal(t, "gcs-bucket", cfg.GCS.BucketName)
	require.Equal(t, "data/", cfg.GCS.BucketPrefix)
	require.Equal(t, gcsHelper.ProviderTypeServiceAccount, cfg.GCS.Client.Provider.Type)
	require.Equal(t, "/etc/sa.json", cfg.GCS.Client.Provider.ServiceAccount.File)
	require.Equal(t, "{}", cfg.GCS.Client.Provider.ServiceAccount.JSON)
}

func TestCLI_Azure_FromFlags(t *testing.T) {
	cmd, c := newCmdWithCLI(t)

	require.NoError(t, cmd.ParseFlags([]string{
		"--storage.v2.type=azureBlobStorage",
		"--storage.v2.azure-blob-storage.client.tenant-id=tenant",
		"--storage.v2.azure-blob-storage.account-name=account",
		"--storage.v2.azure-blob-storage.endpoint=https://account.blob.core.windows.net",
		"--storage.v2.azure-blob-storage.bucket.name=blob",
		"--storage.v2.azure-blob-storage.bucket.prefix=p/",
		"--storage.v2.azure-blob-storage.client.type=secret",
		"--storage.v2.azure-blob-storage.client.secret.client-id=cid",
		"--storage.v2.azure-blob-storage.client.secret.client-id-file=/etc/cid",
		"--storage.v2.azure-blob-storage.client.secret.client-secret=csec",
		"--storage.v2.azure-blob-storage.client.secret.client-secret-file=/etc/csec",
	}))

	cfg, err := c.Configuration(cmd)
	require.NoError(t, err)

	require.Equal(t, ConfigurationTypeAzure, cfg.Type)
	require.Equal(t, "tenant", cfg.AzureBlobStorage.Client.Provider.TenantID)
	require.Equal(t, "account", cfg.AzureBlobStorage.Client.AccountName)
	require.Equal(t, "https://account.blob.core.windows.net", cfg.AzureBlobStorage.Client.Endpoint)
	require.Equal(t, "blob", cfg.AzureBlobStorage.BucketName)
	require.Equal(t, "p/", cfg.AzureBlobStorage.BucketPrefix)
	require.Equal(t, azureHelper.ProviderTypeSecret, cfg.AzureBlobStorage.Client.Provider.Type)
	require.Equal(t, "cid", cfg.AzureBlobStorage.Client.Provider.Secret.ClientID)
	require.Equal(t, "/etc/cid", cfg.AzureBlobStorage.Client.Provider.Secret.ClientIDFile)
	require.Equal(t, "csec", cfg.AzureBlobStorage.Client.Provider.Secret.ClientSecret)
	require.Equal(t, "/etc/csec", cfg.AzureBlobStorage.Client.Provider.Secret.ClientSecretFile)
}

func TestCLI_PrefixIsHonored(t *testing.T) {
	c := NewCLI("custom.prefix")
	cmd := &cobra.Command{Use: "test"}
	require.NoError(t, cli.RegisterFlags(cmd, c))

	require.NoError(t, cmd.ParseFlags([]string{
		"--custom.prefix.s3.bucket.name=alt-bucket",
	}))

	cfg, err := c.Configuration(cmd)
	require.NoError(t, err)
	require.Equal(t, "alt-bucket", cfg.S3.BucketName)
}
