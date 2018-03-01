package mocks

import (
	"k8s.io/client-go/kubernetes/typed/core/v1"
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
