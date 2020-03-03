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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreatePodDNSName returns the DNS of a pod with a given role & id in
// a given deployment.
func CreatePodDNSName(deployment metav1.Object, role, id string) string {
	return CreatePodHostName(deployment.GetName(), role, id) + "." +
		CreateHeadlessServiceName(deployment.GetName()) + "." +
		deployment.GetNamespace() + ".svc"
}

// CreateDatabaseClientServiceDNSName returns the DNS of the database client service.
func CreateDatabaseClientServiceDNSName(deployment metav1.Object) string {
	return CreateDatabaseClientServiceName(deployment.GetName()) + "." +
		deployment.GetNamespace() + ".svc"
}

// CreateSyncMasterClientServiceDNSName returns the DNS of the syncmaster client service.
func CreateSyncMasterClientServiceDNSName(deployment metav1.Object) string {
	return CreateSyncMasterClientServiceName(deployment.GetName()) + "." +
		deployment.GetNamespace() + ".svc"
}
