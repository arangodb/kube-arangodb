//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package reconcile

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"

	core "k8s.io/api/core/v1"

	mlApi "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/ml/storage"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	ContainerName  = "mlstorage"
	ListenPortName = "mlstorage"

	mountNameStorageCredentials = "ml-storage-credentials"
	mountNameStorageCA          = "ml-storage-ca"

	mountPathStorageCredentials = "/secrets/credentials"
	mountPathStorageCA          = "/secrets/ca"
)

func MakeStorageVolumes(storageObj *mlApi.ArangoMLStorage) ([]core.Volume, []core.VolumeMount) {
	var volumes []core.Volume
	var volumeMounts []core.VolumeMount

	spec := storageObj.Spec

	if spec.Backend.S3 != nil {

		s := spec.Backend.S3

		secretObj := s.GetCredentialsSecret()
		if secretObj.GetNamespace(storageObj) != "" {
			panic("not implemented")
		}
		volumes = append(volumes, k8sutil.CreateVolumeWithSecret(mountNameStorageCredentials, secretObj.GetName()))
		volumeMounts = append(volumeMounts, core.VolumeMount{
			Name:      mountNameStorageCredentials,
			MountPath: mountPathStorageCredentials,
		})

		if caSecret := s.GetCASecret(); !caSecret.IsEmpty() {
			if caSecret.GetNamespace(storageObj) != "" {
				panic("not implemented")
			}
			volumes = append(volumes, k8sutil.CreateVolumeWithSecret(mountNameStorageCA, caSecret.GetName()))
			volumeMounts = append(volumeMounts, core.VolumeMount{
				Name:      mountNameStorageCA,
				MountPath: mountPathStorageCA,
			})
		}
	}

	return volumes, volumeMounts
}

func MakeStorageContainer(storageObj *mlApi.ArangoMLStorage, image string) (core.Container, error) {
	storageServiceType := storage.StorageTypeS3Proxy

	spec := storageObj.Spec

	s3Spec := spec.GetBackend().GetS3()
	sidecarSpec := spec.GetMode().GetSidecar()

	endpointURL, _ := url.Parse(s3Spec.GetEndpoint())
	disableSSL := endpointURL.Scheme == "http"

	args := []string{
		string(storageServiceType),
		"--server.address", fmt.Sprintf("0.0.0.0:%d", sidecarSpec.GetListenPort()),

		"--s3.endpoint", s3Spec.GetEndpoint(),
		"--s3.allow-insecure", strconv.FormatBool(s3Spec.GetAllowInsecure()),
		"--s3.disable-ssl", strconv.FormatBool(disableSSL),
		"--s3.region", s3Spec.GetRegion(),
		"--s3.bucket", s3Spec.GetBucketName(),
		"--s3.access-key", filepath.Join(mountPathStorageCredentials, constants.SecretCredentialsAccessKey),
		"--s3.secret-key", filepath.Join(mountPathStorageCredentials, constants.SecretCredentialsSecretKey),
	}

	if !s3Spec.GetCASecret().IsEmpty() {
		args = append(args, "--s3.ca-crt", filepath.Join(mountPathStorageCA, constants.SecretCACertificate))
		args = append(args, "--s3.ca-key", filepath.Join(mountPathStorageCA, constants.SecretCAKey))
	}

	exePath := k8sutil.LifecycleBinary()
	lifecycle, err := k8sutil.NewLifecycleFinalizers()
	if err != nil {
		return core.Container{}, errors.Wrapf(err, "NewLifecycleFinalizers failed")
	}

	c := core.Container{
		Name:    ContainerName,
		Image:   image,
		Command: append([]string{exePath, "ml", "storage"}, args...),
		Env:     k8sutil.GetLifecycleEnv(),
		Ports: []core.ContainerPort{
			{
				Name:          ListenPortName,
				ContainerPort: int32(sidecarSpec.GetListenPort()),
				Protocol:      core.ProtocolTCP,
			},
		},
		Lifecycle:       lifecycle,
		Resources:       k8sutil.ExtractPodAcceptedResourceRequirement(sidecarSpec.GetResources()),
		SecurityContext: createDefaultSecurityContext(),
		ImagePullPolicy: core.PullIfNotPresent,
		VolumeMounts: []core.VolumeMount{
			k8sutil.LifecycleVolumeMount(),
		},
	}
	c.ReadinessProbe = &core.Probe{
		ProbeHandler: core.ProbeHandler{
			GRPC: &core.GRPCAction{
				Port: int32(sidecarSpec.GetListenPort()),
			},
		},
		InitialDelaySeconds: 1,  // Wait 1s before first probe
		TimeoutSeconds:      2,  // Timeout of each probe is 2s
		PeriodSeconds:       30, // Interval between probes is 30s
		SuccessThreshold:    1,  // Single probe is enough to indicate success
		FailureThreshold:    2,  // Need 2 failed probes to consider a failed state
	}

	_, vms := MakeStorageVolumes(storageObj)
	c.VolumeMounts = append(c.VolumeMounts, vms...)

	return c, nil
}

func createDefaultSecurityContext() *core.SecurityContext {
	r := &core.SecurityContext{
		RunAsUser:              util.NewType[int64](shared.DefaultRunAsUser),
		RunAsGroup:             util.NewType[int64](shared.DefaultRunAsGroup),
		RunAsNonRoot:           util.NewType(true),
		ReadOnlyRootFilesystem: util.NewType(true),
		Capabilities: &core.Capabilities{
			Drop: []core.Capability{
				"ALL",
			},
		},
	}
	return r
}
