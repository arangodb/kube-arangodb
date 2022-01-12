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

package mocks

import (
	"k8s.io/apimachinery/pkg/watch"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type coreV1 struct {
	restClient
	secrets map[string]v1.SecretInterface
}

func NewCore() v1.CoreV1Interface {
	return &coreV1{
		secrets: make(map[string]v1.SecretInterface),
	}
}

func nilOrWatch(x interface{}) watch.Interface {
	if s, ok := x.(watch.Interface); ok {
		return s
	}
	return nil
}

func (c *coreV1) ComponentStatuses() v1.ComponentStatusInterface {
	panic("not support")
}

func (c *coreV1) ConfigMaps(namespace string) v1.ConfigMapInterface {
	panic("not support")
}

func (c *coreV1) Endpoints(namespace string) v1.EndpointsInterface {
	panic("not support")
}

func (c *coreV1) Events(namespace string) v1.EventInterface {
	panic("not support")
}

func (c *coreV1) LimitRanges(namespace string) v1.LimitRangeInterface {
	panic("not support")
}

func (c *coreV1) Namespaces() v1.NamespaceInterface {
	panic("not support")
}

func (c *coreV1) Nodes() v1.NodeInterface {
	panic("not support")
}

func (c *coreV1) PersistentVolumes() v1.PersistentVolumeInterface {
	panic("not support")
}

func (c *coreV1) PersistentVolumeClaims(namespace string) v1.PersistentVolumeClaimInterface {
	panic("not support")
}

func (c *coreV1) Pods(namespace string) v1.PodInterface {
	panic("not support")
}

func (c *coreV1) PodTemplates(namespace string) v1.PodTemplateInterface {
	panic("not support")
}

func (c *coreV1) ReplicationControllers(namespace string) v1.ReplicationControllerInterface {
	panic("not support")
}

func (c *coreV1) ResourceQuotas(namespace string) v1.ResourceQuotaInterface {
	panic("not support")
}

func (c *coreV1) Secrets(namespace string) v1.SecretInterface {
	if x, found := c.secrets[namespace]; found {
		return x
	}
	x := NewSecrets()
	c.secrets[namespace] = x
	return x
}

func (c *coreV1) Services(namespace string) v1.ServiceInterface {
	panic("not support")
}

func (c *coreV1) ServiceAccounts(namespace string) v1.ServiceAccountInterface {
	panic("not support")
}
