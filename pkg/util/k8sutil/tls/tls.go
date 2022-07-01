//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package tls

import (
	"net/url"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

type KeyfileInput struct {
	AltNames []string
	Email    []string
}

func (k KeyfileInput) Append(b KeyfileInput) KeyfileInput {
	k.Email = append(k.Email, b.Email...)
	k.AltNames = append(k.AltNames, b.AltNames...)
	return k
}

func GetAltNames(tls api.TLSSpec) (KeyfileInput, error) {
	var k KeyfileInput

	// Load alt names
	dnsNames, ipAddresses, emailAddress, err := tls.GetParsedAltNames()
	if err != nil {
		return k, errors.WithStack(err)
	}

	k.AltNames = append(k.AltNames, dnsNames...)
	k.AltNames = append(k.AltNames, ipAddresses...)

	k.Email = emailAddress

	return k, nil
}

func GetServerAltNames(deployment meta.Object, spec api.DeploymentSpec, tls api.TLSSpec, service *core.Service, group api.ServerGroup, member api.MemberStatus) (KeyfileInput, error) {
	var k KeyfileInput

	k.AltNames = append(k.AltNames,
		k8sutil.CreateDatabaseClientServiceDNSName(deployment),
		k8sutil.CreatePodDNSName(deployment, group.AsRole(), member.ID),
		k8sutil.CreateServiceDNSName(service),
		service.Spec.ClusterIP,
		service.GetName(),
	)

	if spec.ClusterDomain != nil {
		k.AltNames = append(k.AltNames,
			k8sutil.CreateDatabaseClientServiceDNSNameWithDomain(deployment, spec.ClusterDomain),
			k8sutil.CreatePodDNSNameWithDomain(deployment, spec.ClusterDomain, group.AsRole(), member.ID),
			k8sutil.CreateServiceDNSNameWithDomain(service, spec.ClusterDomain))
	}

	if ip := spec.ExternalAccess.GetLoadBalancerIP(); ip != "" {
		k.AltNames = append(k.AltNames, ip)
	}

	if names, err := GetAltNames(tls); err != nil {
		return k, errors.WithStack(err)
	} else {
		k = k.Append(names)
	}

	return k, nil
}

func GetSyncAltNames(deployment meta.Object, spec api.DeploymentSpec, tls api.TLSSpec, group api.ServerGroup, member api.MemberStatus) (KeyfileInput, error) {
	k, err := GetAltNames(tls)
	if err != nil {
		return k, errors.WithStack(err)
	}

	k.AltNames = append(k.AltNames,
		k8sutil.CreateSyncMasterClientServiceName(deployment.GetName()),
		k8sutil.CreateSyncMasterClientServiceDNSNameWithDomain(deployment, spec.ClusterDomain),
		k8sutil.CreatePodDNSNameWithDomain(deployment, spec.ClusterDomain, group.AsRole(), member.ID),
	)

	masterEndpoint := spec.Sync.ExternalAccess.ResolveMasterEndpoint(k8sutil.CreateSyncMasterClientServiceDNSNameWithDomain(deployment, spec.ClusterDomain), shared.ArangoSyncMasterPort)
	for _, ep := range masterEndpoint {
		if u, err := url.Parse(ep); err == nil {
			k.AltNames = append(k.AltNames, u.Hostname())
		}
	}

	return k, nil
}
