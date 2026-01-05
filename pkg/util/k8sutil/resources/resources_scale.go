//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package resources

import (
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func ScaleResources(in core.ResourceRequirements, ratio float64) core.ResourceRequirements {
	var r = core.ResourceRequirements{}

	if l := in.Limits; len(l) > 0 {
		r.Limits = ScaleResourceList(l, ratio)
	}

	if l := in.Requests; len(l) > 0 {
		r.Requests = ScaleResourceList(l, ratio)
	}

	return r
}

func DefaultResourceList(in core.ResourceList, quantity resource.Quantity, resources ...core.ResourceName) core.ResourceList {
	if in == nil {
		in = core.ResourceList{}
	}

	for _, v := range resources {
		if _, ok := in[v]; !ok {
			in[v] = quantity.DeepCopy()
		}
	}

	return in
}

func ScaleResourceList(in core.ResourceList, ratio float64) core.ResourceList {
	var r = core.ResourceList{}

	for k, v := range in {
		q := ScaleQuantity(v, k, ratio)
		if q.IsZero() {
			continue
		}

		r[k] = q
	}

	return r
}

func ScaleQuantity(in resource.Quantity, t core.ResourceName, ratio float64) resource.Quantity {
	switch t {
	case core.ResourceMemory:
		return *resource.NewQuantity(int64(float64(in.Value())*ratio), in.Format)
	case core.ResourceCPU:
		return *resource.NewMilliQuantity(int64(float64(in.MilliValue())*ratio), in.Format)
	default:
		return resource.Quantity{}
	}
}
