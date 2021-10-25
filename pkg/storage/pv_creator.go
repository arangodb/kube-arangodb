//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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

package storage

import (
	"context"
	"crypto/sha1"
	"fmt"
	"math/rand"
	"net"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/dchest/uniuri"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	defaultVolumeSize = int64(8 * 1024 * 1024 * 1024) // 8GB

	// AnnProvisionedBy is the external provisioner annotation in PV object
	AnnProvisionedBy = "pv.kubernetes.io/provisioned-by"
)

var (
	// name of the annotation containing the node name
	nodeNameAnnotation = api.SchemeGroupVersion.Group + "/node-name"
)

// createPVs creates a given number of PersistentVolume's.
func (ls *LocalStorage) createPVs(ctx context.Context, apiObject *api.ArangoLocalStorage, unboundClaims []v1.PersistentVolumeClaim) error {
	log := ls.deps.Log
	// Find provisioner clients
	clients, err := ls.createProvisionerClients()
	if err != nil {
		return errors.WithStack(err)
	}
	if len(clients) == 0 {
		// No provisioners available
		return errors.WithStack(errors.Newf("No ready provisioner endpoints found"))
	}
	// Randomize list
	rand.Shuffle(len(clients), func(i, j int) {
		clients[i], clients[j] = clients[j], clients[i]
	})

	var nodeClientMap map[string]provisioner.API
	for i, claim := range unboundClaims {
		// Find deployment name & role in the claim (if any)
		deplName, role, enforceAniAffinity := getDeploymentInfo(claim)
		allowedClients := clients
		if deplName != "" {
			// Select nodes to choose from such that no volume in group lands on the same node
			if nodeClientMap == nil {
				nodeClientMap = createNodeClientMap(ctx, clients)
			}
			var err error
			allowedClients, err = ls.filterAllowedNodes(nodeClientMap, deplName, role)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to filter allowed nodes")
				continue // We'll try this claim again later
			}
			if !enforceAniAffinity && len(allowedClients) == 0 {
				// No possible nodes found that have no other volume (in same group) on it.
				// We don't have to enforce separate nodes, so use all clients.
				allowedClients = clients
			}
		}

		// Find size of PVC
		volSize := defaultVolumeSize
		if reqStorage, ok := claim.Spec.Resources.Requests[v1.ResourceStorage]; ok {
			if v, ok := reqStorage.AsInt64(); ok && v > 0 {
				volSize = v
			}
		}
		// Create PV
		if err := ls.createPV(ctx, apiObject, allowedClients, i, volSize, claim, deplName, role); err != nil {
			log.Error().Err(err).Msg("Failed to create PersistentVolume")
		}
	}

	return nil
}

