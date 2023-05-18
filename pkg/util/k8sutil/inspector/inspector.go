//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package inspector

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"

	arangoInformer "github.com/arangodb/kube-arangodb/pkg/generated/informers/externalversions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/anonymous"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangoclustersynchronization"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangodeployment"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangomember"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangotask"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/endpoints"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/mods"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/node"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolume"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolumeclaim"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/poddisruptionbudget"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/refresh"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/server"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/serviceaccount"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/servicemonitor"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

type Object interface {
	meta.Object

	GroupVersionKind() schema.GroupVersionKind
}

type Inspector interface {
	SetClient(k kclient.Client)
	Client() kclient.Client

	Namespace() string

	Initialised() bool

	anonymous.Impl

	IsOwnerOf(ctx context.Context, owner Object, obj meta.Object) bool

	AnonymousObjects() []anonymous.Impl

	refresh.Inspector
	throttle.Inspector

	pod.Inspector
	secret.Inspector
	persistentvolumeclaim.Inspector
	service.Inspector
	poddisruptionbudget.Inspector
	servicemonitor.Inspector
	serviceaccount.Inspector
	arangomember.Inspector
	server.Inspector
	endpoints.Inspector

	arangodeployment.Inspector

	node.Inspector
	persistentvolume.Inspector
	arangoclustersynchronization.Inspector
	arangotask.Inspector

	mods.Mods

	RegisterInformers(k8s informers.SharedInformerFactory, arango arangoInformer.SharedInformerFactory)
}
