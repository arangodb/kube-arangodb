//
// DISCLAIMER
//
// Copyright 2023-2025 ArangoDB GmbH, Cologne, Germany
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

package v1alpha1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arangodb/kube-arangodb/pkg/apis/ml"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
)

const (
	ArangoMLVersion = string(utilConstants.VersionV1Alpha1)
)

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme

	SchemeGroupVersion = schema.GroupVersion{Group: ml.ArangoMLGroupName, Version: ArangoMLVersion}
)

// Resource gets an ArangoCluster GroupResource for a specified resource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// addKnownTypes adds the set of types defined in this package to the supplied scheme.
func addKnownTypes(s *runtime.Scheme) error {
	s.AddKnownTypes(SchemeGroupVersion,
		&ArangoMLStorage{},
		&ArangoMLStorageList{},
		&ArangoMLExtension{},
		&ArangoMLExtensionList{},
		&ArangoMLBatchJob{},
		&ArangoMLBatchJobList{},
		&ArangoMLCronJob{},
		&ArangoMLCronJobList{})
	meta.AddToGroupVersion(s, SchemeGroupVersion)
	return nil
}
