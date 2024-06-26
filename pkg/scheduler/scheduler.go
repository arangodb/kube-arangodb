//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package scheduler

import (
	"context"

	core "k8s.io/api/core/v1"

	pbSchedulerV1 "github.com/arangodb/kube-arangodb/integrations/scheduler/v1/definition"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/generators/kubernetes"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func NewScheduler(client kclient.Client, namespace string) Scheduler {
	return scheduler{
		client:    client,
		namespace: namespace,
	}
}

type Scheduler interface {
	Render(ctx context.Context, in *pbSchedulerV1.Spec, templates ...*schedulerApi.ProfileTemplate) (*core.PodTemplateSpec, []string, error)
}

type scheduler struct {
	client    kclient.Client
	namespace string
}

func (s scheduler) Render(ctx context.Context, in *pbSchedulerV1.Spec, templates ...*schedulerApi.ProfileTemplate) (*core.PodTemplateSpec, []string, error) {
	if in == nil {
		return nil, nil, errors.Errorf("Unable to parse nil Spec")
	}

	profileMap, err := kubernetes.MapObjects[*schedulerApi.ArangoProfileList, *schedulerApi.ArangoProfile](ctx, s.client.Arango().SchedulerV1beta1().ArangoProfiles(s.namespace), func(result *schedulerApi.ArangoProfileList) []*schedulerApi.ArangoProfile {
		q := make([]*schedulerApi.ArangoProfile, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})

	if err != nil {
		return nil, nil, err
	}

	var labels map[string]string
	var additionalProfiles []string

	if job := in.Job; job != nil {
		labels = job.Labels
		additionalProfiles = job.Profiles
	}

	if len(in.Containers) == 0 {
		return nil, nil, errors.Errorf("Required at least 1 container")
	}

	profiles := profileMap.AsList().Filter(func(a *schedulerApi.ArangoProfile) bool {
		return a != nil && a.Spec.Template != nil
	}).Filter(func(a *schedulerApi.ArangoProfile) bool {
		if a.Spec.Selectors == nil {
			return false
		}

		if !a.Spec.Selectors.Select(labels) {
			return false
		}

		return true
	})

	for _, name := range additionalProfiles {
		p, ok := profileMap.ByName(name)
		if !ok {
			return nil, nil, errors.Errorf("Profile with name `%s` is missing", name)
		}

		profiles = append(profiles, p)
	}

	profiles = profiles.Unique(func(existing kubernetes.List[*schedulerApi.ArangoProfile], o *schedulerApi.ArangoProfile) bool {
		return existing.Contains(func(a *schedulerApi.ArangoProfile) bool {
			return a.GetName() == o.GetName()
		})
	})

	profiles = profiles.Sort(func(a, b *schedulerApi.ArangoProfile) bool {
		return a.Spec.Template.GetPriority() > b.Spec.Template.GetPriority()
	})

	if err := errors.Errors(kubernetes.Extract(profiles, func(in *schedulerApi.ArangoProfile) error {
		return in.Spec.Validate()
	})...); err != nil {
		return nil, nil, err
	}

	extracted := schedulerApi.ProfileTemplates(kubernetes.Extract(profiles, func(in *schedulerApi.ArangoProfile) *schedulerApi.ProfileTemplate {
		return in.Spec.Template
	}).Append(templates...).Append(baseAsTemplate(in)))

	names := kubernetes.Extract(profiles, func(in *schedulerApi.ArangoProfile) string {
		return in.GetName()
	})

	var pod core.PodTemplateSpec

	if err := extracted.RenderOnTemplate(&pod); err != nil {
		return nil, names, err
	}

	return &pod, names, nil
}
