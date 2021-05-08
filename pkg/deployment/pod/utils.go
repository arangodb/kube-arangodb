package pod

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GenerateMemberEndpoint(services service.Inspector, apiObject meta.Object, spec api.DeploymentSpec, group api.ServerGroup, member api.MemberStatus) (string, error) {
	memberName := member.ArangoMemberName(apiObject.GetName(), group)
	svc, ok := services.Service(memberName)
	if !ok {
		return "", errors.Newf("Service %s not found", memberName)
	}

	return GenerateMemberEndpointFromService(svc, apiObject, spec, group, member)
}

func GenerateMemberEndpointFromService(svc *core.Service, apiObject meta.Object, spec api.DeploymentSpec, group api.ServerGroup, member api.MemberStatus) (string, error) {
	if group.IsArangod() {
		switch method := spec.CommunicationMethod.Get(); method {
		case api.DeploymentCommunicationMethodDNS, api.DeploymentCommunicationMethodIP:
			switch method {
			case api.DeploymentCommunicationMethodDNS:
				return k8sutil.CreateServiceDNSNameWithDomain(svc, spec.ClusterDomain), nil
			case api.DeploymentCommunicationMethodIP:
				if svc.Spec.ClusterIP == "" {
					return "", errors.Newf("ClusterIP of service %s is empty", svc.GetName())
				}

				return svc.Spec.ClusterIP, nil
			case api.DeploymentCommunicationMethodShortDNS:
				return svc.GetName(), nil
			default:
				return "", errors.Newf("Unexpected method %s", method.String())
			}
		default:
			return k8sutil.CreatePodDNSNameWithDomain(apiObject, spec.ClusterDomain, group.AsRole(), member.ID), nil
		}
	} else {
		return k8sutil.CreateSyncMasterClientServiceDNSNameWithDomain(apiObject, spec.ClusterDomain), nil
	}
}
