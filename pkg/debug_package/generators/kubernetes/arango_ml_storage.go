//
// DISCLAIMER
//
// Copyright 2023-2024 ArangoDB GmbH, Cologne, Germany
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

	mlApi "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/cli"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func mlStorages(logger zerolog.Logger, files chan<- shared.File, client kclient.Client) error {
	storages, err := listMLStorages(client)
	if err != nil {
		if kerrors.IsForbiddenOrNotFound(err) {
			return nil
		}

		return err
	}

	if err := errors.ExecuteWithErrorArrayP2(mlStorage, client, files, storages...); err != nil {
		logger.Err(err).Msgf("Error while collecting arango ml storages")
		return err
	}

	return nil
}

func mlStorage(client kclient.Client, files chan<- shared.File, ext *mlApi.ArangoMLStorage) error {
	files <- shared.NewYAMLFile(fmt.Sprintf("kubernetes/arango/ml/storages/%s.yaml", ext.GetName()), func() ([]interface{}, error) {
		return []interface{}{ext}, nil
	})

	return nil
}

func listMLStorages(client kclient.Client) ([]*mlApi.ArangoMLStorage, error) {
	return ListObjects[*mlApi.ArangoMLStorageList, *mlApi.ArangoMLStorage](context.Background(), client.Arango().MlV1beta1().ArangoMLStorages(cli.GetInput().Namespace), func(result *mlApi.ArangoMLStorageList) []*mlApi.ArangoMLStorage {
		q := make([]*mlApi.ArangoMLStorage, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}
