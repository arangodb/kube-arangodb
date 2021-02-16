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

package k8sutil

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"

	"k8s.io/apimachinery/pkg/api/resource"

	core "k8s.io/api/core/v1"
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
	TLSSNIKeyfileVolumeMountDir     = "/secrets/sni"
	ClientAuthCAVolumeMountDir      = "/secrets/client-auth/ca"
	ClusterJWTSecretVolumeMountDir  = "/secrets/cluster/jwt"
	ExporterJWTVolumeMountDir       = "/secrets/exporter/jwt"
	MasterJWTSecretVolumeMountDir   = "/secrets/master/jwt"
)

// IsPodReady returns true if the PodReady condition on
// the given pod is set to true.
func IsPodReady(pod *core.Pod) bool {
	condition := getPodCondition(&pod.Status, core.PodReady)
	return condition != nil && condition.Status == core.ConditionTrue
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

// IsPodSucceeded returns true if the arangodb container of the pod
// has terminated with exit code 0.
func IsPodSucceeded(pod *core.Pod) bool {
	if pod.Status.Phase == core.PodSucceeded {
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
func IsPodFailed(pod *core.Pod) bool {
	if pod.Status.Phase == core.PodFailed {
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

// IsArangoDBImageIDAndVersionPod returns true if the given pod is used for fetching image ID and ArangoDB version of an image
func IsArangoDBImageIDAndVersionPod(p *core.Pod) bool {
	role, found := p.GetLabels()[LabelKeyRole]
	return found && role == ImageIDAndVersionRole
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
func ArangodVolumeMount() core.VolumeMount {
	return core.VolumeMount{
		Name:      ArangodVolumeName,
		MountPath: ArangodVolumeMountDir,
	}
}

// TlsKeyfileVolumeMount creates a volume mount structure for a TLS keyfile.
func TlsKeyfileVolumeMount() core.VolumeMount {
	return core.VolumeMount{
		Name:      TlsKeyfileVolumeName,
		MountPath: TLSKeyfileVolumeMountDir,
		ReadOnly:  true,
	}
}

// ClientAuthCACertificateVolumeMount creates a volume mount structure for a client-auth CA certificate (ca.crt).
func ClientAuthCACertificateVolumeMount() core.VolumeMount {
	return core.VolumeMount{
		Name:      ClientAuthCAVolumeName,
		MountPath: ClientAuthCAVolumeMountDir,
	}
}

// MasterJWTVolumeMount creates a volume mount structure for a master JWT secret (token).
func MasterJWTVolumeMount() core.VolumeMount {
	return core.VolumeMount{
		Name:      MasterJWTSecretVolumeName,
		MountPath: MasterJWTSecretVolumeMountDir,
	}
}

// ClusterJWTVolumeMount creates a volume mount structure for a cluster JWT secret (token).
func ClusterJWTVolumeMount() core.VolumeMount {
	return core.VolumeMount{
		Name:      ClusterJWTSecretVolumeName,
		MountPath: ClusterJWTSecretVolumeMountDir,
	}
}

func ExporterJWTVolumeMount() core.VolumeMount {
	return core.VolumeMount{
		Name:      ExporterJWTVolumeName,
		MountPath: ExporterJWTVolumeMountDir,
		ReadOnly:  true,
	}
}

// RocksdbEncryptionVolumeMount creates a volume mount structure for a RocksDB encryption key.
func RocksdbEncryptionVolumeMount() core.VolumeMount {
	return core.VolumeMount{
		Name:      RocksdbEncryptionVolumeName,
		MountPath: RocksDBEncryptionVolumeMountDir,
	}
}

// RocksdbEncryptionReadOnlyVolumeMount creates a volume mount structure for a RocksDB encryption key.
func RocksdbEncryptionReadOnlyVolumeMount() core.VolumeMount {
	return core.VolumeMount{
		Name:      RocksdbEncryptionVolumeName,
		MountPath: RocksDBEncryptionVolumeMountDir,
		ReadOnly:  true,
	}
}

// ArangodInitContainer creates a container configured to initalize a UUID file.
func ArangodInitContainer(name, id, engine, executable, operatorImage string, requireUUID bool, securityContext *core.SecurityContext) core.Container {
	uuidFile := filepath.Join(ArangodVolumeMountDir, "UUID")
	engineFile := filepath.Join(ArangodVolumeMountDir, "ENGINE")
	var command []string = []string{
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
		VolumeMounts: []core.VolumeMount{
			ArangodVolumeMount(),
		},
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
func NewContainer(args []string, containerCreator interfaces.ContainerCreator) (core.Container, error) {

	liveness, readiness, err := containerCreator.GetProbes()
	if err != nil {
		return core.Container{}, err
	}

	lifecycle, err := containerCreator.GetLifecycle()
	if err != nil {
		return core.Container{}, err
	}

	return core.Container{
		Name:            ServerContainerName,
		Image:           containerCreator.GetImage(),
		Command:         append([]string{containerCreator.GetExecutor()}, args...),
		Ports:           containerCreator.GetPorts(),
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
func NewPod(deploymentName, role, id, podName string, podCreator interfaces.PodCreator) core.Pod {

	hostname := CreatePodHostName(deploymentName, role, id)
	p := core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:       podName,
			Labels:     LabelsForDeployment(deploymentName, role),
			Finalizers: podCreator.GetFinalizers(),
		},
		Spec: core.PodSpec{
			Hostname:           hostname,
			Subdomain:          CreateHeadlessServiceName(deploymentName),
			RestartPolicy:      core.RestartPolicyNever,
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

	return fmt.Sprintf("%0x", sha256.Sum256(data)), nil
}

// CreatePod adds an owner to the given pod and calls the k8s api-server to created it.
// If the pod already exists, nil is returned.
// If another error occurs, that error is returned.
func CreatePod(kubecli kubernetes.Interface, pod *core.Pod, ns string, owner metav1.OwnerReference) (types.UID, error) {
	AddOwnerRefToObject(pod.GetObjectMeta(), &owner)

	if pod, err := kubecli.CoreV1().Pods(ns).Create(pod); err != nil && !IsAlreadyExists(err) {
		return "", errors.WithStack(err)
	} else {
		return pod.UID, nil
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
