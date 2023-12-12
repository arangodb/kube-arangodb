//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

	core "k8s.io/api/core/v1"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
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
func GetArangoDBImageIDFromPod(pod *core.Pod) (string, error) {
	if pod == nil {
		return "", errors.New("failed to get container statuses from nil pod")
	}

	if len(pod.Status.ContainerStatuses) == 0 {
		return "", errors.New("empty list of ContainerStatuses")
	}

	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == shared.ServerContainerName {
			return ConvertImageID2Image(cs.ImageID), nil
		}
	}

	// If Server container is not found use first container
	return ConvertImageID2Image(pod.Status.ContainerStatuses[0].ImageID), nil
}

// GetImageDetails Returns latest defined Image details
func GetImageDetails(images ...*sharedApi.Image) *sharedApi.Image {
	var out *sharedApi.Image

	for _, image := range images {
		if image != nil {
			out = image
		}
	}

	return out
}

// InjectImageDetails injects image details into the Pod definition
func InjectImageDetails(image *sharedApi.Image, pod *core.PodTemplateSpec, containers ...*core.Container) error {
	if image == nil {
		return errors.Newf("Image not found")
	} else if err := image.Validate(); err != nil {
		return errors.Wrapf(err, "Unable to validate image")
	}

	for _, secret := range image.PullSecrets {
		if HasImagePullSecret(pod.Spec.ImagePullSecrets, secret) {
			continue
		}

		pod.Spec.ImagePullSecrets = append(pod.Spec.ImagePullSecrets, core.LocalObjectReference{
			Name: secret,
		})
	}

	for _, container := range containers {
		container.Image = *image.Image

		if ps := image.PullPolicy; ps != nil {
			container.ImagePullPolicy = *ps
		}
	}

	return nil
}

func HasImagePullSecret(secrets []core.LocalObjectReference, secret string) bool {
	for _, sec := range secrets {
		if sec.Name == secret {
			return true
		}
	}

	return false
}
