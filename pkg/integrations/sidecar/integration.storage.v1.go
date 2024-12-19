//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	"net/url"
	"path/filepath"
	"strconv"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/ml/storage"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

type IntegrationStorageV1Community struct {
	Core    *Core
	Storage *platformApi.ArangoPlatformStorage
}

func (i IntegrationStorageV1Community) Name() []string {
	return []string{"STORAGE", "V1"}
}

func (i IntegrationStorageV1Community) Validate() error {
	if i.Storage == nil {
		return errors.Errorf("Storage is nil")
	}

	return nil
}

func (i IntegrationStorageV1Community) Envs() ([]core.EnvVar, error) {
	var envs = []core.EnvVar{
		{
			Name:  "INTEGRATION_STORAGE_V1",
			Value: "true",
		},
	}

	if s3 := i.Storage.Spec.GetBackend().GetS3(); s3 != nil {

		endpointURL, _ := url.Parse(s3.GetEndpoint())
		disableSSL := endpointURL.Scheme == "http"

		envs = append(envs,
			core.EnvVar{
				Name:  "INTEGRATION_STORAGE_V1_TYPE",
				Value: string(storage.S3),
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

	return i.Core.Envs(i, envs...), nil
}

func (i IntegrationStorageV1Community) Volumes() ([]core.Volume, []core.VolumeMount, error) {
	var volumeMounts []core.VolumeMount
	var volumes []core.Volume

	if s := i.Storage.Spec.GetBackend().GetS3(); s != nil {
		secretObj := s.GetCredentialsSecret()
		if secretObj.GetNamespace(i.Storage) != i.Storage.GetNamespace() {
			return nil, nil, errors.New("secrets from different namespace are not supported yet")
		}
		volumes = append(volumes, k8sutil.CreateVolumeWithSecret(mountNameStorageCredentials, secretObj.GetName()))
		volumeMounts = append(volumeMounts, core.VolumeMount{
			Name:      mountNameStorageCredentials,
			MountPath: mountPathStorageCredentials,
		})

		if caSecret := s.GetCASecret(); !caSecret.IsEmpty() {
			if caSecret.GetNamespace(i.Storage) != i.Storage.GetNamespace() {
				return nil, nil, errors.New("secrets from different namespace are not supported yet")
			}
			volumes = append(volumes, k8sutil.CreateVolumeWithSecret(mountNameStorageCA, caSecret.GetName()))
			volumeMounts = append(volumeMounts, core.VolumeMount{
				Name:      mountNameStorageCA,
				MountPath: mountPathStorageCA,
			})
		}
	}

	return volumes, volumeMounts, nil
}

func (i IntegrationStorageV1Community) GlobalEnvs() ([]core.EnvVar, error) {
	var envs []core.EnvVar
	envs = append(envs,
		core.EnvVar{
			Name:  "BUCKET_STORAGE_MODE",
			Value: "bucket",
		},
	)
	if s := i.Storage.Spec.GetBackend().GetS3(); s != nil {
		envs = append(envs,
			core.EnvVar{
				Name:  "BLOB_STORE_CONTAINER",
				Value: s.GetBucketName(),
			},
			core.EnvVar{
				Name:  "BLOB_STORE_PATH",
				Value: "",
			},
		)
	}

	return envs, nil
}
