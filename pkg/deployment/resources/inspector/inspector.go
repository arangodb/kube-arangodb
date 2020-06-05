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
// Author Adam Janikowski
//

package inspector

import (
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func NewInspector(k kubernetes.Interface, namespace string) (Inspector, error) {
	pods, err := podsToMap(k, namespace)
	if err != nil {
		return nil, err
	}

	secrets, err := secretsToMap(k, namespace)
	if err != nil {
		return nil, err
	}

	pvcs, err := pvcsToMap(k, namespace)
	if err != nil {
		return nil, err
	}

	services, err := servicesToMap(k, namespace)
	if err != nil {
		return nil, err
	}

	return NewInspectorFromData(pods, secrets, pvcs, services), nil
}

func NewEmptyInspector() Inspector {
	return NewInspectorFromData(nil, nil, nil, nil)
}

func NewInspectorFromData(pods map[string]*core.Pod, secrets map[string]*core.Secret, pvcs map[string]*core.PersistentVolumeClaim, services map[string]*core.Service) Inspector {
	return &inspector{
		pods:     pods,
		secrets:  secrets,
		pvcs:     pvcs,
		services: services,
	}
}

type Inspector interface {
	Pod(name string) (*core.Pod, bool)
	IteratePods(action PodAction, filters ...PodFilter) error

	Secret(name string) (*core.Secret, bool)
	IterateSecrets(action SecretAction, filters ...SecretFilter) error

	PersistentVolumeClaim(name string) (*core.PersistentVolumeClaim, bool)
	IteratePersistentVolumeClaims(action PersistentVolumeClaimAction, filters ...PersistentVolumeClaimFilter) error

	Service(name string) (*core.Service, bool)
	IterateServices(action ServiceAction, filters ...ServiceFilter) error
}

type inspector struct {
	pods     map[string]*core.Pod
	secrets  map[string]*core.Secret
	pvcs     map[string]*core.PersistentVolumeClaim
	services map[string]*core.Service
}
