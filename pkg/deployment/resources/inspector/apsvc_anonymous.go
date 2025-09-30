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

package inspector

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/anonymous"
	inspectorConstants "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
)

func (p *arangoPlatformServicesInspector) Anonymous(gvk schema.GroupVersionKind) (anonymous.Interface, bool) {
	g := inspectorConstants.ArangoPlatformServiceGKv1Beta1()

	if g.Kind == gvk.Kind && g.Group == gvk.Group {
		switch gvk.Version {
		case inspectorConstants.ArangoPlatformServiceVersionV1Beta1, DefaultVersion:
			if p.v1beta1 == nil || p.v1beta1.err != nil {
				return nil, false
			}
			return anonymous.NewAnonymous[*platformApi.ArangoPlatformService](g, p.state.arangoPlatformServices.v1beta1, p.state.ArangoPlatformServiceModInterface().V1Beta1()), true
		}
	}

	return nil, false
}
