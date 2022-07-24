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

package chaos

import (
	"context"
	"math/rand"

	"github.com/rs/zerolog"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/timer"
)

var (
	chaosMonkeyLogger = logging.Global().RegisterAndGetLogger("chaos-monkey", logging.Info)
)

// Monkey is the service that introduces chaos in the deployment
// if allowed and enabled.
type Monkey struct {
	namespace, name string
	log             logging.Logger
	context         Context
}

func (m Monkey) WrapLogger(in *zerolog.Event) *zerolog.Event {
	return in.Str("namespace", m.namespace).Str("name", m.name)
}

// NewMonkey creates a new chaos monkey with given context.
func NewMonkey(namespace, name string, context Context) *Monkey {
	m := &Monkey{
		context:   context,
		namespace: namespace,
		name:      name,
	}
	m.log = chaosMonkeyLogger.WrapObj(m)
	return m
}

// Run the monkey until the given channel is closed.
func (m Monkey) Run(stopCh <-chan struct{}) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		spec := m.context.GetSpec()
		if spec.Chaos.IsEnabled() {
			// Gamble to set if we must introduce chaos
			chance := float64(spec.Chaos.GetKillPodProbability()) / 100.0
			if rand.Float64() < chance {
				// Let's introduce pod chaos
				if err := m.killRandomPod(ctx); err != nil {
					m.log.Err(err).Info("Failed to kill random pod")
				}
			}
		}

		select {
		case <-timer.After(spec.Chaos.GetInterval()):
			// Continue
		case <-stopCh:
			// We're done
			return
		}
	}
}

// killRandomPod fetches all owned pods and tries to kill one.
func (m Monkey) killRandomPod(ctx context.Context) error {
	pods, err := m.context.GetOwnedPods(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	if len(pods) <= 1 {
		// Not enough pods
		return nil
	}
	p := pods[rand.Intn(len(pods))]
	m.log.Str("pod-name", p.GetName()).Info("Killing pod")
	if err := m.context.DeletePod(ctx, p.GetName(), meta.DeleteOptions{}); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
