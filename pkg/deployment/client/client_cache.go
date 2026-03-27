//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

	adbDriverV2 "github.com/arangodb/go-driver/v2/arangodb"
	adbDriverV2Shared "github.com/arangodb/go-driver/v2/arangodb/shared"
	adbDriverV2Connection "github.com/arangodb/go-driver/v2/connection"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconciler"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

type Cache interface {
	GetAuth() Auth

	Connection(host string) (adbDriverV2Connection.Connection, error)

	Get(ctx context.Context, group api.ServerGroup, id string) (adbDriverV2.Client, error)

	GetConnection(group api.ServerGroup, id string) (adbDriverV2Connection.Connection, error)

	GetDatabase(ctx context.Context) (adbDriverV2.Client, error)
}

type CacheGen interface {
	reconciler.DeploymentEndpoints
	reconciler.DeploymentInfoGetter
}

type Auth func() (adbDriverV2Connection.Authentication, error)

type Config func() (adbDriverV2Connection.HttpConfiguration, error)

func NewClientCache(in CacheGen, auth Auth, config Config) Cache {
	return &cache{
		in:     in,
		auth:   auth,
		config: config,
	}
}

type cache struct {
	mutex sync.Mutex
	in    CacheGen

	auth   Auth
	config Config
}

func (cc *cache) Connection(host string) (adbDriverV2Connection.Connection, error) {
	conn, err := cc.config()
	if err != nil {
		return nil, err
	}

	conn.Endpoint = adbDriverV2Connection.NewRoundRobinEndpoints([]string{host})

	auth, err := cc.auth()
	if err != nil {
		return nil, err
	}

	c := adbDriverV2Connection.NewHttpConnection(conn)

	if err := c.SetAuthentication(auth); err != nil {
		return nil, err
	}

	return c, nil
}

func (cc *cache) Client(host string) (adbDriverV2.Client, error) {
	conn, err := cc.Connection(host)
	if err != nil {
		return nil, err
	}

	return adbDriverV2.NewClient(conn), nil
}

func (cc *cache) extendHost(host string) string {
	scheme := "http"
	if cc.in.GetSpec().TLS.IsSecure() {
		scheme = "https"
	}

	return scheme + "://" + net.JoinHostPort(host, strconv.Itoa(shared.ArangoPort))
}

func (cc *cache) getConnection(group api.ServerGroup, id string) (adbDriverV2Connection.Connection, error) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	m, _, _ := cc.in.GetStatus().Members.ElementByID(id)

	endpoint, err := cc.in.GenerateMemberEndpoint(group, m)
	if err != nil {
		return nil, err
	}

	c, err := cc.Connection(endpoint)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

func (cc *cache) getClient(group api.ServerGroup, id string) (adbDriverV2.Client, error) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	m, _, _ := cc.in.GetStatus().Members.ElementByID(id)

	endpoint, err := cc.in.GenerateMemberEndpoint(group, m)
	if err != nil {
		return nil, err
	}

	c, err := cc.Client(endpoint)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// Get a cached client for the given ID in the given group, creating one
// if needed.
func (cc *cache) GetConnection(group api.ServerGroup, id string) (adbDriverV2Connection.Connection, error) {
	return cc.getConnection(group, id)
}

// Get a cached client for the given ID in the given group, creating one
// if needed.
func (cc *cache) Get(ctx context.Context, group api.ServerGroup, id string) (adbDriverV2.Client, error) {
	client, err := cc.getClient(group, id)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if _, err := client.Version(ctx); err == nil {
		return client, nil
	} else if adbDriverV2Shared.IsUnauthorized(err) {
		return cc.getClient(group, id)
	} else {
		return client, nil
	}
}

func (cc *cache) GetAuth() Auth {
	return cc.auth
}

func (cc *cache) getDatabaseClient() (adbDriverV2.Client, error) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	c, err := cc.Client(cc.extendHost(k8sutil.CreateDatabaseClientServiceDNSName(cc.in.GetAPIObject())))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// GetDatabase returns a cached client for the entire database (cluster coordinators or single server),
// creating one if needed.
func (cc *cache) GetDatabase(ctx context.Context) (adbDriverV2.Client, error) {
	client, err := cc.getDatabaseClient()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if _, err := client.Version(ctx); err == nil {
		return client, nil
	} else if adbDriverV2Shared.IsUnauthorized(err) {
		return cc.getDatabaseClient()
	} else {
		return client, nil
	}
}
