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

package panics

import (
	"sync"

	"github.com/arangodb/kube-arangodb/pkg/generated/metric_descriptions"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/metrics/collector"
	"github.com/arangodb/kube-arangodb/pkg/util/metrics"
)

func init() {
	collector.GetCollector().RegisterMetric(panicsReceived)
}

var (
	panicsReceived = &panicsReceiver{
		panics: map[string]uint{},
	}
)

type panicsReceiver struct {
	panics map[string]uint
	lock   sync.Mutex
}

func (p *panicsReceiver) CollectMetrics(in metrics.PushMetric) {
	for k, v := range p.panics {
		in.Push(metric_descriptions.ArangodbOperatorEnginePanicsRecoveredCounter(float64(v), k))
	}
}

func (p *panicsReceiver) register(section string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.panics[section] = p.panics[section] + 1
}

func recoverWithSection(log logging.Logger, section string, in func() error) error {
	err := recoverPanic(5, in)

	if err != nil {
		if perr, ok := IsPanicError(err); ok {
			panicsReceived.register(section)

			log.Strs("stack", perr.Stack().String()...).Str("section", section).Error("Panic Recovered")
		}
	}

	return err
}

func RecoverWithSection(section string, in func() error) error {
	return recoverWithSection(logger, section, in)
}

func RecoverWithSectionL(log logging.Logger, section string, in func() error) error {
	return recoverWithSection(log, section, in)
}
