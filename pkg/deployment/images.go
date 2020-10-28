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

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"

	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"

	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

var _ interfaces.PodCreator = &ImageUpdatePod{}
var _ interfaces.ContainerCreator = &ArangoDImageUpdateContainer{}

type ImageUpdatePod struct {
	spec      api.DeploymentSpec
	apiObject k8sutil.APIObject
	image     string
}

func (i *ImageUpdatePod) Annotations() map[string]string {
	return nil
}

func (i *ImageUpdatePod) Labels() map[string]string {
	return nil
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
func (d *Deployment) ensureImages(apiObject *api.ArangoDeployment) (bool, bool, error) {
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
	retrySoon, exists, err := ib.Run(ctx)
	if err != nil {
		return retrySoon, exists, maskAny(err)
	}
	return retrySoon, exists, nil
}

// Run creates pods needed to detect ImageID for specified images and puts the found
// image ID's into the status.Images list.
// Returns: retrySoon, error
func (ib *imagesBuilder) Run(ctx context.Context) (bool, bool, error) {
	// Check ArangoDB image
	if _, found := ib.Status.Images.GetByImage(ib.Spec.GetImage()); !found {
		// We need to find the image ID for the ArangoDB image
		retrySoon, err := ib.fetchArangoDBImageIDAndVersion(ctx, ib.Spec.GetImage())
		if err != nil {
			return retrySoon, false, maskAny(err)
		}
		return retrySoon, false, nil
	}

	return false, true, nil
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
		spec:      ib.Spec,
		image:     image,
		apiObject: ib.APIObject,
	}

	pod, err := resources.RenderArangoPod(ib.APIObject, role, id, podName, args, &imagePod)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to render image ID pod")
		return true, maskAny(err)
	}

	if _, err := resources.CreateArangoPod(ib.KubeCli, ib.APIObject, ib.Spec, api.ServerGroupImageDiscovery, pod); err != nil {
		log.Debug().Err(err).Msg("Failed to create image ID pod")
		return true, maskAny(err)
	}
	// Come back soon to inspect the pod
	return true, nil
}

func (a *ArangoDImageUpdateContainer) GetExecutor() string {
	return resources.ArangoDExecutor
}

func (a *ArangoDImageUpdateContainer) GetProbes() (*core.Probe, *core.Probe, error) {
	return nil, nil, nil
}

func (a *ArangoDImageUpdateContainer) GetResourceRequirements() core.ResourceRequirements {
	return core.ResourceRequirements{
		Limits:   make(core.ResourceList),
		Requests: make(core.ResourceList),
	}
}

func (a *ArangoDImageUpdateContainer) GetImage() string {
	return a.image
}

func (a *ArangoDImageUpdateContainer) GetEnvs() []core.EnvVar {
	env := make([]core.EnvVar, 0)

	if a.spec.License.HasSecretName() {
		env = append(env, k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoLicenseKey,
			a.spec.License.GetSecretName(), constants.SecretKeyToken))
	}

	if len(env) > 0 {
		return env
	}

	return nil
}

func (a *ArangoDImageUpdateContainer) GetLifecycle() (*core.Lifecycle, error) {
	return nil, nil
}

func (a *ArangoDImageUpdateContainer) GetImagePullPolicy() core.PullPolicy {
	return a.spec.GetImagePullPolicy()
}

func (i *ImageUpdatePod) GetName() string {
	return i.apiObject.GetName()
}

func (i *ImageUpdatePod) GetRole() string {
	return "id"
}

func (i *ImageUpdatePod) Init(pod *core.Pod) {
	terminationGracePeriodSeconds := int64((time.Second * 30).Seconds())
	pod.Spec.TerminationGracePeriodSeconds = &terminationGracePeriodSeconds
	pod.Spec.PriorityClassName = i.spec.ID.Get().PriorityClassName
}

func (i *ImageUpdatePod) GetImagePullSecrets() []string {
	return i.spec.ImagePullSecrets
}

