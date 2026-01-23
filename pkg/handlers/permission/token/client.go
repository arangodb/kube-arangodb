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

package token

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"

	"github.com/arangodb/go-driver/v2/arangodb"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/client"
	"github.com/arangodb/kube-arangodb/pkg/util/http"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

type clientProvider interface {
	ArangoClient(ctx context.Context, client kubernetes.Interface, depl *api.ArangoDeployment) (arangodb.Client, error)
}

type clientProviderFunc func(ctx context.Context, client kubernetes.Interface, depl *api.ArangoDeployment) (arangodb.Client, error)

func (c clientProviderFunc) ArangoClient(ctx context.Context, client kubernetes.Interface, depl *api.ArangoDeployment) (arangodb.Client, error) {
	return c(ctx, client, depl)
}

func arangoClientProvider(ctx context.Context, c kubernetes.Interface, depl *api.ArangoDeployment) (arangodb.Client, error) {
	return client.NewFactory(client.DirectArangoDBAuthentication(c, depl), client.HTTPClientFactory(
		http.ShortTransport(),
		http.WithTransportTLS(http.Insecure),
	)).Client(ctx, fmt.Sprintf("%s://%s:%d", util.BoolSwitch(depl.GetAcceptedSpec().IsSecure(), "https", "http"), k8sutil.CreateDatabaseClientServiceDNSNameWithDomain(depl, depl.GetAcceptedSpec().ClusterDomain), shared.ArangoPort))
}
