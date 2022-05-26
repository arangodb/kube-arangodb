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

package v2alpha1

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func Test_ACS_KubeConfigSpec(t *testing.T) {
	test := func(t *testing.T, spec *ArangoClusterSynchronizationKubeConfigSpec, error error) {
		err := spec.Validate()

		if error != nil {
			require.EqualError(t, err, error.Error())
		} else {
			require.NoError(t, err)
		}
	}

	type testCase struct {
		spec  *ArangoClusterSynchronizationKubeConfigSpec
		error string
	}

	testCases := map[string]testCase{
		"Nil": {
			error: "KubeConfig Spec cannot be nil",
		},
		"Empty": {
			spec:  &ArangoClusterSynchronizationKubeConfigSpec{},
			error: "Received 3 errors: secretName: Name '' is not a valid resource name, secretKey: Name '' is not a valid resource name, namespace: Name '' is not a valid resource name",
		},
		"Missing key & NS": {
			spec: &ArangoClusterSynchronizationKubeConfigSpec{
				SecretName: "secret",
			},
			error: "Received 2 errors: secretKey: Name '' is not a valid resource name, namespace: Name '' is not a valid resource name",
		},
		"Missing NS": {
			spec: &ArangoClusterSynchronizationKubeConfigSpec{
				SecretName: "secret",
				SecretKey:  "key",
			},
			error: "Received 1 errors: namespace: Name '' is not a valid resource name",
		},
		"Valid": {
			spec: &ArangoClusterSynchronizationKubeConfigSpec{
				SecretName: "secret",
				SecretKey:  "key",
				Namespace:  "ns",
			},
		},
		"Invalid": {
			spec: &ArangoClusterSynchronizationKubeConfigSpec{
				SecretName: "secret_n",
				SecretKey:  "key",
				Namespace:  "ns",
			},
			error: "Received 1 errors: secretName: Name 'secret_n' is not a valid resource name",
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			var err error
			if tc.error != "" {
				err = errors.Errorf(tc.error)
			}
			test(t, tc.spec, err)
		})
	}
}
