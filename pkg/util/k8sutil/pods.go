//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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

package k8sutil

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	InitDataContainerName           = "init-data"
	InitLifecycleContainerName      = "init-lifecycle"
	ServerContainerName             = "server"
	alpineImage                     = "alpine"
	arangodVolumeName               = "arangod-data"
	tlsKeyfileVolumeName            = "tls-keyfile"
	lifecycleVolumeName             = "lifecycle"
	rocksdbEncryptionVolumeName     = "rocksdb-encryption"
	ArangodVolumeMountDir           = "/data"
	RocksDBEncryptionVolumeMountDir = "/secrets/rocksdb/encryption"
	TLSKeyfileVolumeMountDir        = "/secrets/tls"
	LifecycleVolumeMountDir         = "/lifecycle/tools"
)

// EnvValue is a helper structure for environment variable sources.
type EnvValue struct {
	Value      string // If set, the environment value gets this value
	SecretName string // If set, the environment value gets its value from a secret with this name
	SecretKey  string // Key inside secret to fill into the envvar. Only relevant is SecretName is set.
}

// CreateEnvVar creates an EnvVar structure for given key from given EnvValue.
func (v EnvValue) CreateEnvVar(key string) v1.EnvVar {
	ev := v1.EnvVar{
		Name: key,
	}
	if ev.Value != "" {
		ev.Value = v.Value
	} else if v.SecretName != "" {
		ev.ValueFrom = &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: v.SecretName,
				},
				Key: v.SecretKey,
			},
		}
	}
	return ev
}

// IsPodReady returns true if the PodReady condition on
// the given pod is set to true.
func IsPodReady(pod *v1.Pod) bool {
	condition := getPodCondition(&pod.Status, v1.PodReady)
	return condition != nil && condition.Status == v1.ConditionTrue
}

// IsPodSucceeded returns true if all containers of the pod
// have terminated with exit code 0.
func IsPodSucceeded(pod *v1.Pod) bool {
	return pod.Status.Phase == v1.PodSucceeded
}

// IsPodFailed returns true if all containers of the pod
// have terminated and at least one of them wih a non-zero exit code.
func IsPodFailed(pod *v1.Pod) bool {
	return pod.Status.Phase == v1.PodFailed
}

// IsPodScheduled returns true if the pod has been scheduled.
func IsPodScheduled(pod *v1.Pod) bool {
	condition := getPodCondition(&pod.Status, v1.PodScheduled)
	return condition != nil && condition.Status == v1.ConditionTrue
}

// IsPodNotScheduledFor returns true if the pod has not been scheduled
// for longer than the given duration.
func IsPodNotScheduledFor(pod *v1.Pod, timeout time.Duration) bool {
	condition := getPodCondition(&pod.Status, v1.PodScheduled)
	return condition != nil &&
		condition.Status == v1.ConditionFalse &&
		condition.LastTransitionTime.Time.Add(timeout).Before(time.Now())
}

// IsPodMarkedForDeletion returns true if the pod has been marked for deletion.
func IsPodMarkedForDeletion(pod *v1.Pod) bool {
	return pod.DeletionTimestamp != nil
}

// IsArangoDBImageIDAndVersionPod returns true if the given pod is used for fetching image ID and ArangoDB version of an image
func IsArangoDBImageIDAndVersionPod(p v1.Pod) bool {
	role, found := p.GetLabels()[LabelKeyRole]
	return found && role == ImageIDAndVersionRole
}

// getPodCondition returns the condition of given type in the given status.
// If not found, nil is returned.
func getPodCondition(status *v1.PodStatus, condType v1.PodConditionType) *v1.PodCondition {
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
	return CreatePodHostName(deploymentName, role, id) + suffix
}

// CreatePodHostName returns the hostname of the pod for a member with
// a given id in a deployment with a given name.
func CreatePodHostName(deploymentName, role, id string) string {
	return deploymentName + "-" + role + "-" + stripArangodPrefix(id)
}

// CreateTLSKeyfileSecretName returns the name of the Secret that holds the TLS keyfile for a member with
// a given id in a deployment with a given name.
func CreateTLSKeyfileSecretName(deploymentName, role, id string) string {
	return CreatePodName(deploymentName, role, id, "-tls-keyfile")
}

// lifecycleVolumeMounts creates a volume mount structure for shared lifecycle emptyDir.
func lifecycleVolumeMounts() []v1.VolumeMount {
	return []v1.VolumeMount{
		{Name: lifecycleVolumeName, MountPath: LifecycleVolumeMountDir},
	}
}

