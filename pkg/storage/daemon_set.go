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

package storage

import (
	"context"
	"fmt"
	"strconv"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

const (
	roleProvisioner = "provisioner"
)

// ensureDaemonSet ensures that a daemonset is created for the given local storage.
// If it already exists, it is updated.
func (ls *LocalStorage) ensureDaemonSet(apiObject *api.ArangoLocalStorage) error {
	ns := ls.config.Namespace
	c := core.Container{
		Name:            "provisioner",
		Image:           ls.image,
		ImagePullPolicy: ls.imagePullPolicy,
		Args: []string{
			"storage",
			"provisioner",
			"--port=" + strconv.Itoa(provisioner.DefaultPort),
		},
		Ports: []core.ContainerPort{
			core.ContainerPort{
				ContainerPort: int32(provisioner.DefaultPort),
			},
		},
		Env: []core.EnvVar{
			core.EnvVar{
				Name: constants.EnvOperatorNodeName,
				ValueFrom: &core.EnvVarSource{
					FieldRef: &core.ObjectFieldSelector{
						FieldPath: "spec.nodeName",
					},
				},
			},
		},
	}

	if apiObject.Spec.GetPrivileged() {
		c.SecurityContext = &core.SecurityContext{
			Privileged: util.NewType[bool](true),
		}
	}

	dsLabels := k8sutil.LabelsForLocalStorage(apiObject.GetName(), roleProvisioner)
	dsSpec := apps.DaemonSetSpec{
		Selector: &meta.LabelSelector{
			MatchLabels: dsLabels,
		},
		Template: core.PodTemplateSpec{
			ObjectMeta: meta.ObjectMeta{
				Labels: dsLabels,
			},
			Spec: core.PodSpec{
				Containers: []core.Container{
					c,
				},
				NodeSelector:     apiObject.Spec.NodeSelector,
				ImagePullSecrets: ls.imagePullSecrets,
				Priority:         apiObject.Spec.PodCustomization.GetPriority(),
				Tolerations:      apiObject.Spec.Tolerations,
			},
		},
	}

	for i, lp := range apiObject.Spec.LocalPath {
		volName := fmt.Sprintf("local-path-%d", i)
		c := &dsSpec.Template.Spec.Containers[0]
		c.VolumeMounts = append(c.VolumeMounts,
			core.VolumeMount{
				Name:      volName,
				MountPath: lp,
			})
		hostPathType := core.HostPathDirectoryOrCreate
		dsSpec.Template.Spec.Volumes = append(dsSpec.Template.Spec.Volumes, core.Volume{
			Name: volName,
			VolumeSource: core.VolumeSource{
				HostPath: &core.HostPathVolumeSource{
					Path: lp,
					Type: &hostPathType,
				},
			},
		})
	}
	ds := &apps.DaemonSet{
		ObjectMeta: meta.ObjectMeta{
			Name:   apiObject.GetName(),
			Labels: dsLabels,
		},
		Spec: dsSpec,
	}
	// Attach DS to ArangoLocalStorage
	ds.SetOwnerReferences(append(ds.GetOwnerReferences(), apiObject.AsOwner()))
	// Create DS
	if _, err := ls.deps.Client.Kubernetes().AppsV1().DaemonSets(ns).Create(context.Background(), ds, meta.CreateOptions{}); err != nil {
		if kerrors.IsAlreadyExists(err) {
			// Already exists, update it
		} else {
			return errors.WithStack(err)
		}
	} else {
		// We're done
		ls.log.Debug("Created DaemonSet")
		return nil
	}

	// Update existing DS
	attempt := 0
	for {
		attempt++

		// Load current DS
		current, err := ls.deps.Client.Kubernetes().AppsV1().DaemonSets(ns).Get(context.Background(), ds.GetName(), meta.GetOptions{})
		if err != nil {
			return errors.WithStack(err)
		}

		// Update it
		current.Spec = dsSpec
		if _, err := ls.deps.Client.Kubernetes().AppsV1().DaemonSets(ns).Update(context.Background(), current, meta.UpdateOptions{}); kerrors.IsConflict(err) && attempt < 10 {
			ls.log.Err(err).Debug("failed to patch DaemonSet spec")
			// Failed to update, try again
			continue
		} else if err != nil {
			ls.log.Err(err).Debug("failed to patch DaemonSet spec")
			return errors.WithStack(errors.Newf("failed to patch DaemonSet spec: %v", err))
		} else {
			// Update was a success
			ls.log.Debug("Updated DaemonSet")
			return nil
		}
	}
}
