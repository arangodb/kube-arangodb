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

package resources

import (
	"time"

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func (r *Resources) failedContainerHandler(log logging.Logger, memberStatus api.MemberStatus, pod *core.Pod) time.Time {
	last := memberStatus.LastTermination()

	current := last

	for _, c := range pod.Status.InitContainerStatuses {
		if t := c.State.Terminated; t != nil {
			if q := t.FinishedAt.Time; q.After(last) {
				if q.After(current) {
					current = q
				}
			} else {
				continue
			}

			if t.ExitCode == 0 {
				continue
			}

			log.Str("member", memberStatus.ID).
				Str("pod", pod.GetName()).
				Str("container", c.Name).
				Str("uid", string(pod.GetUID())).
				Int32("exit-code", t.ExitCode).
				Str("reason", t.Reason).
				Str("message", t.Message).
				Int32("signal", t.Signal).
				Time("started", t.StartedAt.Time).
				Time("finished", t.FinishedAt.Time).
				Warn("Pod failed in unexpected way: Init Container failed")

			r.metrics.IncMemberInitContainerRestarts(memberStatus.ID, c.Name, t.Reason, t.ExitCode)
		} else if t := c.LastTerminationState.Terminated; t != nil {
			if q := t.FinishedAt.Time; q.After(last) {
				if q.After(current) {
					current = q
				}
			} else {
				continue
			}

			if t.ExitCode == 0 {
				continue
			}

			log.Str("member", memberStatus.ID).
				Str("pod", pod.GetName()).
				Str("container", c.Name).
				Str("uid", string(pod.GetUID())).
				Int32("exit-code", t.ExitCode).
				Str("reason", t.Reason).
				Str("message", t.Message).
				Int32("signal", t.Signal).
				Time("started", t.StartedAt.Time).
				Time("finished", t.FinishedAt.Time).
				Warn("Pod failed in unexpected way: Init Container failed")

			r.metrics.IncMemberInitContainerRestarts(memberStatus.ID, c.Name, t.Reason, t.ExitCode)
		}
	}

	for _, c := range pod.Status.ContainerStatuses {
		if t := c.State.Terminated; t != nil {
			if q := t.FinishedAt.Time; q.After(last) {
				if q.After(current) {
					current = q
				}
			} else {
				continue
			}

			if t.ExitCode == 0 {
				continue
			}

			log.Str("member", memberStatus.ID).
				Str("pod", pod.GetName()).
				Str("container", c.Name).
				Str("uid", string(pod.GetUID())).
				Int32("exit-code", t.ExitCode).
				Str("reason", t.Reason).
				Str("message", t.Message).
				Int32("signal", t.Signal).
				Time("started", t.StartedAt.Time).
				Time("finished", t.FinishedAt.Time).
				Warn("Pod failed in unexpected way: Container failed")

			r.metrics.IncMemberContainerRestarts(memberStatus.ID, c.Name, t.Reason, t.ExitCode)
		} else if t := c.LastTerminationState.Terminated; t != nil {
			if q := t.FinishedAt.Time; q.After(last) {
				if q.After(current) {
					current = q
				}
			} else {
				continue
			}

			if t.ExitCode == 0 {
				continue
			}

			log.Str("member", memberStatus.ID).
				Str("pod", pod.GetName()).
				Str("container", c.Name).
				Str("uid", string(pod.GetUID())).
				Int32("exit-code", t.ExitCode).
				Str("reason", t.Reason).
				Str("message", t.Message).
				Int32("signal", t.Signal).
				Time("started", t.StartedAt.Time).
				Time("finished", t.FinishedAt.Time).
				Warn("Pod failed in unexpected way: Container failed")

			r.metrics.IncMemberContainerRestarts(memberStatus.ID, c.Name, t.Reason, t.ExitCode)
		}
	}

	for _, c := range pod.Status.ContainerStatuses {
		if t := c.State.Terminated; t != nil {
			if q := t.FinishedAt.Time; q.After(last) {
				if q.After(current) {
					current = q
				}
			} else {
				continue
			}

			if t.ExitCode == 0 {
				continue
			}

			log.Str("member", memberStatus.ID).
				Str("pod", pod.GetName()).
				Str("container", c.Name).
				Str("uid", string(pod.GetUID())).
				Int32("exit-code", t.ExitCode).
				Str("reason", t.Reason).
				Str("message", t.Message).
				Int32("signal", t.Signal).
				Time("started", t.StartedAt.Time).
				Time("finished", t.FinishedAt.Time).
				Warn("Pod failed in unexpected way: Ephemeral Container failed")

			r.metrics.IncMemberEphemeralContainerRestarts(memberStatus.ID, c.Name, t.Reason, t.ExitCode)
		} else if t := c.LastTerminationState.Terminated; t != nil {
			if q := t.FinishedAt.Time; q.After(last) {
				if q.After(current) {
					current = q
				}
			} else {
				continue
			}

			if t.ExitCode == 0 {
				continue
			}

			log.Str("member", memberStatus.ID).
				Str("pod", pod.GetName()).
				Str("container", c.Name).
				Str("uid", string(pod.GetUID())).
				Int32("exit-code", t.ExitCode).
				Str("reason", t.Reason).
				Str("message", t.Message).
				Int32("signal", t.Signal).
				Time("started", t.StartedAt.Time).
				Time("finished", t.FinishedAt.Time).
				Warn("Pod failed in unexpected way: Ephemeral Container failed")

			r.metrics.IncMemberEphemeralContainerRestarts(memberStatus.ID, c.Name, t.Reason, t.ExitCode)
		}
	}

	return last
}

func getDefaultRestartPolicy(spec api.ServerGroupSpec) core.RestartPolicy {
	def := util.BoolSwitch(features.RestartPolicyAlways().Enabled(), core.RestartPolicyAlways, core.RestartPolicyNever)

	if r := spec.RestartPolicy; r != nil {
		switch *r {
		case core.RestartPolicyNever, core.RestartPolicyAlways:
			return *r
		default:
			return core.RestartPolicyNever
		}
	}

	return def
}
