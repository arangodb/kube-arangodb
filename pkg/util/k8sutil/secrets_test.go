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

package k8sutil

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/mocks"
)

// TestValidateEncryptionKeySecret tests ValidateEncryptionKeySecret.
func TestValidateEncryptionKeySecret(t *testing.T) {
	cli := mocks.NewCore()
	secrets := cli.Secrets("ns")

	// Prepare mock
	m := mocks.AsMock(secrets)
	m.On("Get", "good", mock.Anything).Return(&v1.Secret{
		Data: map[string][]byte{
			"key": make([]byte, 32),
		},
	}, nil)
	m.On("Get", "no-key", mock.Anything).Return(&v1.Secret{
		Data: map[string][]byte{
			"foo": make([]byte, 32),
		},
	}, nil)
	m.On("Get", "short-key", mock.Anything).Return(&v1.Secret{
		Data: map[string][]byte{
			"key": make([]byte, 31),
		},
	}, nil)
	m.On("Get", "notfound", mock.Anything).Return(nil, apierrors.NewNotFound(schema.GroupResource{}, "notfound"))

	assert.NoError(t, ValidateEncryptionKeySecret(secrets, "good"))
	assert.Error(t, ValidateEncryptionKeySecret(secrets, "no-key"))
	assert.Error(t, ValidateEncryptionKeySecret(secrets, "short-key"))
	assert.True(t, IsNotFound(ValidateEncryptionKeySecret(secrets, "notfound")))
}

// TestCreateEncryptionKeySecret tests CreateEncryptionKeySecret
func TestCreateEncryptionKeySecret(t *testing.T) {
	cli := mocks.NewCore()
	secrets := cli.Secrets("ns")

	// Prepare mock
	m := mocks.AsMock(cli.Secrets("ns"))
	m.On("Create", mock.Anything).Run(func(arg mock.Arguments) {
		s := arg.Get(0).(*v1.Secret)
		if s.GetName() == "good" {
			assert.Equal(t, make([]byte, 32), s.Data["key"])
		} else {
			assert.Fail(t, "Unexpected secret named '%s'", s.GetName())
		}
	}).Return(nil, nil)

	key := make([]byte, 32)
	assert.NoError(t, CreateEncryptionKeySecret(secrets, "good", key))
	key = make([]byte, 31)
	assert.Error(t, CreateEncryptionKeySecret(secrets, "short-key", key))
}

// TestGetTokenSecret tests GetTokenSecret.
func TestGetTokenSecret(t *testing.T) {
	cli := mocks.NewCore()
	secrets := cli.Secrets("ns")

	// Prepare mock
	m := mocks.AsMock(cli.Secrets("ns"))
	m.On("Get", "good", mock.Anything).Return(&v1.Secret{
		Data: map[string][]byte{
			"token": []byte("foo"),
		},
	}, nil)
	m.On("Get", "no-token", mock.Anything).Return(&v1.Secret{
		Data: map[string][]byte{
			"foo": make([]byte, 13),
		},
	}, nil)
	m.On("Get", "notfound", mock.Anything).Return(nil, apierrors.NewNotFound(schema.GroupResource{}, "notfound"))

	token, err := GetTokenSecret(secrets, "good")
	assert.NoError(t, err)
	assert.Equal(t, token, "foo")
	_, err = GetTokenSecret(secrets, "no-token")
	assert.Error(t, err)
	_, err = GetTokenSecret(secrets, "notfound")
	assert.True(t, IsNotFound(err))
}

// TestCreateTokenSecret tests CreateTokenSecret
func TestCreateTokenSecret(t *testing.T) {
	cli := mocks.NewCore()
	secrets := cli.Secrets("ns")

	// Prepare mock
	m := mocks.AsMock(cli.Secrets("ns"))
	m.On("Create", mock.Anything).Run(func(arg mock.Arguments) {
		s := arg.Get(0).(*v1.Secret)
		if s.GetName() == "good" {
			assert.Equal(t, []byte("token"), s.Data["token"])
		} else if s.GetName() == "with-owner" {
			assert.Len(t, s.GetOwnerReferences(), 1)
		} else {
			assert.Fail(t, "Unexpected secret named '%s'", s.GetName())
		}
	}).Return(nil, nil)

	assert.NoError(t, CreateTokenSecret(secrets, "good", "token", nil))
	assert.NoError(t, CreateTokenSecret(secrets, "with-owner", "token", &metav1.OwnerReference{}))
}
