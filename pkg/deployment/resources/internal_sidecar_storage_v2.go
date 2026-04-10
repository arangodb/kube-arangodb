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

package resources

import (
	"net/url"
	"path/filepath"
	"strconv"

	core "k8s.io/api/core/v1"

	pbImplStorageV2 "github.com/arangodb/kube-arangodb/integrations/storage/v2"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/util/aws"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	internalSidecarStorageCredentialsMount = "sidecar-storage-credentials"
	internalSidecarStorageCAMount          = "sidecar-storage-ca"

	internalSidecarStorageCredentialsPath = "/secrets/storage/credentials"
	internalSidecarStorageCAPath          = "/secrets/storage/ca"
)

func internalSidecarStorageV2Args(storage *platformApi.ArangoPlatformStorage) k8sutil.OptionPairs {
	options := k8sutil.CreateOptionPairs(16)

	if s3 := storage.Spec.GetBackend().GetS3(); s3 != nil {
		endpointURL, _ := url.Parse(s3.GetEndpoint())
		disableSSL := endpointURL != nil && endpointURL.Scheme == "http"

		options.Add("--storage.v2.type", string(pbImplStorageV2.ConfigurationTypeS3))
		options.Add("--storage.v2.s3.endpoint", s3.GetEndpoint())
		options.Add("--storage.v2.s3.allow-insecure", strconv.FormatBool(s3.GetAllowInsecure()))
		options.Add("--storage.v2.s3.disable-ssl", strconv.FormatBool(disableSSL))
		options.Add("--storage.v2.s3.region", s3.GetRegion())
		options.Add("--storage.v2.s3.bucket.name", s3.GetBucketName())
		options.Add("--storage.v2.s3.bucket.prefix", s3.GetBucketPrefix())
		options.Add("--storage.v2.s3.provider.type", string(aws.ProviderTypeFile))
		options.Add("--storage.v2.s3.provider.file.access-key", filepath.Join(internalSidecarStorageCredentialsPath, utilConstants.SecretCredentialsAccessKey))
		options.Add("--storage.v2.s3.provider.file.secret-key", filepath.Join(internalSidecarStorageCredentialsPath, utilConstants.SecretCredentialsSecretKey))

		if !s3.GetCASecret().IsEmpty() {
			options.Add("--storage.v2.s3.ca", filepath.Join(internalSidecarStorageCAPath, utilConstants.SecretCACertificate))
		}
	} else if gcs := storage.Spec.GetBackend().GetGCS(); gcs != nil {
		options.Add("--storage.v2.type", string(pbImplStorageV2.ConfigurationTypeGCS))
		options.Add("--storage.v2.gcs.project-id", gcs.GetProjectID())
		options.Add("--storage.v2.gcs.bucket.name", gcs.GetBucketName())
		options.Add("--storage.v2.gcs.bucket.prefix", gcs.GetBucketPrefix())
		options.Add("--storage.v2.gcs.provider.sa.file", filepath.Join(internalSidecarStorageCredentialsPath, utilConstants.SecretCredentialsServiceAccount))
	} else if abs := storage.Spec.GetBackend().GetAzureBlobStorage(); abs != nil {
		options.Add("--storage.v2.type", string(pbImplStorageV2.ConfigurationTypeAzure))
		options.Add("--storage.v2.azure-blob-storage.client.tenant-id", abs.GetTenantID())
		options.Add("--storage.v2.azure-blob-storage.account-name", abs.GetAccountName())
		options.Add("--storage.v2.azure-blob-storage.endpoint", abs.GetEndpoint())
		options.Add("--storage.v2.azure-blob-storage.bucket.name", abs.GetBucketName())
		options.Add("--storage.v2.azure-blob-storage.bucket.prefix", abs.GetBucketPrefix())
		options.Add("--storage.v2.azure-blob-storage.client.secret.client-id-file", filepath.Join(internalSidecarStorageCredentialsPath, utilConstants.SecretCredentialsAzureBlobStorageClientID))
		options.Add("--storage.v2.azure-blob-storage.client.secret.client-secret-file", filepath.Join(internalSidecarStorageCredentialsPath, utilConstants.SecretCredentialsAzureBlobStorageClientSecret))
	}

	return options
}

func internalSidecarStorageV2Volumes(storage *platformApi.ArangoPlatformStorage) ([]core.Volume, []core.VolumeMount) {
	var volumes []core.Volume
	var volumeMounts []core.VolumeMount

	if s3 := storage.Spec.GetBackend().GetS3(); s3 != nil {
		secretObj := s3.GetCredentialsSecret()
		volumes = append(volumes, k8sutil.CreateVolumeWithSecret(internalSidecarStorageCredentialsMount, secretObj.GetName()))
		volumeMounts = append(volumeMounts, core.VolumeMount{
			Name:      internalSidecarStorageCredentialsMount,
			MountPath: internalSidecarStorageCredentialsPath,
		})

		if caSecret := s3.GetCASecret(); !caSecret.IsEmpty() {
			volumes = append(volumes, k8sutil.CreateVolumeWithSecret(internalSidecarStorageCAMount, caSecret.GetName()))
			volumeMounts = append(volumeMounts, core.VolumeMount{
				Name:      internalSidecarStorageCAMount,
				MountPath: internalSidecarStorageCAPath,
			})
		}
	} else if gcs := storage.Spec.GetBackend().GetGCS(); gcs != nil {
		secretObj := gcs.GetCredentialsSecret()
		volumes = append(volumes, k8sutil.CreateVolumeWithSecret(internalSidecarStorageCredentialsMount, secretObj.GetName()))
		volumeMounts = append(volumeMounts, core.VolumeMount{
			Name:      internalSidecarStorageCredentialsMount,
			MountPath: internalSidecarStorageCredentialsPath,
		})
	} else if abs := storage.Spec.GetBackend().GetAzureBlobStorage(); abs != nil {
		secretObj := abs.GetCredentialsSecret()
		volumes = append(volumes, k8sutil.CreateVolumeWithSecret(internalSidecarStorageCredentialsMount, secretObj.GetName()))
		volumeMounts = append(volumeMounts, core.VolumeMount{
			Name:      internalSidecarStorageCredentialsMount,
			MountPath: internalSidecarStorageCredentialsPath,
		})
	}

	return volumes, volumeMounts
}
