package mocks

import (
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type SecretInterface interface {
	corev1.SecretInterface
	MockGetter
}

type secrets struct {
	mock.Mock
}

func NewSecrets() SecretInterface {
	return &secrets{}
}

func nilOrSecret(x interface{}) *v1.Secret {
	if s, ok := x.(*v1.Secret); ok {
		return s
	}
	return nil
}

func nilOrSecretList(x interface{}) *v1.SecretList {
	if s, ok := x.(*v1.SecretList); ok {
		return s
	}
	return nil
}

func (s *secrets) AsMock() *mock.Mock {
	return &s.Mock
}

func (s *secrets) Create(x *v1.Secret) (*v1.Secret, error) {
	args := s.Called(x)
	return nilOrSecret(args.Get(0)), args.Error(1)
}

func (s *secrets) Update(x *v1.Secret) (*v1.Secret, error) {
	args := s.Called(x)
	return nilOrSecret(args.Get(0)), args.Error(1)
}

func (s *secrets) Delete(name string, options *meta_v1.DeleteOptions) error {
	args := s.Called(name, options)
	return args.Error(0)
}

func (s *secrets) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	args := s.Called(options, listOptions)
	return args.Error(0)
}

func (s *secrets) Get(name string, options meta_v1.GetOptions) (*v1.Secret, error) {
	args := s.Called(name, options)
	return nilOrSecret(args.Get(0)), args.Error(1)
}

func (s *secrets) List(opts meta_v1.ListOptions) (*v1.SecretList, error) {
	args := s.Called(opts)
	return nilOrSecretList(args.Get(0)), args.Error(1)
}

func (s *secrets) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	args := s.Called(opts)
	return nilOrWatch(args.Get(0)), args.Error(1)
}

func (s *secrets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Secret, err error) {
	args := s.Called(name, pt, data, subresources)
	return nilOrSecret(args.Get(0)), args.Error(1)
}
