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

package metric_descriptions

import (
	"github.com/arangodb/kube-arangodb/pkg/util/metrics"
)

var (
	arangodbOperatorMembersConditions = metrics.NewDescription("arangodb_operator_members_conditions", "Representation of the ArangoMember condition state (true/false)", []string{`namespace`, `name`, `member`, `condition`}, nil)
)

func init() {
	registerDescription(arangodbOperatorMembersConditions)
}

func NewArangodbOperatorMembersConditionsGaugeFactory() metrics.FactoryGauge[ArangodbOperatorMembersConditionsInput] {
	return metrics.NewFactoryGauge[ArangodbOperatorMembersConditionsInput]()
}

func NewArangodbOperatorMembersConditionsInput(namespace string, name string, member string, condition string) ArangodbOperatorMembersConditionsInput {
	return ArangodbOperatorMembersConditionsInput{
		Namespace: namespace,
		Name:      name,
		Member:    member,
		Condition: condition,
	}
}

type ArangodbOperatorMembersConditionsInput struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Member    string `json:"member"`
	Condition string `json:"condition"`
}

func (i ArangodbOperatorMembersConditionsInput) Gauge(value float64) metrics.Metric {
	return ArangodbOperatorMembersConditionsGauge(value, i.Namespace, i.Name, i.Member, i.Condition)
}

func (i ArangodbOperatorMembersConditionsInput) Desc() metrics.Description {
	return ArangodbOperatorMembersConditions()
}

func ArangodbOperatorMembersConditions() metrics.Description {
	return arangodbOperatorMembersConditions
}

func ArangodbOperatorMembersConditionsGauge(value float64, namespace string, name string, member string, condition string) metrics.Metric {
	return ArangodbOperatorMembersConditions().Gauge(value, namespace, name, member, condition)
}
