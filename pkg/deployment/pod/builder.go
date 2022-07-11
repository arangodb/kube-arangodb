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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/go-driver"

	deploymentApi "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
)

type Input struct {
	ApiObject    meta.Object
	Deployment   deploymentApi.DeploymentSpec
	Status       deploymentApi.DeploymentStatus
	GroupSpec    deploymentApi.ServerGroupSpec
	Group        deploymentApi.ServerGroup
	Version      driver.Version
	Member       deploymentApi.MemberStatus
	ArangoMember deploymentApi.ArangoMember
	Enterprise   bool
	AutoUpgrade  bool
}

type Builder interface {
	Args(i Input) k8sutil.OptionPairs
	Volumes(i Input) ([]core.Volume, []core.VolumeMount)
	Envs(i Input) []core.EnvVar
	Verify(i Input, cachedStatus interfaces.Inspector) error
}
