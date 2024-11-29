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

package helm

import (
	"helm.sh/helm/v3/pkg/chart"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ChartData interface {
	Chart() *chart.Chart

	Platform() (*Platform, error)
}

type chartData struct {
	chart *chart.Chart
}

func (c chartData) Platform() (*Platform, error) {
	return extractPlatform(c.chart)
}

func (c chartData) Chart() *chart.Chart {
	return c.chart
}

func newChartFromData(data []byte) (ChartData, error) {
	var c chartData

	v, err := newChartReaderFromBytes(data)
	if err != nil {
		return nil, errors.Errorf("Unable to load chart")
	}

	c.chart = v

	return c, nil
}
