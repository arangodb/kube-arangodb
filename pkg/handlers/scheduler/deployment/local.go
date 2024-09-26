//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package deployment

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arangodb/kube-arangodb/pkg/apis/scheduler"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
)

func GVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   Group(),
		Version: Version(),
		Kind:    Kind(),
	}
}

func Kind() string {
	return scheduler.DeploymentResourceKind
}

func Group() string {
	return schedulerApi.SchemeGroupVersion.Group
}

func Version() string {
	return schedulerApi.SchemeGroupVersion.Version
}
