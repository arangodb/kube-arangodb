//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package k8sutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/handlers/utils"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	podv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/pod/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

const (
	ServerContainerConditionContainersNotReady = "ContainersNotReady"
	ServerContainerConditionPrefix             = "containers with unready status: "
)

// GetAnyVolumeByName returns the volume in the given volumes with the given name.
// Returns false if not found.
func GetAnyVolumeByName(volumes []core.Volume, name string) (core.Volume, bool) {
	for _, c := range volumes {
		if c.Name == name {
			return c, true
		}
	}
	return core.Volume{}, false
}

// GetAnyVolumeMountByName returns the volumemount in the given volumemountss with the given name.
// Returns false if not found.
func GetAnyVolumeMountByName(volumes []core.VolumeMount, name string) (core.VolumeMount, bool) {
	for _, c := range volumes {
		if c.Name == name {
			return c, true
		}
	}
	return core.VolumeMount{}, false
}

// IsPodReady returns true if the PodReady condition on
// the given pod is set to true.
func IsPodReady(pod *core.Pod) bool {
	condition := getPodCondition(&pod.Status, core.PodReady)
	return condition != nil && condition.Status == core.ConditionTrue
}

func IsContainerStarted(pod *core.Pod, container string) bool {
	for _, c := range pod.Status.ContainerStatuses {
		if c.Name != container {
			continue
		}

		return c.State.Terminated != nil || c.State.Running != nil
	}

	return false
}

// AreContainersReady checks whether Pod is considered as ready.
// Returns true if the PodReady condition on the given pod is set to true,
// or all provided containers' names are running and are not in the list of failed containers.
func AreContainersReady(pod *core.Pod, coreContainers utils.StringList) bool {
	condition := getPodCondition(&pod.Status, core.PodReady)
	if condition == nil {
		return false
	}

	if condition.Status == core.ConditionTrue {
		return true
	}

	// Check if all required containers are running.
	for _, c := range coreContainers {
		if !IsContainerRunning(pod, c) {
			return false
		}
	}

	// From here on all required containers are running, but unready condition must be checked additionally.
	switch condition.Reason {
	case ServerContainerConditionContainersNotReady:
		unreadyContainers, ok := extractContainerNamesFromConditionMessage(condition.Message)

		if !ok {
			return false
		}

		for _, c := range coreContainers {
			if unreadyContainers.Has(c) {
				// The container is on the list with unready containers.
				return false
			}
		}

		return true
	}

	return false
}

func extractContainerNamesFromConditionMessage(msg string) (utils.StringList, bool) {
	if !strings.HasPrefix(msg, ServerContainerConditionPrefix) {
		return nil, false
	}

	unreadyContainers := strings.TrimPrefix(msg, ServerContainerConditionPrefix)

	if !strings.HasPrefix(unreadyContainers, "[") {
		return nil, false
	}

	if !strings.HasSuffix(unreadyContainers, "]") {
		return nil, false
	}

	unreadyContainers = strings.TrimPrefix(strings.TrimSuffix(unreadyContainers, "]"), "[")

	unreadyContainersList := utils.StringList(strings.Split(unreadyContainers, " "))

	return unreadyContainersList, true
}

// GetPodByName returns pod if it exists among the pods' list
// Returns false if not found.
func GetPodByName(pods []core.Pod, podName string) (core.Pod, bool) {
	for _, pod := range pods {
		if pod.GetName() == podName {
			return pod, true
		}
	}
	return core.Pod{}, false
}

// IsPodServerContainerRunning returns true if the arangodb container of the pod is still running
func IsPodServerContainerRunning(pod *core.Pod) bool {
	return IsContainerRunning(pod, shared.ServerContainerName)
}

// IsContainerRunning returns true if the container of the pod is still running
func IsContainerRunning(pod *core.Pod, name string) bool {
	for _, c := range pod.Status.ContainerStatuses {
		if c.Name != name {
			continue
		}

		if c.State.Running == nil {
			return false
		}

		return true
	}
	return false
}

// IsPodSucceeded returns true when all core containers are terminated wih a zero exit code,
// or the whole pod has been succeeded.
func IsPodSucceeded(pod *core.Pod, coreContainers utils.StringList) bool {
	if pod.Status.Phase == core.PodSucceeded {
		return true
	}

	core, succeeded := 0, 0
	for _, c := range pod.Status.ContainerStatuses {
		if !coreContainers.Has(c.Name) {
			// It is not core container, so check next one status.
			continue
		}

		core++
		if t := c.State.Terminated; t != nil && t.ExitCode == 0 {
			succeeded++
		}
	}

	if core > 0 && core == succeeded {
		// If there are some core containers and all of them succeeded then return that the whole pod succeeded.
		return true
	}

	return false
}

