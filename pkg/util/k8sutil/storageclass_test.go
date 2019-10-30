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
	"errors"
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/util/retry"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/storage/v1"
	er "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestStorageClassIsDefault(t *testing.T) {
	testCases := []struct {
		Name         string
		StorageClass v1.StorageClass
		IsDefault    bool
	}{
		{
			Name: "Storage class without annotations",
			StorageClass: v1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{},
			},
			IsDefault: false,
		},
		{
			Name: "Storage class with empty annotations",
			StorageClass: v1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			IsDefault: false,
		},
		{
			Name: "Storage class without default",
			StorageClass: v1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						annStorageClassIsDefault: "false",
					},
				},
			},
			IsDefault: false,
		},
		{
			Name: "Storage class with invalid value in annotation",
			StorageClass: v1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						annStorageClassIsDefault: "foo",
					},
				},
			},
			IsDefault: false,
		},
		{
			Name: "Default storage class exits",
			StorageClass: v1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						annStorageClassIsDefault: "true",
					},
				},
			},
			IsDefault: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			result := StorageClassIsDefault(&testCase.StorageClass)
			assert.Equal(t, testCase.IsDefault, result, "StorageClassIsDefault failed. Expected %v, got %v for %#v",
				testCase.IsDefault, result, testCase.StorageClass)
		})
	}
}

func TestPatchStorageClassIsDefault(t *testing.T) {
	// Arrange
	resourceName := "storageclasses"
	testCases := []struct {
		Name              string
		StorageClassName  string
		ExpectedErr       error
		Reactor           func(action k8stesting.Action) (handled bool, ret runtime.Object, err error)
		ReactorActionVerb string
	}{
		{
			Name:             "Set storage class is set to default",
			StorageClassName: "test",
		},
		{
			Name:             "Storage class does not exist",
			StorageClassName: "invalid",
			ExpectedErr:      er.NewNotFound(v1.Resource(resourceName), "invalid"),
		},
		{
			Name:             "Can not get storage class from kubernetes",
			StorageClassName: "test",
			Reactor: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, retry.Permanent(errors.New("test"))
			},
			ReactorActionVerb: "get",
			ExpectedErr:       errors.New("test"),
		},
		{
			Name:             "Can not update storage class",
			StorageClassName: "test",
			Reactor: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, errors.New("test")
			},
			ReactorActionVerb: "update",
			ExpectedErr:       errors.New("test"),
		},
		{
			Name:             "Can not update Storage class due to permanent conflict",
			StorageClassName: "test",
			Reactor: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil,
					retry.Permanent(er.NewConflict(v1.Resource(resourceName), "test", nil))
			},
			ReactorActionVerb: "update",
			ExpectedErr:       er.NewConflict(v1.Resource(resourceName), "test", nil),
		},
	}

	for _, testCase := range testCases {
		//nolint:scopelint
		t.Run(testCase.Name, func(t *testing.T) {
			// Arrange
			var err error

			clientSet := fake.NewSimpleClientset()
			storageSet := clientSet.StorageV1()
			_, err = storageSet.StorageClasses().Create(&v1.StorageClass{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			})
			require.NoError(t, err)

			if testCase.Reactor != nil {
				clientSet.PrependReactor(testCase.ReactorActionVerb, resourceName, testCase.Reactor)
			}

			// Act
			err = PatchStorageClassIsDefault(storageSet, testCase.StorageClassName, true)

			// Assert
			if testCase.ExpectedErr != nil {
				require.EqualError(t, err, testCase.ExpectedErr.Error())
				return
			}

			assert.NoError(t, err)
		})
	}

}
