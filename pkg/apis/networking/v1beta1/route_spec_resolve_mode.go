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

package v1beta1

import (
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

type ArangoRouteSpecResolveMode string

const (
	ArangoRouteSpecResolveModeDefault                            = ArangoRouteSpecResolveModeDNS
	ArangoRouteSpecResolveModeDNS     ArangoRouteSpecResolveMode = "dns"
	ArangoRouteSpecResolveModeIP      ArangoRouteSpecResolveMode = "ip"
)

func (a *ArangoRouteSpecResolveMode) Get() ArangoRouteSpecResolveMode {
	if a == nil {
		return ArangoRouteSpecResolveModeDefault
	}

	return ArangoRouteSpecResolveMode(strings.ToLower(string(*a)))
}

func (a *ArangoRouteSpecResolveMode) String() string {
	return string(a.Get())
}

func (a *ArangoRouteSpecResolveMode) Validate() error {
	switch x := a.Get(); x {
	case ArangoRouteSpecResolveModeDNS, ArangoRouteSpecResolveModeIP:
		return nil
	default:
		return errors.Errorf("Invalid resolve mode: %s", x.String())
	}
}
