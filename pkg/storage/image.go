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

package storage

import (
	"context"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// getMyImage fetched the docker image from my own pod
func (l *LocalStorage) getMyImage() (string, core.PullPolicy, []core.LocalObjectReference, error) {
	return l.getImage(l.config.Namespace, l.config.PodName, l.deps.Client.Kubernetes())
}

func (l *LocalStorage) getImage(ns, name string, client kubernetes.Interface) (string, core.PullPolicy, []core.LocalObjectReference, error) {
	p, err := client.CoreV1().Pods(ns).Get(context.Background(), name, meta.GetOptions{})
	if err != nil {
		l.log.Err(err).Str("pod-name", name).Debug("Failed to get my own pod")
		return "", "", nil, errors.WithStack(err)
	}

	c := p.Spec.Containers[0]

	return c.Image, c.ImagePullPolicy, p.Spec.ImagePullSecrets, nil
}
