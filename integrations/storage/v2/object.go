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

package v2

import (
	"context"
	"net/url"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	pbImplStorageV2Shared "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared"
	pbImplStorageV2SharedGCS "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared/gcs"
	pbImplStorageV2SharedS3 "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared/s3"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	awsHelper "github.com/arangodb/kube-arangodb/pkg/util/aws"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	gcsHelper "github.com/arangodb/kube-arangodb/pkg/util/gcs"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func NewIOFromObject(ctx context.Context, client kclient.Client, in *platformApi.ArangoPlatformStorage) (pbImplStorageV2Shared.IO, error) {
	if err := in.Spec.Validate(); err != nil {
		return nil, err
	}

	if backend := in.Spec.Backend; backend != nil {
		if s3Spec := backend.S3; s3Spec != nil {
			var config awsHelper.Config

			if v := s3Spec.CASecret; v != nil {
				secret, err := client.Kubernetes().CoreV1().Secrets(v.GetNamespace(in)).Get(ctx, v.GetName(), meta.GetOptions{})
				if err != nil {
					return nil, errors.WithMessage(err, "Failed to get S3 secret")
				}

				q, ok := secret.Data[utilConstants.SecretCACertificate]
				if !ok {
					return nil, errors.WithMessagef(err, "Failed to get S3 secret %s data: Key %s not found", secret.GetName(), utilConstants.SecretCACertificate)
				}

				config.TLS.CABytes = [][]byte{q}
			}

			if v := s3Spec.CredentialsSecret; v != nil {
				secret, err := client.Kubernetes().CoreV1().Secrets(v.GetNamespace(in)).Get(ctx, v.GetName(), meta.GetOptions{})
				if err != nil {
					return nil, errors.WithMessage(err, "Failed to get S3 secret")
				}

				sk, ok := secret.Data[utilConstants.SecretCredentialsSecretKey]
				if !ok {
					return nil, errors.Errorf("Failed to get S3 secret %s data: Key %s not found", secret.GetName(), utilConstants.SecretCredentialsSecretKey)
				}

				ak, ok := secret.Data[utilConstants.SecretCredentialsAccessKey]
				if !ok {
					return nil, errors.Errorf("Failed to get S3 secret %s data: Key %s not found", secret.GetName(), utilConstants.SecretCredentialsAccessKey)
				}

				config.Provider.Static.AccessKeyID = string(ak)
				config.Provider.Static.SecretAccessKey = string(sk)
				config.Provider.Type = awsHelper.ProviderTypeStatic
			}

			if v := s3Spec.AllowInsecure; v != nil {
				config.TLS.Insecure = *v
			}

			{
				if e := s3Spec.GetEndpoint(); e != "" {
					endpointURL, err := url.Parse(s3Spec.GetEndpoint())

					if err != nil {
						return nil, errors.WithMessagef(err, "Failed to parse url: %s", s3Spec.GetEndpoint())
					}
					disableSSL := endpointURL.Scheme == "http"

					config.DisableSSL = disableSSL
				}
			}

			config.Endpoint = s3Spec.GetEndpoint()
			config.Region = s3Spec.GetRegion()

			var cfg pbImplStorageV2SharedS3.Configuration

			cfg.BucketName = s3Spec.GetBucketName()
			cfg.BucketPrefix = s3Spec.GetBucketPrefix()
			cfg.Client = config

			return cfg.New()
		}

		if gcsSpec := backend.GCS; gcsSpec != nil {
			var config gcsHelper.Config

			if v := gcsSpec.CredentialsSecret; v != nil {
				secret, err := client.Kubernetes().CoreV1().Secrets(v.GetNamespace(in)).Get(ctx, v.GetName(), meta.GetOptions{})
				if err != nil {
					return nil, errors.WithMessage(err, "Failed to get GCS secret")
				}

				sk, ok := secret.Data[utilConstants.SecretCredentialsServiceAccount]
				if !ok {
					return nil, errors.Errorf("Failed to get GCS secret %s data: Key %s not found", secret.GetName(), utilConstants.SecretCredentialsServiceAccount)
				}

				config.Provider.ServiceAccount.JSON = string(sk)
				config.Provider.Type = gcsHelper.ProviderTypeServiceAccount
			}

			config.ProjectID = gcsSpec.GetProjectID()

			var cfg pbImplStorageV2SharedGCS.Configuration

			cfg.BucketName = gcsSpec.GetBucketName()
			cfg.BucketPrefix = gcsSpec.GetBucketPrefix()
			cfg.Client = config

			return cfg.New(ctx)
		}
	}

	return nil, errors.Errorf("Unable to init the storage")
}
