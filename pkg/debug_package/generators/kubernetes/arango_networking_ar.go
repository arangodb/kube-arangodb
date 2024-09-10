//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package kubernetes

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/cli"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func networkingArangoRoutes(logger zerolog.Logger, files chan<- shared.File, client kclient.Client) error {
	arangoRoutes, err := listNetowkingArangoRoutes(client)
	if err != nil {
		if kerrors.IsForbiddenOrNotFound(err) {
			return nil
		}

		return err
	}

	if err := errors.ExecuteWithErrorArrayP2(networkingArangoRoute, client, files, arangoRoutes...); err != nil {
		logger.Err(err).Msgf("Error while collecting networking arango routes")
		return err
	}

	return nil
}

func networkingArangoRoute(client kclient.Client, files chan<- shared.File, ext *networkingApi.ArangoRoute) error {
	files <- shared.NewYAMLFile(fmt.Sprintf("kubernetes/arango/networking/arangoroutes/%s.yaml", ext.GetName()), func() ([]interface{}, error) {
		return []interface{}{ext}, nil
	})

	return nil
}

func listNetowkingArangoRoutes(client kclient.Client) ([]*networkingApi.ArangoRoute, error) {
	return ListObjects[*networkingApi.ArangoRouteList, *networkingApi.ArangoRoute](context.Background(), client.Arango().NetworkingV1alpha1().ArangoRoutes(cli.GetInput().Namespace), func(result *networkingApi.ArangoRouteList) []*networkingApi.ArangoRoute {
		q := make([]*networkingApi.ArangoRoute, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}
