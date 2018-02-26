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
	"net"
	nhttp "net/http"
	"strconv"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"k8s.io/client-go/kubernetes"

	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
	"github.com/arangodb/k8s-operator/pkg/util/k8sutil"
)

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
func CreateArangodClient(kubecli kubernetes.Interface, apiObject *api.ArangoDeployment, group api.ServerGroup, id string) (driver.Client, error) {
	// Create connection
	dnsName := k8sutil.CreatePodDNSName(apiObject, group.AsRole(), id)
	c, err := createArangodClientForDNSName(kubecli, apiObject, dnsName)
	if err != nil {
		return nil, maskAny(err)
	}
	return c, nil
}

// CreateArangodDatabaseClient creates a go-driver client for accessing the entire cluster (or single server).
func CreateArangodDatabaseClient(kubecli kubernetes.Interface, apiObject *api.ArangoDeployment) (driver.Client, error) {
	// Create connection
	dnsName := k8sutil.CreateDatabaseClientServiceDNSName(apiObject)
	c, err := createArangodClientForDNSName(kubecli, apiObject, dnsName)
	if err != nil {
		return nil, maskAny(err)
	}
	return c, nil
}

// CreateArangodClientForDNSName creates a go-driver client for a given DNS name.
func createArangodClientForDNSName(kubecli kubernetes.Interface, apiObject *api.ArangoDeployment, dnsName string) (driver.Client, error) {
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
		s, err := k8sutil.GetJWTSecret(kubecli, apiObject.Spec.Authentication.JWTSecretName, apiObject.GetNamespace())
		if err != nil {
			return nil, maskAny(err)
		}
		jwt, err := CreateArangodJwtAuthorizationHeader(s)
		if err != nil {
			return nil, maskAny(err)
		}
		config.Authentication = driver.RawAuthentication(jwt)
	}
	c, err := driver.NewClient(config)
	if err != nil {
		return nil, maskAny(err)
	}
	return c, nil
}
