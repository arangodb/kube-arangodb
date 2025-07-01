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

package platform

import (
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
)

func getHelmPackages(files ...string) (helm.Package, error) {
	if len(files) == 0 {
		return helm.Package{}, nil
	}

	pkgs := make([]helm.Package, len(files))

	for id := range pkgs {
		p, err := util.JsonOrYamlUnmarshalFile[helm.Package](files[id])
		if err != nil {
			return helm.Package{}, err
		}
		pkgs[id] = p
	}

	if len(pkgs) == 1 {
		return pkgs[0], nil
	}

	v, err := helm.NewMergeValues(helm.MergeMaps, pkgs...)
	if err != nil {
		return helm.Package{}, err
	}

	p, err := util.JSONRemarshal[helm.Values, helm.Package](v)
	if err != nil {
		return helm.Package{}, err
	}

	if err := p.Validate(); err != nil {
		return helm.Package{}, err
	}

	return p, nil
}
