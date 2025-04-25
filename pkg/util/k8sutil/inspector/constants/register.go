//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package constants

import (
	"reflect"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type registration struct {
	GVK schema.GroupVersionKind
	GVR schema.GroupVersionResource
}

var (
	registerer = util.NewRegisterer[reflect.Type, registration]()
)

func register[T meta.Object](GVK schema.GroupVersionKind, GVR schema.GroupVersionResource) {
	registerer.MustRegister(util.TypeOf[T](), registration{
		GVK: GVK,
		GVR: GVR,
	})
}

func getRegistration(t reflect.Type) (registration, bool) {
	for _, i := range registerer.Items() {
		if i.K == t {
			return i.V, true
		}
	}

	return registration{}, false
}

func ExtractGVK(t reflect.Type) (schema.GroupVersionKind, bool) {
	k, v := getRegistration(t)
	return k.GVK, v
}

func ExtractGVKFromObject[T meta.Object](obj T) (schema.GroupVersionKind, bool) {
	return ExtractGVK(reflect.TypeOf(obj))
}

func ExtractGVR(t reflect.Type) (schema.GroupVersionResource, bool) {
	k, v := getRegistration(t)
	return k.GVR, v
}

func ExtractGVRFromObject[T meta.Object](obj T) (schema.GroupVersionResource, bool) {
	return ExtractGVR(reflect.TypeOf(obj))
}
