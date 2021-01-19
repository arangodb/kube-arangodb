//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package k8sutil

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func appendDeploymentClusterDomain(dns string, domain *string) string {
	if domain == nil || *domain == "" {
		return dns
	}

	return fmt.Sprintf("%s.%s", dns, *domain)
}

// CreatePodDNSName returns the DNS of a pod with a given role & id in
// a given deployment.
func CreatePodDNSName(deployment metav1.Object, role, id string) string {
	return fmt.Sprintf("%s.%s.%s.svc", CreatePodHostName(deployment.GetName(), role, id), CreateHeadlessServiceName(deployment.GetName()), deployment.GetNamespace())
}

// CreatePodDNSName returns the DNS of a pod with a given role & id in
// a given deployment.
func CreatePodDNSNameWithDomain(deployment metav1.Object, domain *string, role, id string) string {
	return appendDeploymentClusterDomain(CreatePodDNSName(deployment, role, id), domain)
}

// CreateDatabaseClientServiceDNSNameWithDomain returns the DNS of the database client service.
func CreateDatabaseClientServiceDNSNameWithDomain(deployment metav1.Object, domain *string) string {
	return appendDeploymentClusterDomain(CreateDatabaseClientServiceDNSName(deployment), domain)
}

// CreateDatabaseClientServiceDNSName returns the DNS of the database client service.
func CreateDatabaseClientServiceDNSName(deployment metav1.Object) string {
	return fmt.Sprintf("%s.%s.svc", CreateDatabaseClientServiceName(deployment.GetName()), deployment.GetNamespace())
}

// CreateSyncMasterClientServiceDNSNameWithDomain returns the DNS of the syncmaster client service.
func CreateSyncMasterClientServiceDNSNameWithDomain(deployment metav1.Object, domain *string) string {
	return appendDeploymentClusterDomain(CreateSyncMasterClientServiceDNSName(deployment), domain)
}

// CreateSyncMasterClientServiceDNSName returns the DNS of the syncmaster client service.
func CreateSyncMasterClientServiceDNSName(deployment metav1.Object) string {
	return fmt.Sprintf("%s.%s.svc", CreateSyncMasterClientServiceName(deployment.GetName()), deployment.GetNamespace())
}