// arangodVolumeMounts creates a volume mount structure for arangod.
func arangodVolumeMounts() []v1.VolumeMount {
	return []v1.VolumeMount{
		{Name: arangodVolumeName, MountPath: ArangodVolumeMountDir},
	}
}

// tlsKeyfileVolumeMounts creates a volume mount structure for a TLS keyfile.
func tlsKeyfileVolumeMounts() []v1.VolumeMount {
	return []v1.VolumeMount{
		{
			Name:      tlsKeyfileVolumeName,
			MountPath: TLSKeyfileVolumeMountDir,
		},
	}
}

// rocksdbEncryptionVolumeMounts creates a volume mount structure for a RocksDB encryption key.
func rocksdbEncryptionVolumeMounts() []v1.VolumeMount {
	return []v1.VolumeMount{
		{
			Name:      rocksdbEncryptionVolumeName,
			MountPath: RocksDBEncryptionVolumeMountDir,
		},
	}
}

// arangodInitContainer creates a container configured to
// initalize a UUID file.
func arangodInitContainer(name, id, engine string, requireUUID bool) v1.Container {
	uuidFile := filepath.Join(ArangodVolumeMountDir, "UUID")
	engineFile := filepath.Join(ArangodVolumeMountDir, "ENGINE")
	var command string
	if requireUUID {
		command = strings.Join([]string{
			// Files must exist
			fmt.Sprintf("test -f %s", uuidFile),
			fmt.Sprintf("test -f %s", engineFile),
			// Content must match
			fmt.Sprintf("grep -q %s %s", id, uuidFile),
			fmt.Sprintf("grep -q %s %s", engine, engineFile),
		}, " && ")

	} else {
		command = fmt.Sprintf("test -f %s || echo '%s' > %s", uuidFile, id, uuidFile)
	}
	c := v1.Container{
		Command: []string{
			"/bin/sh",
			"-c",
			command,
		},
		Name:         name,
		Image:        alpineImage,
		VolumeMounts: arangodVolumeMounts(),
	}
	return c
}

// arangodContainer creates a container configured to run `arangod`.
func arangodContainer(image string, imagePullPolicy v1.PullPolicy, args []string, env map[string]EnvValue, livenessProbe *HTTPProbeConfig, readinessProbe *HTTPProbeConfig,
	lifecycle *v1.Lifecycle, lifecycleEnvVars []v1.EnvVar) v1.Container {
	c := v1.Container{
		Command:         append([]string{"/usr/sbin/arangod"}, args...),
		Name:            ServerContainerName,
		Image:           image,
		ImagePullPolicy: imagePullPolicy,
		Lifecycle:       lifecycle,
		Ports: []v1.ContainerPort{
			{
				Name:          "server",
				ContainerPort: int32(ArangoPort),
				Protocol:      v1.ProtocolTCP,
			},
		},
		VolumeMounts: arangodVolumeMounts(),
	}
	for k, v := range env {
		c.Env = append(c.Env, v.CreateEnvVar(k))
	}
	if livenessProbe != nil {
		c.LivenessProbe = livenessProbe.Create()
	}
	if readinessProbe != nil {
		c.ReadinessProbe = readinessProbe.Create()
	}
	if lifecycle != nil {
		c.Env = append(c.Env, lifecycleEnvVars...)
		c.VolumeMounts = append(c.VolumeMounts, lifecycleVolumeMounts()...)
	}

	return c
}

// arangosyncContainer creates a container configured to run `arangosync`.
func arangosyncContainer(image string, imagePullPolicy v1.PullPolicy, args []string, env map[string]EnvValue, livenessProbe *HTTPProbeConfig,
	lifecycle *v1.Lifecycle, lifecycleEnvVars []v1.EnvVar) v1.Container {
	c := v1.Container{
		Command:         append([]string{"/usr/sbin/arangosync"}, args...),
		Name:            ServerContainerName,
		Image:           image,
		ImagePullPolicy: imagePullPolicy,
		Lifecycle:       lifecycle,
		Ports: []v1.ContainerPort{
			{
				Name:          "server",
				ContainerPort: int32(ArangoPort),
				Protocol:      v1.ProtocolTCP,
			},
		},
	}
	for k, v := range env {
		c.Env = append(c.Env, v.CreateEnvVar(k))
	}
	if livenessProbe != nil {
		c.LivenessProbe = livenessProbe.Create()
	}
	if lifecycle != nil {
		c.Env = append(c.Env, lifecycleEnvVars...)
		c.VolumeMounts = append(c.VolumeMounts, lifecycleVolumeMounts()...)
	}

	return c
}

