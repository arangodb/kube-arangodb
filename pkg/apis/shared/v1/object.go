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

package v1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type Object struct {
	// Name of the object
	Name string `json:"name"`

	// Namespace of the object. Should default to the namespace of the parent object
	Namespace *string `json:"namespace,omitempty"`
}

func (o *Object) IsEmpty() bool {
	return o == nil ||
		(o.Name == "" && o.Namespace != nil)
}

func (o *Object) GetName() string {
	if o == nil {
		return ""
	}

	return o.Name
}

func (o *Object) GetNamespace(obj meta.Object) string {
	if o != nil {
		if n := o.Namespace; n != nil {
			return *n
		}
	}

	return obj.GetNamespace()
}

func (o *Object) Validate() error {
	if o == nil {
		o = &Object{}
	}

	var errs []error
	if o.Name == "" {
		errs = append(errs, shared.PrefixResourceErrors("name", errors.New("must be not empty")))
	}
	if o.Namespace != nil && *o.Namespace == "" {
		errs = append(errs, shared.PrefixResourceErrors("namespace", errors.New("must be nil or non-empty string")))
	}

	return shared.WithErrors(errs...)
}
