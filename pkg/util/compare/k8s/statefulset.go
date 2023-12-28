//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package k8s

import (
	apps "k8s.io/api/apps/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func ChecksumStatefulSet(s *apps.StatefulSet) (string, error) {
	return checksumStatefulSetSpec(&s.Spec)
}

func checksumStatefulSetSpec(s *apps.StatefulSetSpec) (string, error) {
	parts := map[string]interface{}{
		"replicas":        s.Replicas,
		"serviceName":     s.ServiceName,
		"minReadySeconds": s.MinReadySeconds,
		"selector":        s.Selector,
		"template":        s.Template,
		// add here more fields when needed
	}
	return util.SHA256FromJSON(parts)
}
