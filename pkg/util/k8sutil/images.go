//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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
	goStrings "strings"

	core "k8s.io/api/core/v1"

	schedulerContainerResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container/resources"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
)

const (
	dockerPullableImageIDPrefix = "docker-pullable://"
)

// ConvertImageID2Image converts a ImageID from a ContainerStatus to an Image that can be used
// in a Container specification.
func ConvertImageID2Image(imageID string) string {
	if goStrings.HasPrefix(imageID, dockerPullableImageIDPrefix) {
		return imageID[len(dockerPullableImageIDPrefix):]
	}
	return imageID
}

// GetArangoDBImageIDFromPod returns the ArangoDB specific image from a pod
func GetArangoDBImageIDFromPod(pod *core.Pod, names ...string) (string, error) {
	if pod == nil {
		return "", errors.New("failed to get container statuses from nil pod")
	}

	// First try to find container by name
	if image, ok := GetArangoDBImageIDFromContainerStatuses(pod.Status.ContainerStatuses, names...); ok {
		return image, nil
	}

	if image, ok := GetArangoDBImageFromContainers(pod.Spec.Containers, names...); ok {
		return image, nil
	}

	if cs := pod.Status.ContainerStatuses; len(cs) > 0 {
		if image := cs[0].ImageID; image != "" {
			if disc := ConvertImageID2Image(image); disc != "" {
				return disc, nil
			}
		}
	}
	if cs := pod.Spec.Containers; len(cs) > 0 {
		if image := cs[0].Image; image != "" {
			return image, nil
		}
	}

	return "", errors.Errorf("Unable to find image from pod")
}

// GetArangoDBImageIDFromContainerStatuses returns the ArangoDB specific image from a container statuses
func GetArangoDBImageIDFromContainerStatuses(containers []core.ContainerStatus, names ...string) (string, bool) {
	for _, name := range names {
		if id := kresources.GetContainerStatusIDByName(containers, name); id != -1 {
			if image := containers[id].ImageID; image != "" {
				if disc := ConvertImageID2Image(image); disc != "" {
					return disc, true
				}
			}
		}
	}

	return "", false
}

// GetArangoDBImageFromContainers returns the ArangoDB specific image from a container specs
func GetArangoDBImageFromContainers(containers []core.Container, names ...string) (string, bool) {
	for _, name := range names {
		if id := kresources.GetContainerIDByName(containers, name); id != -1 {
			if image := containers[id].Image; image != "" {
				return image, true
			}
		}
	}

	return "", false
}

// GetImageDetails Returns latest defined Image details
func GetImageDetails(images ...*schedulerContainerResourcesApi.Image) *schedulerContainerResourcesApi.Image {
	var out *schedulerContainerResourcesApi.Image

	for _, image := range images {
		if image != nil {
			out = image
		}
	}

	return out
}
