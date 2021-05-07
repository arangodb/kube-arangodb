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

package mocks

import (
	"context"

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

func (s *secrets) Create(_ context.Context, x *v1.Secret, _ meta_v1.CreateOptions) (*v1.Secret, error) {
	args := s.Called(x)
	return nilOrSecret(args.Get(0)), args.Error(1)
}

func (s *secrets) Update(_ context.Context, x *v1.Secret, _ meta_v1.UpdateOptions) (*v1.Secret, error) {
	args := s.Called(x)
	return nilOrSecret(args.Get(0)), args.Error(1)
}

func (s *secrets) Delete(_ context.Context, name string, options meta_v1.DeleteOptions) error {
	args := s.Called(name, options)
	return args.Error(0)
}

func (s *secrets) DeleteCollection(_ context.Context, options meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	args := s.Called(options, listOptions)
	return args.Error(0)
}

func (s *secrets) Get(_ context.Context, name string, options meta_v1.GetOptions) (*v1.Secret, error) {
	args := s.Called(name, options)
	return nilOrSecret(args.Get(0)), args.Error(1)
}

func (s *secrets) List(_ context.Context, opts meta_v1.ListOptions) (*v1.SecretList, error) {
	args := s.Called(opts)
	return nilOrSecretList(args.Get(0)), args.Error(1)
}

func (s *secrets) Watch(_ context.Context, opts meta_v1.ListOptions) (watch.Interface, error) {
	args := s.Called(opts)
	return nilOrWatch(args.Get(0)), args.Error(1)
}

func (s *secrets) Patch(_ context.Context, name string, pt types.PatchType, data []byte,
	options meta_v1.PatchOptions, subresources ...string) (result *v1.Secret, err error) {
	args := s.Called(name, pt, data, subresources)
	return nilOrSecret(args.Get(0)), args.Error(1)
}
