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
	"path/filepath"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	ServerContainerName             = "server"
	ExporterContainerName           = "exporter"
	ArangodVolumeName               = "arangod-data"
	TlsKeyfileVolumeName            = "tls-keyfile"
	ClientAuthCAVolumeName          = "client-auth-ca"
	ClusterJWTSecretVolumeName      = "cluster-jwt"
	MasterJWTSecretVolumeName       = "master-jwt"
	RocksdbEncryptionVolumeName     = "rocksdb-encryption"
	ExporterJWTVolumeName           = "exporter-jwt"
	ArangodVolumeMountDir           = "/data"
	RocksDBEncryptionVolumeMountDir = "/secrets/rocksdb/encryption"
	TLSKeyfileVolumeMountDir        = "/secrets/tls"
	ClientAuthCAVolumeMountDir      = "/secrets/client-auth/ca"
	ClusterJWTSecretVolumeMountDir  = "/secrets/cluster/jwt"
	ExporterJWTVolumeMountDir       = "/secrets/exporter/jwt"
	MasterJWTSecretVolumeMountDir   = "/secrets/master/jwt"
)

type PodCreator interface {
	Init(*v1.Pod)
	GetVolumes() ([]v1.Volume, []v1.VolumeMount)
	GetSidecars(*v1.Pod)
	GetInitContainers() ([]v1.Container, error)
	GetFinalizers() []string
	GetTolerations() []v1.Toleration
	GetNodeSelector() map[string]string
	GetServiceAccountName() string
	GetAffinityRole() string
	GetContainerCreator() ContainerCreator
	GetImagePullSecrets() []string
	IsDeploymentMode() bool
}

type ContainerCreator interface {
	GetExecutor() string
	GetProbes() (*v1.Probe, *v1.Probe, error)
	GetResourceRequirements() v1.ResourceRequirements
	GetLifecycle() (*v1.Lifecycle, error)
	GetImagePullPolicy() v1.PullPolicy
	GetImage() string
	GetEnvs() []v1.EnvVar
	GetSecurityContext() *v1.SecurityContext
}

// IsPodReady returns true if the PodReady condition on
// the given pod is set to true.
func IsPodReady(pod *v1.Pod) bool {
	condition := getPodCondition(&pod.Status, v1.PodReady)
	return condition != nil && condition.Status == v1.ConditionTrue
}

// GetPodByName returns pod if it exists among the pods' list
// Returns false if not found.
func GetPodByName(pods []v1.Pod, podName string) (v1.Pod, bool) {
	for _, pod := range pods {
		if pod.GetName() == podName {
			return pod, true
		}
	}
	return v1.Pod{}, false
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

// ArangodVolumeMount creates a volume mount structure for arangod.
func ArangodVolumeMount() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      ArangodVolumeName,
		MountPath: ArangodVolumeMountDir,
	}
}

// TlsKeyfileVolumeMount creates a volume mount structure for a TLS keyfile.
func TlsKeyfileVolumeMount() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      TlsKeyfileVolumeName,
		MountPath: TLSKeyfileVolumeMountDir,
	}
}

// ClientAuthCACertificateVolumeMount creates a volume mount structure for a client-auth CA certificate (ca.crt).
func ClientAuthCACertificateVolumeMount() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      ClientAuthCAVolumeName,
		MountPath: ClientAuthCAVolumeMountDir,
	}
}

// MasterJWTVolumeMount creates a volume mount structure for a master JWT secret (token).
func MasterJWTVolumeMount() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      MasterJWTSecretVolumeName,
		MountPath: MasterJWTSecretVolumeMountDir,
	}
}

// ClusterJWTVolumeMount creates a volume mount structure for a cluster JWT secret (token).
func ClusterJWTVolumeMount() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      ClusterJWTSecretVolumeName,
		MountPath: ClusterJWTSecretVolumeMountDir,
	}
}

func ExporterJWTVolumeMount() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      ExporterJWTVolumeName,
		MountPath: ExporterJWTVolumeMountDir,
	}
}

// RocksdbEncryptionVolumeMount creates a volume mount structure for a RocksDB encryption key.
func RocksdbEncryptionVolumeMount() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      RocksdbEncryptionVolumeName,
		MountPath: RocksDBEncryptionVolumeMountDir,
	}
}

