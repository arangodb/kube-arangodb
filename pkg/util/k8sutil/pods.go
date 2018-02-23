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
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	arangodVolumeName     = "arangod-data"
	ArangodVolumeMountDir = "/data"
)

// CreatePodName returns the name of the pod for a member with
// a given id in a deployment with a given name.
func CreatePodName(deploymentName, role, id string) string {
	return deploymentName + "-" + role + "-" + id
}

// arangodVolumeMounts creates a volume mount structure for arangod.
func arangodVolumeMounts() []v1.VolumeMount {
	return []v1.VolumeMount{
		{Name: arangodVolumeName, MountPath: ArangodVolumeMountDir},
	}
}

// arangodContainer creates a container configured to run `arangod`.
func arangodContainer(name string, image string, imagePullPolicy v1.PullPolicy, args []string, env map[string]string, livenessProbe *HTTPProbeConfig, readinessProbe *HTTPProbeConfig) v1.Container {
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
		c.Env = append(c.Env, v1.EnvVar{
			Name:  k,
			Value: v,
		})
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
func arangosyncContainer(name string, image string, imagePullPolicy v1.PullPolicy, args []string, env map[string]string, livenessProbe *HTTPProbeConfig) v1.Container {
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
		c.Env = append(c.Env, v1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	if livenessProbe != nil {
		c.LivenessProbe = livenessProbe.Create()
	}

	return c
}

// newPod creates a basic Pod for given settings.
func newPod(deploymentName, ns, role, id string) v1.Pod {
	name := CreatePodName(deploymentName, role, id)
	p := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: LabelsForDeployment(deploymentName, role),
		},
		Spec: v1.PodSpec{
			Hostname:  name,
			Subdomain: CreateHeadlessServiceName(deploymentName),
		},
	}
	return p
}

// CreateArangodPod creates a Pod that runs `arangod`.
// If the pod already exists, nil is returned.
// If another error occurs, that error is returned.
func CreateArangodPod(kubecli kubernetes.Interface, deployment metav1.Object, role, id, pvcName, image string, imagePullPolicy v1.PullPolicy,
	args []string, env map[string]string, livenessProbe *HTTPProbeConfig, readinessProbe *HTTPProbeConfig, owner metav1.OwnerReference) error {
	// Prepare basic pod
	p := newPod(deployment.GetName(), deployment.GetNamespace(), role, id)

	// Add arangod container
	c := arangodContainer(p.GetName(), image, imagePullPolicy, args, env, livenessProbe, readinessProbe)
	p.Spec.Containers = append(p.Spec.Containers, c)

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

	if err := createPod(kubecli, &p, deployment.GetNamespace(), owner); err != nil {
		return maskAny(err)
	}
	return nil
}

// CreateArangoSyncPod creates a Pod that runs `arangosync`.
// If the pod already exists, nil is returned.
// If another error occurs, that error is returned.
func CreateArangoSyncPod(kubecli kubernetes.Interface, deployment metav1.Object, role, id, image string, imagePullPolicy v1.PullPolicy,
	args []string, env map[string]string, livenessProbe *HTTPProbeConfig, owner metav1.OwnerReference) error {
	// Prepare basic pod
	p := newPod(deployment.GetName(), deployment.GetNamespace(), role, id)

	// Add arangosync container
	c := arangosyncContainer(p.GetName(), image, imagePullPolicy, args, env, livenessProbe)
	p.Spec.Containers = append(p.Spec.Containers, c)

	if err := createPod(kubecli, &p, deployment.GetNamespace(), owner); err != nil {
		return maskAny(err)
	}
	return nil
}

// createPod adds an owner to the given pod and calls the k8s api-server to created it.
// If the pod already exists, nil is returned.
// If another error occurs, that error is returned.
func createPod(kubecli kubernetes.Interface, pod *v1.Pod, ns string, owner metav1.OwnerReference) error {
	addOwnerRefToObject(pod.GetObjectMeta(), owner)
	if _, err := kubecli.CoreV1().Pods(ns).Create(pod); err != nil && !IsAlreadyExists(err) {
		return maskAny(err)
	}
	return nil
}