// IsPodFailed returns true when one of the core containers is terminated wih a non-zero exit code,
// or the whole pod has been failed.
func IsPodFailed(pod *core.Pod, coreContainers utils.StringList) bool {
	if pod.Status.Phase == core.PodFailed {
		return true
	}

	allCore, succeeded, failed := 0, 0, 0
	for _, c := range pod.Status.ContainerStatuses {
		if !coreContainers.Has(c.Name) {
			// It is not core container, so check next one status.
			continue
		}

		allCore++
		if t := c.State.Terminated; t != nil {
			// A core container is terminated.
			if t.ExitCode != 0 {
				failed++
			} else {
				succeeded++
			}
		}
	}

	if failed == 0 && succeeded == 0 {
		// All core containers are not terminated.
		return false
	}

	if failed > 0 {
		// Some (or all) core containers have been terminated.
		// Some other core containers can be still running or succeeded,
		// but the whole pod is considered as failed.
		return true
	} else if allCore == succeeded {
		// All core containers are succeeded, so the pod is not failed.
		// The function `IsPodSucceeded` should recognize it in next iteration.
		return false
	}

	// Some core containers are succeeded, but not all of them.
	return true
}

// IsContainerFailed returns true if the arangodb container
// has terminated wih a non-zero exit code.
func IsContainerFailed(container *core.ContainerStatus) bool {
	if c := container.State.Terminated; c != nil {
		if c.ExitCode != 0 {
			return true
		}
	}

	return false
}

// IsPodScheduled returns true if the pod has been scheduled.
func IsPodScheduled(pod *core.Pod) bool {
	condition := getPodCondition(&pod.Status, core.PodScheduled)
	return condition != nil && condition.Status == core.ConditionTrue
}

// IsPodNotScheduledFor returns true if the pod has not been scheduled
// for longer than the given duration.
func IsPodNotScheduledFor(pod *core.Pod, timeout time.Duration) bool {
	condition := getPodCondition(&pod.Status, core.PodScheduled)
	return condition != nil &&
		condition.Status == core.ConditionFalse &&
		condition.LastTransitionTime.Time.Add(timeout).Before(time.Now())
}

// IsPodMarkedForDeletion returns true if the pod has been marked for deletion.
func IsPodMarkedForDeletion(pod *core.Pod) bool {
	return pod.DeletionTimestamp != nil
}

// IsPodTerminating returns true if the pod has been marked for deletion
// but is still running.
func IsPodTerminating(pod *core.Pod) bool {
	return IsPodMarkedForDeletion(pod) && pod.Status.Phase == core.PodRunning
}

// getPodCondition returns the condition of given type in the given status.
// If not found, nil is returned.
func getPodCondition(status *core.PodStatus, condType core.PodConditionType) *core.PodCondition {
	for i := range status.Conditions {
		if status.Conditions[i].Type == condType {
			return &status.Conditions[i]
		}
	}
	return nil
}

// CreatePodName returns the name of the pod for a member with
// a given id in a deployment with a given name.
func CreatePodName(deploymentName, role, id, suffix string) string {
	if len(suffix) > 0 && suffix[0] != '-' {
		suffix = "-" + suffix
	}
	return shared.CreatePodHostName(deploymentName, role, id) + suffix
}

// CreateTLSKeyfileSecretName returns the name of the Secret that holds the TLS keyfile for a member with
// a given id in a deployment with a given name.
func CreateTLSKeyfileSecretName(deploymentName, role, id string) string {
	return AppendTLSKeyfileSecretPostfix(CreatePodName(deploymentName, role, id, ""))
}

// AppendTLSKeyfileSecretPostfix returns the name of the Secret extended with TLS keyfile postfix.
func AppendTLSKeyfileSecretPostfix(name string) string {
	return fmt.Sprintf("%s-tls-keyfile", name)
}

// ArangodVolumeMount creates a volume mount structure for arangod.
func ArangodVolumeMount() core.VolumeMount {
	return core.VolumeMount{
		Name:      shared.ArangodVolumeName,
		MountPath: shared.ArangodVolumeMountDir,
	}
}

// TlsKeyfileVolumeMount creates a volume mount structure for a TLS keyfile.
func TlsKeyfileVolumeMount() core.VolumeMount {
	return core.VolumeMount{
		Name:      shared.TlsKeyfileVolumeName,
		MountPath: shared.TLSKeyfileVolumeMountDir,
		ReadOnly:  true,
	}
}

