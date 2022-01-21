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
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	dockerPullableImageIDPrefix = "docker-pullable://"
)

// ConvertImageID2Image converts a ImageID from a ContainerStatus to an Image that can be used
// in a Container specification.
func ConvertImageID2Image(imageID string) string {
	if strings.HasPrefix(imageID, dockerPullableImageIDPrefix) {
		return imageID[len(dockerPullableImageIDPrefix):]
	}
	return imageID
}

// GetArangoDBImageIDFromPod returns the ArangoDB specific image from a pod
func GetArangoDBImageIDFromPod(pod *corev1.Pod) (string, error) {
	if pod == nil {
		return "", errors.New("failed to get container statuses from nil pod")
	}

	if len(pod.Status.ContainerStatuses) == 0 {
		return "", errors.New("empty list of ContainerStatuses")
	}

	rawImageID := pod.Status.ContainerStatuses[0].ImageID
	if len(pod.Status.ContainerStatuses) > 1 {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if strings.Contains(containerStatus.ImageID, "arango") {
				rawImageID = containerStatus.ImageID
			}
		}
	}

	return ConvertImageID2Image(rawImageID), nil
}
