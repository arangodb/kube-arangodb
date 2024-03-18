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

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1"
	schedulerContainerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/container"
	schedulerContainerResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/container/resources"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

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

func render(t *testing.T, s Scheduler, in Request, templates ...*schedulerApi.ProfileTemplate) validatorExec {
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

func Test_NoProfiles(t *testing.T) {
	render(t, newScheduler(t, tests.NewMetaObjectInDefaultNamespace[*schedulerApi.ArangoProfile](t, "test")), Request{})(func(t *testing.T, err error, template *core.PodTemplateSpec, accepted []string) {
		require.NoError(t, err)

		require.Len(t, accepted, 0)

		tests.GetContainerByNameT(t, template.Spec.Containers, DefaultContainerName)
	})
}

func Test_MissingSelectedProfile(t *testing.T) {
	render(t, newScheduler(t, tests.NewMetaObjectInDefaultNamespace[*schedulerApi.ArangoProfile](t, "test")), Request{
		Profiles: []string{"missing"},
	})(func(t *testing.T, err error, template *core.PodTemplateSpec, accepted []string) {
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
	})), Request{})(func(t *testing.T, err error, template *core.PodTemplateSpec, accepted []string) {
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
	})), Request{})(func(t *testing.T, err error, template *core.PodTemplateSpec, accepted []string) {
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
	})), Request{})(func(t *testing.T, err error, template *core.PodTemplateSpec, accepted []string) {
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
	})), Request{
		Labels: map[string]string{
			"ml.arangodb.com/type": "training",
		},
	}, nil)(func(t *testing.T, err error, template *core.PodTemplateSpec, accepted []string) {
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
		})), Request{
		Labels: map[string]string{
			"ml.arangodb.com/type": "training",
		},
	})(func(t *testing.T, err error, template *core.PodTemplateSpec, accepted []string) {
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
