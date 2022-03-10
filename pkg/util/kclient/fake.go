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

package kclient

import (
	"sync"

	versionedFake "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/fake"
	monitoringFake "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/fake"
	apiextensionsclientFake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	"k8s.io/apimachinery/pkg/runtime"
	kubernetesFake "k8s.io/client-go/kubernetes/fake"
)

func NewFakeClient() Client {
	return NewStaticClient(kubernetesFake.NewSimpleClientset(), apiextensionsclientFake.NewSimpleClientset(), versionedFake.NewSimpleClientset(), monitoringFake.NewSimpleClientset())
}

type FakeClientBuilder interface {
	Kubernetes(objects ...runtime.Object) FakeClientBuilder
	KubernetesExtensions(objects ...runtime.Object) FakeClientBuilder
	Arango(objects ...runtime.Object) FakeClientBuilder
	Monitoring(objects ...runtime.Object) FakeClientBuilder

	Client() Client
}

func NewFakeClientBuilder() FakeClientBuilder {
	return &fakeClientBuilder{}
}

type fakeClientBuilder struct {
	lock sync.Mutex

	kubernetes           []runtime.Object
	kubernetesExtensions []runtime.Object
	arango               []runtime.Object
	monitoring           []runtime.Object
}

func (f *fakeClientBuilder) Kubernetes(objects ...runtime.Object) FakeClientBuilder {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.kubernetes = append(f.kubernetes, objects...)

	return f
}

func (f *fakeClientBuilder) KubernetesExtensions(objects ...runtime.Object) FakeClientBuilder {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.kubernetesExtensions = append(f.kubernetesExtensions, objects...)

	return f
}

func (f *fakeClientBuilder) Arango(objects ...runtime.Object) FakeClientBuilder {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.arango = append(f.arango, objects...)

	return f
}

func (f *fakeClientBuilder) Monitoring(objects ...runtime.Object) FakeClientBuilder {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.monitoring = append(f.monitoring, objects...)

	return f
}

func (f *fakeClientBuilder) Client() Client {
	return NewStaticClient(
		kubernetesFake.NewSimpleClientset(f.kubernetes...),
		apiextensionsclientFake.NewSimpleClientset(f.kubernetesExtensions...),
		versionedFake.NewSimpleClientset(f.arango...),
		monitoringFake.NewSimpleClientset(f.monitoring...))
}
