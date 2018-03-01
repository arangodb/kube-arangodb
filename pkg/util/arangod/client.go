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

package arangod

import (
	"context"
	"fmt"
	"net"
	nhttp "net/http"
	"strconv"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
	"github.com/arangodb/k8s-operator/pkg/util/k8sutil"
)

type (
	// skipAuthenticationKey is the context key used to indicate NOT setting any authentication
	skipAuthenticationKey struct{}
	// requireAuthenticationKey is the context key used to indicate that authentication is required
	requireAuthenticationKey struct{}
)

// WithSkipAuthentication prepares a context that when given to functions in
// this file will avoid creating any authentication for arango clients.
func WithSkipAuthentication(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipAuthenticationKey{}, true)
}

// WithRequireAuthentication prepares a context that when given to functions in
// this file will fail when authentication is not available.
func WithRequireAuthentication(ctx context.Context) context.Context {
	return context.WithValue(ctx, requireAuthenticationKey{}, true)
}

var (
	sharedHTTPTransport = &nhttp.Transport{
		Proxy: nhttp.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
)

// CreateArangodClient creates a go-driver client for a specific member in the given group.
func CreateArangodClient(ctx context.Context, cli corev1.CoreV1Interface, apiObject *api.ArangoDeployment, group api.ServerGroup, id string) (driver.Client, error) {
	// Create connection
	dnsName := k8sutil.CreatePodDNSName(apiObject, group.AsRole(), id)
	c, err := createArangodClientForDNSName(ctx, cli, apiObject, dnsName)
	if err != nil {
		return nil, maskAny(err)
	}
	return c, nil
}

// CreateArangodDatabaseClient creates a go-driver client for accessing the entire cluster (or single server).
func CreateArangodDatabaseClient(ctx context.Context, cli corev1.CoreV1Interface, apiObject *api.ArangoDeployment) (driver.Client, error) {
	// Create connection
	dnsName := k8sutil.CreateDatabaseClientServiceDNSName(apiObject)
	c, err := createArangodClientForDNSName(ctx, cli, apiObject, dnsName)
	if err != nil {
		return nil, maskAny(err)
	}
	return c, nil
}

// CreateArangodClientForDNSName creates a go-driver client for a given DNS name.
func createArangodClientForDNSName(ctx context.Context, cli corev1.CoreV1Interface, apiObject *api.ArangoDeployment, dnsName string) (driver.Client, error) {
	scheme := "http"
	connConfig := http.ConnectionConfig{
		Endpoints: []string{scheme + "://" + net.JoinHostPort(dnsName, strconv.Itoa(k8sutil.ArangoPort))},
		Transport: sharedHTTPTransport,
	}
	// TODO deal with TLS
	conn, err := http.NewConnection(connConfig)
	if err != nil {
		return nil, maskAny(err)
	}

	// Create client
	config := driver.ClientConfig{
		Connection: conn,
	}
	if apiObject.Spec.IsAuthenticated() {
		// Authentication is enabled.
		// Should we skip using it?
		if ctx.Value(skipAuthenticationKey{}) == nil {
			s, err := k8sutil.GetJWTSecret(cli, apiObject.Spec.Authentication.JWTSecretName, apiObject.GetNamespace())
			if err != nil {
				return nil, maskAny(err)
			}
			jwt, err := CreateArangodJwtAuthorizationHeader(s)
			if err != nil {
				return nil, maskAny(err)
			}
			config.Authentication = driver.RawAuthentication(jwt)
		}
	} else {
		// Authentication is not enabled.
		if ctx.Value(requireAuthenticationKey{}) != nil {
			// Context requires authentication
			return nil, maskAny(fmt.Errorf("Authentication is required by context, but not provided in API object"))
		}
	}
	c, err := driver.NewClient(config)
	if err != nil {
		return nil, maskAny(err)
	}
	return c, nil
}
