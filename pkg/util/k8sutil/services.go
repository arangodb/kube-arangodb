//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Tomasz Mielech
//

package k8sutil

import (
	"context"
	"math/rand"
	"net"
	"strconv"
	"strings"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// CreateExporterClientServiceName returns the name of the service used by arangodb-exporter clients for the given
// deployment name.
func CreateExporterClientServiceName(deploymentName string) string {
	return deploymentName + "-exporter"
}

// CreateExporterService
func CreateExporterService(ctx context.Context, cachedStatus service.Inspector, svcs service.ModInterface,
	deployment metav1.Object, owner metav1.OwnerReference) (string, bool, error) {
	deploymentName := deployment.GetName()
	svcName := CreateExporterClientServiceName(deploymentName)

	selectorLabels := LabelsForExporterServiceSelector(deploymentName)

	if _, exists := cachedStatus.Service(svcName); exists {
		return svcName, false, nil
	}

	svc := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   svcName,
			Labels: LabelsForExporterService(deploymentName),
		},
		Spec: core.ServiceSpec{
			ClusterIP: core.ClusterIPNone,
			Ports: []core.ServicePort{
				core.ServicePort{
					Name:     "exporter",
					Protocol: core.ProtocolTCP,
					Port:     ArangoExporterPort,
				},
			},
			Selector: selectorLabels,
		},
	}
	AddOwnerRefToObject(svc.GetObjectMeta(), &owner)
	if _, err := svcs.Create(ctx, svc, metav1.CreateOptions{}); IsAlreadyExists(err) {
		return svcName, false, nil
	} else if err != nil {
		return svcName, false, errors.WithStack(err)
	}
	return svcName, true, nil
}

// CreateHeadlessService prepares and creates a headless service in k8s, used to provide a stable
// DNS name for all pods.
// If the service already exists, nil is returned.
// If another error occurs, that error is returned.
// The returned bool is true if the service is created, or false when the service already existed.
func CreateHeadlessService(ctx context.Context, svcs service.ModInterface, deployment metav1.Object,
	owner metav1.OwnerReference) (string, bool, error) {
	deploymentName := deployment.GetName()
	svcName := CreateHeadlessServiceName(deploymentName)
	ports := []core.ServicePort{
		core.ServicePort{
			Name:     "server",
			Protocol: core.ProtocolTCP,
			Port:     ArangoPort,
		},
	}
	publishNotReadyAddresses := true
	serviceType := core.ServiceTypeClusterIP
	newlyCreated, err := createService(ctx, svcs, svcName, deploymentName, ClusterIPNone, "", serviceType, ports, "", nil, publishNotReadyAddresses, owner)
	if err != nil {
		return "", false, errors.WithStack(err)
	}
	return svcName, newlyCreated, nil
}

// CreateDatabaseClientService prepares and creates a service in k8s, used by database clients within the k8s cluster.
// If the service already exists, nil is returned.
// If another error occurs, that error is returned.
// The returned bool is true if the service is created, or false when the service already existed.
func CreateDatabaseClientService(ctx context.Context, svcs service.ModInterface, deployment metav1.Object, single bool,
	owner metav1.OwnerReference) (string, bool, error) {
	deploymentName := deployment.GetName()
	svcName := CreateDatabaseClientServiceName(deploymentName)
	ports := []core.ServicePort{
		core.ServicePort{
			Name:     "server",
			Protocol: core.ProtocolTCP,
			Port:     ArangoPort,
		},
	}
	var role string
	if single {
		role = "single"
	} else {
		role = "coordinator"
	}
	serviceType := core.ServiceTypeClusterIP
	publishNotReadyAddresses := false
	newlyCreated, err := createService(ctx, svcs, svcName, deploymentName, "", role, serviceType, ports, "", nil, publishNotReadyAddresses, owner)
	if err != nil {
		return "", false, errors.WithStack(err)
	}
	return svcName, newlyCreated, nil
}

