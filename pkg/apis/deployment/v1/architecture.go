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

package v1

import (
	"runtime"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

type ArangoDeploymentArchitecture []ArangoDeploymentArchitectureType

func (a ArangoDeploymentArchitecture) GetDefault() ArangoDeploymentArchitectureType {
	if len(a) == 0 {
		return ArangoDeploymentArchitectureDefault
	}

	return a[0]
}

func (a ArangoDeploymentArchitecture) Validate() error {
	for id := range a {
		if err := a[id].Validate(); err != nil {
			return errors.WithStack(errors.Wrapf(err, "%d", id))
		}
	}

	return nil
}

func (a ArangoDeploymentArchitecture) IsArchAllowed(arch ArangoDeploymentArchitectureType) bool {
	for id := range a {
		if a[id] == arch {
			return true
		}
	}

	return false
}

func (a ArangoDeploymentArchitecture) AsNodeSelectorRequirement() core.NodeSelectorTerm {
	var archs []string

	if len(a) == 0 {
		archs = append(archs, ArangoDeploymentArchitectureDefault.String())
	} else {
		for _, arch := range a {
			archs = append(archs, arch.String())
		}
	}

	return core.NodeSelectorTerm{
		MatchExpressions: []core.NodeSelectorRequirement{
			{
				Key:      shared.NodeArchAffinityLabel,
				Operator: "In",
				Values:   archs,
			},
		},
	}
}

type ArangoDeploymentArchitectureType string

const (
	// ArangoDeploymentArchitectureAMD64 define const for architecture for amd64
	ArangoDeploymentArchitectureAMD64 ArangoDeploymentArchitectureType = "amd64"
	// ArangoDeploymentArchitectureARM64 define const for architecture for arm64
	ArangoDeploymentArchitectureARM64 ArangoDeploymentArchitectureType = "arm64"

	// ArangoDeploymentArchitectureDefault define default architecture used by Operator
	ArangoDeploymentArchitectureDefault = ArangoDeploymentArchitectureAMD64

	// ArangoDeploymentArchitectureCurrent define current Operator architecture
	ArangoDeploymentArchitectureCurrent = ArangoDeploymentArchitectureType(runtime.GOARCH)
)

func (a ArangoDeploymentArchitectureType) Validate() error {
	switch q := a; q {
	case ArangoDeploymentArchitectureAMD64, ArangoDeploymentArchitectureARM64:
		return nil
	default:
		return errors.Errorf("Unknown architecture type %s", q)
	}
}

func (a ArangoDeploymentArchitectureType) String() string {
	return string(a)
}

func (a *ArangoDeploymentArchitectureType) Default(def ArangoDeploymentArchitectureType) ArangoDeploymentArchitectureType {
	if a == nil {
		return def
	}

	return *a
}

func (a ArangoDeploymentArchitectureType) AsNodeSelectorRequirement() core.NodeSelectorTerm {
	return core.NodeSelectorTerm{
		MatchExpressions: []core.NodeSelectorRequirement{
			{
				Key:      shared.NodeArchAffinityLabel,
				Operator: "In",
				Values:   []string{a.String()},
			},
		},
	}
}

func (a ArangoDeploymentArchitectureType) IsArchMismatch(deploymentArch ArangoDeploymentArchitecture, memberArch *ArangoDeploymentArchitectureType) bool {
	if a.Validate() != nil {
		return false
	}

	if deploymentArch.IsArchAllowed(a) {
		if memberArch == nil {
			return true
		} else if a != *memberArch {
			return true
		}
	}
	return false
}

func GetAllArchFromNodeSelector(selectors []core.NodeSelectorTerm) map[ArangoDeploymentArchitectureType]bool {
	result := make(map[ArangoDeploymentArchitectureType]bool)
	for _, selector := range selectors {
		if selector.MatchExpressions != nil {
			for _, req := range selector.MatchExpressions {
				if req.Key == shared.NodeArchAffinityLabel || req.Key == shared.NodeArchAffinityLabelBeta {
					for _, arch := range req.Values {
						result[ArangoDeploymentArchitectureType(arch)] = true
					}
				}
			}
		}
	}
	return result
}

func (a *ArangoDeploymentArchitectureType) Equal(other *ArangoDeploymentArchitectureType) bool {
	if a == nil && other == nil {
		return true
	} else if a == nil || other == nil {
		return false
	} else if a == other {
		return true
	}
	return false
}
