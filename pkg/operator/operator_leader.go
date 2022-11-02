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

package operator

import (
	"context"
	"fmt"
	"os"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/probe"
	"github.com/arangodb/kube-arangodb/pkg/util/retry"
)

// runLeaderElection performs a leader election on a lock with given name in
// the namespace that the operator is deployed in.
// When the leader election is won, the given callback is called.
// When the leader election is was won once, but then the leadership is lost, the process is killed.
// The given ready probe is set, as soon as this process became the leader, or a new leader
// is detected.
func (o *Operator) runLeaderElection(lockName, label string, onStart func(stop <-chan struct{}), readyProbe *probe.ReadyProbe) {
	namespace := o.Config.Namespace
	kubecli := o.Dependencies.Client.Kubernetes()
	log := o.log.Str("lock-name", lockName)
	eventTarget := o.getLeaderElectionEventTarget(log)
	recordEvent := func(reason, message string) {
		if eventTarget != nil {
			o.Dependencies.EventRecorder.Event(eventTarget, core.EventTypeNormal, reason, message)
		}
	}
	rl, err := resourcelock.New(resourcelock.EndpointsResourceLock,
		namespace,
		lockName,
		kubecli.CoreV1(),
		kubecli.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity:      o.Config.ID,
			EventRecorder: o.Dependencies.EventRecorder,
		})
	if err != nil {
		log.Err(err).Fatal("Failed to create resource lock")
	}

	ctx := context.Background()
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: 15 * time.Second,
		RenewDeadline: 10 * time.Second,
		RetryPeriod:   2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				recordEvent("Leader Election Won", fmt.Sprintf("Pod %s is running as leader", o.Config.PodName))
				readyProbe.SetReady()
				if err := o.setRoleLabel(log, label, constants.LabelRoleLeader); err != nil {
					log.Error("Cannot set leader role on Pod. Terminating process")
					os.Exit(2)
				}
				onStart(ctx.Done())
			},
			OnStoppedLeading: func() {
				recordEvent("Stop Leading", fmt.Sprintf("Pod %s is stopping to run as leader", o.Config.PodName))
				log.Info("Stop leading. Terminating process")
				os.Exit(1)
			},
			OnNewLeader: func(identity string) {
				log.Str("identity", identity).Info("New leader detected")
				readyProbe.SetReady()
			},
		},
	})
}

func (o *Operator) runWithoutLeaderElection(lockName, label string, onStart func(stop <-chan struct{}), readyProbe *probe.ReadyProbe) {
	log := o.log.Str("lock-name", lockName)
	eventTarget := o.getLeaderElectionEventTarget(log)
	recordEvent := func(reason, message string) {
		if eventTarget != nil {
			o.Dependencies.EventRecorder.Event(eventTarget, core.EventTypeNormal, reason, message)
		}
	}
	ctx := context.Background()

	recordEvent("Leader Election Skipped", fmt.Sprintf("Pod %s is running as leader", o.Config.PodName))
	readyProbe.SetReady()
	if err := o.setRoleLabel(log, label, constants.LabelRoleLeader); err != nil {
		log.Error("Cannot set leader role on Pod. Terminating process")
		os.Exit(2)
	}
	onStart(ctx.Done())
}

// getLeaderElectionEventTarget returns the object that leader election related
// events will be added to.
func (o *Operator) getLeaderElectionEventTarget(log logging.Logger) runtime.Object {
	ns := o.Config.Namespace
	kubecli := o.Dependencies.Client.Kubernetes()
	pods := kubecli.CoreV1().Pods(ns)
	log = log.Str("pod-name", o.Config.PodName)
	pod, err := pods.Get(context.Background(), o.Config.PodName, meta.GetOptions{})
	if err != nil {
		log.Err(err).Error("Cannot find Pod containing this operator")
		return nil
	}
	rSet, err := k8sutil.GetPodOwner(kubecli, pod, ns)
	if err != nil {
		log.Err(err).Error("Cannot find ReplicaSet owning the Pod containing this operator")
		return pod
	}
	if rSet == nil {
		log.Error("Pod containing this operator has no ReplicaSet owner")
		return pod
	}
	log = log.Str("replicaSet-name", rSet.Name)
	depl, err := k8sutil.GetReplicaSetOwner(kubecli, rSet, ns)
	if err != nil {
		log.Err(err).Error("Cannot find Deployment owning the ReplicataSet that owns the Pod containing this operator")
		return rSet
	}
	if rSet == nil {
		log.Error("ReplicaSet that owns the Pod containing this operator has no Deployment owner")
		return rSet
	}
	return depl
}

// setRoleLabel sets a label with key `role` and given value in the pod metadata.
func (o *Operator) setRoleLabel(log logging.Logger, label, role string) error {
	ns := o.Config.Namespace
	kubecli := o.Dependencies.Client.Kubernetes()
	pods := kubecli.CoreV1().Pods(ns)
	log = log.Str("pod-name", o.Config.PodName)
	op := func() error {
		pod, err := pods.Get(context.Background(), o.Config.PodName, meta.GetOptions{})
		if kerrors.IsNotFound(err) {
			log.Err(err).Error("Pod not found, so we cannot set its role label")
			return retry.Permanent(errors.WithStack(err))
		} else if err != nil {
			return errors.WithStack(err)
		}
		labels := pod.ObjectMeta.GetLabels()
		if labels == nil {
			labels = make(map[string]string)
		}
		labels[label] = role
		pod.ObjectMeta.SetLabels(labels)
		if _, err := pods.Update(context.Background(), pod, meta.UpdateOptions{}); kerrors.IsConflict(err) {
			// Retry it
			return errors.WithStack(err)
		} else if err != nil {
			log.Err(err).Error("Failed to update Pod wrt 'role' label")
			return retry.Permanent(errors.WithStack(err))
		}
		return nil
	}
	if err := retry.Retry(op, time.Second*15); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
