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

package operator

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
)

func Test_Operator_InformerProcessing(t *testing.T) {
	// Arrange
	name := string(uuid.NewUUID())
	o := NewOperator(name, name, name)
	size := 64

	objects := make([]string, size)
	for id := range objects {
		objects[id] = randomString(10)
	}

	m, i := mockSimpleObject(name, true)
	require.NoError(t, o.RegisterHandler(m))

	client := fake.NewSimpleClientset()
	informer := informers.NewSharedInformerFactory(client, 0)

	require.NoError(t, o.RegisterInformer(informer.Core().V1().Pods().Informer(), "", "v1", "pods"))
	require.NoError(t, o.RegisterStarter(informer))

	stopCh := make(chan struct{})

	// Act
	require.NoError(t, o.Start(4, stopCh))

	for _, name := range objects {
		_, err := client.CoreV1().Pods("test").Create(context.Background(), &core.Pod{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "Pod",
			},
			ObjectMeta: meta.ObjectMeta{
				Name: name,
			},
		}, meta.CreateOptions{})
		require.NoError(t, err)
	}

	// Assert
	res := waitForItems(t, i, size)
	assert.Len(t, res, size)

	time.Sleep(50 * time.Millisecond)
	assert.Len(t, i, 0)

	close(stopCh)
	close(i)
}

func Test_Operator_MultipleInformers(t *testing.T) {
	// Arrange
	name := string(uuid.NewUUID())
	o := NewOperator(name, name, name)
	size := 16

	objects := make([]string, size)
	for id := range objects {
		objects[id] = randomString(10)
	}

	m, i := mockSimpleObject(name, true)
	require.NoError(t, o.RegisterHandler(m))

	client := fake.NewSimpleClientset()
	informer := informers.NewSharedInformerFactory(client, 0)

	require.NoError(t, o.RegisterInformer(informer.Core().V1().Pods().Informer(), "", "v1", "pods"))
	require.NoError(t, o.RegisterInformer(informer.Core().V1().Nodes().Informer(), "", "v1", "nodes"))
	require.NoError(t, o.RegisterStarter(informer))

	stopCh := make(chan struct{})

	// Act
	require.NoError(t, o.Start(4, stopCh))

	for _, name := range objects {
		_, err := client.CoreV1().Pods("test").Create(context.Background(), &core.Pod{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "Pod",
			},
			ObjectMeta: meta.ObjectMeta{
				Name: name,
			},
		}, meta.CreateOptions{})
		require.NoError(t, err)

		_, err = client.CoreV1().Nodes().Create(context.Background(), &core.Node{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "Node",
			},
			ObjectMeta: meta.ObjectMeta{
				Name: name,
			},
		}, meta.CreateOptions{})
		require.NoError(t, err)
	}

	// Assert
	res := waitForItems(t, i, size*2)
	assert.Len(t, res, size*2)

	time.Sleep(50 * time.Millisecond)
	assert.Len(t, i, 0)

	close(stopCh)
	close(i)
}

func Test_Operator_MultipleInformers_IgnoredTypes(t *testing.T) {
	// Arrange
	name := string(uuid.NewUUID())
	o := NewOperator(name, name, name)
	size := 16

	objects := make([]string, size)
	for id := range objects {
		objects[id] = randomString(10)
	}

	m, i := mockSimpleObject(name, true)
	require.NoError(t, o.RegisterHandler(m))

	client := fake.NewSimpleClientset()
	informer := informers.NewSharedInformerFactory(client, 0)

	require.NoError(t, o.RegisterInformer(informer.Core().V1().Pods().Informer(), "", "v1", "pods"))
	require.NoError(t, o.RegisterStarter(informer))

	stopCh := make(chan struct{})

	// Act
	require.NoError(t, o.Start(4, stopCh))

	for _, name := range objects {
		_, err := client.CoreV1().Pods("test").Create(context.Background(), &core.Pod{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "Pod",
			},
			ObjectMeta: meta.ObjectMeta{
				Name: name,
			},
		}, meta.CreateOptions{})
		require.NoError(t, err)

		_, err = client.CoreV1().Nodes().Create(context.Background(), &core.Node{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "Node",
			},
			ObjectMeta: meta.ObjectMeta{
				Name: name,
			},
		}, meta.CreateOptions{})
		require.NoError(t, err)
	}

	// Assert
	res := waitForItems(t, i, size)
	assert.Len(t, res, size)

	time.Sleep(50 * time.Millisecond)
	assert.Len(t, i, 0)

	close(stopCh)
	close(i)
}

