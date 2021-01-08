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
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getMyImage fetched the docker image from my own pod
func (l *LocalStorage) getMyImage() (string, v1.PullPolicy, error) {
	log := l.deps.Log
	ns := l.config.Namespace

	p, err := l.deps.KubeCli.CoreV1().Pods(ns).Get(l.config.PodName, metav1.GetOptions{})
	if err != nil {
		log.Debug().Err(err).Str("pod-name", l.config.PodName).Msg("Failed to get my own pod")
		return "", "", errors.WithStack(err)
	}

	c := p.Spec.Containers[0]
	return c.Image, c.ImagePullPolicy, nil
}
