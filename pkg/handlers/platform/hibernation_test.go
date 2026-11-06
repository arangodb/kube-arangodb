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
	"testing"

	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiPlatform "github.com/arangodb/kube-arangodb/pkg/apis/platform"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// chartDep returns a chart in namespace whose platform requirements declare the given chart names as
// dependencies.
func chartDep(namespace, name string, requires ...string) *platformApi.ArangoPlatformChart {
	reqs := platformApi.ChartDetailsPlatformRequirements{}
	for _, r := range requires {
		reqs[r] = util.VersionConstrain(">=0.0.0")
	}

	return &platformApi.ArangoPlatformChart{
		ObjectMeta: meta.ObjectMeta{Namespace: namespace, Name: name},
		Status: platformApi.ArangoPlatformChartStatus{
			Info: &platformApi.ChartStatusInfo{
				Details: &platformApi.ChartDetails{
					Name:     name,
					Platform: &platformApi.ChartDetailsPlatform{Requirements: reqs},
				},
			},
		},
	}
}

func svc(namespace, name, chart string, hibernated bool) *platformApi.ArangoPlatformService {
	s := &platformApi.ArangoPlatformService{
		ObjectMeta: meta.ObjectMeta{Namespace: namespace, Name: name},
	}
	s.Spec.Chart = &sharedApi.Object{Name: chart}
	s.Status.Conditions.Update(platformApi.HibernatedCondition, hibernated, "", "")
	return s
}

func wf(namespace, name, chart string, hibernated bool) *platformApi.ArangoPlatformWorkflow {
	w := &platformApi.ArangoPlatformWorkflow{
		ObjectMeta: meta.ObjectMeta{Namespace: namespace, Name: name},
	}
	w.Spec.Chart = &sharedApi.Object{Name: chart}
	w.Status.Conditions.Update(platformApi.HibernatedCondition, hibernated, "", "")
	return w
}

type fixture struct {
	charts    []*platformApi.ArangoPlatformChart
	services  []*platformApi.ArangoPlatformService
	workflows []*platformApi.ArangoPlatformWorkflow
}

func (f fixture) cache() Cache {
	c := NewCache(0)
	for _, ch := range f.charts {
		c.Charts().Set(operation.ItemName{Namespace: ch.GetNamespace(), Name: ch.GetName()}, ch)
	}
	for _, s := range f.services {
		c.Services().Set(operation.ItemName{Namespace: s.GetNamespace(), Name: s.GetName()}, s)
	}
	for _, w := range f.workflows {
		c.Workflows().Set(operation.ItemName{Namespace: w.GetNamespace(), Name: w.GetName()}, w)
	}
	return c
}

func serviceSubject(name, chart string, requires []string, hibernate, hibernated bool) HibernationSubject {
	return HibernationSubject{
		Kind:       apiPlatform.ArangoPlatformServiceResourceKind,
		Namespace:  ns,
		Name:       name,
		Chart:      chart,
		Requires:   requires,
		Hibernate:  hibernate,
		Hibernated: hibernated,
	}
}

const ns = "default"

// Graph used across the cases: service "a" (chart "app") depends on service "b" (chart "base"); the "app"
// chart declares a requirement on the "base" chart.
func baseFixture(aHibernated, bHibernated bool) fixture {
	return fixture{
		charts: []*platformApi.ArangoPlatformChart{
			chartDep(ns, "app", "base"),
			chartDep(ns, "base"),
		},
		services: []*platformApi.ArangoPlatformService{
			svc(ns, "a", "app", aHibernated),
			svc(ns, "b", "base", bHibernated),
		},
	}
}

