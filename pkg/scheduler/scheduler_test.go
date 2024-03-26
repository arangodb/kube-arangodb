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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	pbSchedulerV1 "github.com/arangodb/kube-arangodb/integrations/scheduler/v1/definition"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1"
	schedulerContainerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/container"
	schedulerContainerResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/container/resources"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

const DefaultContainerName = "job"

func newScheduler(t *testing.T, objects ...*schedulerApi.ArangoProfile) Scheduler {
	client := kclient.NewFakeClientBuilder().Client()

	objs := make([]interface{}, len(objects))
	for id := range objs {
		objs[id] = &objects[id]
	}

	tests.CreateObjects(t, client.Kubernetes(), client.Arango(), objs...)

	return NewScheduler(client, tests.FakeNamespace)
}

type validatorExec func(in validator)

type validator func(t *testing.T, err error, template *core.PodTemplateSpec, accepted []string)

func getRequest(in ...func(obj *pbSchedulerV1.Spec)) *pbSchedulerV1.Spec {
	var r pbSchedulerV1.Spec
	for _, i := range in {
		i(&r)
	}

	return &r
}

func withProfiles(profiles ...string) func(obj *pbSchedulerV1.Spec) {
	return func(obj *pbSchedulerV1.Spec) {
		if obj.Job == nil {
			obj.Job = &pbSchedulerV1.JobBase{}
		}

		obj.Job.Profiles = append(obj.Job.Profiles, profiles...)
	}
}

func withLabels(labels map[string]string) func(obj *pbSchedulerV1.Spec) {
	return func(obj *pbSchedulerV1.Spec) {
		if obj.Job == nil {
			obj.Job = &pbSchedulerV1.JobBase{}
		}

		if obj.Job.Labels == nil {
			obj.Job.Labels = make(map[string]string)
		}

		for k, v := range labels {
			obj.Job.Labels[k] = v
		}
	}
}

func withDefaultContainer(in ...func(obj *pbSchedulerV1.ContainerBase)) func(obj *pbSchedulerV1.Spec) {
	return func(obj *pbSchedulerV1.Spec) {
		if obj.Containers == nil {
			obj.Containers = make(map[string]*pbSchedulerV1.ContainerBase)
		}

		var c pbSchedulerV1.ContainerBase

		for _, i := range in {
			i(&c)
		}

		obj.Containers[DefaultContainerName] = &c
	}
}

func render(t *testing.T, s Scheduler, in *pbSchedulerV1.Spec, templates ...*schedulerApi.ProfileTemplate) validatorExec {
	pod, accepted, err := s.Render(context.Background(), in, templates...)
	t.Logf("Accepted templates: %s", strings.Join(accepted, ", "))
	if err != nil {
		return runValidate(t, err, pod, accepted)
	}
	require.NoError(t, err)

	data, err := yaml.Marshal(pod)
	require.NoError(t, err)

	t.Logf("Rendered Template:\n%s", string(data))

	return runValidate(t, nil, pod, accepted)
}

func runValidate(t *testing.T, err error, template *core.PodTemplateSpec, accepted []string) validatorExec {
	return func(in validator) {
		t.Run("Validate", func(t *testing.T) {
			in(t, err, template, accepted)
		})
	}
}

func Test_Nil(t *testing.T) {
	render(t, newScheduler(t, tests.NewMetaObjectInDefaultNamespace[*schedulerApi.ArangoProfile](t, "test")), nil)(func(t *testing.T, err error, template *core.PodTemplateSpec, accepted []string) {
		require.EqualError(t, err, "Unable to parse nil Spec")
	})
}

func Test_NoProfiles(t *testing.T) {
	render(t, newScheduler(t, tests.NewMetaObjectInDefaultNamespace[*schedulerApi.ArangoProfile](t, "test")), &pbSchedulerV1.Spec{})(func(t *testing.T, err error, template *core.PodTemplateSpec, accepted []string) {
		require.EqualError(t, err, "Required at least 1 container")
	})
}

