//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package pod

import (
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	core "k8s.io/api/core/v1"
)

func IsJWTEnabled(i Input) bool {
	return i.Deployment.Authentication.IsAuthenticated()
}

func MultiJWT(i Input) bool {
	return i.Version.CompareTo("3.7.0") < 0
}

func JWT() Builder {
	return jwt{}
}

type jwt struct{}

func (j jwt) Args(i Input) k8sutil.OptionPairs {
	panic("implement me")
}

func (j jwt) Volumes(i Input) ([]core.Volume, []core.VolumeMount) {
	panic("implement me")
}

func (j jwt) Verify(i Input, s k8sutil.SecretInterface) error {
	panic("implement me")
}
