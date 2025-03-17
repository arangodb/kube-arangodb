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

package shared

import (
	"context"
	"sync"
	"time"

	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
)

func WithCache(in Evaluator, ttl time.Duration) Evaluator {
	return &cacheTTL{
		in:    in,
		cache: map[string]cacheTTLItem{},
		ttl:   ttl,
	}
}

type cacheTTL struct {
	lock sync.RWMutex

	in Evaluator

	ttl time.Duration

	cache map[string]cacheTTLItem
}

type cacheTTLItem struct {
	TTL      time.Time
	Response *pbAuthorizationV1.AuthorizationV1PermissionResponse
}

func (c *cacheTTL) Evaluate(ctx context.Context, request *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	if v, ok := c.extract(request); ok {
		return v, nil
	}

	return c.refresh(ctx, request)
}

func (c *cacheTTL) extract(request *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	v, ok := c.cache[request.Hash()]
	if !ok {
		return nil, false
	}

	if v.TTL.After(time.Now()) {
		return v.Response, true
	}

	return nil, false
}

func (c *cacheTTL) refresh(ctx context.Context, request *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	v, ok := c.cache[request.Hash()]
	if ok {
		if v.TTL.After(time.Now()) {
			return v.Response, nil
		}
	}

	resp, err := c.in.Evaluate(ctx, request)
	if err != nil {
		return nil, err
	}

	c.cache[request.Hash()] = cacheTTLItem{
		TTL:      time.Now().Add(c.ttl),
		Response: resp,
	}

	return resp, nil
}