func (i *ImageUpdatePod) GetContainerCreator() interfaces.ContainerCreator {
	return &ArangoDImageUpdateContainer{
		spec:  i.spec,
		image: i.image,
	}
}

func (i *ImageUpdatePod) GetAffinityRole() string {
	return ""
}

func (i *ImageUpdatePod) GetVolumes() ([]core.Volume, []core.VolumeMount) {
	var volumes []core.Volume
	var volumeMounts []core.VolumeMount

	volumes = append(volumes, k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName))
	volumeMounts = append(volumeMounts, k8sutil.ArangodVolumeMount())

	return volumes, volumeMounts
}

func (i *ImageUpdatePod) GetSidecars(*core.Pod) {
}

func (i *ImageUpdatePod) GetInitContainers() ([]core.Container, error) {
	return nil, nil
}

func (i *ImageUpdatePod) GetFinalizers() []string {
	return nil
}

func (i *ImageUpdatePod) GetTolerations() []core.Toleration {
	shortDur := k8sutil.TolerationDuration{
		Forever:  false,
		TimeSpan: time.Second * 5,
	}

	tolerations := make([]core.Toleration, 0, 3+len(i.spec.ID.Get().Tolerations))

	if idTolerations := i.spec.ID.Get().Tolerations; len(idTolerations) > 0 {
		for _, toleration := range idTolerations {
			tolerations = k8sutil.AddTolerationIfNotFound(tolerations, toleration)
		}
	}

	tolerations = k8sutil.AddTolerationIfNotFound(tolerations,
		k8sutil.NewNoExecuteToleration(k8sutil.TolerationKeyNodeNotReady, shortDur))
	tolerations = k8sutil.AddTolerationIfNotFound(tolerations,
		k8sutil.NewNoExecuteToleration(k8sutil.TolerationKeyNodeUnreachable, shortDur))
	tolerations = k8sutil.AddTolerationIfNotFound(tolerations,
		k8sutil.NewNoExecuteToleration(k8sutil.TolerationKeyNodeAlphaUnreachable, shortDur))

	return tolerations
}

func (i *ImageUpdatePod) IsDeploymentMode() bool {
	return true
}

func (i *ImageUpdatePod) GetNodeSelector() map[string]string {
	return i.spec.ID.Get().NodeSelector
}

func (i *ImageUpdatePod) GetServiceAccountName() string {
	return ""
}

func (a *ArangoDImageUpdateContainer) GetPorts() []core.ContainerPort {
	return []core.ContainerPort{
		{
			Name:          "server",
			ContainerPort: int32(k8sutil.ArangoPort),
			Protocol:      core.ProtocolTCP,
		},
	}
}

func (a *ArangoDImageUpdateContainer) GetSecurityContext() *core.SecurityContext {
	// Default security context
	var v api.ServerGroupSpecSecurityContext
	return v.NewSecurityContext()
}

func (i *ImageUpdatePod) GetPodAntiAffinity() *core.PodAntiAffinity {
	a := core.PodAntiAffinity{}

	pod.AppendPodAntiAffinityDefault(i, &a)

	pod.MergePodAntiAffinity(&a, i.spec.ID.Get().AntiAffinity)

	return pod.ReturnPodAntiAffinityOrNil(a)
}

func (i *ImageUpdatePod) GetPodAffinity() *core.PodAffinity {
	a := core.PodAffinity{}

	pod.MergePodAffinity(&a, i.spec.ID.Get().Affinity)

	return pod.ReturnPodAffinityOrNil(a)
}

func (i *ImageUpdatePod) GetNodeAffinity() *core.NodeAffinity {
	a := core.NodeAffinity{}

	pod.AppendNodeSelector(&a)

	pod.MergeNodeAffinity(&a, i.spec.ID.Get().NodeAffinity)

	return pod.ReturnNodeAffinityOrNil(a)
}

func (i *ImageUpdatePod) Validate(cachedStatus inspector.Inspector) error {
	return nil
}

func (i *ImageUpdatePod) ApplyPodSpec(spec *core.PodSpec) error {
	return nil
}
