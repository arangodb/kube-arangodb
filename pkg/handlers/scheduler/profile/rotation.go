//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package profile

import (
	"context"
	"fmt"
	goStrings "strings"
	"time"

	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/scheduler"
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
)

// profileRotationInterval is how often the operator scans ArangoProfiles for pods whose applied profile
// checksum has drifted from the current resolved checksum. Mirrors the ArangoBackup/workflow discovery
// loops.
const profileRotationInterval = time.Minute

// Start launches the profile-rotation scan loop as a background goroutine. Registered via
// operator.RegisterStarter.
func (h *handler) Start(stopCh <-chan struct{}) {
	go h.runProfileRotation(stopCh)
}

func (h *handler) runProfileRotation(stopCh <-chan struct{}) {
	t := time.NewTicker(profileRotationInterval)
	defer t.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-t.C:
			if !features.ProfileRotation().Enabled() {
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), profileRotationInterval)
			if err := h.rotateProfiles(ctx); err != nil {
				logger.Err(err).Warn("ArangoProfile rotation scan failed")
			}
			cancel()
		}
	}
}

// rotateProfiles scans each ArangoDeployment in the operator namespace and evicts its profile-injected
// pods whose stamped profile checksum (profiles.arangodb.com/checksum) no longer matches the current
// resolved checksum, so a changed profile propagates to running pods. The admission policy webhook
// injects profiles and stamps the checksum on CREATE only, so without this a profile change never reaches
// running pods.
//
// The profile list is fetched once per scan; then, per deployment, the pods it owns are listed and each
// one is resolved against that list the same way the webhook does. ArangoDeployment member pods are
// rotated through the deployment plan and are not the target.
func (h *handler) rotateProfiles(ctx context.Context) error {
	ns := h.operator.Namespace()

	profileList, err := h.listProfiles(ctx, ns)
	if err != nil {
		return err
	}
	if len(profileList) == 0 {
		return nil
	}

	deployments, err := h.client.DatabaseV1().ArangoDeployments(ns).List(ctx, meta.ListOptions{})
	if err != nil {
		return err
	}

	for i := range deployments.Items {
		h.rotateDeployment(ctx, ns, deployments.Items[i].GetName(), profileList)
	}

	return nil
}

// listProfiles returns every ArangoProfile in the namespace as a slice, so a single scan resolves many
// pods against one list instead of re-listing per pod.
func (h *handler) listProfiles(ctx context.Context, ns string) ([]*schedulerApi.ArangoProfile, error) {
	l, err := h.client.SchedulerV1beta1().ArangoProfiles(ns).List(ctx, meta.ListOptions{})
	if err != nil {
		return nil, err
	}

	out := make([]*schedulerApi.ArangoProfile, len(l.Items))
	for i := range l.Items {
		out[i] = &l.Items[i]
	}

	return out, nil
}

// rotateDeployment evicts the given deployment's profile-injected pods whose resolved profile checksum
// has drifted from the stamped value. Candidate pods are the deployment's pods (profiles.arangodb.com/
// deployment label) that also carry profiles.arangodb.com/applied=true; for each, the applicable profiles
// are filtered from profileList and the checksum rendered exactly as the webhook does.
func (h *handler) rotateDeployment(ctx context.Context, ns, deployment string, profileList []*schedulerApi.ArangoProfile) {
	pods, err := h.kubeClient.CoreV1().Pods(ns).List(ctx, meta.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", utilConstants.ProfilesDeployment, deployment),
	})
	if err != nil {
		logger.Err(err).Warn("Unable to list pods for deployment %s/%s", ns, deployment)
		return
	}

	for i := range pods.Items {
		p := &pods.Items[i]

		if p.GetDeletionTimestamp() != nil {
			continue
		}

		annotations := p.GetAnnotations()
		if annotations[utilConstants.ProfilesAnnotationApplied] != "true" {
			continue
		}

		if _, ok := annotations[utilConstants.ProfilesAnnotationChecksum]; !ok {
			continue
		}

		labels := p.GetLabels()

		_, current, err := scheduler.ProfilesFromList(profileList, labels, profilesForPod(labels, annotations)...)
		if err != nil {
			logger.Err(err).Warn("Unable to resolve ArangoProfiles for pod %s/%s", ns, p.GetName())
			continue
		}

		reason, changed := profileChecksumMismatch(annotations, current)
		if !changed {
			continue
		}

		h.eventRecorder.Normal(p, "ProfileChanged", "Evicting pod: %s", reason)

		if err := h.evictPod(ctx, p); err != nil {
			logger.Err(err).Warn("Unable to evict pod %s/%s on ArangoProfile change", ns, p.GetName())
			h.eventRecorder.Warning(p, "EvictionFailed", "Unable to evict pod on ArangoProfile change: %s", err.Error())
		}
	}
}

// evictPod evicts the pod via the Eviction API, so its PodDisruptionBudget is respected and its owning
// controller recreates it (re-admitting it through the profile webhook with the current profile).
func (h *handler) evictPod(ctx context.Context, p *core.Pod) error {
	return h.kubeClient.CoreV1().Pods(p.GetNamespace()).EvictV1(ctx, &policy.Eviction{
		ObjectMeta: meta.ObjectMeta{
			Name:      p.GetName(),
			Namespace: p.GetNamespace(),
		},
	})
}

// profilesForPod returns the explicit profile names to resolve for a pod, mirroring the admission policy
// webhook (pkg/handlers/scheduler/webhooks/policies): they come from the profiles.arangodb.com/profiles
// label and annotation, trimmed, with empty entries dropped.
func profilesForPod(labels, annotations map[string]string) []string {
	allProfiles := util.FlattenLists(goStrings.Split(labels[utilConstants.ProfilesList], ","), goStrings.Split(annotations[utilConstants.ProfilesList], ","))
	return util.FilterList(util.FormatList(allProfiles, func(s string) string {
		return goStrings.TrimSpace(s)
	}), func(s string) bool {
		return s != ""
	})
}

// profileChecksumMismatch reports whether a pod was profile-injected and its stamped profiles checksum
// (profiles.arangodb.com/checksum) differs from current, the freshly resolved checksum, returning a
// human-readable reason when it does. A pod without profiles.arangodb.com/applied=true or without a
// stamped checksum is never a mismatch, so pods that were never profile-injected are never evicted.
func profileChecksumMismatch(annotations map[string]string, current string) (string, bool) {
	if annotations[utilConstants.ProfilesAnnotationApplied] != "true" {
		return "", false
	}

	stored, ok := annotations[utilConstants.ProfilesAnnotationChecksum]
	if !ok {
		return "", false
	}

	if current == stored {
		return "", false
	}

	return fmt.Sprintf("ArangoProfile checksum changed (%s -> %s)", stored, current), true
}
