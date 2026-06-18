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

package shared

import (
	"context"
	"fmt"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/compare/k8s"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helpers"
)

const (
	SecretKeyArangoDBHosts      = "ARANGODB_HOSTS"
	SecretKeyArangoDBTLS        = "ARANGODB_TLS"
	SecretKeyArangoDBCACertFile = "ARANGODB_CA_CERT_FILE"
)

func OperatorSecret(handler helpers.Updator[*core.Secret], ctx context.Context, namespace, prefix string, parent k8sutil.K8SObject, ref **sharedApi.Object, data map[string][]byte) (*core.Secret, bool, error) {
	return handler.OperatorUpdate(ctx, namespace, parent, ref, func(ctx context.Context, ref *sharedApi.Object) (*core.Secret, bool, string, error) {
		expectedSecret := &core.Secret{
			ObjectMeta: meta.ObjectMeta{
				GenerateName: prefix,
				Namespace:    namespace,
				OwnerReferences: []meta.OwnerReference{
					parent.AsOwner(),
				},
			},
			Data: data,
			Type: core.SecretTypeOpaque,
		}
		expectedChecksum, err := k8s.CoreSecretChecksum(expectedSecret)
		if err != nil {
			return nil, false, "", err
		}
		return expectedSecret, false, expectedChecksum, nil
	})
}
func OperatorDeploymentAccessKeys(handler helpers.Updator[*core.Secret], ctx context.Context, namespace, prefix string, parent k8sutil.K8SObject, ref **sharedApi.Object, deployment *api.ArangoDeployment, hosts, tls, tlsCertFile string) (*core.Secret, bool, error) {
	data, err := RenderDeploymentAccessKeys(deployment, hosts, tls, tlsCertFile)
	if err != nil {
		return nil, false, err
	}
	return OperatorSecret(handler, ctx, namespace, prefix, parent, ref, data)
}
func RenderDeploymentAccessKeys(deployment *api.ArangoDeployment, hosts, tls, tlsCertFile string) (map[string][]byte, error) {
	data := map[string][]byte{}
	data[hosts] = []byte(fmt.Sprintf("%s://%s.%s.svc:%d", util.BoolSwitch(deployment.GetAcceptedSpec().TLS.IsSecure(), "https", "http"), deployment.GetName(), deployment.GetNamespace(), shared.ArangoPort))
	// TLS
	if deployment.GetAcceptedSpec().TLS.IsSecure() {
		data[tls] = []byte("1")
		data[tlsCertFile] = []byte(fmt.Sprintf("/etc/arangodb/tls/%s", resources.CACertName))
	} else {
		data[tls] = []byte("0")
	}
	return data, nil
}
