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

package sidecar

import (
	"net/url"
	"path/filepath"
	"strconv"

	core "k8s.io/api/core/v1"

	pbImplStorageV1 "github.com/arangodb/kube-arangodb/integrations/storage/v1"
	mlApi "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1beta1"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/util/aws"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

type IntegrationStorageV1 struct {
	Core            *Core
	MLStorage       *mlApi.ArangoMLStorage
	PlatformStorage *platformApi.ArangoPlatformStorage
}

func (i IntegrationStorageV1) Name() []string {
	return []string{"STORAGE", "V1"}
}

func (i IntegrationStorageV1) Validate() error {
	if i.MLStorage == nil && i.PlatformStorage == nil {
		return errors.Errorf("MLStorage and PlatformStorage are nil")
	}
	if i.MLStorage != nil && i.PlatformStorage != nil {
		return errors.Errorf("Only one of MLStorage and PlatformStorage can be set")
	}

	return nil
}

func (i IntegrationStorageV1) Envs() ([]core.EnvVar, error) {
	var envs = []core.EnvVar{
		{
			Name:  "INTEGRATION_STORAGE_V1",
			Value: "true",
		},
	}

	if storage := i.MLStorage; storage != nil {
		if s3 := storage.Spec.GetBackend().GetS3(); s3 != nil {

			endpointURL, _ := url.Parse(s3.GetEndpoint())
			disableSSL := endpointURL.Scheme == "http"

			envs = append(envs,
				core.EnvVar{
					Name:  "INTEGRATION_STORAGE_V1_TYPE",
					Value: string(pbImplStorageV1.ConfigurationTypeS3),
				},
				core.EnvVar{
					Name:  "INTEGRATION_STORAGE_V1_S3_ENDPOINT",
					Value: s3.GetEndpoint(),
				},
				core.EnvVar{
					Name:  "INTEGRATION_STORAGE_V1_S3_ALLOW_INSECURE",
					Value: strconv.FormatBool(s3.GetAllowInsecure()),
				},
				core.EnvVar{
					Name:  "INTEGRATION_STORAGE_V1_S3_DISABLE_SSL",
					Value: strconv.FormatBool(disableSSL),
				},
				core.EnvVar{
					Name:  "INTEGRATION_STORAGE_V1_S3_REGION",
					Value: s3.GetRegion(),
				},
				core.EnvVar{
					Name:  "INTEGRATION_STORAGE_V1_S3_BUCKET",
					Value: storage.Spec.GetBucketName(),
				},
				core.EnvVar{
					Name:  "INTEGRATION_STORAGE_V1_S3_PROVIDER_TYPE",
					Value: string(aws.ProviderTypeFile),
				},
				core.EnvVar{
					Name:  "INTEGRATION_STORAGE_V1_S3_SECRET_KEY",
					Value: filepath.Join(mountPathStorageCredentials, constants.SecretCredentialsSecretKey),
				},
				core.EnvVar{
					Name:  "INTEGRATION_STORAGE_V1_S3_ACCESS_KEY",
					Value: filepath.Join(mountPathStorageCredentials, constants.SecretCredentialsAccessKey),
				},
			)

			if !s3.GetCASecret().IsEmpty() {

				envs = append(envs,
					core.EnvVar{
						Name:  "INTEGRATION_STORAGE_V1_S3_CA_CRT",
						Value: filepath.Join(mountPathStorageCA, constants.SecretCACertificate),
					},
					core.EnvVar{
						Name:  "INTEGRATION_STORAGE_V1_S3_CA_KEY",
						Value: filepath.Join(mountPathStorageCA, constants.SecretCAKey),
					},
				)
			}
		}
	}

	if storage := i.PlatformStorage; storage != nil {
		if s3 := storage.Spec.GetBackend().GetS3(); s3 != nil {

			endpointURL, _ := url.Parse(s3.GetEndpoint())
			disableSSL := endpointURL.Scheme == "http"

			envs = append(envs,
				core.EnvVar{
					Name:  "INTEGRATION_STORAGE_V1_TYPE",
					Value: string(pbImplStorageV1.ConfigurationTypeS3),
				},
				core.EnvVar{
					Name:  "INTEGRATION_STORAGE_V1_S3_ENDPOINT",
					Value: s3.GetEndpoint(),
				},
				core.EnvVar{
					Name:  "INTEGRATION_STORAGE_V1_S3_ALLOW_INSECURE",
					Value: strconv.FormatBool(s3.GetAllowInsecure()),
				},
				core.EnvVar{
					Name:  "INTEGRATION_STORAGE_V1_S3_DISABLE_SSL",
					Value: strconv.FormatBool(disableSSL),
				},
				core.EnvVar{
					Name:  "INTEGRATION_STORAGE_V1_S3_REGION",
					Value: s3.GetRegion(),
				},
				core.EnvVar{
					Name:  "INTEGRATION_STORAGE_V1_S3_BUCKET",
					Value: s3.GetBucketName(),
				},
				core.EnvVar{
					Name:  "INTEGRATION_STORAGE_V1_S3_PROVIDER_TYPE",
					Value: string(aws.ProviderTypeFile),
				},
				core.EnvVar{
					Name:  "INTEGRATION_STORAGE_V1_S3_SECRET_KEY",
					Value: filepath.Join(mountPathStorageCredentials, constants.SecretCredentialsSecretKey),
				},
				core.EnvVar{
					Name:  "INTEGRATION_STORAGE_V1_S3_ACCESS_KEY",
					Value: filepath.Join(mountPathStorageCredentials, constants.SecretCredentialsAccessKey),
				},
			)

			if !s3.GetCASecret().IsEmpty() {

				envs = append(envs,
					core.EnvVar{
						Name:  "INTEGRATION_STORAGE_V1_S3_CA_CRT",
						Value: filepath.Join(mountPathStorageCA, constants.SecretCACertificate),
					},
					core.EnvVar{
						Name:  "INTEGRATION_STORAGE_V1_S3_CA_KEY",
						Value: filepath.Join(mountPathStorageCA, constants.SecretCAKey),
					},
				)
			}
		}
	}

	return i.Core.Envs(i, envs...), nil
}

func (i IntegrationStorageV1) Volumes() ([]core.Volume, []core.VolumeMount, error) {
	var volumeMounts []core.VolumeMount
	var volumes []core.Volume

	if storage := i.MLStorage; storage != nil {
		if s := storage.Spec.GetBackend().GetS3(); s != nil {
			secretObj := s.GetCredentialsSecret()
			if secretObj.GetNamespace(storage) != storage.GetNamespace() {
				return nil, nil, errors.New("secrets from different namespace are not supported yet")
			}
			volumes = append(volumes, k8sutil.CreateVolumeWithSecret(mountNameStorageCredentials, secretObj.GetName()))
			volumeMounts = append(volumeMounts, core.VolumeMount{
				Name:      mountNameStorageCredentials,
				MountPath: mountPathStorageCredentials,
			})

			if caSecret := s.GetCASecret(); !caSecret.IsEmpty() {
				if caSecret.GetNamespace(storage) != storage.GetNamespace() {
					return nil, nil, errors.New("secrets from different namespace are not supported yet")
				}
				volumes = append(volumes, k8sutil.CreateVolumeWithSecret(mountNameStorageCA, caSecret.GetName()))
				volumeMounts = append(volumeMounts, core.VolumeMount{
					Name:      mountNameStorageCA,
					MountPath: mountPathStorageCA,
				})
			}
		}
	}

	if storage := i.PlatformStorage; storage != nil {
		if s := storage.Spec.GetBackend().GetS3(); s != nil {
			secretObj := s.GetCredentialsSecret()
			if secretObj.GetNamespace(storage) != storage.GetNamespace() {
				return nil, nil, errors.New("secrets from different namespace are not supported yet")
			}
			volumes = append(volumes, k8sutil.CreateVolumeWithSecret(mountNameStorageCredentials, secretObj.GetName()))
			volumeMounts = append(volumeMounts, core.VolumeMount{
				Name:      mountNameStorageCredentials,
				MountPath: mountPathStorageCredentials,
			})

			if caSecret := s.GetCASecret(); !caSecret.IsEmpty() {
				if caSecret.GetNamespace(storage) != storage.GetNamespace() {
					return nil, nil, errors.New("secrets from different namespace are not supported yet")
				}
				volumes = append(volumes, k8sutil.CreateVolumeWithSecret(mountNameStorageCA, caSecret.GetName()))
				volumeMounts = append(volumeMounts, core.VolumeMount{
					Name:      mountNameStorageCA,
					MountPath: mountPathStorageCA,
				})
			}
		}
	}

	return volumes, volumeMounts, nil
}

func (i IntegrationStorageV1) GlobalEnvs() ([]core.EnvVar, error) {
	if storage := i.MLStorage; storage != nil {
		return []core.EnvVar{
			{
				Name:  "BUCKET_STORAGE_MODE",
				Value: "bucket",
			},
			{
				Name:  "BLOB_STORE_CONTAINER",
				Value: storage.Spec.GetBucketName(),
			},
			{
				Name:  "BLOB_STORE_PATH",
				Value: storage.Spec.GetBucketPath(),
			},
		}, nil
	}

	if storage := i.PlatformStorage; storage != nil {
		return []core.EnvVar{
			{
				Name:  "BUCKET_STORAGE_MODE",
				Value: "bucket",
			},
			{
				Name:  "BLOB_STORE_CONTAINER",
				Value: storage.Spec.GetBackend().GetS3().GetBucketName(),
			},
			{
				Name:  "BLOB_STORE_PATH",
				Value: storage.Spec.GetBackend().GetS3().GetBucketPrefix(),
			},
		}, nil
	}

	return nil, nil
}
