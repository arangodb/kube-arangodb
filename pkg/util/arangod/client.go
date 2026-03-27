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

package arangod

import (
	"context"
	"fmt"
	"net"
	goHttp "net/http"
	"strconv"

	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"

	adbDriverV2 "github.com/arangodb/go-driver/v2/arangodb"
	adbDriverV2Connection "github.com/arangodb/go-driver/v2/connection"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
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

func sharedHTTPTransport() goHttp.RoundTripper {
	return operatorHTTP.Transport()
}

func sharedHTTPSTransport() goHttp.RoundTripper {
	return operatorHTTP.Transport(operatorHTTP.WithTransportTLS(operatorHTTP.Insecure))
}

func sharedHTTPTransportShortTimeout() goHttp.RoundTripper {
	return operatorHTTP.RoundTripperWithShortTransport()
}

func sharedHTTPSTransportShortTimeout() goHttp.RoundTripper {
	return operatorHTTP.RoundTripperWithShortTransport(operatorHTTP.WithTransportTLS(operatorHTTP.Insecure))
}

// CreateArangodClient creates a go-driver client for a specific member in the given group.
func CreateArangodClient(ctx context.Context, cli typedCore.CoreV1Interface, apiObject *api.ArangoDeployment, group api.ServerGroup, id string, asyncSupport bool) (adbDriverV2.Client, error) {
	// Create connection
	dnsName := k8sutil.CreatePodDNSNameWithDomain(apiObject, apiObject.GetAcceptedSpec().ClusterDomain, group.AsRole(), id)
	c, err := createArangodClientForDNSName(ctx, cli, apiObject, dnsName, false, asyncSupport)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// CreateArangodDatabaseClient creates a go-driver client for accessing the entire cluster (or single server).
func CreateArangodDatabaseClient(ctx context.Context, cli typedCore.CoreV1Interface, apiObject *api.ArangoDeployment, shortTimeout bool, asyncSupport bool) (adbDriverV2.Client, error) {
	// Create connection
	dnsName := k8sutil.CreateDatabaseClientServiceDNSNameWithDomain(apiObject, apiObject.GetAcceptedSpec().ClusterDomain)
	c, err := createArangodClientForDNSName(ctx, cli, apiObject, dnsName, shortTimeout, asyncSupport)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// CreateArangodImageIDClient creates a go-driver client for an ArangoDB instance
// running in an Image-ID pod.
func CreateArangodImageIDClient(ctx context.Context, deployment k8sutil.APIObject, ip string, asyncSupport bool) (adbDriverV2.Client, error) {
	// Create connection
	c, err := createArangodClientForDNSName(ctx, nil, nil, ip, false, asyncSupport)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// CreateArangodClientForDNSName creates a go-driver client for a given DNS name.
func createArangodClientForDNSName(ctx context.Context, cli typedCore.CoreV1Interface, apiObject *api.ArangoDeployment, dnsName string, shortTimeout bool, asyncSupport bool) (adbDriverV2.Client, error) {
	connConfig := createArangodHTTPConfigForDNSNames(apiObject, []string{dnsName}, shortTimeout)
	// TODO deal with TLS with proper CA checking
	conn := adbDriverV2Connection.NewHttpConnection(connConfig)

	auth, err := createArangodClientAuthentication(ctx, cli, apiObject)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if err := conn.SetAuthentication(auth); err != nil {
		return nil, errors.WithStack(err)
	}

	if asyncSupport {
		// Wrap connection with async wrapper
		conn = adbDriverV2Connection.NewConnectionAsyncWrapper(conn)
	}

	// Create client
	return adbDriverV2.NewClient(conn), nil
}

// createArangodHTTPConfigForDNSNames creates a go-driver HTTP connection config for a given DNS names.
func createArangodHTTPConfigForDNSNames(apiObject *api.ArangoDeployment, dnsNames []string, shortTimeout bool) adbDriverV2Connection.HttpConfiguration {
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
	connConfig := adbDriverV2Connection.HttpConfiguration{
		Transport: transport(),
		Endpoint: adbDriverV2Connection.NewRoundRobinEndpoints(util.FormatList(dnsNames, func(a string) string {
			return scheme + "://" + net.JoinHostPort(a, strconv.Itoa(shared.ArangoPort))
		})),
	}
	return connConfig
}

// createArangodClientAuthentication creates a go-driver authentication for the servers in the given deployment.
func createArangodClientAuthentication(ctx context.Context, cli typedCore.CoreV1Interface, apiObject *api.ArangoDeployment) (adbDriverV2Connection.Authentication, error) {
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
			jwt, err := utilToken.NewClaims().With(utilToken.WithDefaultClaims(), utilToken.WithServerID("kube-arangodb")).Sign(s)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			return adbDriverV2Connection.NewHeaderAuth("Authorization", fmt.Sprintf("bearer %s", jwt)), nil
		}
	} else {
		// Authentication is not enabled.
		if ctx.Value(requireAuthenticationKey{}) != nil {
			// Context requires authentication
			return nil, errors.WithStack(errors.Errorf("Authentication is required by context, but not provided in API object"))
		}
	}
	return nil, nil
}