// newLifecycle creates a lifecycle structure with preStop handler.
func newLifecycle() (*v1.Lifecycle, []v1.EnvVar, []v1.Volume, error) {
	binaryPath, err := os.Executable()
	if err != nil {
		return nil, nil, nil, maskAny(err)
	}
	exePath := filepath.Join(LifecycleVolumeMountDir, filepath.Base(binaryPath))
	lifecycle := &v1.Lifecycle{
		PreStop: &v1.Handler{
			Exec: &v1.ExecAction{
				Command: append([]string{exePath}, "lifecycle", "preStop"),
			},
		},
	}
	envVars := []v1.EnvVar{
		v1.EnvVar{
			Name: constants.EnvOperatorPodName,
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		v1.EnvVar{
			Name: constants.EnvOperatorPodNamespace,
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
	}
	vols := []v1.Volume{
		v1.Volume{
			Name: lifecycleVolumeName,
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
	}
	return lifecycle, envVars, vols, nil
}

// initLifecycleContainer creates an init-container to copy the lifecycle binary
// to a shared volume.
func initLifecycleContainer(image string) (v1.Container, error) {
	binaryPath, err := os.Executable()
	if err != nil {
		return v1.Container{}, maskAny(err)
	}
	c := v1.Container{
		Command:         append([]string{binaryPath}, "lifecycle", "copy", "--target", LifecycleVolumeMountDir),
		Name:            InitLifecycleContainerName,
		Image:           image,
		ImagePullPolicy: v1.PullIfNotPresent,
		VolumeMounts:    lifecycleVolumeMounts(),
	}
	return c, nil
}

// newPod creates a basic Pod for given settings.
func newPod(deploymentName, ns, role, id, podName string, finalizers []string, tolerations []v1.Toleration) v1.Pod {
	hostname := CreatePodHostName(deploymentName, role, id)
	p := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:       podName,
			Labels:     LabelsForDeployment(deploymentName, role),
			Finalizers: finalizers,
		},
		Spec: v1.PodSpec{
			Hostname:      hostname,
			Subdomain:     CreateHeadlessServiceName(deploymentName),
			RestartPolicy: v1.RestartPolicyNever,
			Tolerations:   tolerations,
		},
	}
	return p
}

// CreateArangodPod creates a Pod that runs `arangod`.
// If the pod already exists, nil is returned.
// If another error occurs, that error is returned.
func CreateArangodPod(kubecli kubernetes.Interface, developmentMode bool, deployment APIObject,
	role, id, podName, pvcName, image, lifecycleImage string, imagePullPolicy v1.PullPolicy,
	engine string, requireUUID bool, terminationGracePeriod time.Duration,
	args []string, env map[string]EnvValue, finalizers []string,
	livenessProbe *HTTPProbeConfig, readinessProbe *HTTPProbeConfig, tolerations []v1.Toleration,
	tlsKeyfileSecretName, rocksdbEncryptionSecretName string) error {
	// Prepare basic pod
	p := newPod(deployment.GetName(), deployment.GetNamespace(), role, id, podName, finalizers, tolerations)
	terminationGracePeriodSeconds := int64(math.Ceil(terminationGracePeriod.Seconds()))
	p.Spec.TerminationGracePeriodSeconds = &terminationGracePeriodSeconds

	// Add lifecycle container
	var lifecycle *v1.Lifecycle
	var lifecycleEnvVars []v1.EnvVar
	var lifecycleVolumes []v1.Volume
	if lifecycleImage != "" {
		c, err := initLifecycleContainer(lifecycleImage)
		if err != nil {
			return maskAny(err)
		}
		p.Spec.InitContainers = append(p.Spec.InitContainers, c)
		lifecycle, lifecycleEnvVars, lifecycleVolumes, err = newLifecycle()
		if err != nil {
			return maskAny(err)
		}
	}

	// Add arangod container
	c := arangodContainer(image, imagePullPolicy, args, env, livenessProbe, readinessProbe, lifecycle, lifecycleEnvVars)
	if tlsKeyfileSecretName != "" {
		c.VolumeMounts = append(c.VolumeMounts, tlsKeyfileVolumeMounts()...)
	}
	if rocksdbEncryptionSecretName != "" {
		c.VolumeMounts = append(c.VolumeMounts, rocksdbEncryptionVolumeMounts()...)
	}
	p.Spec.Containers = append(p.Spec.Containers, c)

	// Add UUID init container
	p.Spec.InitContainers = append(p.Spec.InitContainers, arangodInitContainer("uuid", id, engine, requireUUID))

	// Add volume
	if pvcName != "" {
		// Create PVC
		vol := v1.Volume{
			Name: arangodVolumeName,
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvcName,
				},
			},
		}
		p.Spec.Volumes = append(p.Spec.Volumes, vol)
	} else {
		// Create emptydir volume
		vol := v1.Volume{
			Name: arangodVolumeName,
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		}
		p.Spec.Volumes = append(p.Spec.Volumes, vol)
	}

	// TLS keyfile secret mount (if any)
	if tlsKeyfileSecretName != "" {
		vol := v1.Volume{
			Name: tlsKeyfileVolumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: tlsKeyfileSecretName,
				},
			},
		}
		p.Spec.Volumes = append(p.Spec.Volumes, vol)
	}

	// RocksDB encryption secret mount (if any)
	if rocksdbEncryptionSecretName != "" {
		vol := v1.Volume{
			Name: rocksdbEncryptionVolumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: rocksdbEncryptionSecretName,
				},
			},
		}
		p.Spec.Volumes = append(p.Spec.Volumes, vol)
	}

	// Lifecycle volumes (if any)
	p.Spec.Volumes = append(p.Spec.Volumes, lifecycleVolumes...)

	// Add (anti-)affinity
	p.Spec.Affinity = createAffinity(deployment.GetName(), role, !developmentMode, "")

	if err := createPod(kubecli, &p, deployment.GetNamespace(), deployment.AsOwner()); err != nil {
		return maskAny(err)
	}
	return nil
}

