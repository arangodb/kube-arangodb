//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package upgrade

import (
	"sort"
	"sync"

	"github.com/pkg/errors"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
)

var (
	upgradeLock sync.Mutex
	upgrades    Upgrades
)

func registerUpgrade(u Upgrade) {
	upgradeLock.Lock()
	defer upgradeLock.Unlock()

	upgrades = append(upgrades, u)
}

func RunUpgrade(obj api.ArangoDeployment, status *api.DeploymentStatus, cache interfaces.Inspector) (bool, error) {
	upgradeLock.Lock()
	defer upgradeLock.Unlock()

	if changed, err := upgrades.Execute(obj, status, cache); err != nil {
		return false, err
	} else if changed {
		v := upgrades[len(upgrades)-1].Version()
		status.Version = &v
		return true, nil
	} else {
		return false, nil
	}
}

type Upgrades []Upgrade

func (u Upgrades) Execute(obj api.ArangoDeployment, status *api.DeploymentStatus, cache interfaces.Inspector) (bool, error) {
	z := u.Sort()

	if err := z.Verify(); err != nil {
		return false, err
	}

	var v api.Version
	if status != nil && status.Version != nil {
		v = *status.Version
	}

	var changed bool

	for _, up := range z {
		if up.Version().Compare(v) < 0 {
			continue
		}
		changed = true
		if err := up.ArangoDeployment(obj, status, cache); err != nil {
			return false, err
		}
	}

	return changed, nil
}

func (u Upgrades) Verify() error {
	v := map[int]map[int]map[int][]int{}

	for _, z := range u {
		ver := z.Version()

		l1 := v[ver.Major]
		if l1 == nil {
			l1 = map[int]map[int][]int{}
		}

		l2 := l1[ver.Minor]
		if l2 == nil {
			l2 = map[int][]int{}
		}

		l3 := l2[ver.Patch]

		l3 = append(l3, ver.ID)

		l2[ver.Patch] = l3

		l1[ver.Minor] = l2

		v[ver.Major] = l1
	}

	for major, majorV := range v {
		for minor, minorV := range majorV {
			for patch, patchV := range minorV {
				for id := range patchV {
					if id+1 != patchV[id] {
						return errors.Errorf("Invalid version in %d.%d.%d - got %d, expected %d", major, minor, patch, patchV[id], id+1)
					}
				}
			}
		}
	}

	return nil
}

func (u Upgrades) Copy() Upgrades {
	c := make(Upgrades, len(u))

	copy(c, u)

	return c
}

func (u Upgrades) Sort() Upgrades {
	sort.Slice(u, func(i, j int) bool {
		return u[i].Version().Compare(u[j].Version()) < 0
	})
	return u
}

type Upgrade interface {
	Version() api.Version

	ArangoDeployment(obj api.ArangoDeployment, status *api.DeploymentStatus, cache interfaces.Inspector) error
}
