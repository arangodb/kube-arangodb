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
	"google.golang.org/protobuf/types/known/timestamppb"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	pbSchedulerV2 "github.com/arangodb/kube-arangodb/integrations/scheduler/v2/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
)

func newChartReleaseFromHelmRelease(in *helm.Release) *pbSchedulerV2.SchedulerV2Release {
	if in == nil {
		return nil
	}

	var rel pbSchedulerV2.SchedulerV2Release

	rel.Name = in.Name
	rel.Namespace = in.Namespace
	rel.Values = in.Values
	rel.Version = int64(in.Version)
	rel.Labels = in.Labels
	rel.Info = newChartReleaseInfoFromHelmRelease(in.Info)

	return &rel
}

func newChartReleaseInfoFromHelmRelease(in helm.ReleaseInfo) *pbSchedulerV2.SchedulerV2ReleaseInfo {
	var rel pbSchedulerV2.SchedulerV2ReleaseInfo

	rel.FirstDeployed = timestamppb.New(in.FirstDeployed)
	rel.LastDeployed = timestamppb.New(in.LastDeployed)
	rel.Deleted = timestamppb.New(in.Deleted)
	rel.Description = in.Description
	rel.Notes = in.Notes
	rel.Status = pbSchedulerV2.FromHelmStatus(in.Status)
	rel.Resources = newChartReleaseResourcesInfoFromHelmRelease(in.Resources)

	return &rel
}

func newChartReleaseResourcesInfoFromHelmRelease(in helm.Resources) []*pbSchedulerV2.SchedulerV2ReleaseInfoResource {
	if len(in) == 0 {
		return nil
	}

	var r = make([]*pbSchedulerV2.SchedulerV2ReleaseInfoResource, len(in))

	for id := range r {
		r[id] = newChartReleaseResourceInfoFromHelmRelease(in[id])
	}

	return r
}

func newChartReleaseResourceInfoFromHelmRelease(in helm.Resource) *pbSchedulerV2.SchedulerV2ReleaseInfoResource {
	var r pbSchedulerV2.SchedulerV2ReleaseInfoResource

	r.Gvk = &pbSchedulerV2.SchedulerV2GVK{
		Group:   in.Group,
		Version: in.Version,
		Kind:    in.Kind,
	}

	r.Name = in.Name
	r.Namespace = in.Namespace

	return &r
}

func newKubernetesApiResourceFromDiscoveryResource(in meta.APIResource) *pbSchedulerV2.SchedulerV2DiscoverAPIResource {
	return &pbSchedulerV2.SchedulerV2DiscoverAPIResource{
		Name:               in.Name,
		SingularName:       in.SingularName,
		Namespaced:         in.Namespaced,
		Group:              in.Group,
		Version:            in.Version,
		Kind:               in.Kind,
		Verbs:              in.Verbs,
		ShortNames:         in.ShortNames,
		Categories:         in.Categories,
		StorageVersionHash: in.StorageVersionHash,
	}
}

func newReleaseInfoResourceObjectsFromResourceObjects(in []helm.ResourceObject) []*pbSchedulerV2.SchedulerV2ReleaseInfoResourceObject {
	res := make([]*pbSchedulerV2.SchedulerV2ReleaseInfoResourceObject, len(in))

	for id := range res {
		res[id] = newReleaseInfoResourceObjectFromResourceObject(in[id])
	}

	return res
}

func newReleaseInfoResourceObjectFromResourceObject(in helm.ResourceObject) *pbSchedulerV2.SchedulerV2ReleaseInfoResourceObject {
	var r pbSchedulerV2.SchedulerV2ReleaseInfoResourceObject

	r.Resource = newChartReleaseResourceInfoFromHelmRelease(in.Resource)

	if d := in.Object; d != nil {
		r.Data = &pbSchedulerV2.SchedulerV2ReleaseInfoResourceObjectData{
			Data: d.Data,
		}
	}

	return &r
}