func Test_EvaluateHibernated(t *testing.T) {
	t.Run("Hibernate dependency blocked while dependent is awake", func(t *testing.T) {
		c := baseFixture(false, false).cache()

		// b (base) is required by a (app), which is still awake, so b must not hibernate yet.
		subject := serviceSubject("b", "base", nil, true, false)
		require.False(t, EvaluateHibernated(c, subject))
	})

	t.Run("Hibernate dependency allowed once dependent is hibernated", func(t *testing.T) {
		c := baseFixture(true, false).cache()

		subject := serviceSubject("b", "base", nil, true, false)
		require.True(t, EvaluateHibernated(c, subject))
	})

	t.Run("Hibernate dependent (leaf) allowed immediately", func(t *testing.T) {
		c := baseFixture(false, false).cache()

		// Nothing depends on a (app), so it may hibernate right away.
		subject := serviceSubject("a", "app", []string{"base"}, true, false)
		require.True(t, EvaluateHibernated(c, subject))
	})

	t.Run("Already hibernated stays hibernated", func(t *testing.T) {
		c := baseFixture(false, true).cache()

		// Even though a still depends on b, an already-hibernated b stays hibernated.
		subject := serviceSubject("b", "base", nil, true, true)
		require.True(t, EvaluateHibernated(c, subject))
	})

	t.Run("Wake blocked while dependency is still hibernated", func(t *testing.T) {
		c := baseFixture(true, true).cache()

		// a wants to wake but b (which a requires) is still hibernated, so a stays hibernated.
		subject := serviceSubject("a", "app", []string{"base"}, false, true)
		require.True(t, EvaluateHibernated(c, subject))
	})

	t.Run("Wake allowed once dependency is awake", func(t *testing.T) {
		c := baseFixture(true, false).cache()

		subject := serviceSubject("a", "app", []string{"base"}, false, true)
		require.False(t, EvaluateHibernated(c, subject))
	})

	t.Run("Wake dependency (root) allowed immediately", func(t *testing.T) {
		c := baseFixture(true, true).cache()

		// b depends on nothing, so it may wake right away.
		subject := serviceSubject("b", "base", nil, false, true)
		require.False(t, EvaluateHibernated(c, subject))
	})

	t.Run("Already awake stays awake", func(t *testing.T) {
		c := baseFixture(false, false).cache()

		subject := serviceSubject("a", "app", []string{"base"}, false, false)
		require.False(t, EvaluateHibernated(c, subject))
	})

	t.Run("Wake with missing dependency provider is allowed", func(t *testing.T) {
		c := fixture{
			charts:   []*platformApi.ArangoPlatformChart{chartDep(ns, "app", "missing")},
			services: []*platformApi.ArangoPlatformService{svc(ns, "a", "app", true)},
		}.cache()

		// No service or workflow provides the "missing" chart, so nothing blocks waking.
		subject := serviceSubject("a", "app", []string{"missing"}, false, true)
		require.False(t, EvaluateHibernated(c, subject))
	})

	t.Run("Dependent in another namespace does not block hibernation", func(t *testing.T) {
		c := fixture{
			charts: []*platformApi.ArangoPlatformChart{
				chartDep("other", "app", "base"),
				chartDep(ns, "base"),
			},
			services: []*platformApi.ArangoPlatformService{
				svc("other", "a", "app", false),
				svc(ns, "b", "base", false),
			},
		}.cache()

		subject := serviceSubject("b", "base", nil, true, false)
		require.True(t, EvaluateHibernated(c, subject))
	})

	t.Run("Workflow dependent blocks service hibernation", func(t *testing.T) {
		c := fixture{
			charts: []*platformApi.ArangoPlatformChart{
				chartDep(ns, "flow", "base"),
				chartDep(ns, "base"),
			},
			services:  []*platformApi.ArangoPlatformService{svc(ns, "b", "base", false)},
			workflows: []*platformApi.ArangoPlatformWorkflow{wf(ns, "w", "flow", false)},
		}.cache()

		// The workflow w (chart flow) depends on the base chart provided by service b.
		subject := serviceSubject("b", "base", nil, true, false)
		require.False(t, EvaluateHibernated(c, subject))
	})

	t.Run("Hibernated workflow dependency blocks service wake", func(t *testing.T) {
		c := fixture{
			charts: []*platformApi.ArangoPlatformChart{
				chartDep(ns, "app", "flow"),
				chartDep(ns, "flow"),
			},
			services:  []*platformApi.ArangoPlatformService{svc(ns, "a", "app", true)},
			workflows: []*platformApi.ArangoPlatformWorkflow{wf(ns, "w", "flow", true)},
		}.cache()

		// Service a requires the flow chart provided by the still-hibernated workflow w.
		subject := serviceSubject("a", "app", []string{"flow"}, false, true)
		require.True(t, EvaluateHibernated(c, subject))
	})
}
