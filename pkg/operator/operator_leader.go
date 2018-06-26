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
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/probe"
)

// runLeaderElection performs a leader election on a lock with given name in
// the namespace that the operator is deployed in.
// When the leader election is won, the given callback is called.
// When the leader election is was won once, but then the leadership is lost, the process is killed.
// The given ready probe is set, as soon as this process became the leader, or a new leader
// is detected.
func (o *Operator) runLeaderElection(lockName string, onStart func(stop <-chan struct{}), readyProbe *probe.ReadyProbe) {
	namespace := o.Config.Namespace
	kubecli := o.Dependencies.KubeCli
	log := o.log.With().Str("lock-name", lockName).Logger()
	eventTarget := o.getLeaderElectionEventTarget(log)
	recordEvent := func(reason, message string) {
		if eventTarget != nil {
			o.Dependencies.EventRecorder.Event(eventTarget, v1.EventTypeNormal, reason, message)
		}
	}
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
			OnStartedLeading: func(stop <-chan struct{}) {
				recordEvent("Leader Election Won", fmt.Sprintf("Pod %s is running as leader", o.Config.PodName))
				readyProbe.SetReady()
				onStart(stop)
			},
			OnStoppedLeading: func() {
				recordEvent("Stop Leading", fmt.Sprintf("Pod %s is stopping to run as leader", o.Config.PodName))
				log.Info().Msg("Stop leading. Terminating process")
				os.Exit(1)
			},
			OnNewLeader: func(identity string) {
				log.Info().Str("identity", identity).Msg("New leader detected")
				readyProbe.SetReady()
			},
		},
	})
}

// getLeaderElectionEventTarget returns the object that leader election related
// events will be added to.
func (o *Operator) getLeaderElectionEventTarget(log zerolog.Logger) runtime.Object {
	ns := o.Config.Namespace
	kubecli := o.Dependencies.KubeCli
	pods := kubecli.CoreV1().Pods(ns)
	log = log.With().Str("pod-name", o.Config.PodName).Logger()
	pod, err := pods.Get(o.Config.PodName, metav1.GetOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Cannot find Pod containing this operator")
		return nil
	}
	rSet, err := k8sutil.GetPodOwner(kubecli, pod, ns)
	if err != nil {
		log.Error().Err(err).Msg("Cannot find ReplicaSet owning the Pod containing this operator")
		return pod
	}
	if rSet == nil {
		log.Error().Msg("Pod containing this operator has no ReplicaSet owner")
		return pod
	}
	log = log.With().Str("replicaSet-name", rSet.Name).Logger()
	depl, err := k8sutil.GetReplicaSetOwner(kubecli, rSet, ns)
	if err != nil {
		log.Error().Err(err).Msg("Cannot find Deployment owning the ReplicataSet that owns the Pod containing this operator")
		return rSet
	}
	if rSet == nil {
		log.Error().Msg("ReplicaSet that owns the Pod containing this operator has no Deployment owner")
		return rSet
	}
	return depl
}
