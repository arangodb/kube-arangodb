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

package types

import "github.com/arangodb/kube-arangodb/pkg/util"

func (a *Context) Hash() string {
	if a == nil {
		return ""
	}

	return util.SHA256FromStringArray(
		util.SHA256FromHashStringMap(a.GetParameters()),
	)
}

func (a *ContextParameter) Hash() string {
	if v := a.GetValues(); len(v) > 0 {
		return util.SHA256FromStringArray(v...)
	}

	return ""
}

func (x *Context) GetContext() map[string][]string {
	if x == nil {
		return nil
	}

	r := make(map[string][]string, len(x.GetParameters()))

	for k, v := range x.GetParameters() {
		r[k] = v.GetValues()
	}

	return r
}
