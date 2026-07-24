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

	pbImplStorageV2SharedAzureBlobStorage "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared/abs"
	azureHelper "github.com/arangodb/kube-arangodb/pkg/util/azure"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
)

func newAzureCLI(prefix string) azureCLI {
	return azureCLI{
		prefix: prefix,

		tenantID: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.client.tenant-id", prefix),
			Description: "Azure Client Tenant ID",
			Default:     "",
		},
		accountName: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.account-name", prefix),
			Description: "AzureBlobStorage Account Name",
			Default:     "",
		},
		endpoint: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.endpoint", prefix),
			Description: "AzureBlobStorage Endpoint",
			Default:     "",
		},
		bucketName: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.bucket.name", prefix),
			Description: "AzureBlobStorage Bucket name",
			Default:     "",
		},
		bucketPrefix: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.bucket.prefix", prefix),
			Description: "AzureBlobStorage Bucket Prefix",
			Default:     "",
		},
		clientType: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.client.type", prefix),
			Description: "Azure Client Provider type",
			Default:     string(azureHelper.ProviderTypeSecret),
		},
		clientID: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.client.secret.client-id", prefix),
			Description: "Azure ClientID",
			Default:     "",
		},
		clientIDFile: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.client.secret.client-id-file", prefix),
			Description: "Azure ClientID File",
			Default:     "",
		},
		clientSecret: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.client.secret.client-secret", prefix),
			Description: "Azure ClientSecret",
			Default:     "",
		},
		clientSecretFile: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.client.secret.client-secret-file", prefix),
			Description: "Azure ClientSecret File",
			Default:     "",
		},
		certClientID: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.client.certificate.client-id", prefix),
			Description: "Azure Certificate ClientID",
			Default:     "",
		},
		certClientIDFile: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.client.certificate.client-id-file", prefix),
			Description: "Azure Certificate ClientID File",
			Default:     "",
		},
		certificate: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.client.certificate.certificate", prefix),
			Description: "Azure Client Certificate (PEM or PKCS#12 bundle)",
			Default:     "",
		},
		certificateFile: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.client.certificate.certificate-file", prefix),
			Description: "Azure Client Certificate File (PEM or PKCS#12 bundle)",
			Default:     "",
		},
		certificateKey: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.client.certificate.key", prefix),
			Description: "Azure Client Certificate private key (PEM), when supplied separately from the certificate",
			Default:     "",
		},
		certificateKeyFile: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.client.certificate.key-file", prefix),
			Description: "Azure Client Certificate private key File (PEM), when supplied separately from the certificate",
			Default:     "",
		},
		certificatePassword: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.client.certificate.password", prefix),
			Description: "Azure Client Certificate Password",
			Default:     "",
		},
		certificatePasswordFile: cli.Flag[string]{
			Name:        fmt.Sprintf("%s.client.certificate.password-file", prefix),
			Description: "Azure Client Certificate Password File",
			Default:     "",
		},
	}
}

type azureCLI struct {
	prefix string

	tenantID         cli.Flag[string]
	accountName      cli.Flag[string]
	endpoint         cli.Flag[string]
	bucketName       cli.Flag[string]
	bucketPrefix     cli.Flag[string]
	clientType       cli.Flag[string]
	clientID         cli.Flag[string]
	clientIDFile     cli.Flag[string]
	clientSecret     cli.Flag[string]
	clientSecretFile cli.Flag[string]

	certClientID            cli.Flag[string]
	certClientIDFile        cli.Flag[string]
	certificate             cli.Flag[string]
	certificateFile         cli.Flag[string]
	certificateKey          cli.Flag[string]
	certificateKeyFile      cli.Flag[string]
	certificatePassword     cli.Flag[string]
	certificatePasswordFile cli.Flag[string]
}

func (a azureCLI) GetName() string {
	return a.prefix
}

func (a azureCLI) Register(cmd *cobra.Command) error {
	return cli.RegisterFlags(
		cmd,
		a.tenantID,
		a.accountName,
		a.endpoint,
		a.bucketName,
		a.bucketPrefix,
		a.clientType,
		a.clientID,
		a.clientIDFile,
		a.clientSecret,
		a.clientSecretFile,
		a.certClientID,
		a.certClientIDFile,
		a.certificate,
		a.certificateFile,
		a.certificateKey,
		a.certificateKeyFile,
		a.certificatePassword,
		a.certificatePasswordFile,
	)
}

func (a azureCLI) Validate(cmd *cobra.Command) error {
	return nil
}

func (a azureCLI) Configuration(cmd *cobra.Command) (pbImplStorageV2SharedAzureBlobStorage.Configuration, error) {
	tenantID, err := a.tenantID.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedAzureBlobStorage.Configuration{}, err
	}
	accountName, err := a.accountName.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedAzureBlobStorage.Configuration{}, err
	}
	endpoint, err := a.endpoint.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedAzureBlobStorage.Configuration{}, err
	}
	bucketName, err := a.bucketName.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedAzureBlobStorage.Configuration{}, err
	}
	bucketPrefix, err := a.bucketPrefix.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedAzureBlobStorage.Configuration{}, err
	}
	clientType, err := a.clientType.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedAzureBlobStorage.Configuration{}, err
	}
	clientID, err := a.clientID.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedAzureBlobStorage.Configuration{}, err
	}
	clientIDFile, err := a.clientIDFile.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedAzureBlobStorage.Configuration{}, err
	}
	clientSecret, err := a.clientSecret.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedAzureBlobStorage.Configuration{}, err
	}
	clientSecretFile, err := a.clientSecretFile.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedAzureBlobStorage.Configuration{}, err
	}
	certClientID, err := a.certClientID.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedAzureBlobStorage.Configuration{}, err
	}
	certClientIDFile, err := a.certClientIDFile.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedAzureBlobStorage.Configuration{}, err
	}
	certificate, err := a.certificate.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedAzureBlobStorage.Configuration{}, err
	}
	certificateFile, err := a.certificateFile.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedAzureBlobStorage.Configuration{}, err
	}
	certificateKey, err := a.certificateKey.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedAzureBlobStorage.Configuration{}, err
	}
	certificateKeyFile, err := a.certificateKeyFile.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedAzureBlobStorage.Configuration{}, err
	}
	certificatePassword, err := a.certificatePassword.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedAzureBlobStorage.Configuration{}, err
	}
	certificatePasswordFile, err := a.certificatePasswordFile.Get(cmd)
	if err != nil {
		return pbImplStorageV2SharedAzureBlobStorage.Configuration{}, err
	}

	return pbImplStorageV2SharedAzureBlobStorage.Configuration{
		BucketName:   bucketName,
		BucketPrefix: bucketPrefix,
		Client: azureHelper.Config{
			AccountName: accountName,
			Endpoint:    endpoint,
			Provider: azureHelper.Provider{
				Type:     azureHelper.ProviderType(clientType),
				TenantID: tenantID,
				Secret: azureHelper.ProviderSecret{
					ClientID:         clientID,
					ClientIDFile:     clientIDFile,
					ClientSecret:     clientSecret,
					ClientSecretFile: clientSecretFile,
				},
				Certificate: azureHelper.ProviderCertificate{
					ClientID:        certClientID,
					ClientIDFile:    certClientIDFile,
					Certificate:     certificate,
					CertificateFile: certificateFile,
					Key:             certificateKey,
					KeyFile:         certificateKeyFile,
					Password:        certificatePassword,
					PasswordFile:    certificatePasswordFile,
				},
			},
		},
	}, nil
}
