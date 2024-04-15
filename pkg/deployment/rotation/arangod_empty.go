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

package rotation

import (
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/compare"
)

func compareAndAssignEmptyField[T interface{}](spec, status *T) (*T, bool, error) {
	if equal, err := util.CompareJSON(spec, status); err != nil {
		return nil, false, err
	} else if !equal {
		if equal, err := util.CompareJSONP(spec, status); err != nil {
			return nil, false, err
		} else if equal {
			return spec, true, nil
		}
	}

	return nil, false, nil
}

func comparePodEmptyFields(_ api.DeploymentSpec, _ api.ServerGroup, spec, status *core.PodTemplateSpec) compare.Func {
	return func(builder api.ActionBuilder) (mode compare.Mode, plan api.Plan, e error) {
		if obj, replace, err := compareAndAssignEmptyField(spec.Spec.SecurityContext, status.Spec.SecurityContext); err != nil {
			e = err
			return
		} else if replace {
			mode = mode.And(compare.SilentRotation)
			status.Spec.SecurityContext = obj.DeepCopy()
		}
		if equal, err := util.CompareJSON(spec.Spec.SecurityContext, status.Spec.SecurityContext); err != nil {
			e = err
			return
		} else if !equal {
			if equal, err := util.CompareJSONP(spec.Spec.SecurityContext, status.Spec.SecurityContext); err != nil {
				e = err
				return
			} else if equal {
				mode = mode.And(compare.SilentRotation)
				status.Spec.SecurityContext = spec.Spec.SecurityContext.DeepCopy()
			}
		}
		return
	}
}
