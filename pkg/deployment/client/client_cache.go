//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	"net"
	"strconv"
	"sync"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/agency"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconciler"
	"github.com/arangodb/kube-arangodb/pkg/handlers/utils"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/conn"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

type Cache interface {
	GetAuth() conn.Auth

	Connection(ctx context.Context, host string) (driver.Connection, error)

	GetRaw(group api.ServerGroup, id string) (conn.Connection, error)

	Get(ctx context.Context, group api.ServerGroup, id string) (driver.Client, error)
	GetDatabase(ctx context.Context) (driver.Client, error)
	GetAgency(ctx context.Context, agencyIDs ...string) (agency.Agency, error)
}

type CacheGen interface {
	reconciler.DeploymentEndpoints
	reconciler.DeploymentInfoGetter
}

func NewClientCache(in CacheGen, factory conn.Factory) Cache {
	return &cache{
		in:      in,
		factory: factory,
	}
}

type cache struct {
	mutex sync.Mutex
	in    CacheGen

	factory conn.Factory
}

func (cc *cache) GetRaw(group api.ServerGroup, id string) (conn.Connection, error) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	m, _, _ := cc.in.GetStatus().Members.ElementByID(id)

	endpoint, err := cc.in.GenerateMemberEndpoint(group, m)
	if err != nil {
		return nil, err
	}

	return cc.factory.RawConnection(cc.extendHost(m.GetEndpoint(endpoint)))
}

func (cc *cache) Connection(ctx context.Context, host string) (driver.Connection, error) {
	return cc.factory.Connection(host)
}

func (cc *cache) extendHost(host string) string {
	scheme := "http"
	if cc.in.GetSpec().TLS.IsSecure() {
		scheme = "https"
	}

	return scheme + "://" + net.JoinHostPort(host, strconv.Itoa(shared.ArangoPort))
}

func (cc *cache) getClient(group api.ServerGroup, id string) (driver.Client, error) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	m, _, _ := cc.in.GetStatus().Members.ElementByID(id)

	endpoint, err := cc.in.GenerateMemberEndpoint(group, m)
	if err != nil {
		return nil, err
	}

	c, err := cc.factory.Client(cc.extendHost(m.GetEndpoint(endpoint)))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// Get a cached client for the given ID in the given group, creating one
// if needed.
func (cc *cache) Get(ctx context.Context, group api.ServerGroup, id string) (driver.Client, error) {
	client, err := cc.getClient(group, id)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if _, err := client.Version(ctx); err == nil {
		return client, nil
	} else if driver.IsUnauthorized(err) {
		return cc.getClient(group, id)
	} else {
		return client, nil
	}
}

func (cc *cache) GetAuth() conn.Auth {
	return cc.factory.GetAuth()
}

func (cc *cache) getDatabaseClient() (driver.Client, error) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	c, err := cc.factory.Client(cc.extendHost(k8sutil.CreateDatabaseClientServiceDNSName(cc.in.GetAPIObject())))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// GetDatabase returns a cached client for the entire database (cluster coordinators or single server),
// creating one if needed.
func (cc *cache) GetDatabase(ctx context.Context) (driver.Client, error) {
	client, err := cc.getDatabaseClient()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if _, err := client.Version(ctx); err == nil {
		return client, nil
	} else if driver.IsUnauthorized(err) {
		return cc.getDatabaseClient()
	} else {
		return client, nil
	}
}

// GetAgency returns a cached client for the agency
func (cc *cache) GetAgency(_ context.Context, agencyIDs ...string) (agency.Agency, error) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	// Not found, create a new client
	var dnsNames []string
	for _, m := range cc.in.GetStatus().Members.Agents {
		if len(agencyIDs) > 0 {
			if !utils.StringList(agencyIDs).Has(m.ID) {
				continue
			}
		}

		endpoint, err := cc.in.GenerateMemberEndpoint(api.ServerGroupAgents, m)
		if err != nil {
			return nil, err
		}

		dnsNames = append(dnsNames, cc.extendHost(m.GetEndpoint(endpoint)))
	}

	if len(dnsNames) == 0 {
		return nil, errors.Newf("There is no DNS Name")
	}

	c, err := cc.factory.Agency(dnsNames...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}
