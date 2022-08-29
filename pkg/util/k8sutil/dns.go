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

package k8sutil

import (
	"fmt"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

func appendDeploymentClusterDomain(dns string, domain *string) string {
	if domain == nil || *domain == "" {
		return dns
	}

	return fmt.Sprintf("%s.%s", dns, *domain)
}

// CreatePodDNSName returns the DNS of a pod with a given role & id in
// a given deployment.
func CreatePodDNSName(deployment meta.Object, role, id string) string {
	return fmt.Sprintf("%s.%s.%s.svc", shared.CreatePodHostName(deployment.GetName(), role, id), CreateHeadlessServiceName(deployment.GetName()), deployment.GetNamespace())
}

// CreatePodDNSName returns the DNS of a pod with a given role & id in
// a given deployment.
func CreatePodDNSNameWithDomain(deployment meta.Object, domain *string, role, id string) string {
	return appendDeploymentClusterDomain(CreatePodDNSName(deployment, role, id), domain)
}

// CreateServiceDNSName returns the DNS of a service.
func CreateServiceDNSName(svc *core.Service) string {
	return fmt.Sprintf("%s.%s.svc", svc.GetName(), svc.GetNamespace())
}

// CreateServiceDNSNameWithDomain returns the DNS of a service extended with domain.
func CreateServiceDNSNameWithDomain(svc *core.Service, domain *string) string {
	return appendDeploymentClusterDomain(CreateServiceDNSName(svc), domain)
}

// CreateDatabaseClientServiceDNSNameWithDomain returns the DNS of the database client service.
func CreateDatabaseClientServiceDNSNameWithDomain(deployment meta.Object, domain *string) string {
	return appendDeploymentClusterDomain(CreateDatabaseClientServiceDNSName(deployment), domain)
}

// CreateDatabaseClientServiceDNSName returns the DNS of the database client service.
func CreateDatabaseClientServiceDNSName(deployment meta.Object) string {
	return fmt.Sprintf("%s.%s.svc", CreateDatabaseClientServiceName(deployment.GetName()), deployment.GetNamespace())
}

// CreateSyncMasterClientServiceDNSNameWithDomain returns the DNS of the syncmaster client service.
func CreateSyncMasterClientServiceDNSNameWithDomain(deployment meta.Object, domain *string) string {
	return appendDeploymentClusterDomain(CreateSyncMasterClientServiceDNSName(deployment), domain)
}

// CreateSyncMasterClientServiceDNSName returns the DNS of the syncmaster client service.
func CreateSyncMasterClientServiceDNSName(deployment meta.Object) string {
	return fmt.Sprintf("%s.%s.svc", CreateSyncMasterClientServiceName(deployment.GetName()), deployment.GetNamespace())
}
