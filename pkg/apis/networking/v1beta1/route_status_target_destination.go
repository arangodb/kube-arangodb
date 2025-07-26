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

package v1beta1

import (
	"fmt"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type ArangoRouteStatusTargetDestinations []ArangoRouteStatusTargetDestination

func (a ArangoRouteStatusTargetDestinations) Hash() string {
	return util.SHA256FromExtract(func(t ArangoRouteStatusTargetDestination) string {
		return t.Hash()
	}, a...)
}

type ArangoRouteStatusTargetDestination struct {
	Host string `json:"host,omitempty"`
	Port int32  `json:"port,omitempty"`
}

func (a *ArangoRouteStatusTargetDestination) Hash() string {
	if a == nil {
		return ""
	}
	return util.SHA256FromStringArray(fmt.Sprintf("%s:%d", a.Host, a.Port))
}
