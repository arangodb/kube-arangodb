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

package storage

import (
	"fmt"
	"strconv"

	"k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	roleProvisioner = "provisioner"
)

// ensureDaemonSet ensures that a daemonset is created for the given local storage.
// If it already exists, it is updated.
func (ls *LocalStorage) ensureDaemonSet(apiObject *api.ArangoLocalStorage) error {
	log := ls.deps.Log
	ns := ls.config.Namespace
	c := corev1.Container{
		Name:            "provisioner",
		Image:           ls.image,
		ImagePullPolicy: ls.imagePullPolicy,
		Args: []string{
			"storage",
			"provisioner",
			"--port=" + strconv.Itoa(provisioner.DefaultPort),
		},
		Ports: []corev1.ContainerPort{
			corev1.ContainerPort{
				ContainerPort: int32(provisioner.DefaultPort),
			},
		},
		Env: []corev1.EnvVar{
			corev1.EnvVar{
				Name: constants.EnvOperatorNodeName,
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "spec.nodeName",
					},
				},
			},
		},
	}
	dsLabels := k8sutil.LabelsForLocalStorage(apiObject.GetName(), roleProvisioner)
	dsSpec := v1.DaemonSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: dsLabels,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: dsLabels,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					c,
				},
				NodeSelector: apiObject.Spec.NodeSelector,
			},
		},
	}
	for i, lp := range apiObject.Spec.LocalPath {
		volName := fmt.Sprintf("local-path-%d", i)
		c := &dsSpec.Template.Spec.Containers[0]
		c.VolumeMounts = append(c.VolumeMounts,
			corev1.VolumeMount{
				Name:      volName,
				MountPath: lp,
			})
		hostPathType := corev1.HostPathDirectoryOrCreate
		dsSpec.Template.Spec.Volumes = append(dsSpec.Template.Spec.Volumes, corev1.Volume{
			Name: volName,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: lp,
					Type: &hostPathType,
				},
			},
		})
	}
	ds := &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:   apiObject.GetName(),
			Labels: dsLabels,
		},
		Spec: dsSpec,
	}
	// Attach DS to ArangoLocalStorage
	ds.SetOwnerReferences(append(ds.GetOwnerReferences(), apiObject.AsOwner()))
	// Create DS
	if _, err := ls.deps.KubeCli.AppsV1().DaemonSets(ns).Create(ds); err != nil {
		if k8sutil.IsAlreadyExists(err) {
			// Already exists, update it
		} else {
			return maskAny(err)
		}
	} else {
		// We're done
		log.Debug().Msg("Created DaemonSet")
		return nil
	}

	// Update existing DS
	attempt := 0
	for {
		attempt++

		// Load current DS
		current, err := ls.deps.KubeCli.AppsV1().DaemonSets(ns).Get(ds.GetName(), metav1.GetOptions{})
		if err != nil {
			return maskAny(err)
		}

		// Update it
		current.Spec = dsSpec
		if _, err := ls.deps.KubeCli.AppsV1().DaemonSets(ns).Update(current); k8sutil.IsConflict(err) && attempt < 10 {
			// Failed to update, try again
			continue
		} else if err != nil {
			ls.deps.Log.Debug().Err(err).Msg("failed to patch DaemonSet spec")
			return maskAny(fmt.Errorf("failed to patch DaemonSet spec: %v", err))
		} else {
			// Update was a success
			log.Debug().Msg("Updated DaemonSet")
			return nil
		}
	}
}
