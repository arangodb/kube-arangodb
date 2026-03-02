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
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"

	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
	pbImplAuthorizationV1Shared "github.com/arangodb/kube-arangodb/integrations/authorization/v1/shared"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	sidecarSvcAuthzDefinition "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/definition"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func NewClient(ctx context.Context, c sidecarSvcAuthzDefinition.AuthorizationPoolServiceClient) Client {
	client := &client{
		client: c,
		closed: make(chan struct{}),
	}

	go client.run(ctx)

	return client
}

type Client interface {
	pbImplAuthorizationV1Shared.Plugin

	Wait(ctx context.Context) bool
}

type client struct {
	lock sync.RWMutex

	setLock sync.Mutex

	client sidecarSvcAuthzDefinition.AuthorizationPoolServiceClient

	closed chan struct{}

	revision uint64

	cache *cache

	policies clientSet[*sidecarSvcAuthzTypes.Policy]
	roles    clientSet[*sidecarSvcAuthzTypes.Role]
}

func (c *client) Revision() uint64 {
	return c.revision
}

func (c *client) Evaluate(ctx context.Context, req *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
	policies := c.get().extractUserPolicies(req.GetUser(), req.GetRoles()...)

	context := req.GetContext().GetContext()

	var allowed bool

	for _, policy := range policies {
		if a, err := policy.Evaluate(req.GetAction(), req.GetResource(), context); err != nil {
			if sidecarSvcAuthzTypes.IsPermissionDenied(err) {
				return &pbAuthorizationV1.AuthorizationV1PermissionResponse{
					Message: "Explicit deny",
					Effect:  pbAuthorizationV1.AuthorizationV1Effect_Deny,
				}, nil
			}
		} else if a {
			allowed = true
		}
	}

	if allowed {
		return &pbAuthorizationV1.AuthorizationV1PermissionResponse{
			Message: "Access Granted",
			Effect:  pbAuthorizationV1.AuthorizationV1Effect_Allow,
		}, nil
	}
	return &pbAuthorizationV1.AuthorizationV1PermissionResponse{
		Message: "Permission denied",
		Effect:  pbAuthorizationV1.AuthorizationV1Effect_Deny,
	}, nil
}

func (c *client) get() *cache {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.cache
}

func (c *client) Wait(ctx context.Context) bool {
	tickerT := time.NewTicker(50 * time.Millisecond)
	defer tickerT.Stop()

	for {
		select {
		case <-c.closed:
			return false
		case <-ctx.Done():
			return false
		case <-tickerT.C:
			if c.Ready(ctx) == nil {
				return true
			}
		}
	}
}

func (c *client) Ready(ctx context.Context) error {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return errors.Errors(c.policies.ready(), c.roles.ready(), util.BoolSwitch(c.cache == nil, errors.Errorf("nil cache"), nil))
}

func (c *client) setRoles(items map[string]*sidecarSvcAuthzTypes.Role) {
	c.setLock.Lock()
	defer c.setLock.Unlock()

	c.revision += 1

	cp := make(map[string]*sidecarSvcAuthzTypes.Role, len(items))
	for k, v := range items {
		cp[k] = v
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	cache := newCache(c.policies.items, cp)

	c.roles.set(cp)
	c.cache = &cache
}

func (c *client) setPolicies(items map[string]*sidecarSvcAuthzTypes.Policy) {
	c.setLock.Lock()
	defer c.setLock.Unlock()

	c.revision += 1

	cp := make(map[string]*sidecarSvcAuthzTypes.Policy, len(items))
	for k, v := range items {
		cp[k] = v
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	cache := newCache(cp, c.roles.items)

	c.policies.set(cp)
	c.cache = &cache
}

type clientSet[T proto.Message] struct {
	items map[string]T

	updated time.Time
}

func (c *clientSet[T]) ready() error {
	if time.Since(c.updated) > time.Minute {
		return errors.Errorf("Timeout exceeded while waiting for permissions")
	}

	return nil
}

func (c *clientSet[T]) set(items map[string]T) {
	c.items = items
	c.updated = time.Now()
}

func (c *client) run(ctx context.Context) {
	defer close(c.closed)
	for {
		if err := c.runE(ctx); err != nil {
			logger.Err(err).Warn("Authorization pool client error")
		}

		select {
		case <-ctx.Done():
			return
		default:
			continue
		}
	}
}

func (c *client) runE(ctx context.Context) error {
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return c.runPoliciesE(gctx)
	})

	g.Go(func() error {
		return c.runRolesE(gctx)
	})

	return g.Wait()
}

func (c *client) runPoliciesE(ctx context.Context) error {
	policies := map[string]*sidecarSvcAuthzTypes.Policy{}

	var index uint32

	{
		response, err := c.client.GetPolicy(ctx, &pbSharedV1.Empty{})
		if err != nil {
			return err
		}

		for {
			spec, err := response.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			for _, item := range spec.GetItems() {
				policies[item.GetName()] = item.GetItem()
				index = item.GetIndex()
			}
		}
	}

	c.setPolicies(policies)

	logger.Trace("Policies init complete")

	for {
		changes, err := c.client.PoolPolicyChanges(ctx, &sidecarSvcAuthzDefinition.AuthorizationPoolRequest{
			Start:   index,
			Timeout: durationpb.New(15 * time.Second),
		})
		if err != nil {
			return err
		}

		for {
			spec, err := changes.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			logger.Int("items", len(spec.Items)).Trace("Received policy update")

			for _, item := range spec.GetItems() {
				policies[item.GetName()] = item.GetItem()
				index = item.GetIndex()
			}
		}

		c.setPolicies(policies)
	}
}

func (c *client) runRolesE(ctx context.Context) error {
	roles := map[string]*sidecarSvcAuthzTypes.Role{}

	var index uint32

	{
		response, err := c.client.GetRole(ctx, &pbSharedV1.Empty{})
		if err != nil {
			return err
		}

		for {
			spec, err := response.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			for _, item := range spec.GetItems() {
				roles[item.GetName()] = item.GetItem()
				index = item.GetIndex()
			}
		}
	}

	c.setRoles(roles)

	logger.Trace("Roles init complete")

	for {
		changes, err := c.client.PoolRoleChanges(ctx, &sidecarSvcAuthzDefinition.AuthorizationPoolRequest{
			Start:   index,
			Timeout: durationpb.New(15 * time.Second),
		})
		if err != nil {
			return err
		}

		for {
			spec, err := changes.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			logger.Int("items", len(spec.Items)).Trace("Received roles update")

			for _, item := range spec.GetItems() {
				roles[item.GetName()] = item.GetItem()
				index = item.GetIndex()
			}
		}

		c.setRoles(roles)
	}
}
