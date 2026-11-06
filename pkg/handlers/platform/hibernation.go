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

package platform

import (
	apiPlatform "github.com/arangodb/kube-arangodb/pkg/apis/platform"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
)

// HibernationSubject describes the platform object whose effective hibernation state is being evaluated.
// A service or workflow "provides" the chart named in its spec (Chart) and "depends on" the charts named
// by the custom requirement keys of that chart (Requires).
type HibernationSubject struct {
	// Kind is the resource kind of the subject (ArangoPlatformService or ArangoPlatformWorkflow), used to
	// tell the subject apart from a same-named object of the other kind.
	Kind string
	// Namespace is the namespace of the subject; only objects in the same namespace participate.
	Namespace string
	// Name is the name of the subject.
	Name string
	// Chart is the name of the chart the subject provides (spec.chart).
	Chart string
	// Requires lists the chart names the subject depends on (the custom requirement keys of its chart).
	Requires []string
	// Hibernate is the requested hibernation state (spec.hibernate).
	Hibernate bool
	// Hibernated is the subject's current Hibernated condition state.
	Hibernated bool
}

// EvaluateHibernated returns the effective Hibernated state for the subject, honoring dependency order
// derived from chart requirements:
//
//   - A subject will not become hibernated while any still-awake service or workflow depends on its
//     chart (so a required object stays up until everything that depends on it has hibernated).
//   - A subject will not wake up while any service or workflow it depends on is still hibernated (so an
//     object comes back only once the objects it requires are awake again).
//
// Once a transition is blocked the subject keeps its current state and is re-evaluated on the next
// reconcile, when the state of the objects it is waiting on may have changed.
func EvaluateHibernated(c Cache, subject HibernationSubject) bool {
	if subject.Hibernate {
		// Hibernation requested: stay hibernated once settled, otherwise only settle when nothing that
		// depends on this object is still awake.
		return subject.Hibernated || canHibernate(c, subject)
	}

	// Wake requested: stay awake once settled, otherwise only wake when nothing this object depends on is
	// still hibernated.
	return subject.Hibernated && !canWake(c, subject)
}

// canHibernate reports whether the subject may hibernate: true when no still-awake service or workflow in
// the namespace depends on the subject's chart.
func canHibernate(c Cache, subject HibernationSubject) bool {
	if subject.Chart == "" {
		return true
	}

	blocked := false
	forEachMember(c, subject.Namespace, func(m hibernationMember) {
		if m.sameAs(subject) || m.hibernated {
			return
		}

		if contains(m.requires, subject.Chart) {
			blocked = true
		}
	})

	return !blocked
}

// canWake reports whether the subject may wake: true when none of the services or workflows providing a
// chart the subject depends on is still hibernated.
func canWake(c Cache, subject HibernationSubject) bool {
	if len(subject.Requires) == 0 {
		return true
	}

	required := make(map[string]bool, len(subject.Requires))
	for _, r := range subject.Requires {
		required[r] = true
	}

	blocked := false
	forEachMember(c, subject.Namespace, func(m hibernationMember) {
		if m.sameAs(subject) || !m.hibernated {
			return
		}

		if m.chart != "" && required[m.chart] {
			blocked = true
		}
	})

	return !blocked
}

// hibernationMember is a service or workflow participating in the hibernation dependency graph.
type hibernationMember struct {
	kind       string
	namespace  string
	name       string
	chart      string
	requires   []string
	hibernated bool
}

func (m hibernationMember) sameAs(subject HibernationSubject) bool {
	return m.kind == subject.Kind && m.namespace == subject.Namespace && m.name == subject.Name
}

// forEachMember calls fn for every cached service and workflow in the namespace.
func forEachMember(c Cache, namespace string, fn func(hibernationMember)) {
	for _, s := range c.Services().List() {
		if s.GetNamespace() != namespace {
			continue
		}

		fn(hibernationMember{
			kind:       apiPlatform.ArangoPlatformServiceResourceKind,
			namespace:  s.GetNamespace(),
			name:       s.GetName(),
			chart:      s.Spec.Chart.GetName(),
			requires:   chartRequirementCharts(c, s.GetNamespace(), s.Spec.Chart.GetName()),
			hibernated: s.Status.Conditions.IsTrue(platformApi.HibernatedCondition),
		})
	}

	for _, w := range c.Workflows().List() {
		if w.GetNamespace() != namespace {
			continue
		}

		fn(hibernationMember{
			kind:       apiPlatform.ArangoPlatformWorkflowResourceKind,
			namespace:  w.GetNamespace(),
			name:       w.GetName(),
			chart:      w.Spec.Chart.GetName(),
			requires:   chartRequirementCharts(c, w.GetNamespace(), w.Spec.Chart.GetName()),
			hibernated: w.Status.Conditions.IsTrue(platformApi.HibernatedCondition),
		})
	}
}

// chartRequirementCharts returns the custom requirement keys (the names of other charts this chart
// depends on) of the named chart, resolved through the shared chart cache. Reserved requirement keys
// (ArangoDB, Operator) are not chart dependencies and are skipped.
func chartRequirementCharts(c Cache, namespace, chartName string) []string {
	if chartName == "" {
		return nil
	}

	chart, ok := c.Charts().Get(operation.ItemName{Namespace: namespace, Name: chartName})
	if !ok {
		return nil
	}

	var requirements platformApi.ChartDetailsPlatformRequirements
	if info := chart.Status.Info; info != nil {
		requirements = info.Details.GetPlatform().GetRequirements()
	}

	return CustomRequirementCharts(requirements)
}

func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}

	return false
}
