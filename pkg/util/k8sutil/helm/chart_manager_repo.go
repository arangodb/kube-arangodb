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
	"slices"
	"sort"

	"helm.sh/helm/v3/pkg/repo"
)

type ChartManagerRepo interface {
	Versions() []string

	Get(version string) (ChartManagerRepoVersion, bool)
	Latest() (ChartManagerRepoVersion, bool)
}

type chartManagerRepo struct {
	manager *manager
	name    string
}

func (c chartManagerRepo) Versions() []string {
	c.manager.lock.Lock()
	defer c.manager.lock.Unlock()

	r, ok := c.manager.index.Entries[c.name]
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

func (c chartManagerRepo) Get(version string) (ChartManagerRepoVersion, bool) {
	if version == "latest" {
		return c.Latest()
	}

	r, ok := c.manager.index.Entries[c.name]
	if !ok {
		return nil, false
	}

	for _, z := range r {
		if z.Version == version {
			return chartManagerRepoVersion{
				manager: c.manager,
				chart:   z,
			}, true
		}
	}

	return nil, false
}

func (c chartManagerRepo) Latest() (ChartManagerRepoVersion, bool) {
	c.manager.lock.Lock()
	defer c.manager.lock.Unlock()

	if v := c.find(func(a, b *repo.ChartVersion) int {
		return a.Created.Compare(b.Created) * -1
	}); v == nil {
		return nil, false
	} else {
		return chartManagerRepoVersion{
			manager: c.manager,
			chart:   v,
		}, true
	}
}

func (c chartManagerRepo) find(predicate func(a, b *repo.ChartVersion) int) *repo.ChartVersion {
	r, ok := c.manager.index.Entries[c.name]
	if !ok {
		return nil
	}

	if len(r) == 0 {
		return nil
	}

	if len(r) == 1 {
		return r[0]
	}

	if predicate == nil {
		return r[0]
	}

	z := make(repo.ChartVersions, len(r))
	copy(z, r)

	slices.SortStableFunc(z, predicate)

	return z[0]
}