func Test_Operator_MultipleInformers_MultipleHandlers(t *testing.T) {
	// Arrange
	name := string(uuid.NewUUID())
	o := NewOperator(name, name, name)
	size := 16

	objects := make([]string, size)
	for id := range objects {
		objects[id] = randomString(10)
	}

	mp, ip := mockSimpleObjectFunc(name, func(item operation.Item) bool {
		return item.Kind == "pods"
	})
	require.NoError(t, o.RegisterHandler(mp))

	mn, in := mockSimpleObjectFunc(name, func(item operation.Item) bool {
		return item.Kind == "nodes"
	})
	require.NoError(t, o.RegisterHandler(mn))

	ms, is := mockSimpleObjectFunc(name, func(item operation.Item) bool {
		return item.Kind == "services"
	})
	require.NoError(t, o.RegisterHandler(ms))

	md, id := mockSimpleObject(name, true)
	require.NoError(t, o.RegisterHandler(md))

	client := fake.NewSimpleClientset()
	informer := informers.NewSharedInformerFactory(client, 0)

	require.NoError(t, o.RegisterInformer(informer.Core().V1().Pods().Informer(), "", "v1", "pods"))
	require.NoError(t, o.RegisterInformer(informer.Core().V1().Nodes().Informer(), "", "v1", "nodes"))
	require.NoError(t, o.RegisterInformer(informer.Core().V1().Services().Informer(), "", "v1", "services"))
	require.NoError(t, o.RegisterInformer(informer.Core().V1().ServiceAccounts().Informer(), "", "v1", "sa"))
	require.NoError(t, o.RegisterStarter(informer))

	stopCh := make(chan struct{})

	// Act
	require.NoError(t, o.Start(4, stopCh))

	for _, name := range objects {
		_, err := client.CoreV1().Pods("test").Create(context.Background(), &core.Pod{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "Pod",
			},
			ObjectMeta: meta.ObjectMeta{
				Name: name,
			},
		}, meta.CreateOptions{})
		require.NoError(t, err)

		_, err = client.CoreV1().Nodes().Create(context.Background(), &core.Node{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "Node",
			},
			ObjectMeta: meta.ObjectMeta{
				Name: name,
			},
		}, meta.CreateOptions{})
		require.NoError(t, err)

		_, err = client.CoreV1().Services("test").Create(context.Background(), &core.Service{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "Service",
			},
			ObjectMeta: meta.ObjectMeta{
				Name: name,
			},
		}, meta.CreateOptions{})
		require.NoError(t, err)

		_, err = client.CoreV1().ServiceAccounts("test").Create(context.Background(), &core.ServiceAccount{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
			},
			ObjectMeta: meta.ObjectMeta{
				Name: name,
			},
		}, meta.CreateOptions{})
		require.NoError(t, err)
	}

	// Assert
	assert.Len(t, waitForItems(t, ip, size), size)
	assert.Len(t, waitForItems(t, in, size), size)
	assert.Len(t, waitForItems(t, is, size), size)
	assert.Len(t, waitForItems(t, id, size), size)

	time.Sleep(50 * time.Millisecond)
	assert.Len(t, ip, 0)
	assert.Len(t, in, 0)
	assert.Len(t, is, 0)
	assert.Len(t, id, 0)

	close(stopCh)
	close(ip)
	close(in)
	close(is)
	close(id)
}

func Test_Operator_InformerProcessing_Namespaced(t *testing.T) {
	// Arrange
	name := string(uuid.NewUUID())
	o := NewOperator(name, name, name)
	size := 16

	objects := make([]string, size)
	for id := range objects {
		objects[id] = randomString(10)
	}

	m, i := mockSimpleObject(name, true)
	require.NoError(t, o.RegisterHandler(m))

	client := fake.NewSimpleClientset()
	informer := informers.NewSharedInformerFactoryWithOptions(client, 0, informers.WithNamespace(objects[0]))

	require.NoError(t, o.RegisterInformer(informer.Core().V1().Pods().Informer(), "", "v1", "pods"))
	require.NoError(t, o.RegisterStarter(informer))

	stopCh := make(chan struct{})

	// Act
	require.NoError(t, o.Start(4, stopCh))

	for _, name := range objects {
		_, err := client.CoreV1().Pods(name).Create(context.Background(), &core.Pod{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "Pod",
			},
			ObjectMeta: meta.ObjectMeta{
				Name:      name,
				Namespace: name,
			},
		}, meta.CreateOptions{})
		require.NoError(t, err)
	}

	// Assert
	res := waitForItems(t, i, 1)
	assert.Len(t, res, 1)

	time.Sleep(50 * time.Millisecond)
	assert.Len(t, i, 0)

	close(stopCh)
	close(i)
}
