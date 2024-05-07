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

	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helpers"
)

func CoreService(in *core.Service) *core.Service {
	return FilterP(in, func(in *core.Service) *core.Service {
		return &core.Service{
			ObjectMeta: ObjectMetaFilter(in.ObjectMeta),
			Spec: Filter(in.Spec, func(in core.ServiceSpec) core.ServiceSpec {
				return core.ServiceSpec{
					Type:     in.Type,
					Ports:    in.Ports,
					Selector: in.Selector,
				}
			}),
		}
	})
}

func CoreServiceChecksum(in *core.Service) (string, error) {
	return util.SHA256FromJSON(CoreService(in))
}

func CoreServiceImmutableRecreate(logger logging.Logger) helpers.Decision[*core.Service] {
	return helpers.NewImmutableFields[*core.Service](func(a, b *core.Service, changes map[string]string) helpers.Action {
		if len(changes) > 0 {
			s := logger

			for k, v := range changes {
				s = s.Str(fmt.Sprintf("field%s", k), v)
			}

			s.Info("Replace of Service %s required", a.GetName())
			return helpers.ActionReplace
		}

		return helpers.ActionOK
	}).Evaluate(
		helpers.SubCompare[*core.Service, core.ServiceSpec](".spec", func(in *core.Service) core.ServiceSpec {
			if in == nil {
				return core.ServiceSpec{}
			}
			return in.Spec
		}, helpers.FromSimpleCompareStack(".type", func(a, b core.ServiceSpec) (string, bool) {
			return "ServiceType change requires object recreation", a.Type != b.Type
		})),
	)
}
