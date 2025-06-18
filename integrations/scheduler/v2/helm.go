//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package v2

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	pbSchedulerV2 "github.com/arangodb/kube-arangodb/integrations/scheduler/v2/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
)

func (i *implementation) Alive(ctx context.Context, in *pbSharedV1.Empty) (*pbSharedV1.Empty, error) {
	if err := i.client.Alive(ctx); err != nil {
		logger.Err(err).Warn("Helm is not alive")
		return nil, status.Errorf(codes.Unavailable, "Service is not alive")
	}

	return &pbSharedV1.Empty{}, nil
}

func (i *implementation) List(ctx context.Context, in *pbSchedulerV2.SchedulerV2ListRequest) (*pbSchedulerV2.SchedulerV2ListResponse, error) {
	var mods []util.Mod[action.List]

	mods = append(mods, in.GetOptions().Options()...)
	mods = append(mods, func(action *action.List) {
		var s = labels.NewSelector()
		if action.Selector != "" {
			if n, err := labels.Parse(action.Selector); err == nil {
				s = n
			}
		}

		if r, err := labels.NewRequirement(LabelArangoDBDeploymentName, selection.DoubleEquals, []string{i.cfg.Deployment}); err != nil {
			logger.Err(err).Warn("Unable to render selector")
		} else if r != nil {
			s = s.Add(*r)
		}

		action.Selector = s.String()
	})

	resp, err := i.client.List(ctx, mods...)
	if err != nil {
		logger.Err(err).Warn("Unable to run action: List")
		return nil, status.Errorf(codes.Internal, "Unable to run action: List: %s", err.Error())
	}

	releases := make(map[string]*pbSchedulerV2.SchedulerV2Release, len(resp))

	for _, r := range resp {
		releases[r.Name] = newChartReleaseFromHelmRelease(&r)
	}

	return &pbSchedulerV2.SchedulerV2ListResponse{
		Releases: releases,
	}, nil
}

func (i *implementation) Status(ctx context.Context, in *pbSchedulerV2.SchedulerV2StatusRequest) (*pbSchedulerV2.SchedulerV2StatusResponse, error) {
	if in.GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Name cannot be empty")
	}

	resp, err := i.client.Status(ctx, in.GetName())
	if err != nil {
		logger.Err(err).Warn("Unable to run action: Status")
		return nil, status.Errorf(codes.Internal, "Unable to run action: Status: %s", err.Error())
	}

	return &pbSchedulerV2.SchedulerV2StatusResponse{
		Release: newChartReleaseFromHelmRelease(resp),
	}, nil
}
func (i *implementation) StatusObjects(ctx context.Context, in *pbSchedulerV2.SchedulerV2StatusObjectsRequest) (*pbSchedulerV2.SchedulerV2StatusObjectsResponse, error) {
	if in.GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Name cannot be empty")
	}

	resp, objs, err := i.client.StatusObjects(ctx, in.GetName())
	if err != nil {
		logger.Err(err).Warn("Unable to run action: Status")
		return nil, status.Errorf(codes.Internal, "Unable to run action: Status: %s", err.Error())
	}

	return &pbSchedulerV2.SchedulerV2StatusObjectsResponse{
		Release: newChartReleaseFromHelmRelease(resp),
		Objects: newReleaseInfoResourceObjectsFromResourceObjects(objs),
	}, nil
}

func (i *implementation) Install(ctx context.Context, in *pbSchedulerV2.SchedulerV2InstallRequest) (*pbSchedulerV2.SchedulerV2InstallResponse, error) {
	if in.GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Name cannot be empty")
	}

	var mods []util.Mod[action.Install]

	mods = append(mods, in.GetOptions().Options()...)
	mods = append(mods, func(action *action.Install) {
		action.ReleaseName = in.GetName()
		action.Namespace = i.cfg.Namespace

		if action.Labels == nil {
			action.Labels = map[string]string{}
		}

		action.Labels[LabelArangoDBDeploymentName] = i.cfg.Deployment
	})

	resp, err := i.client.Install(ctx, in.GetChart(), in.GetValues(), mods...)
	if err != nil {
		logger.Err(err).Warn("Unable to run action: Install")
		return nil, status.Errorf(codes.Internal, "Unable to run action: Install: %s", err.Error())
	}

	return &pbSchedulerV2.SchedulerV2InstallResponse{
		Release: newChartReleaseFromHelmRelease(resp),
	}, nil
}

func (i *implementation) Upgrade(ctx context.Context, in *pbSchedulerV2.SchedulerV2UpgradeRequest) (*pbSchedulerV2.SchedulerV2UpgradeResponse, error) {
	if in.GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Name cannot be empty")
	}

	var mods []util.Mod[action.Upgrade]

	mods = append(mods, in.GetOptions().Options()...)
	mods = append(mods, func(action *action.Upgrade) {
		action.Namespace = i.cfg.Namespace

		if action.Labels == nil {
			action.Labels = map[string]string{}
		}

		action.Labels[LabelArangoDBDeploymentName] = i.cfg.Deployment
	})

	resp, err := i.client.Upgrade(ctx, in.GetName(), in.GetChart(), in.GetValues(), mods...)
	if err != nil {
		logger.Err(err).Warn("Unable to run action: Upgrade")
		return nil, status.Errorf(codes.Internal, "Unable to run action: Upgrade: %s", err.Error())
	}

	var r pbSchedulerV2.SchedulerV2UpgradeResponse

	if q := resp.Before; q != nil {
		r.Before = newChartReleaseFromHelmRelease(q)
	}

	if q := resp.After; q != nil {
		r.After = newChartReleaseFromHelmRelease(q)
	}

	return &r, nil
}

