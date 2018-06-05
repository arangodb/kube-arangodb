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

package replication

import (
	"net"
	"strconv"

	certificates "github.com/arangodb-helper/go-certificates"
	"github.com/arangodb/arangosync/client"
	"github.com/arangodb/arangosync/tasks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// createSyncMasterClient creates an arangosync client for the given endpoint.
func (dr *DeploymentReplication) createSyncMasterClient(epSpec api.EndpointSpec) (client.API, error) {
	log := dr.deps.Log

	// Endpoint
	source, err := dr.createArangoSyncEndpoint(epSpec)
	if err != nil {
		return nil, maskAny(err)
	}

	// Authentication
	insecureSkipVerify := true
	tlsAuth := tasks.TLSAuthentication{}
	authJWTSecretName, tlsCASecretName, err := dr.getEndpointSecretNames(epSpec)
	if err != nil {
		return nil, maskAny(err)
	}
	jwtSecret := ""
	if authJWTSecretName != "" {
		var err error
		jwtSecret, err = k8sutil.GetTokenSecret(dr.deps.KubeCli.CoreV1(), authJWTSecretName, dr.apiObject.GetNamespace())
		if err != nil {
			return nil, maskAny(err)
		}
	}
	if tlsCASecretName != "" {
		caCert, err := k8sutil.GetCACertficateSecret(dr.deps.KubeCli.CoreV1(), tlsCASecretName, dr.apiObject.GetNamespace())
		if err != nil {
			return nil, maskAny(err)
		}
		tlsAuth.CACertificate = caCert
	}
	auth := client.NewAuthentication(tlsAuth, jwtSecret)

	// Create client
	c, err := dr.clientCache.GetClient(log, source, auth, insecureSkipVerify)
	if err != nil {
		return nil, maskAny(err)
	}
	return c, nil
}

// createArangoSyncEndpoint creates the endpoints for the given spec.
func (dr *DeploymentReplication) createArangoSyncEndpoint(epSpec api.EndpointSpec) (client.Endpoint, error) {
	if epSpec.HasDeploymentName() {
		deploymentName := epSpec.GetDeploymentName()
		depls := dr.deps.CRCli.DatabaseV1alpha().ArangoDeployments(dr.apiObject.GetNamespace())
		depl, err := depls.Get(deploymentName, metav1.GetOptions{})
		if err != nil {
			dr.deps.Log.Debug().Err(err).Str("deployment", deploymentName).Msg("Failed to get deployment")
			return nil, maskAny(err)
		}
		dnsName := k8sutil.CreateSyncMasterClientServiceDNSName(depl)
		return client.Endpoint{"https://" + net.JoinHostPort(dnsName, strconv.Itoa(k8sutil.ArangoSyncMasterPort))}, nil
	}
	return client.Endpoint(epSpec.MasterEndpoint), nil
}

// createArangoSyncTLSAuthentication creates the authentication needed to authenticate
// the destination syncmaster at the source syncmaster.
func (dr *DeploymentReplication) createArangoSyncTLSAuthentication(spec api.DeploymentReplicationSpec) (client.TLSAuthentication, error) {
	// Fetch keyfile
	keyFileContent, err := k8sutil.GetTLSKeyfileSecret(dr.deps.KubeCli.CoreV1(), spec.Authentication.GetClientAuthSecretName(), dr.apiObject.GetNamespace())
	if err != nil {
		return client.TLSAuthentication{}, maskAny(err)
	}
	kf, err := certificates.NewKeyfile(keyFileContent)
	if err != nil {
		return client.TLSAuthentication{}, maskAny(err)
	}

	// Fetch TLS CA certificate for source
	_, tlsCASecretName, err := dr.getEndpointSecretNames(spec.Source)
	if err != nil {
		return client.TLSAuthentication{}, maskAny(err)
	}
	caCert, err := k8sutil.GetCACertficateSecret(dr.deps.KubeCli.CoreV1(), tlsCASecretName, dr.apiObject.GetNamespace())
	if err != nil {
		return client.TLSAuthentication{}, maskAny(err)
	}

	// Create authentication
	result := client.TLSAuthentication{
		TLSClientAuthentication: tasks.TLSClientAuthentication{
			ClientCertificate: kf.EncodeCertificates(),
			ClientKey:         kf.EncodePrivateKey(),
		},
		CACertificate: caCert,
	}
	return result, nil
}

// getEndpointSecretNames returns the names of secrets that hold the JWT token, TLS ca.crt.
func (dr *DeploymentReplication) getEndpointSecretNames(epSpec api.EndpointSpec) (authJWTSecretName, tlsCASecretName string, err error) {
	if epSpec.HasDeploymentName() {
		deploymentName := epSpec.GetDeploymentName()
		depls := dr.deps.CRCli.DatabaseV1alpha().ArangoDeployments(dr.apiObject.GetNamespace())
		depl, err := depls.Get(deploymentName, metav1.GetOptions{})
		if err != nil {
			dr.deps.Log.Debug().Err(err).Str("deployment", deploymentName).Msg("Failed to get deployment")
			return "", "", maskAny(err)
		}
		return depl.Spec.Sync.Authentication.GetJWTSecretName(), depl.Spec.Sync.TLS.GetCASecretName(), nil
	}
	return epSpec.Authentication.GetJWTSecretName(), epSpec.TLS.GetCASecretName(), nil
}
