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
	goHttp "net/http"
	"strconv"

	adbDriverV2 "github.com/arangodb/go-driver/v2/arangodb"
	adbDriverV2Shared "github.com/arangodb/go-driver/v2/arangodb/shared"
	adbDriverV2Connection "github.com/arangodb/go-driver/v2/connection"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconciler"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

type Cache interface {
	GetAuth(ctx context.Context) (adbDriverV2Connection.Authentication, error)

	Connection(host string) (adbDriverV2Connection.Connection, error)

	Get(ctx context.Context, group api.ServerGroup, id string) (adbDriverV2.Client, error)

	GetConnection(group api.ServerGroup, id string) (adbDriverV2Connection.Connection, error)

	GetDatabase(ctx context.Context) (adbDriverV2.Client, error)
}

type CacheGen interface {
	reconciler.DeploymentEndpoints
	reconciler.DeploymentInfoGetter
}

func NewClientCache(in CacheGen, auth cache.Object[adbDriverV2Connection.Authentication], config cache.Object[goHttp.RoundTripper]) Cache {
	return &cacheObject{
		in:     in,
		auth:   auth,
		config: config,
	}
}

type cacheObject struct {
	in CacheGen

	auth   cache.Object[adbDriverV2Connection.Authentication]
	config cache.Object[goHttp.RoundTripper]
}

func (cc *cacheObject) Connection(host string) (adbDriverV2Connection.Connection, error) {
	transport, err := cc.config.Get(shutdown.Context())
	if err != nil {
		return nil, err
	}

	auth, err := cc.auth.Get(shutdown.Context())
	if err != nil {
		return nil, err
	}

	conn := adbDriverV2Connection.HttpConfiguration{
		Authentication:     auth,
		Endpoint:           cc.extendHosts(host),
		Transport:          transport,
		DontFollowRedirect: true,
	}

	c := adbDriverV2Connection.NewHttpConnection(conn)

	if err := c.SetAuthentication(auth); err != nil {
		return nil, err
	}

	return c, nil
}

func (cc *cacheObject) Client(host string) (adbDriverV2.Client, error) {
	conn, err := cc.Connection(host)
	if err != nil {
		return nil, err
	}

	return adbDriverV2.NewClient(conn), nil
}

func (cc *cacheObject) extendHosts(hosts ...string) adbDriverV2Connection.Endpoint {
	scheme := "http"
	if cc.in.GetSpec().TLS.IsSecure() {
		scheme = "https"
	}

	return adbDriverV2Connection.NewRoundRobinEndpoints(util.FormatList(hosts, func(host string) string {
		return scheme + "://" + net.JoinHostPort(host, strconv.Itoa(shared.ArangoPort))
	}))

}

func (cc *cacheObject) getConnection(group api.ServerGroup, id string) (adbDriverV2Connection.Connection, error) {
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

func (cc *cacheObject) getClient(group api.ServerGroup, id string) (adbDriverV2.Client, error) {
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
func (cc *cacheObject) GetConnection(group api.ServerGroup, id string) (adbDriverV2Connection.Connection, error) {
	return cc.getConnection(group, id)
}

// Get a cached client for the given ID in the given group, creating one
// if needed.
func (cc *cacheObject) Get(ctx context.Context, group api.ServerGroup, id string) (adbDriverV2.Client, error) {
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

func (cc *cacheObject) GetAuth(ctx context.Context) (adbDriverV2Connection.Authentication, error) {
	return cc.auth.Get(ctx)
}

func (cc *cacheObject) getDatabaseClient() (adbDriverV2.Client, error) {
	c, err := cc.Client(k8sutil.CreateDatabaseClientServiceDNSName(cc.in.GetAPIObject()))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// GetDatabase returns a cached client for the entire database (cluster coordinators or single server),
// creating one if needed.
func (cc *cacheObject) GetDatabase(ctx context.Context) (adbDriverV2.Client, error) {
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
