//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
	"net"
	"strconv"
	"sync"

	"github.com/pkg/errors"

	"github.com/arangodb/go-driver/agency"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/conn"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

type clientCache struct {
	mutex           sync.Mutex
	clients         map[string]driver.Client
	apiObjectGetter func() *api.ArangoDeployment

	databaseClient driver.Client

	factory conn.Factory
}

func newClientCache(apiObjectGetter func() *api.ArangoDeployment, factory conn.Factory) *clientCache {
	return &clientCache{
		clients:         make(map[string]driver.Client),
		apiObjectGetter: apiObjectGetter,
		factory:         factory,
	}
}

func (cc *clientCache) extendHost(host string) string {
	scheme := "http"
	if cc.apiObjectGetter().Spec.TLS.IsSecure() {
		scheme = "https"
	}

	return scheme + "://" + net.JoinHostPort(host, strconv.Itoa(k8sutil.ArangoPort))
}

func (cc *clientCache) getClient(ctx context.Context, group api.ServerGroup, id string) (driver.Client, error) {
	key := fmt.Sprintf("%d-%s", group, id)
	c, found := cc.clients[key]
	if found {
		return c, nil
	}

	// Not found, create a new client
	c, err := cc.factory.Client(cc.extendHost(k8sutil.CreatePodDNSName(cc.apiObjectGetter(), group.AsRole(), id)))
	if err != nil {
		return nil, maskAny(err)
	}
	cc.clients[key] = c
	return c, nil
}

func (cc *clientCache) get(ctx context.Context, group api.ServerGroup, id string) (driver.Client, error) {
	client, err := cc.getClient(ctx, group, id)
	if err != nil {
		return nil, maskAny(err)
	}

	if _, err := client.Version(ctx); err == nil {
		return client, nil
	} else if driver.IsUnauthorized(err) {
		delete(cc.clients, fmt.Sprintf("%d-%s", group, id))
		return cc.getClient(ctx, group, id)
	} else {
		return client, nil
	}
}

// Get a cached client for the given ID in the given group, creating one
// if needed.
func (cc *clientCache) Get(ctx context.Context, group api.ServerGroup, id string) (driver.Client, error) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	return cc.get(ctx, group, id)
}

func (cc *clientCache) getDatabaseClient() (driver.Client, error) {
	if c := cc.databaseClient; c != nil {
		return c, nil
	}

	// Not found, create a new client
	c, err := cc.factory.Client(cc.extendHost(k8sutil.CreateDatabaseClientServiceDNSName(cc.apiObjectGetter())))
	if err != nil {
		return nil, maskAny(err)
	}
	cc.databaseClient = c
	return c, nil
}

func (cc *clientCache) getDatabase(ctx context.Context) (driver.Client, error) {
	client, err := cc.getDatabaseClient()
	if err != nil {
		return nil, maskAny(err)
	}

	if _, err := client.Version(ctx); err == nil {
		return client, nil
	} else if driver.IsUnauthorized(err) {
		cc.databaseClient = nil
		return cc.getDatabaseClient()
	} else {
		return client, nil
	}
}

// GetDatabase returns a cached client for the entire database (cluster coordinators or single server),
// creating one if needed.
func (cc *clientCache) GetDatabase(ctx context.Context) (driver.Client, error) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	return cc.getDatabase(ctx)
}

func (cc *clientCache) getAgencyClient() (agency.Agency, error) {
	// Not found, create a new client
	var dnsNames []string
	for _, m := range cc.apiObjectGetter().Status.Members.Agents {
		dnsNames = append(dnsNames, cc.extendHost(k8sutil.CreatePodDNSName(cc.apiObjectGetter(), api.ServerGroupAgents.AsRole(), m.ID)))
	}

	if len(dnsNames) == 0 {
		return nil, errors.Errorf("There is no DNS Name")
	}

	c, err := cc.factory.Agency(dnsNames...)
	if err != nil {
		return nil, maskAny(err)
	}
	return c, nil
}

// GetDatabase returns a cached client for the agency
func (cc *clientCache) GetAgency(ctx context.Context) (agency.Agency, error) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	return cc.getAgencyClient()
}
