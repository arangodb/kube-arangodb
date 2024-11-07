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
	"net/url"
	"path/filepath"
	"strconv"

	core "k8s.io/api/core/v1"

	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/ml/storage"
	"github.com/arangodb/kube-arangodb/pkg/util/aws"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	mountNameStorageCredentials = "integration-credentials"
	mountNameStorageCA          = "integration-ca"

	mountPathStorageCredentials = "/secrets/credentials"
	mountPathStorageCA          = "/secrets/ca"
)

type IntegrationStorageV2 struct {
	Core    *Core
	Storage *platformApi.ArangoPlatformStorage
}

func (i IntegrationStorageV2) Name() []string {
	return []string{"STORAGE", "V2"}
}

func (i IntegrationStorageV2) Validate() error {
	if i.Storage == nil {
		return errors.Errorf("Storage is nil")
	}

	if err := i.Storage.Spec.Validate(); err != nil {
		return errors.Wrap(err, "Storage failed")
	}

	if !i.Storage.Status.Conditions.IsTrue(platformApi.ReadyCondition) {
		return errors.Errorf("Storage is not Ready")
	}

	return nil
}

func (i IntegrationStorageV2) Envs() ([]core.EnvVar, error) {
	var envs = []core.EnvVar{
		{
			Name:  "INTEGRATION_STORAGE_V2",
			Value: "true",
		},
	}

	if s3 := i.Storage.Spec.GetBackend().GetS3(); s3 != nil {

		endpointURL, _ := url.Parse(s3.GetEndpoint())
		disableSSL := endpointURL.Scheme == "http"

		envs = append(envs,
			core.EnvVar{
				Name:  "INTEGRATION_STORAGE_V2_TYPE",
				Value: string(storage.S3),
			},
			core.EnvVar{
				Name:  "INTEGRATION_STORAGE_V2_S3_ENDPOINT",
				Value: s3.GetEndpoint(),
			},
			core.EnvVar{
				Name:  "INTEGRATION_STORAGE_V2_S3_ALLOW_INSECURE",
				Value: strconv.FormatBool(s3.GetAllowInsecure()),
			},
			core.EnvVar{
				Name:  "INTEGRATION_STORAGE_V2_S3_DISABLE_SSL",
				Value: strconv.FormatBool(disableSSL),
			},
			core.EnvVar{
				Name:  "INTEGRATION_STORAGE_V2_S3_REGION",
				Value: s3.GetRegion(),
			},
			core.EnvVar{
				Name:  "INTEGRATION_STORAGE_V2_S3_BUCKET_NAME",
				Value: s3.GetBucketName(),
			},
			core.EnvVar{
				Name:  "INTEGRATION_STORAGE_V2_S3_BUCKET_PREFIX",
				Value: s3.GetBucketPrefix(),
			},
			core.EnvVar{
				Name:  "INTEGRATION_STORAGE_V2_S3_PROVIDER_TYPE",
				Value: string(aws.ProviderTypeFile),
			},
			core.EnvVar{
				Name:  "INTEGRATION_STORAGE_V2_S3_PROVIDER_FILE_SECRET_KEY",
				Value: filepath.Join(mountPathStorageCredentials, constants.SecretCredentialsSecretKey),
			},
			core.EnvVar{
				Name:  "INTEGRATION_STORAGE_V2_S3_PROVIDER_FILE_ACCESS_KEY",
				Value: filepath.Join(mountPathStorageCredentials, constants.SecretCredentialsAccessKey),
			},
		)

		if !s3.GetCASecret().IsEmpty() {

			envs = append(envs,
				core.EnvVar{
					Name:  "INTEGRATION_STORAGE_V2_S3_CA",
					Value: filepath.Join(mountPathStorageCA, constants.SecretCACertificate),
				},
			)
		}
	}

	return i.Core.Envs(i, envs...), nil
}

func (i IntegrationStorageV2) GlobalEnvs() ([]core.EnvVar, error) {
	return nil, nil
}

func (i IntegrationStorageV2) Volumes() ([]core.Volume, []core.VolumeMount, error) {
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
