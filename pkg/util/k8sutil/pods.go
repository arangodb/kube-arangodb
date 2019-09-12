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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	InitDataContainerName           = "init-data"
	InitLifecycleContainerName      = "init-lifecycle"
	ServerContainerName             = "server"
	ExporterContainerName           = "exporter"
	arangodVolumeName               = "arangod-data"
	tlsKeyfileVolumeName            = "tls-keyfile"
	lifecycleVolumeName             = "lifecycle"
	clientAuthCAVolumeName          = "client-auth-ca"
	clusterJWTSecretVolumeName      = "cluster-jwt"
	masterJWTSecretVolumeName       = "master-jwt"
	rocksdbEncryptionVolumeName     = "rocksdb-encryption"
	exporterJWTVolumeName           = "exporter-jwt"
	ArangodVolumeMountDir           = "/data"
	RocksDBEncryptionVolumeMountDir = "/secrets/rocksdb/encryption"
	JWTSecretFileVolumeMountDir     = "/secrets/jwt"
	TLSKeyfileVolumeMountDir        = "/secrets/tls"
	LifecycleVolumeMountDir         = "/lifecycle/tools"
	ClientAuthCAVolumeMountDir      = "/secrets/client-auth/ca"
	ClusterJWTSecretVolumeMountDir  = "/secrets/cluster/jwt"
	ExporterJWTVolumeMountDir       = "/secrets/exporter/jwt"
	MasterJWTSecretVolumeMountDir   = "/secrets/master/jwt"
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

// IsPodSucceeded returns true if the arangodb container of the pod
// has terminated with exit code 0.
func IsPodSucceeded(pod *v1.Pod) bool {
	if pod.Status.Phase == v1.PodSucceeded {
		return true
	} else {
		for _, c := range pod.Status.ContainerStatuses {
			if c.Name != ServerContainerName {
				continue
			}

			t := c.State.Terminated
			if t != nil {
				return t.ExitCode == 0
			}
		}
		return false
	}
}