// ArangodInitContainer creates a container configured to initalize a UUID file.
func ArangodInitContainer(name, id, engine, alpineImage string, requireUUID bool, securityContext *v1.SecurityContext) v1.Container {
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
		Name:  name,
		Image: alpineImage,
		Command: []string{
			"/bin/sh",
			"-c",
			command,
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("100m"),
				v1.ResourceMemory: resource.MustParse("10Mi"),
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("100m"),
				v1.ResourceMemory: resource.MustParse("50Mi"),
			},
		},
		VolumeMounts: []v1.VolumeMount{
			ArangodVolumeMount(),
		},
		SecurityContext: securityContext,
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

// NewContainer creates a container for specified creator
func NewContainer(args []string, containerCreator ContainerCreator) (v1.Container, error) {

	liveness, readiness, err := containerCreator.GetProbes()
	if err != nil {
		return v1.Container{}, err
	}

	lifecycle, err := containerCreator.GetLifecycle()
	if err != nil {
		return v1.Container{}, err
	}

	return v1.Container{
		Name:    ServerContainerName,
		Image:   containerCreator.GetImage(),
		Command: append([]string{containerCreator.GetExecutor()}, args...),
		Ports: []v1.ContainerPort{
			{
				Name:          "server",
				ContainerPort: int32(ArangoPort),
				Protocol:      v1.ProtocolTCP,
			},
		},
		Env:             containerCreator.GetEnvs(),
		Resources:       containerCreator.GetResourceRequirements(),
		LivenessProbe:   liveness,
		ReadinessProbe:  readiness,
		Lifecycle:       lifecycle,
		ImagePullPolicy: containerCreator.GetImagePullPolicy(),
		SecurityContext: containerCreator.GetSecurityContext(),
	}, nil
}

// NewPod creates a basic Pod for given settings.
func NewPod(deploymentName, role, id, podName string, podCreator PodCreator) v1.Pod {

	hostname := CreatePodHostName(deploymentName, role, id)
	p := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:       podName,
			Labels:     LabelsForDeployment(deploymentName, role),
			Finalizers: podCreator.GetFinalizers(),
		},
		Spec: v1.PodSpec{
			Hostname:           hostname,
			Subdomain:          CreateHeadlessServiceName(deploymentName),
			RestartPolicy:      v1.RestartPolicyNever,
			Tolerations:        podCreator.GetTolerations(),
			ServiceAccountName: podCreator.GetServiceAccountName(),
			NodeSelector:       podCreator.GetNodeSelector(),
		},
	}

	// Add ImagePullSecrets
	imagePullSecrets := podCreator.GetImagePullSecrets()
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

// CreatePod adds an owner to the given pod and calls the k8s api-server to created it.
// If the pod already exists, nil is returned.
// If another error occurs, that error is returned.
func CreatePod(kubecli kubernetes.Interface, pod *v1.Pod, ns string, owner metav1.OwnerReference) error {
	addOwnerRefToObject(pod.GetObjectMeta(), &owner)
	if _, err := kubecli.CoreV1().Pods(ns).Create(pod); err != nil && !IsAlreadyExists(err) {
		return maskAny(err)
	}
	return nil
}

func CreateVolumeEmptyDir(name string) v1.Volume {
	return v1.Volume{
		Name: name,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}
}

func CreateVolumeWithSecret(name, secretName string) v1.Volume {
	return v1.Volume{
		Name: name,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	}
}

func CreateVolumeWithPersitantVolumeClaim(name, claimName string) v1.Volume {
	return v1.Volume{
		Name: name,
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: claimName,
			},
		},
	}
}

func CreateEnvFieldPath(name, fieldPath string) v1.EnvVar {
	return v1.EnvVar{
		Name: name,
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: fieldPath,
			},
		},
	}
}

func CreateEnvSecretKeySelector(name, SecretKeyName, secretKey string) v1.EnvVar {
	return v1.EnvVar{
		Name:  name,
		Value: "",
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: SecretKeyName,
				},
				Key: secretKey,
			},
		},
	}
}
