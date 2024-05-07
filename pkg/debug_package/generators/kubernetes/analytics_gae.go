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

	analyticsApi "github.com/arangodb/kube-arangodb/pkg/apis/analytics/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/cli"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func analyticsGAEs(logger zerolog.Logger, files chan<- shared.File, client kclient.Client) error {
	extensions, err := listAnalyticsGAEs(client)
	if err != nil {
		if kerrors.IsForbiddenOrNotFound(err) {
			return nil
		}

		return err
	}

	if err := errors.ExecuteWithErrorArrayP2(analyticsGAE, client, files, extensions...); err != nil {
		logger.Err(err).Msgf("Error while collecting arango ml extensions")
		return err
	}

	return nil
}

func analyticsGAE(client kclient.Client, files chan<- shared.File, ext *analyticsApi.GraphAnalyticsEngine) error {
	files <- shared.NewYAMLFile(fmt.Sprintf("kubernetes/arango/analytics/gaes/%s.yaml", ext.GetName()), func() ([]interface{}, error) {
		return []interface{}{ext}, nil
	})

	return nil
}

func listAnalyticsGAEs(client kclient.Client) ([]*analyticsApi.GraphAnalyticsEngine, error) {
	return ListObjects[*analyticsApi.GraphAnalyticsEngineList, *analyticsApi.GraphAnalyticsEngine](context.Background(), client.Arango().AnalyticsV1alpha1().GraphAnalyticsEngines(cli.GetInput().Namespace), func(result *analyticsApi.GraphAnalyticsEngineList) []*analyticsApi.GraphAnalyticsEngine {
		q := make([]*analyticsApi.GraphAnalyticsEngine, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}
