//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

package metric_descriptions

import (
	"github.com/arangodb/kube-arangodb/pkg/util/metrics"
)

var (
	arangodbOperatorMembersUnexpectedContainerExitCodes = metrics.NewDescription("arangodb_operator_members_unexpected_container_exit_codes", "Counter of unexpected restarts in pod (Containers/InitContainers/EphemeralContainers)", []string{`namespace`, `name`, `member`, `container`, `container_type`, `code`, `reason`}, nil)
)

func init() {
	registerDescription(arangodbOperatorMembersUnexpectedContainerExitCodes)
}

func NewArangodbOperatorMembersUnexpectedContainerExitCodesCounterFactory() metrics.FactoryCounter[ArangodbOperatorMembersUnexpectedContainerExitCodesInput] {
	return metrics.NewFactoryCounter[ArangodbOperatorMembersUnexpectedContainerExitCodesInput]()
}

func NewArangodbOperatorMembersUnexpectedContainerExitCodesInput(namespace string, name string, member string, container string, containerType string, code string, reason string) ArangodbOperatorMembersUnexpectedContainerExitCodesInput {
	return ArangodbOperatorMembersUnexpectedContainerExitCodesInput{
		Namespace:     namespace,
		Name:          name,
		Member:        member,
		Container:     container,
		ContainerType: containerType,
		Code:          code,
		Reason:        reason,
	}
}

type ArangodbOperatorMembersUnexpectedContainerExitCodesInput struct {
	Namespace     string `json:"namespace"`
	Name          string `json:"name"`
	Member        string `json:"member"`
	Container     string `json:"container"`
	ContainerType string `json:"containerType"`
	Code          string `json:"code"`
	Reason        string `json:"reason"`
}

func (i ArangodbOperatorMembersUnexpectedContainerExitCodesInput) Counter(value float64) metrics.Metric {
	return ArangodbOperatorMembersUnexpectedContainerExitCodesCounter(value, i.Namespace, i.Name, i.Member, i.Container, i.ContainerType, i.Code, i.Reason)
}

func (i ArangodbOperatorMembersUnexpectedContainerExitCodesInput) Desc() metrics.Description {
	return ArangodbOperatorMembersUnexpectedContainerExitCodes()
}

func ArangodbOperatorMembersUnexpectedContainerExitCodes() metrics.Description {
	return arangodbOperatorMembersUnexpectedContainerExitCodes
}

func ArangodbOperatorMembersUnexpectedContainerExitCodesCounter(value float64, namespace string, name string, member string, container string, containerType string, code string, reason string) metrics.Metric {
	return ArangodbOperatorMembersUnexpectedContainerExitCodes().Counter(value, namespace, name, member, container, containerType, code, reason)
}
