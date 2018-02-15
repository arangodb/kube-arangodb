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
)

const (
	arangodVolumeName     = "arangod-data"
	arangodVolumeMountDir = "/data"
)

// CreatePodName returns the name of the pod for a member with
// a given id in a deployment with a given name.
func CreatePodName(deploymentName, role, id string) string {
	return deploymentName + "-" + role + "-" + id
}

// arangodVolumeMounts creates a volume mount structure for arangod.
func arangodVolumeMounts() []v1.VolumeMount {
	return []v1.VolumeMount{
		{Name: arangodVolumeName, MountPath: arangodVolumeMountDir},
	}
}

// arangodContainer creates a container configured to run `arangod`.
func arangodContainer(name string, args []string, image string) v1.Container {
	c := v1.Container{
		Command: append([]string{"/usr/sbin/arangod"}, args...),
		Name:    name,
		Image:   image,
		Ports: []v1.ContainerPort{
			{
				Name:          "server",
				ContainerPort: int32(ArangoPort),
				Protocol:      v1.ProtocolTCP,
			},
		},
		VolumeMounts: arangodVolumeMounts(),
	}

	return c
}

// arangodPod creates a container configured to run `arangod`.
func arangodPod(clusterName, name string, args []string, image string) v1.Pod {
	p := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				arangodContainer(name, args, image),
			},
			Hostname:  name,
			Subdomain: clusterName,
		},
	}

	return p
}
