//
// DISCLAIMER
//
// Copyright 2023-2026 ArangoDB GmbH, Cologne, Germany
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

	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func kubernetesCoreServiceEndpointSlices(ctx context.Context, logger zerolog.Logger, client kclient.Client, files chan<- shared.File, item *core.Service) error {
	files <- shared.NewYAMLFile("endpoint-slice.yaml", func() ([]interface{}, error) {
		ep, err := client.Kubernetes().DiscoveryV1().EndpointSlices(item.GetNamespace()).List(ctx, meta.ListOptions{
			LabelSelector: labels.SelectorFromSet(map[string]string{
				discovery.LabelServiceName: item.GetName(),
			}).String(),
		})
		if err != nil {
			return nil, err
		}

		return util.FormatList(ep.Items, func(a discovery.EndpointSlice) any {
			ep := a.DeepCopy()
			ep.ManagedFields = nil
			return ep
		}), nil
	})
	return nil
}
