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

package v2

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/runtime/schema"

	pbSchedulerV2 "github.com/arangodb/kube-arangodb/integrations/scheduler/v2/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
)

func (i *implementation) DiscoverAPIResources(ctx context.Context, in *pbSchedulerV2.SchedulerV2DiscoverAPIResourcesRequest) (*pbSchedulerV2.SchedulerV2DiscoverAPIResourcesResponse, error) {
	if in.GetVersion() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Version cannot be empty")
	}

	resp, err := i.client.DiscoverKubernetesApiVersions(schema.GroupVersion{
		Group:   in.GetGroup(),
		Version: in.GetVersion(),
	})
	if err != nil {
		logger.Err(err).Warn("Unable to run action: DiscoverAPIResources")
		return nil, status.Errorf(codes.Internal, "Unable to run action: DiscoverAPIResources: %s", err.Error())
	}

	resources := make([]*pbSchedulerV2.SchedulerV2DiscoverAPIResource, 0, len(resp))

	for _, v := range resp {
		resources = append(resources, newKubernetesApiResourceFromDiscoveryResource(v))
	}

	return &pbSchedulerV2.SchedulerV2DiscoverAPIResourcesResponse{
		Resources: resources,
	}, nil
}

func (i *implementation) DiscoverAPIResource(ctx context.Context, in *pbSchedulerV2.SchedulerV2DiscoverAPIResourceRequest) (*pbSchedulerV2.SchedulerV2DiscoverAPIResourceResponse, error) {
	if in.GetVersion() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Version cannot be empty")
	}
	if in.GetKind() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Kind cannot be empty")
	}

	resp, err := i.client.DiscoverKubernetesApiVersionKind(schema.GroupVersionKind{
		Group:   in.GetGroup(),
		Version: in.GetVersion(),
		Kind:    in.GetKind(),
	})
	if err != nil {
		logger.Err(err).Warn("Unable to run action: DiscoverAPIResource")
		return nil, status.Errorf(codes.Internal, "Unable to run action: DiscoverAPIResource: %s", err.Error())
	}

	if resp != nil {
		return &pbSchedulerV2.SchedulerV2DiscoverAPIResourceResponse{
			Resource: newKubernetesApiResourceFromDiscoveryResource(*resp),
		}, nil
	}

	return &pbSchedulerV2.SchedulerV2DiscoverAPIResourceResponse{}, nil
}

func (i *implementation) KubernetesGet(ctx context.Context, in *pbSchedulerV2.SchedulerV2KubernetesGetRequest) (*pbSchedulerV2.SchedulerV2KubernetesGetResponse, error) {
	reqs := make(helm.Resources, len(in.GetResources()))

	for id := range reqs {
		reqs[id] = in.GetResources()[id].AsHelmResource()
	}

	resp, err := i.client.NativeGet(ctx, reqs...)
	if err != nil {
		logger.Err(err).Warn("Unable to run action: KubernetesGet")
		return nil, status.Errorf(codes.Internal, "Unable to run action: KubernetesGet: %s", err.Error())
	}

	return &pbSchedulerV2.SchedulerV2KubernetesGetResponse{
		Objects: newReleaseInfoResourceObjectsFromResourceObjects(resp),
	}, nil
}
