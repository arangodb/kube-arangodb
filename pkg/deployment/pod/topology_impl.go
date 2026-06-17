//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package pod

import (
	"fmt"

	core "k8s.io/api/core/v1"

	deplTopology "github.com/arangodb/kube-arangodb/pkg/deployment/topology"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
)

func (t topology) Args(i Input) k8sutil.OptionPairs {
	return nil
}

func (t topology) Volumes(i Input) ([]core.Volume, []core.VolumeMount) {
	return nil, nil
}

func (t topology) Envs(i Input) []core.EnvVar {
	top := i.Member.Topology

	if top == nil || !i.Status.Topology.IsTopologyOwned(top) {
		return nil
	}

	return []core.EnvVar{
		{
			Name:  deplTopology.ArangoDBZone,
			Value: fmt.Sprintf("%d", top.Zone),
		},
	}
}

func (t topology) Verify(i Input, cachedStatus interfaces.Inspector) error {
	return nil
}
