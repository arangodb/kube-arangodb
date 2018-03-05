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
	"fmt"

	"k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/k8s-operator/pkg/apis/storage/v1alpha"
	"github.com/arangodb/k8s-operator/pkg/util/k8sutil"
)

// ensureDaemonSet ensures that a daemonset is created for the given local storage.
func (l *LocalStorage) ensureDaemonSet(apiObject *api.ArangoLocalStorage) error {
	ns := apiObject.GetNamespace()
	c := corev1.Container{
		Image: l.image,
		Args: []string{
			"storage",
			"provisioner",
			"--storage-class-name=" + apiObject.Spec.StorageClass.Name,
		},
	}
	ds := &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: apiObject.GetName(),
		},
		Spec: v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						c,
					},
					NodeSelector: apiObject.Spec.NodeSelector,
				},
			},
		},
	}
	for i, lp := range apiObject.Spec.LocalPath {
		volName := fmt.Sprintf("local-path-%d", i)
		c.Args = append(c.Args, fmt.Sprintf("--local-path=%s", lp))
		c.VolumeMounts = append(c.VolumeMounts,
			corev1.VolumeMount{
				Name: volName,
			})
		hostPathType := corev1.HostPathDirectoryOrCreate
		ds.Spec.Template.Spec.Volumes = append(ds.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: volName,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: lp,
					Type: &hostPathType,
				},
			},
		})
	}
	if _, err := l.deps.KubeCli.AppsV1().DaemonSets(ns).Create(ds); !k8sutil.IsAlreadyExists(err) && err != nil {
		return maskAny(err)
	}
	// TODO
	return nil
}
