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

package poddisruptionbudget

import (
	policy "k8s.io/api/policy/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
)

func List(filter ...generic.Filter[*policy.PodDisruptionBudget]) generic.ExtractorList[*policy.PodDisruptionBudgetList, *policy.PodDisruptionBudget] {
	return func(in *policy.PodDisruptionBudgetList) []*policy.PodDisruptionBudget {
		ret := make([]*policy.PodDisruptionBudget, 0, len(in.Items))

		for _, el := range in.Items {
			z := el.DeepCopy()
			if !generic.FilterObject(z, filter...) {
				continue
			}

			ret = append(ret, z)
		}

		return ret
	}
}
