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
	"bytes"
	"io"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

type Chart []byte

func (c Chart) Get() (*chart.Chart, error) {
	return newChartReaderFromBytes(c)
}

func newChartReaderFromBytes(in []byte) (*chart.Chart, error) {
	return newChartReader(bytes.NewBuffer(in))
}

func newChartReader(in io.Reader) (*chart.Chart, error) {
	files, err := loader.LoadArchiveFiles(in)
	if err != nil {
		return nil, err
	}

	return loader.LoadFiles(files)
}
