//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package pod

import (
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service"
)

func GenerateMemberEndpoint(services service.Inspector, apiObject meta.Object, spec api.DeploymentSpec, group api.ServerGroup, member api.MemberStatus) (string, error) {
	memberName := member.ArangoMemberName(apiObject.GetName(), group)
	svc, ok := services.Service().V1().GetSimple(memberName)
	if !ok {
		return "", errors.Newf("Service %s not found", memberName)
	}

	return GenerateMemberEndpointFromService(svc, apiObject, spec, group, member)
}

func GenerateMemberEndpointFromService(svc *core.Service, apiObject meta.Object, spec api.DeploymentSpec, group api.ServerGroup, member api.MemberStatus) (string, error) {
	if group.IsArangod() {
		switch method := spec.CommunicationMethod.Get(); method {
		case api.DeploymentCommunicationMethodDNS, api.DeploymentCommunicationMethodHeadlessDNS:
			return k8sutil.CreateServiceDNSNameWithDomain(svc, spec.ClusterDomain), nil
		case api.DeploymentCommunicationMethodIP:
			if svc.Spec.ClusterIP == "" {
				return "", errors.Newf("ClusterIP of service %s is empty", svc.GetName())
			}

			if svc.Spec.ClusterIP == core.ClusterIPNone {
				return k8sutil.CreateServiceDNSNameWithDomain(svc, spec.ClusterDomain), nil
			}

			return svc.Spec.ClusterIP, nil
		case api.DeploymentCommunicationMethodShortDNS:
			return svc.GetName(), nil
		default:
			return k8sutil.CreatePodDNSNameWithDomain(apiObject, spec.ClusterDomain, group.AsRole(), member.ID), nil
		}
	} else {
		return k8sutil.CreateSyncMasterClientServiceDNSNameWithDomain(apiObject, spec.ClusterDomain), nil
	}
}