// IsPodFailed returns true if the arangodb container of the pod
// has terminated wih a non-zero exit code.
func IsPodFailed(pod *v1.Pod) bool {
	if pod.Status.Phase == v1.PodFailed {
		return true
	} else {
		for _, c := range pod.Status.ContainerStatuses {
			if c.Name != ServerContainerName {
				continue
			}

			t := c.State.Terminated
			if t != nil {
				return t.ExitCode != 0
			}
		}

		return false
	}
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

// IsPodTerminating returns true if the pod has been marked for deletion
// but is still running.
func IsPodTerminating(pod *v1.Pod) bool {
	return IsPodMarkedForDeletion(pod) && pod.Status.Phase == v1.PodRunning
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

// clientAuthCACertificateVolumeMounts creates a volume mount structure for a client-auth CA certificate (ca.crt).
func clientAuthCACertificateVolumeMounts() []v1.VolumeMount {
	return []v1.VolumeMount{
		{
			Name:      clientAuthCAVolumeName,
			MountPath: ClientAuthCAVolumeMountDir,
		},
	}
}

// masterJWTVolumeMounts creates a volume mount structure for a master JWT secret (token).
func masterJWTVolumeMounts() []v1.VolumeMount {
	return []v1.VolumeMount{
		{
			Name:      masterJWTSecretVolumeName,
			MountPath: MasterJWTSecretVolumeMountDir,
		},
	}
}

// clusterJWTVolumeMounts creates a volume mount structure for a cluster JWT secret (token).
func clusterJWTVolumeMounts() []v1.VolumeMount {
	return []v1.VolumeMount{
		{
			Name:      clusterJWTSecretVolumeName,
			MountPath: ClusterJWTSecretVolumeMountDir,
		},
	}
}

func exporterJWTVolumeMounts() []v1.VolumeMount {
	return []v1.VolumeMount{
		{
			Name:      exporterJWTVolumeName,
			MountPath: ExporterJWTVolumeMountDir,
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
func arangodInitContainer(name, id, engine, alpineImage string, requireUUID bool) v1.Container {
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

// ExtractPodResourceRequirement filters resource requirements for Pods.
func ExtractPodResourceRequirement(resources v1.ResourceRequirements) v1.ResourceRequirements {

	filterStorage := func(list v1.ResourceList) v1.ResourceList {
		newlist := make(v1.ResourceList)
		for k, v := range list {
			if k != v1.ResourceCPU && k != v1.ResourceMemory {
				continue
			}
			newlist[k] = v
		}
		return newlist
	}

	return v1.ResourceRequirements{
		Limits:   filterStorage(resources.Limits),
		Requests: filterStorage(resources.Requests),
	}
}

// arangodContainer creates a container configured to run `arangod`.
func arangodContainer(image string, imagePullPolicy v1.PullPolicy, args []string, env map[string]EnvValue, livenessProbe *HTTPProbeConfig, readinessProbe *HTTPProbeConfig,
	lifecycle *v1.Lifecycle, lifecycleEnvVars []v1.EnvVar, resources v1.ResourceRequirements, noFilterResources bool) v1.Container {
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
	if noFilterResources {
		c.Resources = resources // if volumeclaimtemplate is specified
	} else {
		c.Resources = ExtractPodResourceRequirement(resources) // Storage is handled via pvcs
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
	lifecycle *v1.Lifecycle, lifecycleEnvVars []v1.EnvVar, resources v1.ResourceRequirements) v1.Container {
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
		Resources: resources,
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

func arangodbexporterContainer(image string, imagePullPolicy v1.PullPolicy, args []string, env map[string]EnvValue, livenessProbe *HTTPProbeConfig) v1.Container {
	c := v1.Container{
		Command:         append([]string{"/app/arangodb-exporter"}, args...),
		Name:            ExporterContainerName,
		Image:           image,
		ImagePullPolicy: v1.PullIfNotPresent,
		Ports: []v1.ContainerPort{
			{
				Name:          "exporter",
				ContainerPort: int32(ArangoExporterPort),
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
		v1.EnvVar{
			Name: constants.EnvOperatorNodeName,
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "spec.nodeName",
				},
			},
		},
		v1.EnvVar{
			Name: constants.EnvOperatorNodeNameArango,
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "spec.nodeName",
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
func newPod(deploymentName, ns, role, id, podName string, imagePullSecrets []string, finalizers []string, tolerations []v1.Toleration, serviceAccountName string, nodeSelector map[string]string) v1.Pod {
	hostname := CreatePodHostName(deploymentName, role, id)
	p := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:       podName,
			Labels:     LabelsForDeployment(deploymentName, role),
			Finalizers: finalizers,
		},
		Spec: v1.PodSpec{
			Hostname:           hostname,
			Subdomain:          CreateHeadlessServiceName(deploymentName),
			RestartPolicy:      v1.RestartPolicyNever,
			Tolerations:        tolerations,
			ServiceAccountName: serviceAccountName,
			NodeSelector:       nodeSelector,
		},
	}

	// Add ImagePullSecrets
	if imagePullSecrets != nil {
		imagePullSecretsReference := make([]v1.LocalObjectReference, len(imagePullSecrets))
		for id := range imagePullSecrets {
			imagePullSecretsReference[id] = v1.LocalObjectReference{
				Name: imagePullSecrets[id],
			}
		}
		p.Spec.ImagePullSecrets = imagePullSecretsReference
	}

	return p
}

// ArangodbExporterContainerConf contains configuration of the exporter container
type ArangodbExporterContainerConf struct {
	Args               []string
	Env                map[string]EnvValue
	JWTTokenSecretName string
	LivenessProbe      *HTTPProbeConfig
	Image              string
}

// CreateArangodPod creates a Pod that runs `arangod`.
// If the pod already exists, nil is returned.
// If another error occurs, that error is returned.
func CreateArangodPod(kubecli kubernetes.Interface, developmentMode bool, deployment APIObject,
	role, id, podName, pvcName, image, lifecycleImage, alpineImage string,
	imagePullPolicy v1.PullPolicy, imagePullSecrets []string,
	engine string, requireUUID bool, terminationGracePeriod time.Duration,
	args []string, env map[string]EnvValue, finalizers []string,
	livenessProbe *HTTPProbeConfig, readinessProbe *HTTPProbeConfig, tolerations []v1.Toleration, serviceAccountName string,
	tlsKeyfileSecretName, rocksdbEncryptionSecretName string, clusterJWTSecretName string, nodeSelector map[string]string,
	podPriorityClassName string, resources v1.ResourceRequirements, exporter *ArangodbExporterContainerConf, sidecars []v1.Container, vct *v1.PersistentVolumeClaim) error {

	// Prepare basic pod
	p := newPod(deployment.GetName(), deployment.GetNamespace(), role, id, podName, imagePullSecrets, finalizers, tolerations, serviceAccountName, nodeSelector)
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
	c := arangodContainer(image, imagePullPolicy, args, env, livenessProbe, readinessProbe, lifecycle, lifecycleEnvVars, resources, vct != nil)
	if tlsKeyfileSecretName != "" {
		c.VolumeMounts = append(c.VolumeMounts, tlsKeyfileVolumeMounts()...)
	}
	if rocksdbEncryptionSecretName != "" {
		c.VolumeMounts = append(c.VolumeMounts, rocksdbEncryptionVolumeMounts()...)
	}
	if clusterJWTSecretName != "" {
		c.VolumeMounts = append(c.VolumeMounts, clusterJWTVolumeMounts()...)
	}

	p.Spec.Containers = append(p.Spec.Containers, c)

	// Add arangodb exporter container
	if exporter != nil {
		c = arangodbexporterContainer(exporter.Image, imagePullPolicy, exporter.Args, exporter.Env, exporter.LivenessProbe)
		if exporter.JWTTokenSecretName != "" {
			c.VolumeMounts = append(c.VolumeMounts, exporterJWTVolumeMounts()...)
		}
		if tlsKeyfileSecretName != "" {
			c.VolumeMounts = append(c.VolumeMounts, tlsKeyfileVolumeMounts()...)
		}
		p.Spec.Containers = append(p.Spec.Containers, c)
		p.Labels[LabelKeyArangoExporter] = "yes"
	}

	// Add sidecars
	if len(sidecars) > 0 {
		p.Spec.Containers = append(p.Spec.Containers, sidecars...)
	}

	// Add priorityClassName
	p.Spec.PriorityClassName = podPriorityClassName

	// Add UUID init container
	if alpineImage != "" {
		p.Spec.InitContainers = append(p.Spec.InitContainers, arangodInitContainer("uuid", id, engine, alpineImage, requireUUID))
	}

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

	// Exporter Token Mount
	if exporter != nil && exporter.JWTTokenSecretName != "" {
		vol := v1.Volume{
			Name: exporterJWTVolumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: exporter.JWTTokenSecretName,
				},
			},
		}
		p.Spec.Volumes = append(p.Spec.Volumes, vol)
	}

	// Cluster JWT secret mount (if any)
	if clusterJWTSecretName != "" {
		vol := v1.Volume{
			Name: clusterJWTSecretVolumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: clusterJWTSecretName,
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
func CreateArangoSyncPod(kubecli kubernetes.Interface, developmentMode bool, deployment APIObject, role, id, podName, image, lifecycleImage string,
	imagePullPolicy v1.PullPolicy, imagePullSecrets []string,
	terminationGracePeriod time.Duration, args []string, env map[string]EnvValue, livenessProbe *HTTPProbeConfig, tolerations []v1.Toleration, serviceAccountName string,
	tlsKeyfileSecretName, clientAuthCASecretName, masterJWTSecretName, clusterJWTSecretName, affinityWithRole string, nodeSelector map[string]string,
	podPriorityClassName string, resources v1.ResourceRequirements, sidecars []v1.Container) error {
	// Prepare basic pod
	p := newPod(deployment.GetName(), deployment.GetNamespace(), role, id, podName, imagePullSecrets, nil, tolerations, serviceAccountName, nodeSelector)
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

	// Lifecycle volumes (if any)
	p.Spec.Volumes = append(p.Spec.Volumes, lifecycleVolumes...)

	// Add arangosync container
	c := arangosyncContainer(image, imagePullPolicy, args, env, livenessProbe, lifecycle, lifecycleEnvVars, resources)
	if tlsKeyfileSecretName != "" {
		c.VolumeMounts = append(c.VolumeMounts, tlsKeyfileVolumeMounts()...)
	}
	if clientAuthCASecretName != "" {
		c.VolumeMounts = append(c.VolumeMounts, clientAuthCACertificateVolumeMounts()...)
	}
	if masterJWTSecretName != "" {
		c.VolumeMounts = append(c.VolumeMounts, masterJWTVolumeMounts()...)
	}
	if clusterJWTSecretName != "" {
		c.VolumeMounts = append(c.VolumeMounts, clusterJWTVolumeMounts()...)
	}
	p.Spec.Containers = append(p.Spec.Containers, c)

	// Add sidecars
	if len(sidecars) > 0 {
		p.Spec.Containers = append(p.Spec.Containers, sidecars...)
	}

	// Add priorityClassName
	p.Spec.PriorityClassName = podPriorityClassName

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

	// Client Authentication certificate secret mount (if any)
	if clientAuthCASecretName != "" {
		vol := v1.Volume{
			Name: clientAuthCAVolumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: clientAuthCASecretName,
				},
			},
		}
		p.Spec.Volumes = append(p.Spec.Volumes, vol)
	}

	// Master JWT secret mount (if any)
	if masterJWTSecretName != "" {
		vol := v1.Volume{
			Name: masterJWTSecretVolumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: masterJWTSecretName,
				},
			},
		}
		p.Spec.Volumes = append(p.Spec.Volumes, vol)
	}

	// Cluster JWT secret mount (if any)
	if clusterJWTSecretName != "" {
		vol := v1.Volume{
			Name: clusterJWTSecretVolumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: clusterJWTSecretName,
				},
			},
		}
		p.Spec.Volumes = append(p.Spec.Volumes, vol)
	}

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
