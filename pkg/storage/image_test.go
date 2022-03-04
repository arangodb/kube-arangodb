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

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestGetMyImage tests getMyImage() method
func TestGetMyImage(t *testing.T) {
	testNamespace := "testNs"
	testPodName := "testPodname"
	testImage := "test-image"
	testPullSecrets := []v1.LocalObjectReference{
		{
			Name: "custom-docker",
		},
	}

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testPodName,
			Namespace: testNamespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            "test",
					Image:           testImage,
					ImagePullPolicy: v1.PullAlways,
				},
			},
			ImagePullSecrets: testPullSecrets,
		},
	}

	ls := &LocalStorage{
		deps: Dependencies{
			Client: kclient.NewFakeClient(),
		},
		config: Config{
			Namespace: testNamespace,
			PodName:   testPodName,
		},
	}

	// prepare mock
	if _, err := ls.deps.Client.Kubernetes().CoreV1().Pods(testNamespace).Create(context.Background(), &pod, metav1.CreateOptions{}); err != nil {
		require.NoError(t, err)
	}

	image, pullPolicy, pullSecrets, err := ls.getMyImage()
	require.NoError(t, err)
	require.Equal(t, image, testImage)
	require.Equal(t, pullPolicy, v1.PullAlways)
	require.Equal(t, pullSecrets, testPullSecrets)
}
