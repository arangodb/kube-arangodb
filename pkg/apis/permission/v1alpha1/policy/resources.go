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

package policy

import (
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

type Resources []Resource

func (a Resources) Hash() string {
	return util.SHA256FromExtract(func(v Resource) string { return v.Hash() }, a...)
}

func (a Resources) Validate() error {
	return shared.ValidateInterfaceList(a)
}

type Resource string

func (a Resource) Hash() string {
	return util.SHA256FromString(string(a))
}

func (a Resource) Validate() error {
	return sidecarSvcAuthzTypes.ValidateResource(string(a))
}