// CreateArangoSyncPod creates a Pod that runs `arangosync`.
// If the pod already exists, nil is returned.
// If another error occurs, that error is returned.
func CreateArangoSyncPod(kubecli kubernetes.Interface, developmentMode bool, deployment APIObject, role, id, podName, image, lifecycleImage string, imagePullPolicy v1.PullPolicy,
	terminationGracePeriod time.Duration, args []string, env map[string]EnvValue, livenessProbe *HTTPProbeConfig, tolerations []v1.Toleration, affinityWithRole string) error {
	// Prepare basic pod
	p := newPod(deployment.GetName(), deployment.GetNamespace(), role, id, podName, nil, tolerations)
	terminationGracePeriodSeconds := int64(math.Ceil(terminationGracePeriod.Seconds()))
	p.Spec.TerminationGracePeriodSeconds = &terminationGracePeriodSeconds

	// Add lifecycle container
	var lifecycle *v1.Lifecycle
	var lifecycleEnvVars []v1.EnvVar
	var lifecycleVolumes []v1.Volume
	if lifecycleImage != "" {
		c, err := initLifecycleContainer(lifecycleImage)
		if err != nil {
			return maskAny(err)
		}
		p.Spec.InitContainers = append(p.Spec.InitContainers, c)
		lifecycle, lifecycleEnvVars, lifecycleVolumes, err = newLifecycle()
		if err != nil {
			return maskAny(err)
		}
	}

	// Add arangosync container
	c := arangosyncContainer(image, imagePullPolicy, args, env, livenessProbe, lifecycle, lifecycleEnvVars)
	p.Spec.Containers = append(p.Spec.Containers, c)

	// Lifecycle volumes (if any)
	p.Spec.Volumes = append(p.Spec.Volumes, lifecycleVolumes...)

	// Add (anti-)affinity
	p.Spec.Affinity = createAffinity(deployment.GetName(), role, !developmentMode, affinityWithRole)

	if err := createPod(kubecli, &p, deployment.GetNamespace(), deployment.AsOwner()); err != nil {
		return maskAny(err)
	}
	return nil
}

// createPod adds an owner to the given pod and calls the k8s api-server to created it.
// If the pod already exists, nil is returned.
// If another error occurs, that error is returned.
func createPod(kubecli kubernetes.Interface, pod *v1.Pod, ns string, owner metav1.OwnerReference) error {
	addOwnerRefToObject(pod.GetObjectMeta(), &owner)
	if _, err := kubecli.CoreV1().Pods(ns).Create(pod); err != nil && !IsAlreadyExists(err) {
		return maskAny(err)
	}
	return nil
}
