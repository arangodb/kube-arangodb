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

	"helm.sh/helm/v3/pkg/repo"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
)

type chartManagerRepoVersion struct {
	manager *manager
	chart   *repo.ChartVersion
}

func (m chartManagerRepoVersion) Chart() *repo.ChartVersion {
	return m.chart
}

type ChartManagerRepoVersion interface {
	Get(ctx context.Context) (Chart, error)

	Chart() *repo.ChartVersion
}

func (m chartManagerRepoVersion) Get(ctx context.Context) (Chart, error) {
	if len(m.chart.URLs) == 0 {
		return nil, errors.Errorf("Chart `%s-%s` does not have any urls defined", m.chart.Name, m.chart.Version)
	}

	var errs = make([]error, len(m.chart.URLs))

	for id, url := range m.chart.URLs {
		data, err := operatorHTTP.Download(ctx, m.manager.client, url)
		if err != nil {
			errs[id] = err
			continue
		}

		return data, nil
	}

	return nil, errors.Errors(errs...)
}
