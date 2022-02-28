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

package replication

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

type serverEndpoint struct {
	dr      *DeploymentReplication
	getSpec func() api.EndpointSpec
}

// DeploymentName returns the name of the ArangoDeployment of this endpoint
func (ep serverEndpoint) DeploymentName() string {
	return ep.getSpec().GetDeploymentName()
}

// MasterEndpoint returns the URLs of the custom master endpoint
func (ep serverEndpoint) MasterEndpoint() []string {
	return ep.getSpec().MasterEndpoint
}

// AuthKeyfileSecretName returns the name of a Secret containing the authentication keyfile
// for accessing the syncmaster at this endpoint
func (ep serverEndpoint) AuthKeyfileSecretName() string {
	return ep.getSpec().Authentication.GetKeyfileSecretName()
}

// AuthUserSecretName returns the name of a Secret containing the authentication username+password
// for accessing the syncmaster at this endpoint
func (ep serverEndpoint) AuthUserSecretName() string {
	return ep.getSpec().Authentication.GetUserSecretName()
}

// TLSCACert returns a PEM encoded TLS CA certificate of the syncmaster at this endpoint
func (ep serverEndpoint) TLSCACert() string {
	tlsCASecretName := ep.getSpec().TLS.GetCASecretName()
	secrets := ep.dr.deps.Client.Kubernetes().CoreV1().Secrets(ep.dr.apiObject.GetNamespace())
	caCert, err := k8sutil.GetCACertficateSecret(context.TODO(), secrets, tlsCASecretName)
	if err != nil {
		return ""
	}
	return caCert
}

// TLSCACertSecretName returns the name of a Secret containing the TLS CA certificate of the syncmaster at this endpoint
func (ep serverEndpoint) TLSCACertSecretName() string {
	return ep.getSpec().TLS.GetCASecretName()
}