// ClientAuthCACertificateVolumeMount creates a volume mount structure for a client-auth CA certificate (ca.crt).
func ClientAuthCACertificateVolumeMount() core.VolumeMount {
	return core.VolumeMount{
		Name:      shared.ClientAuthCAVolumeName,
		MountPath: shared.ClientAuthCAVolumeMountDir,
	}
}

// MasterJWTVolumeMount creates a volume mount structure for a master JWT secret (token).
func MasterJWTVolumeMount() core.VolumeMount {
	return core.VolumeMount{
		Name:      shared.MasterJWTSecretVolumeName,
		MountPath: shared.MasterJWTSecretVolumeMountDir,
	}
}

// IsPodAlive returns true if any of the containers within pod is running
func IsPodAlive(pod *core.Pod) bool {
	return IsAnyContainerAlive(pod.Status.ContainerStatuses) ||
		IsAnyContainerAlive(pod.Status.InitContainerStatuses) ||
		IsAnyContainerAlive(pod.Status.EphemeralContainerStatuses)
}

// IsAnyContainerAlive returns true if any of the containers is running
func IsAnyContainerAlive(containers []core.ContainerStatus) bool {
	for _, c := range containers {
		if IsContainerAlive(c) {
			return true
		}
	}

	return false
}

// IsContainerAlive returns true if container is running
func IsContainerAlive(container core.ContainerStatus) bool {
	return container.State.Running != nil
}

// PodStopTime returns time when pod has been stopped
func PodStopTime(pod *core.Pod) time.Time {
	var t time.Time

	if q := ContainersRecentStopTime(pod.Status.ContainerStatuses); q.After(t) {
		t = q
	}

	if q := ContainersRecentStopTime(pod.Status.InitContainerStatuses); q.After(t) {
		t = q
	}

	if q := ContainersRecentStopTime(pod.Status.EphemeralContainerStatuses); q.After(t) {
		t = q
	}

	return t
}

// ContainersRecentStopTime returns most recent termination time of pods
func ContainersRecentStopTime(containers []core.ContainerStatus) time.Time {
	var t time.Time

	for _, c := range containers {
		if v := ContainerStopTime(c); v.After(t) {
			t = v
		}
	}

	return t
}

// ContainerStopTime returns time of the Container stop. If container is running, time.Zero is returned
func ContainerStopTime(container core.ContainerStatus) time.Time {
	if p := container.State.Terminated; p != nil {
		return p.FinishedAt.Time
	}

	return time.Time{}
}

// ClusterJWTVolumeMount creates a volume mount structure for a cluster JWT secret (token).
func ClusterJWTVolumeMount() core.VolumeMount {
	return core.VolumeMount{
		Name:      shared.ClusterJWTSecretVolumeName,
		MountPath: shared.ClusterJWTSecretVolumeMountDir,
	}
}

func ExporterJWTVolumeMount() core.VolumeMount {
	return core.VolumeMount{
		Name:      shared.ExporterJWTVolumeName,
		MountPath: shared.ExporterJWTVolumeMountDir,
		ReadOnly:  true,
	}
}

// RocksdbEncryptionVolumeMount creates a volume mount structure for a RocksDB encryption key.
func RocksdbEncryptionVolumeMount() core.VolumeMount {
	return core.VolumeMount{
		Name:      shared.RocksdbEncryptionVolumeName,
		MountPath: shared.RocksDBEncryptionVolumeMountDir,
	}
}

// RocksdbEncryptionReadOnlyVolumeMount creates a volume mount structure for a RocksDB encryption key.
func RocksdbEncryptionReadOnlyVolumeMount() core.VolumeMount {
	return core.VolumeMount{
		Name:      shared.RocksdbEncryptionVolumeName,
		MountPath: shared.RocksDBEncryptionVolumeMountDir,
		ReadOnly:  true,
	}
}

// ArangodInitContainer creates a container configured to initialize a UUID file.
func ArangodInitContainer(name, id, engine, executable, operatorImage string, requireUUID bool, securityContext *core.SecurityContext) core.Container {
	uuidFile := filepath.Join(shared.ArangodVolumeMountDir, "UUID")
	engineFile := filepath.Join(shared.ArangodVolumeMountDir, "ENGINE")
	var command = []string{
		executable,
		"uuid",
		"--uuid-path",
		uuidFile,
		"--engine-path",
		engineFile,
		"--uuid",
		id,
		"--engine",
		engine,
	}
	if requireUUID {
		command = append(command, "--require")
	}

	volumes := []core.VolumeMount{
		ArangodVolumeMount(),
	}
	return operatorInitContainer(name, operatorImage, command, securityContext, volumes)
}