func (i *implementation) Uninstall(ctx context.Context, in *pbSchedulerV2.SchedulerV2UninstallRequest) (*pbSchedulerV2.SchedulerV2UninstallResponse, error) {
	if in.GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Name cannot be empty")
	}

	var mods []util.Mod[action.Uninstall]

	mods = append(mods, in.GetOptions().Options()...)

	resp, err := i.client.Uninstall(ctx, in.GetName(), mods...)
	if err != nil {
		logger.Err(err).Warn("Unable to run action: Uninstall")
		return nil, status.Errorf(codes.Internal, "Unable to run action: Uninstall: %s", err.Error())
	}

	return &pbSchedulerV2.SchedulerV2UninstallResponse{
		Info:    resp.Info,
		Release: newChartReleaseFromHelmRelease(&resp.Release),
	}, nil
}

func (i *implementation) Test(ctx context.Context, in *pbSchedulerV2.SchedulerV2TestRequest) (*pbSchedulerV2.SchedulerV2TestResponse, error) {
	if in.GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Name cannot be empty")
	}

	var mods []util.Mod[action.ReleaseTesting]

	mods = append(mods, in.GetOptions().Options()...)

	resp, err := i.client.Test(ctx, in.GetName(), mods...)
	if err != nil {
		logger.Err(err).Warn("Unable to run action: Test")
		return nil, status.Errorf(codes.Internal, "Unable to run action: Test: %s", err.Error())
	}

	return &pbSchedulerV2.SchedulerV2TestResponse{
		Release: newChartReleaseFromHelmRelease(resp),
	}, nil
}

func (i *implementation) InstallV2(ctx context.Context, in *pbSchedulerV2.SchedulerV2InstallV2Request) (*pbSchedulerV2.SchedulerV2InstallV2Response, error) {
	if in.GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Name cannot be empty")
	}

	chart, err := i.GetChart(ctx, &pbSchedulerV2.SchedulerV2GetChartRequest{Name: in.GetChart()})
	if err != nil {
		return nil, err
	}

	rawValues := make([]helm.Values, 0, len(in.Values)+1)

	if len(chart.Overrides) > 0 {
		rawValues = append(rawValues, chart.Overrides)
	}

	for _, v := range in.Values {
		if len(v) > 0 {
			rawValues = append(rawValues, v)
		}
	}

	values, err := helm.NewMergeRawValues(helm.MergeMaps, rawValues...)
	if err != nil {
		return nil, err
	}

	var mods []util.Mod[action.Install]

	mods = append(mods, in.GetOptions().Options()...)
	mods = append(mods, func(action *action.Install) {
		action.ReleaseName = in.GetName()
		action.Namespace = i.cfg.Namespace

		if action.Labels == nil {
			action.Labels = map[string]string{}
		}

		action.Labels[LabelArangoDBDeploymentName] = i.cfg.Deployment
	})

	resp, err := i.client.Install(ctx, chart.Chart, values, mods...)
	if err != nil {
		logger.Err(err).Warn("Unable to run action: InstallV2")
		return nil, status.Errorf(codes.Internal, "Unable to run action: InstallV2: %s", err.Error())
	}

	return &pbSchedulerV2.SchedulerV2InstallV2Response{
		Release: newChartReleaseFromHelmRelease(resp),
	}, nil
}

func (i *implementation) UpgradeV2(ctx context.Context, in *pbSchedulerV2.SchedulerV2UpgradeV2Request) (*pbSchedulerV2.SchedulerV2UpgradeV2Response, error) {
	if in.GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Name cannot be empty")
	}

	chart, err := i.GetChart(ctx, &pbSchedulerV2.SchedulerV2GetChartRequest{Name: in.GetChart()})
	if err != nil {
		return nil, err
	}

	rawValues := make([]helm.Values, 0, len(in.Values)+1)

	if len(chart.Overrides) > 0 {
		rawValues = append(rawValues, chart.Overrides)
	}

	for _, v := range in.Values {
		if len(v) > 0 {
			rawValues = append(rawValues, v)
		}
	}

	values, err := helm.NewMergeRawValues(helm.MergeMaps, rawValues...)
	if err != nil {
		return nil, err
	}

	var mods []util.Mod[action.Upgrade]

	mods = append(mods, in.GetOptions().Options()...)
	mods = append(mods, func(action *action.Upgrade) {
		action.Namespace = i.cfg.Namespace

		if action.Labels == nil {
			action.Labels = map[string]string{}
		}

		action.Labels[LabelArangoDBDeploymentName] = i.cfg.Deployment
	})

	resp, err := i.client.Upgrade(ctx, in.GetName(), chart.Chart, values, mods...)
	if err != nil {
		logger.Err(err).Warn("Unable to run action: UpgradeV2")
		return nil, status.Errorf(codes.Internal, "Unable to run action: UpgradeV2: %s", err.Error())
	}

	var r pbSchedulerV2.SchedulerV2UpgradeV2Response

	if q := resp.Before; q != nil {
		r.Before = newChartReleaseFromHelmRelease(q)
	}

	if q := resp.After; q != nil {
		r.After = newChartReleaseFromHelmRelease(q)
	}

	return &r, nil
}
