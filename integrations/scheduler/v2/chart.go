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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	pbSchedulerV2 "github.com/arangodb/kube-arangodb/integrations/scheduler/v2/definition"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func (i *implementation) ListCharts(req *pbSchedulerV2.SchedulerV2ListChartsRequest, server pbSchedulerV2.SchedulerV2_ListChartsServer) error {
	ctx := server.Context()

	var ct string

	for {
		resp, err := i.kclient.Arango().PlatformV1alpha1().ArangoPlatformCharts(i.client.Namespace()).List(ctx, meta.ListOptions{
			Limit:    util.OptionalType(req.Items, 128),
			Continue: ct,
		})
		if err != nil {
			logger.Err(err).Warn("Unable to run action: ListCharts")
			return asGRPCError(err)
		}

		var res pbSchedulerV2.SchedulerV2ListChartsResponse

		for _, item := range resp.Items {
			res.Charts = append(res.Charts, item.GetName())
		}

		if err := server.Send(&res); err != nil {
			logger.Err(err).Warn("Unable to send action response: ListCharts")
			return err
		}

		if resp.Continue == "" {
			return nil
		}

		ct = resp.Continue
	}
}

func (i *implementation) GetChart(ctx context.Context, in *pbSchedulerV2.SchedulerV2GetChartRequest) (*pbSchedulerV2.SchedulerV2GetChartResponse, error) {
	resp, err := i.kclient.Arango().PlatformV1alpha1().ArangoPlatformCharts(i.client.Namespace()).Get(ctx, in.GetName(), meta.GetOptions{})
	if err != nil {
		logger.Err(err).Warn("Unable to run action: GetChart")
		return nil, asGRPCError(err)
	}

	if !resp.Status.Conditions.IsTrue(platformApi.SpecValidCondition) {
		return nil, status.Errorf(codes.Unavailable, "Chart Spec is invalid")
	}

	if info := resp.Status.Info; info == nil {
		return nil, status.Errorf(codes.Unavailable, "Chart Infos are missing")
	} else {
		if !info.Valid {
			if msg := info.Message; msg == "" {
				return nil, status.Errorf(codes.Unavailable, "Chart is not Valid")
			} else {
				return nil, status.Errorf(codes.Unavailable, "Chart is not Valid: %s", msg)
			}
		} else {
			if details := info.Details; details == nil {
				return nil, status.Errorf(codes.Unavailable, "Chart Details are missing")
			} else {
				return &pbSchedulerV2.SchedulerV2GetChartResponse{
					Chart: info.Definition,
					Info:  chartInfoDetailsAsInfo(details),
				}, nil
			}
		}
	}
}

func chartInfoDetailsAsInfo(in *platformApi.ChartDetails) *pbSchedulerV2.SchedulerV2ChartInfo {
	if in == nil {
		return nil
	}

	return &pbSchedulerV2.SchedulerV2ChartInfo{
		Name:    in.Name,
		Version: in.Version,
	}
}
