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

package impl

import (
	"github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/impl/auth_bearer"
	"github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/impl/auth_cookie"
	"github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/impl/auth_required"
	"github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/impl/pass_mode"
	"github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/impl/required"
	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
)

func Factory() pbImplEnvoyAuthV3Shared.Factory {
	return pbImplEnvoyAuthV3Shared.NewFactory(
		required.New,
		auth_bearer.New,
		auth_cookie.New,
		auth_required.New,
		pass_mode.New,
	)
}
