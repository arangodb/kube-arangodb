//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package integration

import (
	"crypto/tls"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/client-go/kubernetes"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authentication"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
	utilTokenLoader "github.com/arangodb/kube-arangodb/pkg/util/token/loader"
)

func NewIntegrationConnection() (grpc.ClientConnInterface, error) {
	addr, ok := utilConstants.CENTRAL_INTEGRATION_SERVICE_ADDRESS.Lookup()
	if !ok {
		return nil, errors.Errorf("Integration Service Address not found")
	}

	var opts []grpc.DialOption

	if v, ok := utilConstants.CENTRAL_INTEGRATION_SECURED.Lookup(); ok && v == "true" {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	opts = append(opts, authentication.NewInterceptorClientOptions(authentication.NewArangoTokenAuthentication())...)

	return grpc.NewClient(addr, opts...)
}

func NewIntegrationConnectionFromDeployment(client kubernetes.Interface, depl *api.ArangoDeployment, mods ...util.ModR[utilToken.Claims]) (*grpc.ClientConn, error) {
	spec := depl.GetAcceptedSpec()

	if !depl.Status.Conditions.IsTrue(api.ConditionTypeGatewaySidecarEnabled) {
		return nil, errors.Errorf("Integration Service is not enabled")
	}

	auth := cache.NewObject[utilToken.Secret](utilTokenLoader.SecretCacheSecretAPI(client.CoreV1().Secrets(depl.GetNamespace()), pod.JWTSecretFolder(depl.GetName()), 15*time.Second))

	var opts []grpc.DialOption

	if spec.IsSecure() {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	opts = append(opts, authentication.NewInterceptorClientOptions(authentication.NewSecretAuthentication(auth, mods...))...)

	return grpc.NewClient(fmt.Sprintf("%s:%d", k8sutil.CreateSidecarClientServiceName(depl.GetName()), shared.InternalSidecarContainerPortGRPC), opts...)
}
