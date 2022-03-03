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

package storage

import (
	"context"
	"testing"

	api "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestEnsureDaemonSet tests ensureDaemonSet() method
func TestEnsureDaemonSet(t *testing.T) {
	testNamespace := "testNs"
	testLsName := "testDsName"

	testPodName := "testPodName"
	testImage := "test-image"

	testPullSecrets := []v1.LocalObjectReference{
		{
			Name: "custom-docker",
		},
	}

	ls := &LocalStorage{
		apiObject: &api.ArangoLocalStorage{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testLsName,
				Namespace: testNamespace,
			},
			Spec: api.LocalStorageSpec{},
		},
		deps: Dependencies{
			Client: kclient.NewFakeClient(),
		},
		config: Config{
			Namespace: testNamespace,
			PodName:   testPodName,
		},
		image:            testImage,
		imagePullSecrets: testPullSecrets,
		imagePullPolicy:  v1.PullAlways,
	}

	err := ls.ensureDaemonSet(ls.apiObject)
	require.NoError(t, err)

	// verify if DaemonSet has been created with correct values
	ds, err := ls.deps.Client.Kubernetes().AppsV1().DaemonSets(testNamespace).Get(context.Background(), testLsName, metav1.GetOptions{})
	require.NoError(t, err)

	pod := ds.Spec.Template.Spec

	require.Equal(t, ds.GetName(), testLsName)
	require.Equal(t, pod.ImagePullSecrets, testPullSecrets)
	require.Equal(t, len(pod.Containers), 1)

	c := pod.Containers[0]
	require.Equal(t, c.Image, testImage)
	require.Equal(t, c.ImagePullPolicy, v1.PullAlways)
}
