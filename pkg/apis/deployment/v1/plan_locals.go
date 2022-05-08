//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package v1

type PlanLocalKey string

func (p PlanLocalKey) String() string {
	return string(p)
}

type PlanLocals map[PlanLocalKey]string

func (p *PlanLocals) Remove(key PlanLocalKey) bool {
	if *p == nil {
		return false
	}

	z := *p

	if _, ok := z[key]; ok {
		delete(z, key)
		*p = z
		return true
	}

	return false
}

func (p PlanLocals) Get(key PlanLocalKey) (string, bool) {
	v, ok := p[key]
	return v, ok
}

func (p PlanLocals) GetWithParent(parent PlanLocals, key PlanLocalKey) (string, bool) {
	v, ok := p[key]
	if ok {
		return v, true
	}
	return parent.Get(key)
}

func (p *PlanLocals) Merge(merger PlanLocals) (changed bool) {
	for k, v := range merger {
		if p.Add(k, v, true) {
			changed = true
		}
	}

	return
}

func (p *PlanLocals) Add(key PlanLocalKey, value string, override bool) bool {
	if value == "" {
		return p.Remove(key)
	}

	if *p == nil {
		*p = PlanLocals{
			key: value,
		}

		return true
	}

	z := *p

	if v, ok := z[key]; ok {
		if v == value {
			return true
		}

		if !override {
			return false
		}
	}

	z[key] = value

	*p = z

	return true
}

func (p PlanLocals) Equal(other PlanLocals) bool {
	if len(p) != len(other) {
		return false
	}

	for k, v := range p {
		if v2, ok := other[k]; !ok || v != v2 {
			return false
		}
	}

	return true
}
