//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package deployment

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

type ImageUpdatePod struct {
	spec  api.DeploymentSpec
	image string
}

type ArangoDImageUpdateContainer struct {
	spec  api.DeploymentSpec
	image string
}

type imagesBuilder struct {
	APIObject      k8sutil.APIObject
	Spec           api.DeploymentSpec
	Status         api.DeploymentStatus
	Log            zerolog.Logger
	KubeCli        kubernetes.Interface
	UpdateCRStatus func(status api.DeploymentStatus) error
}

// ensureImages creates pods needed to detect ImageID for specified images.
// Returns: retrySoon, error
func (d *Deployment) ensureImages(apiObject *api.ArangoDeployment) (bool, error) {
	status, lastVersion := d.GetStatus()
	ib := imagesBuilder{
		APIObject: apiObject,
		Spec:      apiObject.Spec,
		Status:    status,
		Log:       d.deps.Log,
		KubeCli:   d.deps.KubeCli,
		UpdateCRStatus: func(status api.DeploymentStatus) error {
			if err := d.UpdateStatus(status, lastVersion); err != nil {
				return maskAny(err)
			}
			return nil
		},
	}
	ctx := context.Background()
	retrySoon, err := ib.Run(ctx)
	if err != nil {
		return retrySoon, maskAny(err)
	}
	return retrySoon, nil
}

// Run creates pods needed to detect ImageID for specified images and puts the found
// image ID's into the status.Images list.
// Returns: retrySoon, error
func (ib *imagesBuilder) Run(ctx context.Context) (bool, error) {
	result := false
	// Check ArangoDB image
	if _, found := ib.Status.Images.GetByImage(ib.Spec.GetImage()); !found {
		// We need to find the image ID for the ArangoDB image
		retrySoon, err := ib.fetchArangoDBImageIDAndVersion(ctx, ib.Spec.GetImage())
		if err != nil {
			return retrySoon, maskAny(err)
		}
		result = result || retrySoon
	}

	return result, nil
}

// fetchArangoDBImageIDAndVersion checks a running pod for fetching the ID of the given image.
// When no pod exists, it is created, otherwise the ID is fetched & version detected.
// Returns: retrySoon, error
func (ib *imagesBuilder) fetchArangoDBImageIDAndVersion(ctx context.Context, image string) (bool, error) {
	role := k8sutil.ImageIDAndVersionRole
	id := fmt.Sprintf("%0x", sha1.Sum([]byte(image)))[:6]
	podName := k8sutil.CreatePodName(ib.APIObject.GetName(), role, id, "")
	ns := ib.APIObject.GetNamespace()
	log := ib.Log.With().
		Str("pod", podName).
		Str("image", image).
		Logger()

	// Check if pod exists
	if pod, err := ib.KubeCli.CoreV1().Pods(ns).Get(podName, metav1.GetOptions{}); err == nil {
		// Pod found
		if k8sutil.IsPodFailed(pod) {
			// Wait some time before deleting the pod
			if time.Now().After(pod.GetCreationTimestamp().Add(30 * time.Second)) {
				if err := ib.KubeCli.CoreV1().Pods(ns).Delete(podName, nil); err != nil && !k8sutil.IsNotFound(err) {
					log.Warn().Err(err).Msg("Failed to delete Image ID Pod")
					return false, nil
				}
			}
			return false, nil
		}
		if !k8sutil.IsPodReady(pod) {
			log.Debug().Msg("Image ID Pod is not yet ready")
			return true, nil
		}

		if len(pod.Status.ContainerStatuses) == 0 {
			log.Warn().Msg("Empty list of ContainerStatuses")
			return true, nil
		}
		imageID := k8sutil.GetArangoDBImageIDFromPod(pod)
		if imageID == "" {
			// Fall back to specified image
			imageID = image
		}

		// Try fetching the ArangoDB version
		client, err := arangod.CreateArangodImageIDClient(ctx, ib.APIObject, role, id)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to create Image ID Pod client")
			return true, nil
		}
		v, err := client.Version(ctx)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to fetch version from Image ID Pod")
			return true, nil
		}
		version := v.Version
		enterprise := strings.ToLower(v.License) == "enterprise"

		// We have all the info we need now, kill the pod and store the image info.
		if err := ib.KubeCli.CoreV1().Pods(ns).Delete(podName, nil); err != nil && !k8sutil.IsNotFound(err) {
			log.Warn().Err(err).Msg("Failed to delete Image ID Pod")
			return true, nil
		}

		info := api.ImageInfo{
			Image:           image,
			ImageID:         imageID,
			ArangoDBVersion: version,
			Enterprise:      enterprise,
		}
		ib.Status.Images.AddOrUpdate(info)
		if err := ib.UpdateCRStatus(ib.Status); err != nil {
			log.Warn().Err(err).Msg("Failed to save Image Info in CR status")
			return true, maskAny(err)
		}
		// We're done
		log.Debug().
			Str("image-id", imageID).
			Str("arangodb-version", string(version)).
			Msg("Found image ID and ArangoDB version")
		return false, nil
	}
	// Pod cannot be fetched, ensure it is created
	args := []string{
		"--server.authentication=false",
		fmt.Sprintf("--server.endpoint=tcp://%s:%d", ib.Spec.GetListenAddr(), k8sutil.ArangoPort),
		"--database.directory=" + k8sutil.ArangodVolumeMountDir,
		"--log.output=+",
	}

	imagePod := ImageUpdatePod{
		spec:  ib.Spec,
		image: image,
	}

	if err := resources.CreateArangoPod(ib.KubeCli, ib.APIObject, role, id, podName, args, &imagePod); err != nil {
		log.Debug().Err(err).Msg("Failed to create image ID pod")
		return true, maskAny(err)
	}
	// Come back soon to inspect the pod
	return true, nil
}

