//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/handlers/utils"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tolerations"
)

var _ interfaces.PodCreator = &ImageUpdatePod{}
var _ interfaces.ContainerCreator = &ContainerIdentity{}

// ImageUpdatePod describes how to launch the ID ArangoD POD.
type ImageUpdatePod struct {
	spec             api.DeploymentSpec
	status           api.DeploymentStatus
	apiObject        k8sutil.APIObject
	containerCreator *ArangoDIdentity
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
	input     pod.Input
	License   *string
	ipAddress string
}

// ArangoSyncIdentity helps to resolve the ArangoSync identity, e.g.: image ID, version of the entrypoint.
type ArangoSyncIdentity struct {
	interfaces.ContainerCreator
}

type imagesBuilder struct {
	Log            logging.Logger
	Context        resources.Context
	APIObject      k8sutil.APIObject
	Spec           api.DeploymentSpec
	Status         api.DeploymentStatus
	UpdateCRStatus func(status api.DeploymentStatus) error
}

// ensureImages creates pods needed to detect ImageID for specified images.
// Returns: retrySoon, error
func (d *Deployment) ensureImages(ctx context.Context, apiObject *api.ArangoDeployment, cachedStatus inspectorInterface.Inspector) (bool, bool, error) {
	status := d.GetStatus()
	ib := imagesBuilder{
		Context:   d,
		APIObject: apiObject,
		Spec:      apiObject.GetAcceptedSpec(),
		Status:    status,
		Log:       d.log,
		UpdateCRStatus: func(status api.DeploymentStatus) error {
			if err := d.UpdateStatus(ctx, status); err != nil {
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
	role := api.ServerGroupImageDiscovery.AsRole()
	id := fmt.Sprintf("%0x", sha1.Sum([]byte(image)))[:6]
	podName := k8sutil.CreatePodName(ib.APIObject.GetName(), role, id, "")
	log := ib.Log.
		Str("pod", podName).
		Str("image", image)

	// Check if pod exists
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	pod, err := ib.Context.ACS().CurrentClusterCache().Pod().V1().Read().Get(ctxChild, podName, meta.GetOptions{})
	if err == nil {
		// Pod found
		if k8sutil.IsPodFailed(pod, utils.StringList{shared.ServerContainerName}) {
			// Wait some time before deleting the pod
			if time.Now().After(pod.GetCreationTimestamp().Add(30 * time.Second)) {
				err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
					return ib.Context.ACS().CurrentClusterCache().PodsModInterface().V1().Delete(ctxChild, podName, meta.DeleteOptions{})
				})
				if err != nil && !kerrors.IsNotFound(err) {
					log.Err(err).Warn("Failed to delete Image ID Pod")
					return false, nil
				}
			}
			return false, nil
		}
		if !k8sutil.IsPodReady(pod) {
			log.Debug("Image ID Pod is not yet ready")
			return true, nil
		}

		imageID, err := k8sutil.GetArangoDBImageIDFromPod(pod)
		if err != nil {
			log.Err(err).Warn("failed to get image ID from pod")
			return true, nil
		}
		if imageID == "" {
			// Fall back to specified image
			imageID = image
		}

		// Try fetching the ArangoDB version
		client, err := arangod.CreateArangodImageIDClient(ctx, ib.APIObject, pod.Status.PodIP)
		if err != nil {
			log.Err(err).Warn("Failed to create Image ID Pod client")
			return true, nil
		}
		ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
		defer cancel()
		v, err := client.Version(ctxChild)
		if err != nil {
			log.Err(err).Debug("Failed to fetch version from Image ID Pod")
			return true, nil
		}
		version := v.Version
		enterprise := strings.ToLower(v.License) == "enterprise"

		// We have all the info we need now, kill the pod and store the image info.
		err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			return ib.Context.ACS().CurrentClusterCache().PodsModInterface().V1().Delete(ctxChild, podName, meta.DeleteOptions{})
		})
		if err != nil && !kerrors.IsNotFound(err) {
			log.Err(err).Warn("Failed to delete Image ID Pod")
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
			log.Err(err).Warn("Failed to save Image Info in CR status")
			return true, errors.WithStack(err)
		}
		// We're done
		log.
			Str("image-id", imageID).
			Str("arangodb-version", string(version)).
			Debug("Found image ID and ArangoDB version")
		return false, nil
	}

	// Find license
	var license *string
	if s := ib.Spec.License; s.HasSecretName() {
		if secret, ok := cachedStatus.Secret().V1().GetSimple(s.GetSecretName()); ok {
			if _, ok := secret.Data[constants.SecretKeyToken]; ok {
				license = util.NewType[string](s.GetSecretName())
			}
		}
	}

	imagePod := ImageUpdatePod{
		spec:      ib.Spec,
		status:    ib.Status,
		apiObject: ib.APIObject,
		containerCreator: &ArangoDIdentity{
			ContainerCreator: &ContainerIdentity{
				ID:              ib.Spec.ID,
				image:           image,
				imagePullPolicy: ib.Spec.GetImagePullPolicy(),
			},
			License:   license,
			ipAddress: ib.Spec.GetListenAddr(),
		},
	}
	imagePod.containerCreator.input = imagePod.AsInput()

	pod, err = resources.RenderArangoPod(ctx, cachedStatus, ib.APIObject, role, id, podName, &imagePod)
	if err != nil {
		log.Err(err).Debug("Failed to render image ID pod")
		return true, errors.WithStack(err)
	}

	// here we need a pod with selector
	err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		_, _, err := resources.CreateArangoPod(ctxChild, ib.Context.ACS().CurrentClusterCache().PodsModInterface().V1(), ib.APIObject, ib.Spec, api.ServerGroupImageDiscovery, pod)
		return err
	})
	if err != nil {
		log.Err(err).Debug("Failed to create image ID pod")
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

func (i *ImageUpdatePod) GetRestartPolicy() core.RestartPolicy {
	return core.RestartPolicyNever
}

func (i *ImageUpdatePod) GetAffinityRole() string {
	return ""
}

func (i *ImageUpdatePod) GetVolumes() []core.Volume {
	return getVolumes(i.AsInput()).Volumes()
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
	shortDur := tolerations.TolerationDuration{
		Forever:  false,
		TimeSpan: time.Second * 5,
	}

	ts := make([]core.Toleration, 0, 3+len(i.spec.ID.Get().Tolerations))

	if idTolerations := i.spec.ID.Get().Tolerations; len(idTolerations) > 0 {
		for _, toleration := range idTolerations {
			ts = tolerations.AddTolerationIfNotFound(ts, toleration)
		}
	}

	ts = tolerations.AddTolerationIfNotFound(ts,
		tolerations.NewNoExecuteToleration(tolerations.TolerationKeyNodeNotReady, shortDur))
	ts = tolerations.AddTolerationIfNotFound(ts,
		tolerations.NewNoExecuteToleration(tolerations.TolerationKeyNodeUnreachable, shortDur))
	ts = tolerations.AddTolerationIfNotFound(ts,
		tolerations.NewNoExecuteToleration(tolerations.TolerationKeyNodeAlphaUnreachable, shortDur))

	return ts
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
	pod.AppendArchSelector(&a, i.spec.Architecture.AsNodeSelectorRequirement())

	pod.MergeNodeAffinity(&a, i.spec.ID.Get().NodeAffinity)

	return pod.ReturnNodeAffinityOrNil(a)
}

func (i *ImageUpdatePod) Validate(_ interfaces.Inspector) error {
	return nil
}

func (i *ImageUpdatePod) ApplyPodSpec(p *core.PodSpec) error {
	if id := i.spec.ID; id != nil {
		p.SecurityContext = i.spec.ID.SecurityContext.NewPodSecurityContext()
	}
	return nil
}

func (a *ContainerIdentity) GetArgs() ([]string, error) {
	return nil, nil
}

// GetEnvs returns environment variables for identity containers.
func (a *ContainerIdentity) GetEnvs() ([]core.EnvVar, []core.EnvFromSource) {
	return nil, nil
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
	return shared.ServerContainerName
}

func (a *ContainerIdentity) GetPorts() []core.ContainerPort {
	return []core.ContainerPort{
		{
			Name:          shared.ServerContainerName,
			ContainerPort: int32(shared.ArangoPort),
			Protocol:      core.ProtocolTCP,
		},
	}
}

func (a *ContainerIdentity) GetProbes() (*core.Probe, *core.Probe, *core.Probe, error) {
	return nil, nil, nil, nil
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
	options := k8sutil.CreateOptionPairs(64)
	options.Add("--server.authentication", "false")
	options.Addf("--server.endpoint", "tcp://%s:%d", a.ipAddress, shared.ArangoPort)
	options.Add("--database.directory", shared.ArangodVolumeMountDir)
	options.Add("--log.output", "+")

	// Security
	options.Merge(pod.Security().Args(a.input))

	return options.Copy().Sort().AsArgs(), nil
}

// GetEnvs returns environment variables for Arango identity containers.
func (a *ArangoDIdentity) GetEnvs() ([]core.EnvVar, []core.EnvFromSource) {
	env := make([]core.EnvVar, 0)

	// Add advanced check for license
	if l := a.License; l != nil {
		env = append(env, k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoLicenseKey,
			*l, constants.SecretKeyToken))
	}

	if len(env) > 0 {
		return env, nil
	}

	return nil, nil
}

// GetVolumeMounts returns volume mount for the ArangoD data.
func (a *ArangoDIdentity) GetVolumeMounts() []core.VolumeMount {
	return getVolumes(a.input).VolumeMounts()
}

func (a *ImageUpdatePod) AsInput() pod.Input {
	return pod.Input{
		ApiObject:  a.apiObject,
		Deployment: a.spec,
		Status:     a.status,
		Group:      api.ServerGroupImageDiscovery,
	}
}

// GetExecutor returns the fixed path to the ArangoSync binary in the container.
func (a *ArangoSyncIdentity) GetExecutor() string {
	return resources.ArangoSyncExecutor
}

func getVolumes(input pod.Input) pod.Volumes {
	volumes := pod.NewVolumes()
	volumes.AddVolume(k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName))
	volumes.AddVolumeMount(k8sutil.ArangodVolumeMount())

	// Security
	volumes.Append(pod.Security(), input)

	return volumes
}
