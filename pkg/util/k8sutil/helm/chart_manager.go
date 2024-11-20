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
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"

	"helm.sh/helm/v3/pkg/repo"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func NewChartManager(ctx context.Context, client *http.Client, url string) (ChartManager, error) {
	if client == nil {
		client = http.DefaultClient
	}

	m := manager{
		client: client,
		url:    url,
	}

	if err := m.Reload(ctx); err != nil {
		return nil, err
	}

	return &m, nil
}

type ChartManager interface {
	Reload(ctx context.Context) error

	Repositories() []string
	Versions(repo string) []string
	Latest(repo string) (string, bool)
	Chart(ctx context.Context, repo, version string) (Chart, error)
}

type manager struct {
	lock sync.Mutex

	client *http.Client

	url string

	index *repo.IndexFile
}

func (m *manager) Latest(repoName string) (string, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if v := m.latest(repoName); v == nil {
		return "", false
	} else {
		return v.Version, true
	}
}

func (m *manager) latest(repoName string) *repo.ChartVersion {
	r, ok := m.index.Entries[repoName]
	if !ok {
		return nil
	}

	if len(r) == 0 {
		return nil
	}

	if len(r) == 1 {
		return r[0]
	}

	var p = 0

	for id := range r {
		if id == p {
			continue
		}

		if r[id].Created.After(r[p].Created) {
			p = id
		}
	}

	if p == -1 {
		return nil
	}

	return r[p]
}

func (m *manager) Chart(ctx context.Context, repoName, version string) (Chart, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	r, ok := m.index.Entries[repoName]
	if !ok {
		return nil, errors.Errorf("Repo `%s` not found", repoName)
	}

	var ver *repo.ChartVersion

	if version == "latest" {
		ver = m.latest(repoName)
	} else {
		vs, ok := util.PickFromList(r, func(v *repo.ChartVersion) bool {
			if v == nil {
				return false
			}

			return v.Version == version
		})
		if !ok {
			return nil, errors.Errorf("Repo `%s` does not contains version `%s`", repoName, version)
		}
		ver = vs
	}

	if len(ver.URLs) == 0 {
		return nil, errors.Errorf("Chart `%s-%s` does not have any urls defined", repoName, version)
	}

	var errs = make([]error, len(ver.URLs))

	for id, url := range ver.URLs {
		data, err := m.download(ctx, url)
		if err != nil {
			errs[id] = err
			continue
		}

		return data, nil
	}

	return nil, errors.Errors(errs...)
}

func (m *manager) download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		if err := resp.Body.Close(); err != nil {
			return nil, err
		}

		return nil, err
	}

	if err := resp.Body.Close(); err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("Unexpected code: %d", resp.StatusCode)
	}

	return data, nil
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

func (m *manager) Versions(repo string) []string {
	m.lock.Lock()
	defer m.lock.Unlock()

	r, ok := m.index.Entries[repo]
	if !ok {
		return nil
	}

	if len(r) == 0 {
		return []string{}
	}

	s := make([]string, 0, len(r))

	for _, v := range r {
		s = append(s, v.Version)
	}

	sort.Strings(s)

	return s
}

func (m *manager) Reload(ctx context.Context) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	data, err := m.download(ctx, fmt.Sprintf("%s/index.yaml", m.url))
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
