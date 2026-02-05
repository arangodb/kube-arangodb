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

package shared

import (
	"context"
	"sync"

	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
)

func CachedPlugin(parent Plugin) Plugin {
	return &cachedPlugin{
		items:  map[string]*pbAuthorizationV1.AuthorizationV1PermissionResponse{},
		parent: parent,
	}
}

type cachedPlugin struct {
	lock sync.RWMutex

	items map[string]*pbAuthorizationV1.AuthorizationV1PermissionResponse

	parent Plugin

	revision uint64
}

func (c *cachedPlugin) Revision() uint64 {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.parent.Revision()
}

func (c *cachedPlugin) clean() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.revision == c.parent.Revision() {
		return
	}

	c.items = map[string]*pbAuthorizationV1.AuthorizationV1PermissionResponse{}
	c.revision = c.parent.Revision()
}

func (c *cachedPlugin) Ready(ctx context.Context) error {
	return c.parent.Ready(ctx)
}

func (c *cachedPlugin) Evaluate(ctx context.Context, req *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
	hash := req.Hash()

	if c.parent.Revision() != c.revision {
		c.clean()
	}

	if res, ok := c.evaluate(hash); ok {
		return res, nil
	}

	return c.evaluateWithWrite(ctx, hash, req)
}

func (c *cachedPlugin) evaluate(hash string) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	v, ok := c.items[hash]
	return v, ok
}

func (c *cachedPlugin) evaluateWithWrite(ctx context.Context, hash string, req *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if v, ok := c.items[hash]; ok {
		return v, nil
	}

	ret, err := c.parent.Evaluate(ctx, req)
	if err != nil {
		return nil, err
	}

	c.items[hash] = ret

	return ret, nil
}
