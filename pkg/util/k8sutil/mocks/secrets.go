package mocks

import (
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type secrets struct {
	m map[string]v1.Secret
}

func NewSecrets() corev1.SecretInterface {
	return &secrets{
		m: make(map[string]v1.Secret),
	}
}

func (s *secrets) Create(x *v1.Secret) (*v1.Secret, error) {
	if _, found := s.m[x.GetName()]; found {
		return nil, apierrors.NewAlreadyExists(schema.GroupResource{}, x.GetName())
	}
	s.m[x.GetName()] = *x
	return x, nil
}

func (s *secrets) Update(x *v1.Secret) (*v1.Secret, error) {
	if _, found := s.m[x.GetName()]; !found {
		return nil, apierrors.NewNotFound(schema.GroupResource{}, x.GetName())
	}
	s.m[x.GetName()] = *x
	return x, nil
}

func (s *secrets) Delete(name string, options *meta_v1.DeleteOptions) error {
	if _, found := s.m[name]; found {
		delete(s.m, name)
		return nil
	}
	return apierrors.NewNotFound(schema.GroupResource{}, name)
}

func (s *secrets) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	panic("not support")
}

func (s *secrets) Get(name string, options meta_v1.GetOptions) (*v1.Secret, error) {
	x, found := s.m[name]
	if !found {
		return nil, apierrors.NewNotFound(schema.GroupResource{}, name)
	}
	return &x, nil
}

func (s *secrets) List(opts meta_v1.ListOptions) (*v1.SecretList, error) {
	panic("not support")
}

func (s *secrets) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	panic("not support")
}

func (s *secrets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Secret, err error) {
	panic("not support")
}
