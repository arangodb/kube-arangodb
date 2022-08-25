//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package arangod

import (
	"context"
	"crypto/tls"
	"net"
	nhttp "net/http"
	"strconv"
	"time"

	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/arangodb/go-driver/jwt"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
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
			KeepAlive: 90 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	sharedHTTPSTransport = &nhttp.Transport{
		Proxy: nhttp.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 90 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}
	sharedHTTPTransportShortTimeout = &nhttp.Transport{
		Proxy: nhttp.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 100 * time.Millisecond,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       100 * time.Millisecond,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	sharedHTTPSTransportShortTimeout = &nhttp.Transport{
		Proxy: nhttp.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 100 * time.Millisecond,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       100 * time.Millisecond,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}
)

// CreateArangodClient creates a go-driver client for a specific member in the given group.
func CreateArangodClient(ctx context.Context, cli typedCore.CoreV1Interface, apiObject *api.ArangoDeployment, group api.ServerGroup, id string) (driver.Client, error) {
	// Create connection
	dnsName := k8sutil.CreatePodDNSNameWithDomain(apiObject, apiObject.GetAcceptedSpec().ClusterDomain, group.AsRole(), id)
	c, err := createArangodClientForDNSName(ctx, cli, apiObject, dnsName, false)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// CreateArangodDatabaseClient creates a go-driver client for accessing the entire cluster (or single server).
func CreateArangodDatabaseClient(ctx context.Context, cli typedCore.CoreV1Interface, apiObject *api.ArangoDeployment, shortTimeout bool) (driver.Client, error) {
	// Create connection
	dnsName := k8sutil.CreateDatabaseClientServiceDNSNameWithDomain(apiObject, apiObject.GetAcceptedSpec().ClusterDomain)
	c, err := createArangodClientForDNSName(ctx, cli, apiObject, dnsName, shortTimeout)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// CreateArangodImageIDClient creates a go-driver client for an ArangoDB instance
// running in an Image-ID pod.
func CreateArangodImageIDClient(ctx context.Context, deployment k8sutil.APIObject, ip string) (driver.Client, error) {
	// Create connection
	c, err := createArangodClientForDNSName(ctx, nil, nil, ip, false)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// CreateArangodClientForDNSName creates a go-driver client for a given DNS name.
func createArangodClientForDNSName(ctx context.Context, cli typedCore.CoreV1Interface, apiObject *api.ArangoDeployment, dnsName string, shortTimeout bool) (driver.Client, error) {
	connConfig := createArangodHTTPConfigForDNSNames(apiObject, []string{dnsName}, shortTimeout)
	// TODO deal with TLS with proper CA checking
	conn, err := http.NewConnection(connConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Create client
	config := driver.ClientConfig{
		Connection: conn,
	}
	auth, err := createArangodClientAuthentication(ctx, cli, apiObject)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	config.Authentication = auth
	c, err := driver.NewClient(config)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// createArangodHTTPConfigForDNSNames creates a go-driver HTTP connection config for a given DNS names.
func createArangodHTTPConfigForDNSNames(apiObject *api.ArangoDeployment, dnsNames []string, shortTimeout bool) http.ConnectionConfig {
	scheme := "http"
	transport := sharedHTTPTransport
	if shortTimeout {
		transport = sharedHTTPTransportShortTimeout
	}
	if apiObject != nil && apiObject.GetAcceptedSpec().IsSecure() {
		scheme = "https"
		transport = sharedHTTPSTransport
		if shortTimeout {
			transport = sharedHTTPSTransportShortTimeout
		}
	}
	connConfig := http.ConnectionConfig{
		Transport:          transport,
		DontFollowRedirect: true,
	}
	for _, dnsName := range dnsNames {
		connConfig.Endpoints = append(connConfig.Endpoints, scheme+"://"+net.JoinHostPort(dnsName, strconv.Itoa(shared.ArangoPort)))
	}
	return connConfig
}

// createArangodClientAuthentication creates a go-driver authentication for the servers in the given deployment.
func createArangodClientAuthentication(ctx context.Context, cli typedCore.CoreV1Interface, apiObject *api.ArangoDeployment) (driver.Authentication, error) {
	if apiObject != nil && apiObject.GetAcceptedSpec().IsAuthenticated() {
		// Authentication is enabled.
		// Should we skip using it?
		if ctx.Value(skipAuthenticationKey{}) == nil {
			secrets := cli.Secrets(apiObject.GetNamespace())
			ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
			defer cancel()
			s, err := k8sutil.GetTokenSecret(ctxChild, secrets, apiObject.GetAcceptedSpec().Authentication.GetJWTSecretName())
			if err != nil {
				return nil, errors.WithStack(err)
			}
			jwt, err := jwt.CreateArangodJwtAuthorizationHeader(s, "kube-arangodb")
			if err != nil {
				return nil, errors.WithStack(err)
			}
			return driver.RawAuthentication(jwt), nil
		}
	} else {
		// Authentication is not enabled.
		if ctx.Value(requireAuthenticationKey{}) != nil {
			// Context requires authentication
			return nil, errors.WithStack(errors.Newf("Authentication is required by context, but not provided in API object"))
		}
	}
	return nil, nil
}