// ArangodWaiterInitContainer creates a container configured to wait for specific ArangoDeployment to be ready
func ArangodWaiterInitContainer(name, deploymentName, executable, operatorImage string, isSecured bool, securityContext *core.SecurityContext) core.Container {
	var command = []string{
		executable,
		"lifecycle",
		"wait",
		"--deployment-name",
		deploymentName,
	}

	var volumes []core.VolumeMount
	if isSecured {
		volumes = append(volumes, TlsKeyfileVolumeMount())
	}
	return operatorInitContainer(name, operatorImage, command, securityContext, volumes)
}

// createInitContainer creates operator-specific init container
func operatorInitContainer(name, operatorImage string, command []string, securityContext *core.SecurityContext, volumes []core.VolumeMount) core.Container {
	c := core.Container{
		Name:    name,
		Image:   operatorImage,
		Command: command,
		Resources: core.ResourceRequirements{
			Requests: core.ResourceList{
				core.ResourceCPU:    resource.MustParse("100m"),
				core.ResourceMemory: resource.MustParse("10Mi"),
			},
			Limits: core.ResourceList{
				core.ResourceCPU:    resource.MustParse("100m"),
				core.ResourceMemory: resource.MustParse("50Mi"),
			},
		},
		Env: []core.EnvVar{
			{
				Name:  "MY_POD_NAMESPACE",
				Value: os.Getenv(constants.EnvOperatorPodNamespace),
			},
		},
		VolumeMounts:    volumes,
		SecurityContext: securityContext,
	}
	return c
}

// ExtractPodResourceRequirement filters resource requirements for Pods.
func ExtractPodResourceRequirement(resources core.ResourceRequirements) core.ResourceRequirements {

	filterStorage := func(list core.ResourceList) core.ResourceList {
		newlist := make(core.ResourceList)
		if q, ok := list[core.ResourceCPU]; ok {
			newlist[core.ResourceCPU] = q
		}
		if q, ok := list[core.ResourceMemory]; ok {
			newlist[core.ResourceMemory] = q
		}
		return newlist
	}

	return core.ResourceRequirements{
		Limits:   filterStorage(resources.Limits),
		Requests: filterStorage(resources.Requests),
	}
}

// NewContainer creates a container for specified creator
func NewContainer(containerCreator interfaces.ContainerCreator) (core.Container, error) {

	liveness, readiness, startup, err := containerCreator.GetProbes()
	if err != nil {
		return core.Container{}, err
	}

	lifecycle, err := containerCreator.GetLifecycle()
	if err != nil {
		return core.Container{}, err
	}

	args, err := containerCreator.GetArgs()
	if err != nil {
		return core.Container{}, err
	}

	env, envFrom := containerCreator.GetEnvs()
	return core.Container{
		Name:            containerCreator.GetName(),
		Image:           containerCreator.GetImage(),
		Command:         append([]string{containerCreator.GetExecutor()}, args...),
		Ports:           containerCreator.GetPorts(),
		Env:             env,
		EnvFrom:         envFrom,
		Resources:       containerCreator.GetResourceRequirements(),
		LivenessProbe:   liveness,
		ReadinessProbe:  readiness,
		StartupProbe:    startup,
		Lifecycle:       lifecycle,
		ImagePullPolicy: containerCreator.GetImagePullPolicy(),
		SecurityContext: containerCreator.GetSecurityContext(),
		VolumeMounts:    containerCreator.GetVolumeMounts(),
	}, nil
}

// NewPod creates a basic Pod for given settings.
func NewPod(deploymentName, role, id, podName string, podCreator interfaces.PodCreator) core.Pod {

	hostname := shared.CreatePodHostName(deploymentName, role, id)
	p := core.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name:       podName,
			Labels:     LabelsForMember(deploymentName, role, id),
			Finalizers: podCreator.GetFinalizers(),
		},
		Spec: core.PodSpec{
			Hostname:           hostname,
			Subdomain:          CreateHeadlessServiceName(deploymentName),
			RestartPolicy:      podCreator.GetRestartPolicy(),
			Tolerations:        podCreator.GetTolerations(),
			ServiceAccountName: podCreator.GetServiceAccountName(),
			NodeSelector:       podCreator.GetNodeSelector(),
		},
	}

	// Add ImagePullSecrets
	imagePullSecrets := podCreator.GetImagePullSecrets()
	if imagePullSecrets != nil {
		imagePullSecretsReference := make([]core.LocalObjectReference, len(imagePullSecrets))
		for id := range imagePullSecrets {
			imagePullSecretsReference[id] = core.LocalObjectReference{
				Name: imagePullSecrets[id],
			}
		}
		p.Spec.ImagePullSecrets = imagePullSecretsReference
	}

	return p
}

