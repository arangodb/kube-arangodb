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

package tests

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strings"
	"time"

	"github.com/dchest/uniuri"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/arangodb/kube-arangodb/pkg/util/retry"
)

// prepullImage runs a daemonset that pulls a given ArangoDB image on all nodes.
func prepullArangoImage(cli kubernetes.Interface, image, namespace string) error {
	name := "prepuller-" + strings.ToLower(uniuri.NewLen(4))
	dsLabels := map[string]string{
		"app":        "prepuller",
		"image-hash": fmt.Sprintf("%0x", sha1.Sum([]byte(image)))[:10],
	}
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: dsLabels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: dsLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: dsLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:  "arango",
							Image: image,
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name:  "ARANGO_NO_AUTH",
									Value: "1",
								},
							},
						},
					},
				},
			},
		},
	}
	// Create DS
	if _, err := cli.AppsV1().DaemonSets(namespace).Create(context.Background(), ds, metav1.CreateOptions{}); err != nil {
		return maskAny(err)
	}
	// Cleanup on exit
	defer func() {
		cli.AppsV1().DaemonSets(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	}()
	// Now wait for it to be ready
	op := func() error {
		current, err := cli.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return maskAny(err)
		}
		if current.Status.DesiredNumberScheduled > current.Status.NumberReady {
			return maskAny(fmt.Errorf("Expected %d pods to be ready, got %d", current.Status.DesiredNumberScheduled, current.Status.NumberReady))
		}
		return nil
	}
	if err := retry.Retry(op, time.Hour); err != nil {
		return maskAny(err)
	}
	return nil
}
