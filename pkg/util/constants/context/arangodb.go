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

package context

import (
	adbDriverV2 "github.com/arangodb/go-driver/v2/arangodb"

	pbImplAuthorizationV1Shared "github.com/arangodb/kube-arangodb/integrations/authorization/v1/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

const (
	ArangoDBClientCache util.ContextObject[cache.Object[adbDriverV2.Client]] = "arangodb-client-cache-object"

	AuthZClientPlugin util.ContextObject[pbImplAuthorizationV1Shared.Evaluator] = "authz-client-plugin"
)
