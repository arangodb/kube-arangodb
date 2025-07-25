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
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

type ArangoRouteSpecRoute struct {
	// Path specifies the Path route
	Path *string `json:"path,omitempty"`
}

func (a *ArangoRouteSpecRoute) GetPath() string {
	if a == nil || a.Path == nil {
		return ""
	}

	return *a.Path
}

func (a *ArangoRouteSpecRoute) Validate() error {
	if a == nil {
		a = &ArangoRouteSpecRoute{}
	}

	if err := shared.WithErrors(
		shared.ValidateRequiredPath("path", a.Path, shared.ValidateAPIPath),
	); err != nil {
		return err
	}

	return nil
}

func (a *ArangoRouteSpecRoute) AsStatus() ArangoRouteStatusTargetRoute {
	return ArangoRouteStatusTargetRoute{
		Path: a.GetPath(),
	}
}
