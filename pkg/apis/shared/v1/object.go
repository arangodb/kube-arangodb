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

package v1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func NewObjectWithChecksum(object meta.Object, checksum string) Object {
	obj := NewObject(object)
	obj.Checksum = util.NewType(checksum)
	return obj
}

func NewObject(object meta.Object) Object {
	var n Object

	n.Name = object.GetName()
	n.UID = util.NewType(object.GetUID())

	if ns := object.GetNamespace(); ns != "" {
		n.Namespace = util.NewType(ns)
	}

	return n
}

type Object struct {
	// Name of the object
	Name string `json:"name"`

	// Namespace of the object. Should default to the namespace of the parent object
	Namespace *string `json:"namespace,omitempty"`

	// UID keeps the information about object UID
	UID *types.UID `json:"uid,omitempty"`

	// UID keeps the information about object Checksum
	Checksum *string `json:"checksum,omitempty"`
}

func (o *Object) IsEmpty() bool {
	return o == nil ||
		(o.Name == "" && o.Namespace == nil)
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

func (o *Object) GetUID() types.UID {
	if o != nil {
		if n := o.UID; n != nil {
			return *n
		}
	}

	return ""
}

func (o *Object) GetChecksum() string {
	if o != nil {
		if n := o.Checksum; n != nil {
			return *n
		}
	}

	return ""
}

func (o *Object) Equals(obj meta.Object) bool {
	if o == nil {
		return false
	}

	if o.Name != obj.GetName() {
		return false
	}

	if n := o.Namespace; n != nil {
		if *n != obj.GetNamespace() {
			return false
		}
	}

	if n := o.UID; n != nil {
		if *n != obj.GetUID() {
			return false
		}
	}

	return true
}

func (o *Object) Validate() error {
	if o == nil {
		o = &Object{}
	}

	var errs []error
	errs = append(errs, shared.PrefixResourceErrors("name", AsKubernetesResourceName(&o.Name).Validate()))
	if o.Namespace != nil {
		errs = append(errs, shared.PrefixResourceErrors("namespace", AsKubernetesResourceName(o.Namespace).Validate()))
	}
	if u := o.UID; u != nil {
		errs = append(errs, shared.PrefixResourceErrors("uid", shared.ValidateUID(*u)))
	}

	return shared.WithErrors(errs...)
}
