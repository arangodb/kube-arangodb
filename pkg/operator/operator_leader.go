//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package operator

import (
	"time"

	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

func (o *Operator) runLeaderElection(lockName string, onStart func(stop <-chan struct{})) {
	namespace := o.Config.Namespace
	kubecli := o.Dependencies.KubeCli
	log := o.Dependencies.Log.With().Str("lock-name", lockName).Logger()
	rl, err := resourcelock.New(resourcelock.EndpointsResourceLock,
		namespace,
		lockName,
		kubecli.CoreV1(),
		resourcelock.ResourceLockConfig{
			Identity:      o.Config.ID,
			EventRecorder: o.Dependencies.EventRecorder,
		})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create resource lock")
	}

	leaderelection.RunOrDie(leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: 15 * time.Second,
		RenewDeadline: 10 * time.Second,
		RetryPeriod:   2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: onStart,
			OnStoppedLeading: func() {
				log.Info().Msg("Leader election lost")
			},
		},
	})
}
