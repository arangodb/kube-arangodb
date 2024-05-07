//
// DISCLAIMER
//
// Copyright 2023-2024 ArangoDB GmbH, Cologne, Germany
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
	"fmt"

	apps "k8s.io/api/apps/v1"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helpers"
)

func AppsStatefulSet(in *apps.StatefulSet) *apps.StatefulSet {
	return FilterP(in, func(in *apps.StatefulSet) *apps.StatefulSet {
		return &apps.StatefulSet{
			ObjectMeta: ObjectMetaFilter(in.ObjectMeta),
			Spec: Filter(in.Spec, func(in apps.StatefulSetSpec) apps.StatefulSetSpec {
				return apps.StatefulSetSpec{
					Template:        in.Template,
					Replicas:        in.Replicas,
					MinReadySeconds: in.MinReadySeconds,
					Selector:        in.Selector,
					ServiceName:     in.ServiceName,
				}
			}),
		}
	})
}

func AppsStatefulSetChecksum(in *apps.StatefulSet) (string, error) {
	return util.SHA256FromJSON(AppsStatefulSet(in))
}

func AppsStatefulSetRecreate(logger logging.Logger) helpers.Decision[*apps.StatefulSet] {
	return helpers.NewImmutableFields[*apps.StatefulSet](func(a, b *apps.StatefulSet, changes map[string]string) helpers.Action {
		if len(changes) > 0 {
			s := logger

			for k, v := range changes {
				s = s.Str(fmt.Sprintf("field%s", k), v)
			}

			s.Info("Replace of StatefulSet %s required", a.GetName())
			return helpers.ActionReplace
		}

		return helpers.ActionOK
	}).Evaluate()
}
