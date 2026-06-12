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

package token

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	permissionApi "github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
)

// HandleArangoDBPolicy is a no-op since direct policy references were removed
// from the token spec. Policies are now accessed through roles.
func (h *handler) HandleArangoDBPolicy(ctx context.Context, item operation.Item, extension *permissionApi.ArangoPermissionToken, st *permissionApi.ArangoPermissionTokenStatus, depl *api.ArangoDeployment) (bool, error) {
	return false, nil
}
