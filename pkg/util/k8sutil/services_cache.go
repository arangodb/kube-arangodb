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

package k8sutil

import (
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// servicesCache implements a cached version of a ServiceInterface.
// It is NOT go-routine safe.
type servicesCache struct {
	cli   corev1.ServiceInterface
	cache []v1.Service
}

// NewServiceCache creates a cached version of the given ServiceInterface.
func NewServiceCache(cli corev1.ServiceInterface) ServiceInterface {
	return &servicesCache{cli: cli}
}

var (
	serviceGroupResource = schema.GroupResource{
		Group:    v1.GroupName,
		Resource: "Service",
	}
)

func (sc *servicesCache) Create(s *v1.Service) (*v1.Service, error) {
	sc.cache = nil
	result, err := sc.cli.Create(s)
	if err != nil {
		return nil, maskAny(err)
	}
	return result, nil
}

func (sc *servicesCache) Update(s *v1.Service) (*v1.Service, error) {
	sc.cache = nil
	result, err := sc.cli.Update(s)
	if err != nil {
		return nil, maskAny(err)
	}
	return result, nil
}

func (sc *servicesCache) Delete(name string, options *metav1.DeleteOptions) error {
	sc.cache = nil
	if err := sc.cli.Delete(name, options); err != nil {
		return maskAny(err)
	}
	return nil
}

func (sc *servicesCache) Get(name string, options metav1.GetOptions) (*v1.Service, error) {
	if sc.cache == nil {
		list, err := sc.cli.List(metav1.ListOptions{})
		if err != nil {
			return nil, maskAny(err)
		}
		sc.cache = list.Items
	}
	for _, s := range sc.cache {
		if s.GetName() == name {
			return &s, nil
		}
	}
	return nil, maskAny(apierrors.NewNotFound(serviceGroupResource, name))
}