// GetPodSpecChecksum return checksum of requested pod spec based on deployment and group spec
func GetPodSpecChecksum(podSpec core.PodSpec) (string, error) {
	data, err := json.Marshal(podSpec)
	if err != nil {
		return "", err
	}

	return util.SHA256(data), nil
}

// CreatePod adds an owner to the given pod and calls the k8s api-server to created it.
// If the pod already exists, nil is returned.
// If another error occurs, that error is returned.
func CreatePod(ctx context.Context, c podv1.ModInterface, pod *core.Pod, ns string,
	owner meta.OwnerReference) (string, types.UID, error) {
	AddOwnerRefToObject(pod.GetObjectMeta(), &owner)

	if createdPod, err := c.Create(ctx, pod, meta.CreateOptions{}); err != nil {
		if kerrors.IsAlreadyExists(err) {
			return pod.GetName(), "", nil // If pod exists do not return any error but do not record UID (enforced rotation)
		}

		return "", "", errors.WithStack(err)
	} else {
		return createdPod.GetName(), createdPod.GetUID(), nil
	}
}

func CreateVolumeEmptyDir(name string) core.Volume {
	return core.Volume{
		Name: name,
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	}
}

func CreateVolumeWithSecret(name, secretName string) core.Volume {
	return core.Volume{
		Name: name,
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	}
}

func CreateVolumeWithPersitantVolumeClaim(name, claimName string) core.Volume {
	return core.Volume{
		Name: name,
		VolumeSource: core.VolumeSource{
			PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
				ClaimName: claimName,
			},
		},
	}
}

func CreateEnvFieldPath(name, fieldPath string) core.EnvVar {
	return core.EnvVar{
		Name: name,
		ValueFrom: &core.EnvVarSource{
			FieldRef: &core.ObjectFieldSelector{
				FieldPath: fieldPath,
			},
		},
	}
}

func CreateEnvSecretKeySelector(name, SecretKeyName, secretKey string) core.EnvVar {
	return core.EnvVar{
		Name:  name,
		Value: "",
		ValueFrom: &core.EnvVarSource{
			SecretKeyRef: &core.SecretKeySelector{
				LocalObjectReference: core.LocalObjectReference{
					Name: SecretKeyName,
				},
				Key: secretKey,
			},
		},
	}
}

func EnsureFinalizerAbsent(ctx context.Context, pods podv1.Interface, pod *core.Pod, finalizers ...string) error {
	var newFinalizers []string

	c := utils.StringList(finalizers)

	for _, fn := range pod.Finalizers {
		if !c.Has(fn) {
			newFinalizers = append(newFinalizers, fn)
		}
	}

	if len(newFinalizers) == len(pod.Finalizers) {
		return nil
	}

	return SetFinalizers(ctx, pods, pod, newFinalizers...)
}

func EnsureFinalizerPresent(ctx context.Context, pods podv1.Interface, pod *core.Pod, finalizers ...string) error {
	var newFinalizers []string

	newFinalizers = append(newFinalizers, pod.Finalizers...)

	for _, fn := range finalizers {
		if utils.StringList(newFinalizers).Has(fn) {
			continue
		}

		newFinalizers = append(newFinalizers, fn)
	}

	if len(newFinalizers) == len(pod.Finalizers) {
		return nil
	}

	return SetFinalizers(ctx, pods, pod, newFinalizers...)
}

func SetFinalizers(ctx context.Context, pods podv1.Interface, pod *core.Pod, finalizers ...string) error {
	d, err := patch.NewPatch(patch.ItemReplace(patch.NewPath("metadata", "finalizers"), finalizers)).Marshal()
	if err != nil {
		return err
	}

	nctx, c := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer c()

	if _, err := pods.Patch(nctx, pod.GetName(), types.JSONPatchType, d, meta.PatchOptions{}); err != nil {
		return err
	}

	return nil
}

func GetFinalizers(spec api.ServerGroupSpec, group api.ServerGroup) []string {
	var finalizers []string

	if d := spec.GetShutdownDelay(group); d != 0 {
		finalizers = append(finalizers, constants.FinalizerDelayPodTermination)
	}

	if features.GracefulShutdown().Enabled() {
		finalizers = append(finalizers, constants.FinalizerPodGracefulShutdown) // No need for other finalizers, quorum will be managed
	} else {
		switch group {
		case api.ServerGroupAgents:
			finalizers = append(finalizers, constants.FinalizerPodAgencyServing)
		case api.ServerGroupDBServers:
			finalizers = append(finalizers, constants.FinalizerPodDrainDBServer)
		}
	}

	return finalizers
}
