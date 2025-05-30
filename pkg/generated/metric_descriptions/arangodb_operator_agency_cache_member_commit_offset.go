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
	arangodbOperatorAgencyCacheMemberCommitOffset = metrics.NewDescription("arangodb_operator_agency_cache_member_commit_offset", "Determines agency member commit offset", []string{`namespace`, `name`, `agent`}, nil)
)

func init() {
	registerDescription(arangodbOperatorAgencyCacheMemberCommitOffset)
}

func NewArangodbOperatorAgencyCacheMemberCommitOffsetGaugeFactory() metrics.FactoryGauge[ArangodbOperatorAgencyCacheMemberCommitOffsetInput] {
	return metrics.NewFactoryGauge[ArangodbOperatorAgencyCacheMemberCommitOffsetInput]()
}

func NewArangodbOperatorAgencyCacheMemberCommitOffsetInput(namespace string, name string, agent string) ArangodbOperatorAgencyCacheMemberCommitOffsetInput {
	return ArangodbOperatorAgencyCacheMemberCommitOffsetInput{
		Namespace: namespace,
		Name:      name,
		Agent:     agent,
	}
}

type ArangodbOperatorAgencyCacheMemberCommitOffsetInput struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Agent     string `json:"agent"`
}

func (i ArangodbOperatorAgencyCacheMemberCommitOffsetInput) Gauge(value float64) metrics.Metric {
	return ArangodbOperatorAgencyCacheMemberCommitOffsetGauge(value, i.Namespace, i.Name, i.Agent)
}

func (i ArangodbOperatorAgencyCacheMemberCommitOffsetInput) Desc() metrics.Description {
	return ArangodbOperatorAgencyCacheMemberCommitOffset()
}

func ArangodbOperatorAgencyCacheMemberCommitOffset() metrics.Description {
	return arangodbOperatorAgencyCacheMemberCommitOffset
}

func ArangodbOperatorAgencyCacheMemberCommitOffsetGauge(value float64, namespace string, name string, agent string) metrics.Metric {
	return ArangodbOperatorAgencyCacheMemberCommitOffset().Gauge(value, namespace, name, agent)
}
