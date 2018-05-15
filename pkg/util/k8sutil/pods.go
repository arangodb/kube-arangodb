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

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	alpineImage                     = "alpine"
	arangodVolumeName               = "arangod-data"
	tlsKeyfileVolumeName            = "tls-keyfile"
	clientAuthCAVolumeName          = "client-auth-ca"
	clusterJWTSecretVolumeName      = "cluster-jwt"
	masterJWTSecretVolumeName       = "master-jwt"
	rocksdbEncryptionVolumeName     = "rocksdb-encryption"
	ArangodVolumeMountDir           = "/data"
	RocksDBEncryptionVolumeMountDir = "/secrets/rocksdb/encryption"
	TLSKeyfileVolumeMountDir        = "/secrets/tls"
	ClientAuthCAVolumeMountDir      = "/secrets/client-auth/ca"
	ClusterJWTSecretVolumeMountDir  = "/secrets/cluster/jwt"
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
func arangodContainer(name string, image string, imagePullPolicy v1.PullPolicy, args []string, env map[string]EnvValue, livenessProbe *HTTPProbeConfig, readinessProbe *HTTPProbeConfig) v1.Container {
	c := v1.Container{
		Command:         append([]string{"/usr/sbin/arangod"}, args...),
		Name:            name,
		Image:           image,
		ImagePullPolicy: imagePullPolicy,
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

	return c
}

// arangosyncContainer creates a container configured to run `arangosync`.
func arangosyncContainer(name string, image string, imagePullPolicy v1.PullPolicy, args []string, env map[string]EnvValue, livenessProbe *HTTPProbeConfig) v1.Container {
	c := v1.Container{
		Command:         append([]string{"/usr/sbin/arangosync"}, args...),
		Name:            name,
		Image:           image,
		ImagePullPolicy: imagePullPolicy,
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

	return c
}

// newPod creates a basic Pod for given settings.
func newPod(deploymentName, ns, role, id, podName string) v1.Pod {
	hostname := CreatePodHostName(deploymentName, role, id)
	p := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   podName,
			Labels: LabelsForDeployment(deploymentName, role),
		},
		Spec: v1.PodSpec{
			Hostname:      hostname,
			Subdomain:     CreateHeadlessServiceName(deploymentName),
			RestartPolicy: v1.RestartPolicyNever,
		},
	}
	return p
}

// CreateArangodPod creates a Pod that runs `arangod`.
// If the pod already exists, nil is returned.
// If another error occurs, that error is returned.
func CreateArangodPod(kubecli kubernetes.Interface, developmentMode bool, deployment APIObject,
	role, id, podName, pvcName, image string, imagePullPolicy v1.PullPolicy,
	engine string, requireUUID bool,
	args []string, env map[string]EnvValue,
	livenessProbe *HTTPProbeConfig, readinessProbe *HTTPProbeConfig,
	tlsKeyfileSecretName, rocksdbEncryptionSecretName string) error {
	// Prepare basic pod
	p := newPod(deployment.GetName(), deployment.GetNamespace(), role, id, podName)

	// Add arangod container
	c := arangodContainer("arangod", image, imagePullPolicy, args, env, livenessProbe, readinessProbe)
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
func CreateArangoSyncPod(kubecli kubernetes.Interface, developmentMode bool, deployment APIObject, role, id, podName, image string, imagePullPolicy v1.PullPolicy,
	args []string, env map[string]EnvValue, livenessProbe *HTTPProbeConfig, tlsKeyfileSecretName, clientAuthCASecretName, masterJWTSecretName, clusterJWTSecretName, affinityWithRole string) error {
	// Prepare basic pod
	p := newPod(deployment.GetName(), deployment.GetNamespace(), role, id, podName)

	// Add arangosync container
	c := arangosyncContainer("arangosync", image, imagePullPolicy, args, env, livenessProbe)
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
