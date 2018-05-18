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
	certificates "github.com/arangodb-helper/go-certificates"
	"github.com/arangodb/arangosync/client"
	"github.com/arangodb/arangosync/tasks"

	api "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// createSyncMasterClient creates an arangosync client for the given endpoint.
func (dr *DeploymentReplication) createSyncMasterClient(epSpec api.EndpointSpec) (client.API, error) {
	log := dr.deps.Log

	// Endpoint
	source := dr.createArangoSyncEndpoint(epSpec)

	// Authentication
	insecureSkipVerify := true
	tlsAuth := tasks.TLSAuthentication{}
	jwtSecret := ""
	if jwtSecretName := epSpec.Authentication.GetJWTSecretName(); jwtSecretName != "" {
		var err error
		jwtSecret, err = k8sutil.GetJWTSecret(dr.deps.KubeCli.CoreV1(), jwtSecretName, dr.apiObject.GetNamespace())
		if err != nil {
			return nil, maskAny(err)
		}
	}
	if caSecretName := epSpec.TLS.GetCASecretName(); caSecretName != "" {
		caCert, err := k8sutil.GetCACertficateSecret(dr.deps.KubeCli.CoreV1(), caSecretName, dr.apiObject.GetNamespace())
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
func (dr *DeploymentReplication) createArangoSyncEndpoint(epSpec api.EndpointSpec) client.Endpoint {
	// TODO when adding deploymentname to EndpointSpec, reflect that here
	return client.Endpoint(epSpec.MasterEndpoint)
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
	caCert, err := k8sutil.GetCACertficateSecret(dr.deps.KubeCli.CoreV1(), spec.Source.TLS.GetCASecretName(), dr.apiObject.GetNamespace())
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