func Test_MissingSelectedProfile(t *testing.T) {
	render(t, newScheduler(t, tests.NewMetaObjectInDefaultNamespace[*schedulerApi.ArangoProfile](t, "test")),
		getRequest(withProfiles("missing"), withDefaultContainer()),
	)(func(t *testing.T, err error, template *core.PodTemplateSpec, accepted []string) {
		require.EqualError(t, err, "Profile with name `missing` is missing")
	})
}

func Test_SelectorWithoutSelector(t *testing.T) {
	render(t, newScheduler(t, tests.NewMetaObjectInDefaultNamespace[*schedulerApi.ArangoProfile](t, "test", func(t *testing.T, obj *schedulerApi.ArangoProfile) {
		obj.Spec.Template = &schedulerApi.ProfileTemplate{
			Container: &schedulerApi.ProfileContainerTemplate{
				Containers: schedulerContainerApi.Containers{
					DefaultContainerName: {
						Image: &schedulerContainerResourcesApi.Image{
							Image: util.NewType("image:1"),
						},
					},
				},
			},
		}
	})), getRequest(withDefaultContainer(func(obj *pbSchedulerV1.ContainerBase) {
		obj.Image = util.NewType("")
	})))(func(t *testing.T, err error, template *core.PodTemplateSpec, accepted []string) {
		require.NoError(t, err)

		require.Len(t, accepted, 0)

		c := tests.GetContainerByNameT(t, template.Spec.Containers, DefaultContainerName)
		require.Equal(t, "", c.Image)
	})
}

func Test_SelectorWithSelectorAll(t *testing.T) {
	render(t, newScheduler(t, tests.NewMetaObjectInDefaultNamespace[*schedulerApi.ArangoProfile](t, "test", func(t *testing.T, obj *schedulerApi.ArangoProfile) {
		obj.Spec.Selectors = &schedulerApi.ProfileSelectors{
			Label: &meta.LabelSelector{},
		}
		obj.Spec.Template = &schedulerApi.ProfileTemplate{
			Container: &schedulerApi.ProfileContainerTemplate{
				Containers: schedulerContainerApi.Containers{
					DefaultContainerName: {
						Image: &schedulerContainerResourcesApi.Image{
							Image: util.NewType("image:1"),
						},
					},
				},
			},
		}
	})), getRequest(withDefaultContainer()))(func(t *testing.T, err error, template *core.PodTemplateSpec, accepted []string) {
		require.NoError(t, err)

		require.Len(t, accepted, 1)
		require.Equal(t, []string{
			"test",
		}, accepted)

		c := tests.GetContainerByNameT(t, template.Spec.Containers, DefaultContainerName)
		require.Equal(t, "image:1", c.Image)
	})
}

func Test_SelectorWithSpecificSelector_MissingLabel(t *testing.T) {
	render(t, newScheduler(t, tests.NewMetaObjectInDefaultNamespace[*schedulerApi.ArangoProfile](t, "test", func(t *testing.T, obj *schedulerApi.ArangoProfile) {
		obj.Spec.Selectors = &schedulerApi.ProfileSelectors{
			Label: &meta.LabelSelector{
				MatchExpressions: []meta.LabelSelectorRequirement{
					{
						Key:      "ml.arangodb.com/type",
						Operator: meta.LabelSelectorOpExists,
					},
				},
			},
		}
		obj.Spec.Template = &schedulerApi.ProfileTemplate{
			Container: &schedulerApi.ProfileContainerTemplate{
				Containers: schedulerContainerApi.Containers{
					DefaultContainerName: {
						Image: &schedulerContainerResourcesApi.Image{
							Image: util.NewType("image:1"),
						},
					},
				},
			},
		}
	})), getRequest(withDefaultContainer()))(func(t *testing.T, err error, template *core.PodTemplateSpec, accepted []string) {
		require.NoError(t, err)

		require.Len(t, accepted, 0)

		c := tests.GetContainerByNameT(t, template.Spec.Containers, DefaultContainerName)
		require.Equal(t, "", c.Image)
	})
}