// createPV creates a PersistentVolume.
func (ls *LocalStorage) createPV(ctx context.Context, apiObject *api.ArangoLocalStorage, clients []provisioner.API, clientsOffset int, volSize int64, claim v1.PersistentVolumeClaim, deploymentName, role string) error {
	log := ls.deps.Log
	// Try clients
	for clientIdx := 0; clientIdx < len(clients); clientIdx++ {
		client := clients[(clientsOffset+clientIdx)%len(clients)]

		// Try local path within client
		for _, localPathRoot := range apiObject.Spec.LocalPath {
			log := log.With().Str("local-path-root", localPathRoot).Logger()
			info, err := client.GetInfo(ctx, localPathRoot)
			if err != nil {
				log.Error().Err(err).Msg("Failed to get client info")
				continue
			}
			if info.Available < volSize {
				log.Debug().Msg("Not enough available size")
				continue
			}
			// Ok, prepare a directory
			name := strings.ToLower(uniuri.New())
			localPath := filepath.Join(localPathRoot, name)
			log = log.With().Str("local-path", localPath).Logger()
			if err := client.Prepare(ctx, localPath); err != nil {
				log.Error().Err(err).Msg("Failed to prepare local path")
				continue
			}
			// Create a volume
			pvName := strings.ToLower(apiObject.GetName() + "-" + shortHash(info.NodeName) + "-" + name)
			volumeMode := v1.PersistentVolumeFilesystem
			nodeSel := createNodeSelector(info.NodeName)
			pv := &v1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: pvName,
					Annotations: map[string]string{
						AnnProvisionedBy:   storageClassProvisioner,
						nodeNameAnnotation: info.NodeName,
					},
					Labels: map[string]string{
						k8sutil.LabelKeyArangoDeployment: deploymentName,
						k8sutil.LabelKeyRole:             role,
					},
				},
				Spec: v1.PersistentVolumeSpec{
					Capacity: v1.ResourceList{
						v1.ResourceStorage: *resource.NewQuantity(volSize, resource.BinarySI),
					},
					PersistentVolumeReclaimPolicy: v1.PersistentVolumeReclaimRetain,
					PersistentVolumeSource: v1.PersistentVolumeSource{
						Local: &v1.LocalVolumeSource{
							Path: localPath,
						},
					},
					AccessModes: []v1.PersistentVolumeAccessMode{
						v1.ReadWriteOnce,
					},
					StorageClassName: apiObject.Spec.StorageClass.Name,
					VolumeMode:       &volumeMode,
					ClaimRef: &v1.ObjectReference{
						Kind:       "PersistentVolumeClaim",
						APIVersion: "",
						Name:       claim.GetName(),
						Namespace:  claim.GetNamespace(),
						UID:        claim.GetUID(),
					},
					NodeAffinity: &v1.VolumeNodeAffinity{
						Required: nodeSel,
					},
				},
			}
			// Attach PV to ArangoLocalStorage
			pv.SetOwnerReferences(append(pv.GetOwnerReferences(), apiObject.AsOwner()))
			if _, err := ls.deps.KubeCli.CoreV1().PersistentVolumes().Create(context.Background(), pv, metav1.CreateOptions{}); err != nil {
				log.Error().Err(err).Msg("Failed to create PersistentVolume")
				continue
			}
			log.Debug().
				Str("name", pvName).
				Str("node-name", info.NodeName).
				Msg("Created PersistentVolume")

			// Bind claim to volume
			if err := ls.bindClaimToVolume(claim, pv.GetName()); err != nil {
				// Try to delete the PV now
				if err := ls.deps.KubeCli.CoreV1().PersistentVolumes().Delete(context.Background(), pv.GetName(), metav1.DeleteOptions{}); err != nil {
					log.Error().Err(err).Msg("Failed to delete PV after binding PVC failed")
				}
				return errors.WithStack(err)
			}

			return nil
		}
	}
	return errors.WithStack(errors.Newf("No more nodes available"))
}

// createValidEndpointList convers the given endpoints list into
// valid addresses.
func createValidEndpointList(list *v1.EndpointsList) []string {
	result := make([]string, 0, len(list.Items))
	for _, ep := range list.Items {
		for _, subset := range ep.Subsets {
			for _, ip := range subset.Addresses {
				addr := net.JoinHostPort(ip.IP, strconv.Itoa(provisioner.DefaultPort))
				result = append(result, addr)
			}
		}
	}
	sort.Strings(result)
	return result
}

// createNodeAffinity creates a node affinity serialized to string.
func createNodeSelector(nodeName string) *v1.NodeSelector {
	return &v1.NodeSelector{
		NodeSelectorTerms: []v1.NodeSelectorTerm{
			v1.NodeSelectorTerm{
				MatchExpressions: []v1.NodeSelectorRequirement{
					v1.NodeSelectorRequirement{
						Key:      "kubernetes.io/hostname",
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{nodeName},
					},
				},
			},
		},
	}
}

// createNodeClientMap creates a map from node name to API.
// Clients that do not respond properly on a GetNodeInfo request are
// ignored.
func createNodeClientMap(ctx context.Context, clients []provisioner.API) map[string]provisioner.API {
	result := make(map[string]provisioner.API)
	for _, c := range clients {
		if info, err := c.GetNodeInfo(ctx); err == nil {
			result[info.NodeName] = c
		}
	}
	return result
}

