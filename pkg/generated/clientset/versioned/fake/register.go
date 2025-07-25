//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	analyticsv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/analytics/v1alpha1"
	appsv1 "github.com/arangodb/kube-arangodb/pkg/apis/apps/v1"
	backupv1 "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	databasev1 "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	databasev2alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v2alpha1"
	mlv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1alpha1"
	mlv1beta1 "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1beta1"
	networkingv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1alpha1"
	networkingv1beta1 "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1beta1"
	platformv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	platformv1beta1 "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	replicationv1 "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1"
	replicationv2alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/replication/v2alpha1"
	schedulerv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1"
	schedulerv1beta1 "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	storagev1alpha "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var scheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(scheme)

var localSchemeBuilder = runtime.SchemeBuilder{
	analyticsv1alpha1.AddToScheme,
	appsv1.AddToScheme,
	backupv1.AddToScheme,
	databasev1.AddToScheme,
	databasev2alpha1.AddToScheme,
	mlv1alpha1.AddToScheme,
	mlv1beta1.AddToScheme,
	networkingv1alpha1.AddToScheme,
	networkingv1beta1.AddToScheme,
	platformv1alpha1.AddToScheme,
	platformv1beta1.AddToScheme,
	replicationv1.AddToScheme,
	replicationv2alpha1.AddToScheme,
	schedulerv1alpha1.AddToScheme,
	schedulerv1beta1.AddToScheme,
	storagev1alpha.AddToScheme,
}

// AddToScheme adds all types of this clientset into the given scheme. This allows composition
// of clientsets, like in:
//
//	import (
//	  "k8s.io/client-go/kubernetes"
//	  clientsetscheme "k8s.io/client-go/kubernetes/scheme"
//	  aggregatorclientsetscheme "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset/scheme"
//	)
//
//	kclientset, _ := kubernetes.NewForConfig(c)
//	_ = aggregatorclientsetscheme.AddToScheme(clientsetscheme.Scheme)
//
// After this, RawExtensions in Kubernetes types will serialize kube-aggregator types
// correctly.
var AddToScheme = localSchemeBuilder.AddToScheme

func init() {
	v1.AddToGroupVersion(scheme, schema.GroupVersion{Version: "v1"})
	utilruntime.Must(AddToScheme(scheme))
}
