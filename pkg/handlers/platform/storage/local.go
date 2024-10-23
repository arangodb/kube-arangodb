package storage

import (
	"github.com/arangodb/kube-arangodb/pkg/apis/ml"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
)

func Kind() string {
	return ml.ArangoMLStorageResourceKind
}

func Group() string {
	return platformApi.SchemeGroupVersion.Group
}

func Version() string {
	return platformApi.SchemeGroupVersion.Version
}