// getDeploymentInfo returns the name of the deployment that created the given claim,
// the role of the server that the claim is used for and the value for `enforceAntiAffinity`.
// If not found, empty strings are returned.
// Returns deploymentName, role, enforceAntiAffinity.
func getDeploymentInfo(pvc v1.PersistentVolumeClaim) (string, string, bool) {
	deploymentName := pvc.GetLabels()[k8sutil.LabelKeyArangoDeployment]
	role := pvc.GetLabels()[k8sutil.LabelKeyRole]
	enforceAntiAffinity, _ := strconv.ParseBool(pvc.GetAnnotations()[constants.AnnotationEnforceAntiAffinity]) // If annotation empty, this will yield false.
	return deploymentName, role, enforceAntiAffinity
}

// filterAllowedNodes returns those clients that do not yet have a volume for the given deployment name & role.
func (ls *LocalStorage) filterAllowedNodes(clients map[string]provisioner.API, deploymentName, role string) ([]provisioner.API, error) {
	// Find all PVs for given deployment & role
	list, err := ls.deps.KubeCli.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s,%s=%s", k8sutil.LabelKeyArangoDeployment, deploymentName, k8sutil.LabelKeyRole, role),
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	excludedNodes := make(map[string]struct{})
	for _, pv := range list.Items {
		nodeName := pv.GetAnnotations()[nodeNameAnnotation]
		excludedNodes[nodeName] = struct{}{}
	}
	result := make([]provisioner.API, 0, len(clients))
	for nodeName, c := range clients {
		if _, found := excludedNodes[nodeName]; !found {
			result = append(result, c)
		}
	}
	return result, nil
}

// bindClaimToVolume tries to bind the given claim to the volume with given name.
// If the claim has been updated, the function retries several times.
func (ls *LocalStorage) bindClaimToVolume(claim v1.PersistentVolumeClaim, volumeName string) error {
	log := ls.deps.Log.With().Str("pvc-name", claim.GetName()).Str("volume-name", volumeName).Logger()
	pvcs := ls.deps.KubeCli.CoreV1().PersistentVolumeClaims(claim.GetNamespace())

	for attempt := 0; attempt < 10; attempt++ {
		// Backoff if needed
		time.Sleep(time.Millisecond * time.Duration(10*attempt))

		// Fetch latest version of claim
		updated, err := pvcs.Get(context.Background(), claim.GetName(), metav1.GetOptions{})
		if k8sutil.IsNotFound(err) {
			return errors.WithStack(err)
		} else if err != nil {
			log.Warn().Err(err).Msg("Failed to load updated PersistentVolumeClaim")
			continue
		}

		// Check claim. If already bound, bail out
		if !pvcNeedsVolume(*updated) {
			if updated.Spec.VolumeName == volumeName {
				log.Info().Msg("PersistentVolumeClaim already bound to PersistentVolume")
				return nil
			}
			return errors.WithStack(errors.Newf("PersistentVolumeClaim '%s' no longer needs a volume", claim.GetName()))
		}

		// Try to bind
		updated.Spec.VolumeName = volumeName
		if _, err := pvcs.Update(context.Background(), updated, metav1.UpdateOptions{}); k8sutil.IsConflict(err) {
			// Claim modified already, retry
			log.Debug().Err(err).Msg("PersistentVolumeClaim has been modified. Retrying.")
		} else if err != nil {
			log.Error().Err(err).Msg("Failed to bind PVC to volume")
			return errors.WithStack(err)
		}
		log.Debug().Msg("Bound volume to PersistentVolumeClaim")
		return nil
	}
	log.Error().Msg("All attempts to bind PVC to volume failed")
	return errors.WithStack(errors.Newf("All attempts to bind PVC to volume failed"))
}

// shortHash creates a 6 letter hash of the given name.
func shortHash(name string) string {
	h := sha1.Sum([]byte(name))
	return fmt.Sprintf("%0x", h)[:6]
}
