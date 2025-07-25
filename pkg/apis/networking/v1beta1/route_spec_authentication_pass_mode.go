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
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ArangoRouteSpecAuthenticationPassMode string

const (
	ArangoRouteSpecAuthenticationPassModePass     ArangoRouteSpecAuthenticationPassMode = "pass"
	ArangoRouteSpecAuthenticationPassModeOverride ArangoRouteSpecAuthenticationPassMode = "override"
	ArangoRouteSpecAuthenticationPassModeRemove   ArangoRouteSpecAuthenticationPassMode = "remove"
)

func (a *ArangoRouteSpecAuthenticationPassMode) Get() ArangoRouteSpecAuthenticationPassMode {
	if a == nil {
		return ArangoRouteSpecAuthenticationPassModeOverride
	}
	switch v := *a; v {
	case ArangoRouteSpecAuthenticationPassModePass, ArangoRouteSpecAuthenticationPassModeOverride, ArangoRouteSpecAuthenticationPassModeRemove:
		return v
	}

	return ""
}

func (a *ArangoRouteSpecAuthenticationPassMode) Validate() error {
	switch v := a.Get(); v {
	case ArangoRouteSpecAuthenticationPassModePass, ArangoRouteSpecAuthenticationPassModeOverride, ArangoRouteSpecAuthenticationPassModeRemove:
		return nil
	default:
		return errors.Errorf("Invalid AuthPassMode: %s", v)
	}
}

func (a *ArangoRouteSpecAuthenticationPassMode) Hash() string {
	if a == nil {
		return ""
	}

	return util.SHA256FromString(string(*a))
}
