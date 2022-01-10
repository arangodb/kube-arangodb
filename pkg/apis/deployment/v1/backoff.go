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

import (
	"encoding/json"
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ json.Marshaler = BackOff{}

type BackOffKey string

type BackOff map[BackOffKey]meta.Time

func (b BackOff) MarshalJSON() ([]byte, error) {
	r := map[BackOffKey]meta.Time{}

	for k, v := range b {
		if v.IsZero() {
			continue
		}

		r[k] = *v.DeepCopy()
	}

	return json.Marshal(r)
}

func (b BackOff) Process(key BackOffKey) bool {
	if b == nil {
		return true
	}

	if t, ok := b[key]; ok {
		if t.IsZero() {
			return true
		}

		return time.Now().After(t.Time)
	} else {
		return true
	}
}

func (b BackOff) BackOff(key BackOffKey, delay time.Duration) BackOff {
	n := meta.Time{Time: time.Now().Add(delay)}

	z := b.DeepCopy()

	if z == nil {
		return BackOff{
			key: n,
		}
	}

	z[key] = n
	return z
}

func (b BackOff) Combine(a BackOff) BackOff {
	d := b.DeepCopy()
	if d == nil {
		d = BackOff{}
	}

	for k, v := range a {
		d[k] = v
	}

	return d
}

func (b BackOff) CombineLatest(a BackOff) BackOff {
	d := b.DeepCopy()
	if d == nil {
		d = BackOff{}
	}

	for k, v := range a {
		if i, ok := d[k]; !ok || (ok && v.After(i.Time)) {
			d[k] = v
		}
	}

	return d
}

func (b BackOff) Equal(a BackOff) bool {
	if len(b) == 0 && len(a) == 0 {
		return true
	}

	if len(b) != len(a) {
		return false
	}

	for k, v := range b {
		if av, ok := a[k]; !ok || !v.Equal(&av) {
			return false
		}
	}

	return true
}
