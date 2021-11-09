//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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

package deployment

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
)

var _ interfaces.PodCreator = &ImageUpdatePod{}
var _ interfaces.ContainerCreator = &ContainerIdentity{}

// ImageUpdatePod describes how to launch the ID ArangoD POD.
type ImageUpdatePod struct {
	spec             api.DeploymentSpec
	apiObject        k8sutil.APIObject
	containerCreator interfaces.ContainerCreator
}

// ContainerIdentity helps to resolve the container identity, e.g.: image ID, version of the entrypoint.
type ContainerIdentity struct {
	ID              *api.ServerIDGroupSpec
	image           string
	imagePullPolicy core.PullPolicy
}

// ArangoDIdentity helps to resolve the ArangoD identity, e.g.: image ID, version of the entrypoint.
type ArangoDIdentity struct {
	interfaces.ContainerCreator
	License   api.LicenseSpec
	ipAddress string
}

// ArangoSyncIdentity helps to resolve the ArangoSync identity, e.g.: image ID, version of the entrypoint.
type ArangoSyncIdentity struct {
	interfaces.ContainerCreator
}

type imagesBuilder struct {
	Context        resources.Context
	APIObject      k8sutil.APIObject
	Spec           api.DeploymentSpec
	Status         api.DeploymentStatus
	Log            zerolog.Logger
	UpdateCRStatus func(status api.DeploymentStatus) error
}

// ensureImages creates pods needed to detect ImageID for specified images.
// Returns: retrySoon, error
func (d *Deployment) ensureImages(ctx context.Context, apiObject *api.ArangoDeployment, cachedStatus inspectorInterface.Inspector) (bool, bool, error) {
	status, lastVersion := d.GetStatus()
	ib := imagesBuilder{
		Context:   d,
		APIObject: apiObject,
		Spec:      apiObject.Spec,
		Status:    status,
		Log:       d.deps.Log,
		UpdateCRStatus: func(status api.DeploymentStatus) error {
			if err := d.UpdateStatus(ctx, status, lastVersion); err != nil {
				return errors.WithStack(err)
			}
			return nil
		},
	}
	retrySoon, exists, err := ib.Run(ctx, cachedStatus)
	if err != nil {
		return retrySoon, exists, errors.WithStack(err)
	}
	return retrySoon, exists, nil
}

// Run creates pods needed to detect ImageID for specified images and puts the found
// image ID's into the status.Images list.
// Returns: retrySoon, error
func (ib *imagesBuilder) Run(ctx context.Context, cachedStatus inspectorInterface.Inspector) (bool, bool, error) {
	// Check ArangoDB image
	if _, found := ib.Status.Images.GetByImage(ib.Spec.GetImage()); !found {
		// We need to find the image ID for the ArangoDB image
		retrySoon, err := ib.fetchArangoDBImageIDAndVersion(ctx, cachedStatus, ib.Spec.GetImage())
		if err != nil {
			return retrySoon, false, errors.WithStack(err)
		}
		return retrySoon, false, nil
	}

	return false, true, nil
}

// fetchArangoDBImageIDAndVersion checks a running pod for fetching the ID of the given image.
// When no pod exists, it is created, otherwise the ID is fetched & version detected.
// Returns: retrySoon, error
func (ib *imagesBuilder) fetchArangoDBImageIDAndVersion(ctx context.Context, cachedStatus inspectorInterface.Inspector, image string) (bool, error) {
	role := k8sutil.ImageIDAndVersionRole
	id := fmt.Sprintf("%0x", sha1.Sum([]byte(image)))[:6]
	podName := k8sutil.CreatePodName(ib.APIObject.GetName(), role, id, "")
	log := ib.Log.With().
		Str("pod", podName).
		Str("image", image).
		Logger()

	// Check if pod exists
	ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
	defer cancel()
	pod, err := ib.Context.GetCachedStatus().PodReadInterface().Get(ctxChild, podName, metav1.GetOptions{})
	if err == nil {
		// Pod found
		if k8sutil.IsPodFailed(pod) {
			// Wait some time before deleting the pod
			if time.Now().After(pod.GetCreationTimestamp().Add(30 * time.Second)) {
				err := k8sutil.RunWithTimeout(ctx, func(ctxChild context.Context) error {
					return ib.Context.PodsModInterface().Delete(ctxChild, podName, metav1.DeleteOptions{})
				})
				if err != nil && !k8sutil.IsNotFound(err) {
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
		client, err := arangod.CreateArangodImageIDClient(ctx, ib.APIObject, pod.Status.PodIP)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to create Image ID Pod client")
			return true, nil
		}
		ctxChild, cancel = context.WithTimeout(ctx, arangod.GetRequestTimeout())
		defer cancel()
		v, err := client.Version(ctxChild)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to fetch version from Image ID Pod")
			return true, nil
		}
		version := v.Version
		enterprise := strings.ToLower(v.License) == "enterprise"

		// We have all the info we need now, kill the pod and store the image info.
		err = k8sutil.RunWithTimeout(ctx, func(ctxChild context.Context) error {
			return ib.Context.PodsModInterface().Delete(ctxChild, podName, metav1.DeleteOptions{})
		})
		if err != nil && !k8sutil.IsNotFound(err) {
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
			return true, errors.WithStack(err)
		}
		// We're done
		log.Debug().
			Str("image-id", imageID).
			Str("arangodb-version", string(version)).
			Msg("Found image ID and ArangoDB version")
		return false, nil
	}

	imagePod := ImageUpdatePod{
		spec:      ib.Spec,
		apiObject: ib.APIObject,
		containerCreator: &ArangoDIdentity{
			ContainerCreator: &ContainerIdentity{
				ID:              ib.Spec.ID,
				image:           image,
				imagePullPolicy: ib.Spec.GetImagePullPolicy(),
			},
			License:   ib.Spec.License,
			ipAddress: ib.Spec.GetListenAddr(),
		},
	}

	pod, err = resources.RenderArangoPod(ctx, cachedStatus, ib.APIObject, role, id, podName, &imagePod)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to render image ID pod")
		return true, errors.WithStack(err)
	}

	err = k8sutil.RunWithTimeout(ctx, func(ctxChild context.Context) error {
		_, _, err := resources.CreateArangoPod(ctxChild, ib.Context.PodsModInterface(), ib.APIObject, ib.Spec, api.ServerGroupImageDiscovery, pod)
		return err
	})
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create image ID pod")
		return true, errors.WithStack(err)
	}
	// Come back soon to inspect the pod
	return true, nil
}

