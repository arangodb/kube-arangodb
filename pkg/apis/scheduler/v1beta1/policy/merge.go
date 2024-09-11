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

package policy

import (
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type Policy struct {
	// Method defines the merge method
	// +doc/enum: override|Overrides values during configuration merge
	// +doc/enum: append|Appends, if possible, values during configuration merge
	Method *Method `json:"method,omitempty"`
}

func (m *Policy) Validate() error {
	if m == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceErrors("method", shared.ValidateOptionalInterface(m.Method)),
	)
}

func (m *Policy) GetMethod(d Method) Method {
	if m == nil || m.Method == nil {
		return d
	}

	return *m.Method
}

type Method string

func (m *Method) Validate() error {
	if m == nil {
		return nil
	}

	switch v := *m; v {
	case Override, Append:
		return nil
	default:
		return errors.Errorf("Invalid method: %s", v)
	}
}

const (
	Override Method = "override"
	Append   Method = "append"
)
