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

// CreateDatabaseExternalAccessServiceName returns the name of the service used to access the database from
// output the kubernetes cluster.
func CreateDatabaseExternalAccessServiceName(deploymentName string) string {
	return deploymentName + "-ea"
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
// The returned bool is true if the service is created, or false when the service already existed.
func CreateHeadlessService(kubecli kubernetes.Interface, deployment metav1.Object, owner metav1.OwnerReference) (string, bool, error) {
	deploymentName := deployment.GetName()
	svcName := CreateHeadlessServiceName(deploymentName)
	ports := []v1.ServicePort{
		v1.ServicePort{
			Name:     "server",
			Protocol: v1.ProtocolTCP,
			Port:     ArangoPort,
		},
	}
	publishNotReadyAddresses := false
	sessionAffinity := v1.ServiceAffinityNone
	serviceType := v1.ServiceTypeClusterIP
	newlyCreated, err := createService(kubecli, svcName, deploymentName, deployment.GetNamespace(), ClusterIPNone, "", serviceType, ports, "", publishNotReadyAddresses, sessionAffinity, owner)
	if err != nil {
		return "", false, maskAny(err)
	}
	return svcName, newlyCreated, nil
}

// CreateDatabaseClientService prepares and creates a service in k8s, used by database clients within the k8s cluster.
// If the service already exists, nil is returned.
// If another error occurs, that error is returned.
// The returned bool is true if the service is created, or false when the service already existed.
func CreateDatabaseClientService(kubecli kubernetes.Interface, deployment metav1.Object, single bool, owner metav1.OwnerReference) (string, bool, error) {
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
	publishNotReadyAddresses := true
	sessionAffinity := v1.ServiceAffinityClientIP
	serviceType := v1.ServiceTypeClusterIP
	newlyCreated, err := createService(kubecli, svcName, deploymentName, deployment.GetNamespace(), "", role, serviceType, ports, "", publishNotReadyAddresses, sessionAffinity, owner)
	if err != nil {
		return "", false, maskAny(err)
	}
	return svcName, newlyCreated, nil
}

// CreateExternalAccessService prepares and creates a service in k8s, used to access the database/sync from outside k8s cluster.
// If the service already exists, nil is returned.
// If another error occurs, that error is returned.
// The returned bool is true if the service is created, or false when the service already existed.
func CreateExternalAccessService(kubecli kubernetes.Interface, svcName, role string, deployment metav1.Object, serviceType v1.ServiceType, port, nodePort int, loadBalancerIP string, sessionAffinity v1.ServiceAffinity, owner metav1.OwnerReference) (string, bool, error) {
	deploymentName := deployment.GetName()
	ports := []v1.ServicePort{
		v1.ServicePort{
			Name:     "server",
			Protocol: v1.ProtocolTCP,
			Port:     int32(port),
			NodePort: int32(nodePort),
		},
	}
	publishNotReadyAddresses := true
	newlyCreated, err := createService(kubecli, svcName, deploymentName, deployment.GetNamespace(), "", role, serviceType, ports, loadBalancerIP, publishNotReadyAddresses, sessionAffinity, owner)
	if err != nil {
		return "", false, maskAny(err)
	}
	return svcName, newlyCreated, nil
}

// createService prepares and creates a service in k8s.
// If the service already exists, nil is returned.
// If another error occurs, that error is returned.
// The returned bool is true if the service is created, or false when the service already existed.
func createService(kubecli kubernetes.Interface, svcName, deploymentName, ns, clusterIP, role string, serviceType v1.ServiceType,
	ports []v1.ServicePort, loadBalancerIP string, publishNotReadyAddresses bool, sessionAffinity v1.ServiceAffinity, owner metav1.OwnerReference) (bool, error) {
	labels := LabelsForDeployment(deploymentName, role)
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   svcName,
			Labels: labels,
			Annotations: map[string]string{
				// This annotation is deprecated, PublishNotReadyAddresses is
				// used instead. We leave the annotation in for a while.
				// See https://github.com/kubernetes/kubernetes/pull/49061
				TolerateUnreadyEndpointsAnnotation: "true",
			},
		},
		Spec: v1.ServiceSpec{
			Type:                     serviceType,
			Ports:                    ports,
			Selector:                 labels,
			ClusterIP:                clusterIP,
			PublishNotReadyAddresses: publishNotReadyAddresses,
			SessionAffinity:          sessionAffinity,
			LoadBalancerIP:           loadBalancerIP,
		},
	}
	addOwnerRefToObject(svc.GetObjectMeta(), &owner)
	if _, err := kubecli.CoreV1().Services(ns).Create(svc); IsAlreadyExists(err) {
		return false, nil
	} else if err != nil {
		return false, maskAny(err)
	}
	return true, nil
}
