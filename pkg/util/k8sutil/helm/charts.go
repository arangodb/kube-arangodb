//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package helm

import (
	"context"

	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/list"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func GetLocalCharts(ctx context.Context, client kclient.Client, namespace string) (map[string]*platformApi.ArangoPlatformChart, error) {
	l, err := list.ListObjects[*platformApi.ArangoPlatformChartList, *platformApi.ArangoPlatformChart](ctx, client.Arango().PlatformV1alpha1().ArangoPlatformCharts(namespace), func(result *platformApi.ArangoPlatformChartList) []*platformApi.ArangoPlatformChart {
		q := make([]*platformApi.ArangoPlatformChart, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
	if err != nil {
		return nil, err
	}

	return util.ListAsMap(l, func(in *platformApi.ArangoPlatformChart) string {
		return in.GetName()
	}), nil
}
