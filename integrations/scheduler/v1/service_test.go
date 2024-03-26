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

package v1

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	batch "k8s.io/api/batch/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	pbSchedulerV1 "github.com/arangodb/kube-arangodb/integrations/scheduler/v1/definition"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Handler(t *testing.T, ctx context.Context, client kclient.Client, mods ...Mod) svc.Handler {
	handler, err := New(ctx, client, NewConfiguration().With(mods...))
	require.NoError(t, err)

	return handler
}

func Client(t *testing.T, ctx context.Context, client kclient.Client, mods ...Mod) pbSchedulerV1.SchedulerV1Client {
	local := svc.NewService(svc.Configuration{
		Address: "127.0.0.1:0",
	}, Handler(t, ctx, client, mods...))

	start := local.Start(ctx)

	return tgrpc.NewGRPCClient(t, ctx, pbSchedulerV1.NewSchedulerV1Client, start.Address())
}

func pointers[T interface{}](in []T) []interface{} {
	r := make([]interface{}, len(in))

	for id := range in {
		r[id] = &in[id]
	}

	return r
}

func renderJobWithClient(t *testing.T, req *pbSchedulerV1.CreateBatchJobRequest, client kclient.Client, namespace string, profiles ...*schedulerApi.ArangoProfile) (*pbSchedulerV1.CreateBatchJobResponse, *batch.Job, error) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	tests.CreateObjects(t, client.Kubernetes(), client.Arango(), pointers(profiles)...)
	defer tests.DeleteObjects(t, client.Kubernetes(), client.Arango(), pointers(profiles)...)

	z := Client(t, ctx, client, func(c Configuration) Configuration {
		c.Namespace = namespace
		c.VerifyAccess = false
		return c
	})

	resp, err := z.CreateBatchJob(context.Background(), req)
	if err != nil {
		return nil, nil, err
	}

	job, err := client.Kubernetes().BatchV1().Jobs(namespace).Get(ctx, resp.GetName(), meta.GetOptions{})
	require.NoError(t, err)

	job.Status.Succeeded = util.TypeOrDefault(job.Spec.Completions, 1)

	job, err = client.Kubernetes().BatchV1().Jobs(namespace).Update(ctx, job, meta.UpdateOptions{})
	require.NoError(t, err)

	tests.NewTimeout(func() error {
		resp, err := z.GetBatchJob(context.Background(), &pbSchedulerV1.GetBatchJobRequest{
			Name: resp.GetName(),
		})

		if err != nil {
			return err
		}

		if !resp.GetExists() {
			return errors.Errorf("Job does not exist")
		}

		if resp.GetBatchJob().GetStatus().GetSucceeded() > 0 && resp.GetBatchJob().GetSpec().GetCompletions() == resp.GetBatchJob().GetStatus().GetSucceeded() {
			return tests.Interrupt()
		}

		return nil
	}).WithTimeoutT(t, time.Minute, time.Second)

	delResp, err := z.DeleteBatchJob(context.Background(), &pbSchedulerV1.DeleteBatchJobRequest{Name: resp.GetName(), DeleteChildPods: util.NewType(true)})
	require.NoError(t, err)
	require.True(t, delResp.GetExists())

	return resp, job, nil
}

func Test_Service(t *testing.T) {
	kc := kclient.NewFakeClient()

	resp, _, err := renderJobWithClient(t, &pbSchedulerV1.CreateBatchJobRequest{
		Spec: &pbSchedulerV1.Spec{
			Metadata: &pbSchedulerV1.Metadata{
				Name: "test",
			},
			Job: &pbSchedulerV1.JobBase{
				Labels: nil,
				Profiles: []string{
					"test",
				},
			},
			Containers: map[string]*pbSchedulerV1.ContainerBase{
				"example": {
					Image: "ubuntu:20.04",
					Args: []string{
						"/bin/bash",
						"-c",
						"true",
					},
				},
			},
		},
		BatchJob: &pbSchedulerV1.BatchJobSpec{
			Completions: util.NewType[int32](1),
		},
	}, kc, tests.FakeNamespace, tests.NewMetaObject(t, tests.FakeNamespace, "test", func(t *testing.T, obj *schedulerApi.ArangoProfile) {
		obj.Spec = schedulerApi.ProfileSpec{}
	}), tests.NewMetaObject(t, tests.FakeNamespace, "test-select-all", func(t *testing.T, obj *schedulerApi.ArangoProfile) {
		obj.Spec = schedulerApi.ProfileSpec{
			Selectors: &schedulerApi.ProfileSelectors{
				Label: &meta.LabelSelector{},
			},
			Template: &schedulerApi.ProfileTemplate{},
		}
	}))
	require.NoError(t, err)

	println(resp.Name)
	println(strings.Join(resp.Profiles, ", "))
}
