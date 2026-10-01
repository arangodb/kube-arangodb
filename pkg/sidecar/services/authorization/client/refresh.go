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
	"context"
	"io"
	"time"

	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	sidecarSvcAuthzDefinition "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/definition"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
)

// refreshMinInterval bounds how often an on-deny refresh may re-pull the full state. Evaluate is a
// hot path and RBAC legitimately denies constantly, so a refresh runs at most once per interval
// regardless of how many denials occur - just enough to let a false deny from a stale stream
// (e.g. a freshly created role/policy/binding not yet propagated) self-correct quickly.
const refreshMinInterval = time.Second

// tryRefresh re-pulls the authorization state from the pool service when the last refresh is older
// than refreshMinInterval. It returns true only when a refresh actually ran, so the caller knows a
// re-evaluation against fresh state is worthwhile. Concurrent callers are collapsed to one refresh.
func (c *client) tryRefresh(ctx context.Context) bool {
	c.refreshLock.Lock()
	if !c.lastRefresh.IsZero() && time.Since(c.lastRefresh) < refreshMinInterval {
		c.refreshLock.Unlock()
		return false
	}
	c.lastRefresh = time.Now()
	c.refreshLock.Unlock()

	if err := c.Refresh(ctx); err != nil {
		logger.Err(err).Trace("On-deny authorization refresh failed")
		return false
	}
	return true
}

// Refresh re-pulls a fresh snapshot of every pool from the authorization pool service and replaces
// the local cache. It reuses the same Get* snapshot RPCs as the initial sync, so it reflects the
// authoritative current state (pool writes refresh the server pool synchronously).
func (c *client) Refresh(ctx context.Context) error {
	svc, err := c.client.Get(ctx)
	if err != nil {
		return err
	}

	policies, err := refreshPolicies(ctx, svc)
	if err != nil {
		return err
	}

	roles, err := refreshRoles(ctx, svc)
	if err != nil {
		return err
	}

	userRoleBindings, err := refreshUserRoleBindings(ctx, svc)
	if err != nil {
		return err
	}

	groupRoleBindings, err := refreshGroupRoleBindings(ctx, svc)
	if err != nil {
		return err
	}

	c.setPolicies(policies)
	c.setRoles(roles)
	c.setUserRoleBindings(userRoleBindings)
	c.setGroupRoleBindings(groupRoleBindings)

	return nil
}

func refreshPolicies(ctx context.Context, svc sidecarSvcAuthzDefinition.AuthorizationPoolServiceClient) (map[string]*sidecarSvcAuthzTypes.Policy, error) {
	resp, err := svc.GetPolicy(ctx, &pbSharedV1.Empty{})
	if err != nil {
		return nil, err
	}

	out := map[string]*sidecarSvcAuthzTypes.Policy{}
	for {
		spec, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				return out, nil
			}
			return nil, err
		}
		for _, item := range spec.GetItems() {
			if item.GetItem() != nil {
				out[item.GetName()] = item.GetItem()
			}
		}
	}
}

func refreshRoles(ctx context.Context, svc sidecarSvcAuthzDefinition.AuthorizationPoolServiceClient) (map[string]*sidecarSvcAuthzTypes.Role, error) {
	resp, err := svc.GetRole(ctx, &pbSharedV1.Empty{})
	if err != nil {
		return nil, err
	}

	out := map[string]*sidecarSvcAuthzTypes.Role{}
	for {
		spec, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				return out, nil
			}
			return nil, err
		}
		for _, item := range spec.GetItems() {
			if item.GetItem() != nil {
				out[item.GetName()] = item.GetItem()
			}
		}
	}
}

func refreshUserRoleBindings(ctx context.Context, svc sidecarSvcAuthzDefinition.AuthorizationPoolServiceClient) (map[string]*sidecarSvcAuthzTypes.UserRoleBinding, error) {
	resp, err := svc.GetUserRoleBinding(ctx, &pbSharedV1.Empty{})
	if err != nil {
		return nil, err
	}

	out := map[string]*sidecarSvcAuthzTypes.UserRoleBinding{}
	for {
		spec, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				return out, nil
			}
			return nil, err
		}
		for _, item := range spec.GetItems() {
			if item.GetItem() != nil {
				out[item.GetName()] = item.GetItem()
			}
		}
	}
}

func refreshGroupRoleBindings(ctx context.Context, svc sidecarSvcAuthzDefinition.AuthorizationPoolServiceClient) (map[string]*sidecarSvcAuthzTypes.UserRoleBinding, error) {
	resp, err := svc.GetGroupRoleBinding(ctx, &pbSharedV1.Empty{})
	if err != nil {
		return nil, err
	}

	out := map[string]*sidecarSvcAuthzTypes.UserRoleBinding{}
	for {
		spec, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				return out, nil
			}
			return nil, err
		}
		for _, item := range spec.GetItems() {
			if item.GetItem() != nil {
				out[item.GetName()] = item.GetItem()
			}
		}
	}
}
