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

package kerrors

import (
	"fmt"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
)

func NewResourceError(cause error, obj interface{}) error {
	if cause == nil {
		return nil
	}

	if gvk, ok := constants.ExtractGVKFromObject(obj); !ok {
		return cause
	} else {
		if meta, ok := obj.(meta.Object); ok {
			return ResourceError{
				GVK:       gvk,
				Namespace: meta.GetNamespace(),
				Name:      meta.GetName(),
				cause:     cause,
			}
		} else {
			return ResourceError{
				GVK:   gvk,
				cause: cause,
			}
		}
	}
}

type ResourceError struct {
	GVK schema.GroupVersionKind

	Namespace, Name string

	cause error
}

func (r ResourceError) Cause() error  { return r.cause }
func (r ResourceError) Unwrap() error { return r.cause }

func (r ResourceError) Error() string {
	gvk := r.GVK.String()
	if r.GVK.Empty() {
		gvk = "UNKNOWN"
	}

	n := fmt.Sprintf("%s/%s", r.Namespace, r.Name)
	if r.Namespace == "" {
		n = ""
	}

	cause := "Unknown"
	if e := r.cause; e != nil {
		cause = e.Error()
	}

	return fmt.Sprintf("Resource [%s] %s: %s", gvk, n, cause)
}
