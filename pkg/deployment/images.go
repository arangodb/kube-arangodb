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

package deployment

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	dockerPullableImageIDPrefix = "docker-pullable://"
)

type imagesBuilder struct {
	APIObject      k8sutil.APIObject
	Spec           api.DeploymentSpec
	Status         api.DeploymentStatus
	Log            zerolog.Logger
	KubeCli        kubernetes.Interface
	UpdateCRStatus func(status api.DeploymentStatus) error
}

// ensureImages creates pods needed to detect ImageID for specified images.
// Returns: retrySoon, error
func (d *Deployment) ensureImages(apiObject *api.ArangoDeployment) (bool, error) {
	ib := imagesBuilder{
		APIObject: apiObject,
		Spec:      apiObject.Spec,
		Status:    d.status,
		Log:       d.deps.Log,
		KubeCli:   d.deps.KubeCli,
		UpdateCRStatus: func(status api.DeploymentStatus) error {
			d.status = status
			if err := d.updateCRStatus(); err != nil {
				return maskAny(err)
			}
			return nil
		},
	}
	ctx := context.Background()
	retrySoon, err := ib.Run(ctx)
	if err != nil {
		return retrySoon, maskAny(err)
	}
	return retrySoon, nil
}

// Run creates pods needed to detect ImageID for specified images and puts the found
// image ID's into the status.Images list.
// Returns: retrySoon, error
func (ib *imagesBuilder) Run(ctx context.Context) (bool, error) {
	result := false
	// Check ArangoDB image
	if _, found := ib.Status.Images.GetByImage(ib.Spec.GetImage()); !found {
		// We need to find the image ID for the ArangoDB image
		retrySoon, err := ib.fetchArangoDBImageIDAndVersion(ctx, ib.Spec.GetImage())
		if err != nil {
			return retrySoon, maskAny(err)
		}
		result = result || retrySoon
	}

	return result, nil
}

// fetchArangoDBImageIDAndVersion checks a running pod for fetching the ID of the given image.
// When no pod exists, it is created, otherwise the ID is fetched & version detected.
// Returns: retrySoon, error
func (ib *imagesBuilder) fetchArangoDBImageIDAndVersion(ctx context.Context, image string) (bool, error) {
	role := "id"
	id := fmt.Sprintf("%0x", sha1.Sum([]byte(image)))[:6]
	podName := k8sutil.CreatePodName(ib.APIObject.GetName(), role, id)
	ns := ib.APIObject.GetNamespace()
	log := ib.Log.With().
		Str("pod", podName).
		Str("image", image).
		Logger()

	// Check if pod exists
	if pod, err := ib.KubeCli.CoreV1().Pods(ns).Get(podName, metav1.GetOptions{}); err == nil {
		// Pod found
		if !k8sutil.IsPodReady(pod) {
			log.Debug().Msg("Image ID Pod is not yet ready")
			return true, nil
		}
		if len(pod.Status.ContainerStatuses) == 0 {
			log.Warn().Msg("Empty list of ContainerStatuses")
			return true, nil
		}
		imageID := pod.Status.ContainerStatuses[0].ImageID
		if strings.HasPrefix(imageID, dockerPullableImageIDPrefix) {
			imageID = imageID[len(dockerPullableImageIDPrefix):]
		} else if imageID == "" {
			// Fall back to specified image
			imageID = image
		}

		// Try fetching the ArangoDB version
		client, err := arangod.CreateArangodImageIDClient(ctx, ib.APIObject, role, id)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to create Image ID Pod client")
			return true, nil
		}
		v, err := client.Version(ctx)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to fetch version from Image ID Pod")
			return true, nil
		}
		version := v.Version

		// We have all the info we need now, kill the pod and store the image info.
		if err := ib.KubeCli.CoreV1().Pods(ns).Delete(podName, nil); err != nil && !k8sutil.IsNotFound(err) {
			log.Warn().Err(err).Msg("Failed to delete Image ID Pod")
			return true, nil
		}

		info := api.ImageInfo{
			Image:           image,
			ImageID:         imageID,
			ArangoDBVersion: version,
		}
		ib.Status.Images.AddOrUpdate(info)
		if err := ib.UpdateCRStatus(ib.Status); err != nil {
			log.Warn().Err(err).Msg("Failed to save Image Info in CR status")
			return true, maskAny(err)
		}
		// We're done
		log.Debug().
			Str("image-id", imageID).
			Str("arangodb-version", string(version)).
			Msg("Found image ID and ArangoDB version")
		return false, nil
	}
	// Pod cannot be fetched, ensure it is created
	args := []string{
		"--server.authentication=false",
		fmt.Sprintf("--server.endpoint=tcp://[::]:%d", k8sutil.ArangoPort),
	}
	if err := k8sutil.CreateArangodPod(ib.KubeCli, true, ib.APIObject, role, id, "", image, ib.Spec.GetImagePullPolicy(), args, nil, nil, nil, "", ""); err != nil {
		log.Debug().Err(err).Msg("Failed to create image ID pod")
		return true, maskAny(err)
	}
	// Come back soon to inspect the pod
	return true, nil
}
