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

package pod

import (
	core "k8s.io/api/core/v1"

	deploymentApi "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
)

func AutoUpgrade() Builder {
	return autoUpgrade{}
}

type autoUpgrade struct{}

func (u autoUpgrade) Envs(i Input) []core.EnvVar {
	return nil
}

func (u autoUpgrade) Verify(i Input, cachedStatus interfaces.Inspector) error {
	return nil
}

func (u autoUpgrade) Volumes(i Input) ([]core.Volume, []core.VolumeMount) {
	return nil, nil
}

func (u autoUpgrade) Args(i Input) k8sutil.OptionPairs {
	// Always add upgrade flag due to fact it is now only in initContainers
	if i.Version.CompareTo("3.6.0") >= 0 {
		switch i.Group {
		case deploymentApi.ServerGroupCoordinators:
			return k8sutil.NewOptionPair(k8sutil.OptionPair{Key: "--cluster.upgrade", Value: "online"})
		}
	}

	return k8sutil.NewOptionPair(k8sutil.OptionPair{Key: "--database.auto-upgrade", Value: "true"})
}
