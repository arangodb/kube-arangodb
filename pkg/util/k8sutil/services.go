//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateHeadlessServiceName returns the name of the headless service for the given
// deployment name.
func CreateHeadlessServiceName(deploymentName string) string {
	return deploymentName + "-int"
}

// CreateDatabaseClientServiceName returns the name of the service used by database clients for the given
// deployment name.
func CreateDatabaseClientServiceName(deploymentName string) string {
	return deploymentName
}

// CreateSyncMasterClientServiceName returns the name of the service used by syncmaster clients for the given
// deployment name.
func CreateSyncMasterClientServiceName(deploymentName string) string {
	return deploymentName + "-sync"
}

// CreateHeadlessService prepares and creates a headless service in k8s, used to provide a stable
// DNS name for all pods.
// If the service already exists, nil is returned.
// If another error occurs, that error is returned.
func CreateHeadlessService(kubecli kubernetes.Interface, deployment metav1.Object, owner metav1.OwnerReference) (string, error) {
	deploymentName := deployment.GetName()
	svcName := CreateHeadlessServiceName(deploymentName)
	ports := []v1.ServicePort{
		v1.ServicePort{
			Name:     "server",
			Protocol: v1.ProtocolTCP,
			Port:     ArangoPort,
		},
	}
	if err := createService(kubecli, svcName, deploymentName, deployment.GetNamespace(), ClusterIPNone, "", ports, owner); err != nil {
		return "", maskAny(err)
	}
	return svcName, nil
}

// CreateDatabaseClientService prepares and creates a service in k8s, used by database clients within the k8s cluster.
// If the service already exists, nil is returned.
// If another error occurs, that error is returned.
func CreateDatabaseClientService(kubecli kubernetes.Interface, deployment metav1.Object, single bool, owner metav1.OwnerReference) (string, error) {
	deploymentName := deployment.GetName()
	svcName := CreateDatabaseClientServiceName(deploymentName)
	ports := []v1.ServicePort{
		v1.ServicePort{
			Name:     "server",
			Protocol: v1.ProtocolTCP,
			Port:     ArangoPort,
		},
	}
	var role string
	if single {
		role = "single"
	} else {
		role = "coordinator"
	}
	if err := createService(kubecli, svcName, deploymentName, deployment.GetNamespace(), "", role, ports, owner); err != nil {
		return "", maskAny(err)
	}
	return svcName, nil
}

// CreateSyncMasterClientService prepares and creates a service in k8s, used by syncmaster clients within the k8s cluster.
// If the service already exists, nil is returned.
// If another error occurs, that error is returned.
func CreateSyncMasterClientService(kubecli kubernetes.Interface, deployment metav1.Object, owner metav1.OwnerReference) (string, error) {
	deploymentName := deployment.GetName()
	svcName := CreateSyncMasterClientServiceName(deploymentName)
	ports := []v1.ServicePort{
		v1.ServicePort{
			Name:     "server",
			Protocol: v1.ProtocolTCP,
			Port:     ArangoPort,
		},
	}
	if err := createService(kubecli, svcName, deploymentName, deployment.GetNamespace(), "", "syncmaster", ports, owner); err != nil {
		return "", maskAny(err)
	}
	return svcName, nil
}

// createService prepares and creates a service in k8s.
// If the service already exists, nil is returned.
// If another error occurs, that error is returned.
func createService(kubecli kubernetes.Interface, svcName, deploymentName, ns, clusterIP, role string, ports []v1.ServicePort, owner metav1.OwnerReference) error {
	labels := LabelsForDeployment(deploymentName, role)
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   svcName,
			Labels: labels,
			Annotations: map[string]string{
				TolerateUnreadyEndpointsAnnotation: "true",
			},
		},
		Spec: v1.ServiceSpec{
			Ports:     ports,
			Selector:  labels,
			ClusterIP: clusterIP,
		},
	}
	addOwnerRefToObject(svc.GetObjectMeta(), owner)
	if _, err := kubecli.CoreV1().Services(ns).Create(svc); err != nil && !IsAlreadyExists(err) {
		return maskAny(err)
	}
	return nil
}
