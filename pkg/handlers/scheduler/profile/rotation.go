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

// rotateProfiles scans every ArangoProfile in the operator namespace and evicts the pods each one selects
// whose stamped profile checksum (profiles.arangodb.com/checksum) no longer matches the current resolved
// checksum, so a changed profile propagates to running pods. The admission policy webhook injects profiles
// and stamps the checksum on CREATE only, so without this a profile change never reaches running pods.
//
// Candidate pods are the double of a profile's own label selector AND the profiles.arangodb.com/applied
// annotation. ArangoDeployment member pods are rotated through the deployment plan and are not the target.
func (h *handler) rotateProfiles(ctx context.Context) error {
	ns := h.operator.Namespace()

	profiles, err := h.client.SchedulerV1beta1().ArangoProfiles(ns).List(ctx, meta.ListOptions{})
	if err != nil {
		return err
	}

	// A pod may match several profiles; evict it at most once per scan.
	evicted := map[string]bool{}

	for i := range profiles.Items {
		h.rotateProfile(ctx, &profiles.Items[i], evicted)
	}

	return nil
}

func (h *handler) rotateProfile(ctx context.Context, profile *schedulerApi.ArangoProfile, evicted map[string]bool) {
	if profile.Spec.Selectors == nil || profile.Spec.Selectors.Label == nil {
		return
	}

	selector, err := meta.LabelSelectorAsSelector(profile.Spec.Selectors.Label)
	if err != nil {
		logger.Err(err).Warn("Invalid selector on ArangoProfile %s/%s", profile.GetNamespace(), profile.GetName())
		return
	}

	pods, err := h.kubeClient.CoreV1().Pods(profile.GetNamespace()).List(ctx, meta.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		logger.Err(err).Warn("Unable to list pods for ArangoProfile %s/%s", profile.GetNamespace(), profile.GetName())
		return
	}

	for i := range pods.Items {
		p := &pods.Items[i]

		if evicted[p.GetName()] || p.GetDeletionTimestamp() != nil {
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

		_, current, err := scheduler.Profiles(ctx, h.client.SchedulerV1beta1().ArangoProfiles(p.GetNamespace()), labels, profilesForPod(labels, annotations)...)
		if err != nil {
			logger.Err(err).Warn("Unable to resolve ArangoProfiles for pod %s/%s", p.GetNamespace(), p.GetName())
			continue
		}

		reason, changed := profileChecksumMismatch(annotations, current)
		if !changed {
			continue
		}

		h.eventRecorder.Normal(p, "ProfileChanged", "Evicting pod: %s", reason)

		if err := h.evictPod(ctx, p); err != nil {
			logger.Err(err).Warn("Unable to evict pod %s/%s on ArangoProfile change", p.GetNamespace(), p.GetName())
			h.eventRecorder.Warning(p, "EvictionFailed", "Unable to evict pod on ArangoProfile change: %s", err.Error())
			continue
		}

		evicted[p.GetName()] = true
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
