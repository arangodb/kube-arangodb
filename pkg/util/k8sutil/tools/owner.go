package tools

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type OwnerRefObj interface {
	GetOwnerReferences() []meta.OwnerReference
}

func IsOwnerFromRef(owner, ref meta.OwnerReference) bool {
	return owner.UID == ref.UID
}

func IsOwner(ref meta.OwnerReference, object OwnerRefObj) bool {
	for _, ownerRef := range object.GetOwnerReferences() {
		if IsOwnerFromRef(ref, ownerRef) {
			return true
		}
	}

	return false
}

func IsChildResource(kind, name, namespace string, resource meta.Object) bool {
	if resource == nil {
		return false
	}

	if namespace != resource.GetNamespace() {
		return false
	}

	ownerRef := resource.GetOwnerReferences()

	if len(ownerRef) == 0 {
		return false
	}

	for _, owner := range ownerRef {
		if owner.Kind != kind {
			continue
		}

		if owner.Name != name {
			continue
		}

		return true
	}

	return false
}