func Test_SelectorWithSpecificSelector_PresentLabel(t *testing.T) {
	render(t, newScheduler(t, tests.NewMetaObjectInDefaultNamespace[*schedulerApi.ArangoProfile](t, "test", func(t *testing.T, obj *schedulerApi.ArangoProfile) {
		obj.Spec.Selectors = &schedulerApi.ProfileSelectors{
			Label: &meta.LabelSelector{
				MatchExpressions: []meta.LabelSelectorRequirement{
					{
						Key:      "ml.arangodb.com/type",
						Operator: meta.LabelSelectorOpExists,
					},
				},
			},
		}
		obj.Spec.Template = &schedulerApi.ProfileTemplate{
			Container: &schedulerApi.ProfileContainerTemplate{
				Containers: schedulerContainerApi.Containers{
					DefaultContainerName: {
						Image: &schedulerContainerResourcesApi.Image{
							Image: util.NewType("image:1"),
						},
					},
				},
			},
		}
	})), getRequest(withDefaultContainer(), withLabels(map[string]string{
		"ml.arangodb.com/type": "training",
	})), nil)(func(t *testing.T, err error, template *core.PodTemplateSpec, accepted []string) {
		require.NoError(t, err)

		require.Len(t, accepted, 1)
		require.Equal(t, []string{
			"test",
		}, accepted)

		c := tests.GetContainerByNameT(t, template.Spec.Containers, DefaultContainerName)
		require.Equal(t, "image:1", c.Image)
	})
}

func Test_SelectorWithSpecificSelector_PresentLabel_ByPriority(t *testing.T) {
	render(t, newScheduler(t,
		tests.NewMetaObjectInDefaultNamespace[*schedulerApi.ArangoProfile](t, "test", func(t *testing.T, obj *schedulerApi.ArangoProfile) {
			obj.Spec.Selectors = &schedulerApi.ProfileSelectors{
				Label: &meta.LabelSelector{
					MatchExpressions: []meta.LabelSelectorRequirement{
						{
							Key:      "ml.arangodb.com/type",
							Operator: meta.LabelSelectorOpExists,
						},
					},
				},
			}
			obj.Spec.Template = &schedulerApi.ProfileTemplate{
				Priority: util.NewType(1),
				Container: &schedulerApi.ProfileContainerTemplate{
					Containers: schedulerContainerApi.Containers{
						DefaultContainerName: {
							Image: &schedulerContainerResourcesApi.Image{
								Image: util.NewType("image:1"),
							},
						},
					},
				},
			}
		}), tests.NewMetaObjectInDefaultNamespace[*schedulerApi.ArangoProfile](t, "test2", func(t *testing.T, obj *schedulerApi.ArangoProfile) {
			obj.Spec.Selectors = &schedulerApi.ProfileSelectors{
				Label: &meta.LabelSelector{
					MatchExpressions: []meta.LabelSelectorRequirement{
						{
							Key:      "ml.arangodb.com/type",
							Operator: meta.LabelSelectorOpExists,
						},
					},
				},
			}
			obj.Spec.Template = &schedulerApi.ProfileTemplate{
				Priority: util.NewType(2),
				Container: &schedulerApi.ProfileContainerTemplate{
					Containers: schedulerContainerApi.Containers{
						DefaultContainerName: {
							Image: &schedulerContainerResourcesApi.Image{
								Image: util.NewType("image:2"),
							},
						},
					},
				},
			}
		})), getRequest(withDefaultContainer(), withLabels(map[string]string{
		"ml.arangodb.com/type": "training",
	},
	)))(func(t *testing.T, err error, template *core.PodTemplateSpec, accepted []string) {
		require.NoError(t, err)

		require.Len(t, accepted, 2)
		require.Equal(t, []string{
			"test2",
			"test",
		}, accepted)

		c := tests.GetContainerByNameT(t, template.Spec.Containers, DefaultContainerName)
		require.Equal(t, "image:2", c.Image)
	})
}