func (a *ArangoDImageUpdateContainer) GetExecutor() string {
	return resources.ArangoDExecutor
}

func (a *ArangoDImageUpdateContainer) GetProbes() (*v1.Probe, *v1.Probe, error) {
	return nil, nil, nil
}

func (a *ArangoDImageUpdateContainer) GetResourceRequirements() v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Limits:   make(v1.ResourceList),
		Requests: make(v1.ResourceList),
	}
}

func (a *ArangoDImageUpdateContainer) GetImage() string {
	return a.image
}

func (a *ArangoDImageUpdateContainer) GetEnvs() []v1.EnvVar {
	env := make([]v1.EnvVar, 0)

	if a.spec.License.HasSecretName() {
		env = append(env, k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoLicenseKey,
			a.spec.License.GetSecretName(), constants.SecretKeyToken))
	}

	if len(env) > 0 {
		return env
	}

	return nil
}

func (a *ArangoDImageUpdateContainer) GetLifecycle() (*v1.Lifecycle, error) {
	return nil, nil
}

func (a *ArangoDImageUpdateContainer) GetImagePullPolicy() v1.PullPolicy {
	return a.spec.GetImagePullPolicy()
}

func (i *ImageUpdatePod) Init(pod *v1.Pod) {
	terminationGracePeriodSeconds := int64((time.Second * 30).Seconds())
	pod.Spec.TerminationGracePeriodSeconds = &terminationGracePeriodSeconds
}

func (i *ImageUpdatePod) GetImagePullSecrets() []string {
	return i.spec.ImagePullSecrets
}

func (i *ImageUpdatePod) GetContainerCreator() k8sutil.ContainerCreator {
	return &ArangoDImageUpdateContainer{
		spec:  i.spec,
		image: i.image,
	}
}

func (i *ImageUpdatePod) GetAffinityRole() string {
	return ""
}

func (i *ImageUpdatePod) GetVolumes() ([]v1.Volume, []v1.VolumeMount) {
	var volumes []v1.Volume
	var volumeMounts []v1.VolumeMount

	volumes = append(volumes, k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName))
	volumeMounts = append(volumeMounts, k8sutil.ArangodVolumeMount())

	return volumes, volumeMounts
}

func (i *ImageUpdatePod) GetSidecars(*v1.Pod) {
	return
}

func (i *ImageUpdatePod) GetInitContainers() ([]v1.Container, error) {
	return nil, nil
}

func (i *ImageUpdatePod) GetFinalizers() []string {
	return nil
}

func (i *ImageUpdatePod) GetTolerations() []v1.Toleration {

	shortDur := k8sutil.TolerationDuration{
		Forever:  false,
		TimeSpan: time.Second * 5,
	}

	tolerations := make([]v1.Toleration, 0, 2)
	tolerations = k8sutil.AddTolerationIfNotFound(tolerations,
		k8sutil.NewNoExecuteToleration(k8sutil.TolerationKeyNodeNotReady, shortDur))
	tolerations = k8sutil.AddTolerationIfNotFound(tolerations,
		k8sutil.NewNoExecuteToleration(k8sutil.TolerationKeyNodeUnreachable, shortDur))
	tolerations = k8sutil.AddTolerationIfNotFound(tolerations,
		k8sutil.NewNoExecuteToleration(k8sutil.TolerationKeyNodeAlphaUnreachable, shortDur))

	return tolerations
}

func (a *ArangoDImageUpdateContainer) GetSecurityContext() *v1.SecurityContext {
	return nil
}

func (i *ImageUpdatePod) IsDeploymentMode() bool {
	return true
}

func (i *ImageUpdatePod) GetNodeSelector() map[string]string {
	return nil
}

func (i *ImageUpdatePod) GetServiceAccountName() string {
	return ""
}
