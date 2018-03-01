//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package deployment

import (
	"context"
	"fmt"
	"sync"

	"k8s.io/client-go/kubernetes"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
	"github.com/arangodb/k8s-operator/pkg/util/arangod"
)

type clientCache struct {
	mutex          sync.Mutex
	clients        map[string]driver.Client
	kubecli        kubernetes.Interface
	apiObject      *api.ArangoDeployment
	databaseClient driver.Client
}

// newClientCache creates a new client cache
func newClientCache(kubecli kubernetes.Interface, apiObject *api.ArangoDeployment) *clientCache {
	return &clientCache{
		clients:   make(map[string]driver.Client),
		kubecli:   kubecli,
		apiObject: apiObject,
	}
}

// Get a cached client for the given ID in the given group, creating one
// if needed.
func (cc *clientCache) Get(ctx context.Context, group api.ServerGroup, id string) (driver.Client, error) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	key := fmt.Sprintf("%d-%s", group, id)
	c, found := cc.clients[key]
	if found {
		return c, nil
	}

	// Not found, create a new client
	c, err := arangod.CreateArangodClient(ctx, cc.kubecli.CoreV1(), cc.apiObject, group, id)
	if err != nil {
		return nil, maskAny(err)
	}
	cc.clients[key] = c
	return c, nil
}

// GetDatabase returns a cached client for the entire database (cluster coordinators or single server),
// creating one if needed.
func (cc *clientCache) GetDatabase(ctx context.Context) (driver.Client, error) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	if c := cc.databaseClient; c != nil {
		return c, nil
	}

	// Not found, create a new client
	c, err := arangod.CreateArangodDatabaseClient(ctx, cc.kubecli.CoreV1(), cc.apiObject)
	if err != nil {
		return nil, maskAny(err)
	}
	cc.databaseClient = c
	return c, nil
}
