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

package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/dchest/uniuri"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner"
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
		return maskAny(err)
	}
	if len(clients) == 0 {
		// No provisioners available
		return maskAny(fmt.Errorf("No ready provisioner endpoints found"))
	}
	// Randomize list
	rand.Shuffle(len(clients), func(i, j int) {
		clients[i], clients[j] = clients[j], clients[i]
	})

	for i, claim := range unboundClaims {
		// Find size of PVC
		volSize := defaultVolumeSize
		if reqStorage := claim.Spec.Resources.Requests.StorageEphemeral(); reqStorage != nil {
			if v, ok := reqStorage.AsInt64(); ok && v > 0 {
				volSize = v
			}
		}
		// Create PV
		if err := ls.createPV(ctx, apiObject, clients, i, volSize); err != nil {
			log.Error().Err(err).Msg("Failed to create PersistentVolume")
		}
	}

	return nil
}

// createPV creates a PersistentVolume.
func (ls *LocalStorage) createPV(ctx context.Context, apiObject *api.ArangoLocalStorage, clients []provisioner.API, clientsOffset int, volSize int64) error {
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
			pvName := apiObject.GetName() + "-" + name
			volumeMode := v1.PersistentVolumeFilesystem
			nodeAff, err := createNodeAffinity(info.NodeName)
			if err != nil {
				return maskAny(err) // No continue here, since this should just not happen
			}
			pv := &v1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: pvName,
					Annotations: map[string]string{
						AnnProvisionedBy:                      storageClassProvisioner,
						v1.AlphaStorageNodeAffinityAnnotation: nodeAff,
						nodeNameAnnotation:                    info.NodeName,
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
				},
			}
			// Attach PV to ArangoLocalStorage
			pv.SetOwnerReferences(append(pv.GetOwnerReferences(), apiObject.AsOwner()))
			if _, err := ls.deps.KubeCli.CoreV1().PersistentVolumes().Create(pv); err != nil {
				log.Error().Err(err).Msg("Failed to create PersistentVolume")
				continue
			}
			log.Debug().
				Str("name", pvName).
				Str("node-name", info.NodeName).
				Msg("Created PersistentVolume")
			return nil
		}
	}
	return maskAny(fmt.Errorf("No more nodes available"))
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
func createNodeAffinity(nodeName string) (string, error) {
	aff := v1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
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
		},
	}
	encoded, err := json.Marshal(aff)
	if err != nil {
		return "", maskAny(err)
	}
	return string(encoded), nil
}
