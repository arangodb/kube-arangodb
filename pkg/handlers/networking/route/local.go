//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package route

import (
	"github.com/arangodb/kube-arangodb/pkg/apis/networking"
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1beta1"
)

func Kind() string {
	return networking.ArangoRouteResourceKind
}

func Group() string {
	return networkingApi.SchemeGroupVersion.Group
}

func Version() string {
	return networkingApi.SchemeGroupVersion.Version
}
