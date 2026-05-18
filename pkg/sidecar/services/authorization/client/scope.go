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

package client

import (
	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// ScopedPolicies maps group names to their resolved ScopedPolicy.
type ScopedPolicies map[string]ScopedPolicy

// Evaluate iterates groups and returns Allow if any group grants access.
func (s ScopedPolicies) Evaluate(req *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
	for _, g := range s {
		resp, err := g.Evaluate(req)
		if err != nil {
			return nil, err
		}

		if resp.GetEffect() == sidecarSvcAuthzTypes.Effect_Allow {
			return resp, nil
		}
	}

	return &pbAuthorizationV1.AuthorizationV1PermissionResponse{
		Message: "Permission denied",
		Effect:  sidecarSvcAuthzTypes.Effect_Deny,
	}, nil
}

func (s ScopedPolicies) Hash() string {
	if len(s) == 0 {
		return ""
	}

	var hashes []string
	for k, v := range s {
		hashes = append(hashes, util.SHA256FromStringArray(k, v.Hash()))
	}

	return util.SHA256FromStringArray(hashes...)
}

// ScopedPolicy pairs a set of policies with an optional scope boundary.
// The scope restricts the effective permissions: an action is allowed only
// when the policies allow it AND the scope allows it.
type ScopedPolicy struct {
	Policies PolicyList
	Scope    *Policy
}

func (s *ScopedPolicy) Hash() string {
	if s == nil {
		return ""
	}

	return util.SHA256FromStringArray(
		s.Policies.Hash(),
		s.Scope.Hash(),
	)
}

// Evaluate checks whether the policies allow the action and the scope permits it.
// Returns Allow only when both agree. Nil scope means the group is not considered.
// Evaluate checks scope first, then policies. Returns Allow only when both agree.
// Nil scope means the group is not considered.
func (s *ScopedPolicy) Evaluate(req *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
	if s == nil || s.Scope == nil {
		return &pbAuthorizationV1.AuthorizationV1PermissionResponse{
			Message: "Permission denied",
			Effect:  sidecarSvcAuthzTypes.Effect_Deny,
		}, nil
	}

	scopeResp, err := EvaluatePolicies(req, s.Scope)
	if err != nil {
		return nil, err
	}

	if scopeResp.GetEffect() != sidecarSvcAuthzTypes.Effect_Allow {
		return scopeResp, nil
	}

	return EvaluatePolicies(req, s.Policies...)
}
