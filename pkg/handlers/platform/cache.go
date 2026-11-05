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

package platform

import (
	"time"

	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/cache"
)

// Cache is the shared cache of reconciled platform objects, with one object cache per type. Each handler
// keeps its own type current after every reconcile, and the handlers read the other types (for example the
// service and workflow handlers read Charts) to resolve custom chart requirements without re-fetching
// objects from the API server.
type Cache interface {
	// Charts returns the cache of ArangoPlatformChart objects keyed by namespace/name.
	Charts() cache.ObjectCache[*platformApi.ArangoPlatformChart]
	// Services returns the cache of ArangoPlatformService objects keyed by namespace/name.
	Services() cache.ObjectCache[*platformApi.ArangoPlatformService]
	// Workflows returns the cache of ArangoPlatformWorkflow objects keyed by namespace/name.
	Workflows() cache.ObjectCache[*platformApi.ArangoPlatformWorkflow]
}

// NewCache returns an empty Cache whose entries expire ttl after they are stored.
func NewCache(ttl time.Duration) Cache {
	return &platformCache{
		charts:    cache.NewObjectCache[*platformApi.ArangoPlatformChart](ttl),
		services:  cache.NewObjectCache[*platformApi.ArangoPlatformService](ttl),
		workflows: cache.NewObjectCache[*platformApi.ArangoPlatformWorkflow](ttl),
	}
}

type platformCache struct {
	charts    cache.ObjectCache[*platformApi.ArangoPlatformChart]
	services  cache.ObjectCache[*platformApi.ArangoPlatformService]
	workflows cache.ObjectCache[*platformApi.ArangoPlatformWorkflow]
}

func (c *platformCache) Charts() cache.ObjectCache[*platformApi.ArangoPlatformChart] {
	return c.charts
}

func (c *platformCache) Services() cache.ObjectCache[*platformApi.ArangoPlatformService] {
	return c.services
}

func (c *platformCache) Workflows() cache.ObjectCache[*platformApi.ArangoPlatformWorkflow] {
	return c.workflows
}