func (i *ImageUpdatePod) Annotations() map[string]string {
	return nil
}

func (i *ImageUpdatePod) Labels() map[string]string {
	return nil
}

func (i *ImageUpdatePod) GetName() string {
	return i.apiObject.GetName()
}

func (i *ImageUpdatePod) GetRole() string {
	return "id"
}

func (i *ImageUpdatePod) Init(_ context.Context, _ interfaces.Inspector, pod *core.Pod) error {
	terminationGracePeriodSeconds := int64((time.Second * 30).Seconds())
	pod.Spec.TerminationGracePeriodSeconds = &terminationGracePeriodSeconds
	pod.Spec.PriorityClassName = i.spec.ID.Get().PriorityClassName

	return nil
}

func (i *ImageUpdatePod) GetImagePullSecrets() []string {
	return i.spec.ImagePullSecrets
}

func (i *ImageUpdatePod) GetContainerCreator() interfaces.ContainerCreator {
	return i.containerCreator
}

func (i *ImageUpdatePod) GetAffinityRole() string {
	return ""
}

func (i *ImageUpdatePod) GetVolumes() []core.Volume {
	return getVolumes().Volumes()
}

func (i *ImageUpdatePod) GetSidecars(*core.Pod) error {
	return nil
}

func (i *ImageUpdatePod) GetInitContainers(cachedStatus interfaces.Inspector) ([]core.Container, error) {
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
	return i.spec.ID.GetServiceAccountName()
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

func (i *ImageUpdatePod) Validate(_ interfaces.Inspector) error {
	return nil
}

func (i *ImageUpdatePod) ApplyPodSpec(_ *core.PodSpec) error {
	return nil
}

func (a *ContainerIdentity) GetArgs() ([]string, error) {
	return nil, nil
}

func (a *ContainerIdentity) GetEnvs() []core.EnvVar {
	return nil
}

func (a *ContainerIdentity) GetExecutor() string {
	return a.ID.GetEntrypoint(resources.ArangoDExecutor)
}

func (a *ContainerIdentity) GetImage() string {
	return a.image
}

func (a *ContainerIdentity) GetImagePullPolicy() core.PullPolicy {
	return a.imagePullPolicy
}

func (a *ContainerIdentity) GetLifecycle() (*core.Lifecycle, error) {
	return nil, nil
}

func (a *ContainerIdentity) GetName() string {
	return k8sutil.ServerContainerName
}

func (a *ContainerIdentity) GetPorts() []core.ContainerPort {
	return []core.ContainerPort{
		{
			Name:          k8sutil.ServerContainerName,
			ContainerPort: int32(k8sutil.ArangoPort),
			Protocol:      core.ProtocolTCP,
		},
	}
}

func (a *ContainerIdentity) GetProbes() (*core.Probe, *core.Probe, error) {
	return nil, nil, nil
}

func (a *ContainerIdentity) GetResourceRequirements() core.ResourceRequirements {
	return a.ID.GetResources()
}

func (a *ContainerIdentity) GetSecurityContext() *core.SecurityContext {
	return a.ID.Get().SecurityContext.NewSecurityContext()
}

// GetVolumeMounts returns nil for the basic container identity.
func (a *ContainerIdentity) GetVolumeMounts() []core.VolumeMount {
	return nil
}

// GetArgs returns the list of arguments for the ArangoD container identification.
func (a *ArangoDIdentity) GetArgs() ([]string, error) {
	return []string{
		"--server.authentication=false",
		fmt.Sprintf("--server.endpoint=tcp://%s:%d", a.ipAddress, k8sutil.ArangoPort),
		"--database.directory=" + k8sutil.ArangodVolumeMountDir,
		"--log.output=+",
	}, nil
}

func (a *ArangoDIdentity) GetEnvs() []core.EnvVar {
	env := make([]core.EnvVar, 0)

	if a.License.HasSecretName() {
		env = append(env, k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoLicenseKey,
			a.License.GetSecretName(), constants.SecretKeyToken))
	}

	if len(env) > 0 {
		return env
	}

	return nil
}

// GetVolumeMounts returns volume mount for the ArangoD data.
func (a *ArangoDIdentity) GetVolumeMounts() []core.VolumeMount {
	return getVolumes().VolumeMounts()
}

// GetExecutor returns the fixed path to the ArangoSync binary in the container.
func (a *ArangoSyncIdentity) GetExecutor() string {
	return resources.ArangoSyncExecutor
}

func getVolumes() pod.Volumes {
	volumes := pod.NewVolumes()
	volumes.AddVolume(k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName))
	volumes.AddVolumeMount(k8sutil.ArangodVolumeMount())

	return volumes
}
