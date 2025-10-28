//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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
	"fmt"
	goHttp "net/http"
	"sort"
	"sync"

	"helm.sh/helm/v3/pkg/repo"

	"github.com/arangodb/kube-arangodb/pkg/util"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
)

func NewChartManager(ctx context.Context, client *goHttp.Client, format string, args ...interface{}) (ChartManager, error) {
	if client == nil {
		client = goHttp.DefaultClient
	}

	m := manager{
		client: client,
		url:    fmt.Sprintf(format, args...),
	}

	if err := m.Reload(ctx); err != nil {
		return nil, err
	}

	return &m, nil
}

type ChartManager interface {
	Reload(ctx context.Context) error

	Repositories() []string

	Get(name string) (ChartManagerRepo, bool)
}

type manager struct {
	lock sync.Mutex

	client *goHttp.Client

	url string

	index *repo.IndexFile
}

func (m *manager) Get(name string) (ChartManagerRepo, bool) {
	for entry := range m.index.Entries {
		if entry == name {
			return chartManagerRepo{
				manager: m,
				name:    name,
			}, true
		}
	}

	return nil, false
}

func (m *manager) Repositories() []string {
	m.lock.Lock()
	defer m.lock.Unlock()

	var s = make([]string, 0, len(m.index.Entries))

	for v := range m.index.Entries {
		s = append(s, v)
	}

	sort.Strings(s)

	return s
}

func (m *manager) Reload(ctx context.Context) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	data, err := operatorHTTP.Download(ctx, m.client, "%s", m.url)
	if err != nil {
		return err
	}

	idx, err := util.JsonOrYamlUnmarshal[repo.IndexFile](data)
	if err != nil {
		return err
	}

	idx.SortEntries()

	m.index = &idx

	return nil
}