// CreateExternalAccessService prepares and creates a service in k8s, used to access the database/sync from outside k8s cluster.
// If the service already exists, nil is returned.
// If another error occurs, that error is returned.
// The returned bool is true if the service is created, or false when the service already existed.
func CreateExternalAccessService(ctx context.Context, svcs service.ModInterface, svcName, role string,
	deployment metav1.Object, serviceType core.ServiceType, port, nodePort int, loadBalancerIP string,
	loadBalancerSourceRanges []string, owner metav1.OwnerReference) (string, bool, error) {
	deploymentName := deployment.GetName()
	ports := []core.ServicePort{
		core.ServicePort{
			Name:     "server",
			Protocol: core.ProtocolTCP,
			Port:     int32(port),
			NodePort: int32(nodePort),
		},
	}
	publishNotReadyAddresses := false
	newlyCreated, err := createService(ctx, svcs, svcName, deploymentName, "", role, serviceType, ports, loadBalancerIP, loadBalancerSourceRanges, publishNotReadyAddresses, owner)
	if err != nil {
		return "", false, errors.WithStack(err)
	}
	return svcName, newlyCreated, nil
}

// createService prepares and creates a service in k8s.
// If the service already exists, nil is returned.
// If another error occurs, that error is returned.
// The returned bool is true if the service is created, or false when the service already existed.
func createService(ctx context.Context, svcs service.ModInterface, svcName, deploymentName, clusterIP, role string,
	serviceType core.ServiceType, ports []core.ServicePort, loadBalancerIP string, loadBalancerSourceRanges []string,
	publishNotReadyAddresses bool, owner metav1.OwnerReference) (bool, error) {
	labels := LabelsForDeployment(deploymentName, role)
	svc := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        svcName,
			Labels:      labels,
			Annotations: map[string]string{},
		},
		Spec: core.ServiceSpec{
			Type:                     serviceType,
			Ports:                    ports,
			Selector:                 labels,
			ClusterIP:                clusterIP,
			PublishNotReadyAddresses: publishNotReadyAddresses,
			LoadBalancerIP:           loadBalancerIP,
			LoadBalancerSourceRanges: loadBalancerSourceRanges,
		},
	}
	AddOwnerRefToObject(svc.GetObjectMeta(), &owner)
	if _, err := svcs.Create(ctx, svc, metav1.CreateOptions{}); IsAlreadyExists(err) {
		return false, nil
	} else if err != nil {
		return false, errors.WithStack(err)
	}
	return true, nil
}

// CreateServiceURL creates a URL used to reach the given service.
func CreateServiceURL(svc core.Service, scheme string, portPredicate func(core.ServicePort) bool, nodeFetcher func() (core.NodeList, error)) (string, error) {
	var port int32
	var nodePort int32
	portFound := false
	for _, p := range svc.Spec.Ports {
		if portPredicate == nil || portPredicate(p) {
			port = p.Port
			nodePort = p.NodePort
			portFound = true
			break
		}
	}
	if !portFound {
		return "", errors.WithStack(errors.Newf("Cannot find port in service '%s.%s'", svc.GetName(), svc.GetNamespace()))
	}

	var host string
	switch svc.Spec.Type {
	case core.ServiceTypeLoadBalancer:
		for _, x := range svc.Status.LoadBalancer.Ingress {
			if x.IP != "" {
				host = x.IP
				break
			} else if x.Hostname != "" {
				host = x.Hostname
				break
			}
		}
		if host == "" {
			host = svc.Spec.LoadBalancerIP
		}
	case core.ServiceTypeNodePort:
		if nodePort > 0 {
			port = nodePort
		}
		nodeList, err := nodeFetcher()
		if err != nil {
			return "", errors.WithStack(err)
		}
		if len(nodeList.Items) == 0 {
			return "", errors.WithStack(errors.Newf("No nodes found"))
		}
		node := nodeList.Items[rand.Intn(len(nodeList.Items))]
		if len(node.Status.Addresses) > 0 {
			host = node.Status.Addresses[0].Address
		}
	case core.ServiceTypeClusterIP:
		if svc.Spec.ClusterIP != "None" {
			host = svc.Spec.ClusterIP
		}
	default:
		return "", errors.WithStack(errors.Newf("Unknown service type '%s' in service '%s.%s'", svc.Spec.Type, svc.GetName(), svc.GetNamespace()))
	}
	if host == "" {
		return "", errors.WithStack(errors.Newf("Cannot find host for service '%s.%s'", svc.GetName(), svc.GetNamespace()))
	}
	if !strings.HasSuffix(scheme, "://") {
		scheme = scheme + "://"
	}
	return scheme + net.JoinHostPort(host, strconv.Itoa(int(port))), nil
}
