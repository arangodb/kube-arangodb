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

package v1alpha1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func FromObject(in meta.Object) *Object {
	if in == nil {
		return nil
	}

	return &Object{
		Name:      in.GetName(),
		Namespace: util.NewType(in.GetNamespace()),
		UID:       util.NewType(in.GetUID()),
	}
}

type Object struct {
	// Name of the object
	Name string `json:"name"`

	// Namespace of the object. Should default to the namespace of the parent object
	Namespace *string `json:"namespace,omitempty"`

	// UID keeps the information about object UID
	UID *types.UID `json:"uid,omitempty"`
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
