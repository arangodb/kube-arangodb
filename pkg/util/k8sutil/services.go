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

package k8sutil

import (
	"context"
	"math/rand"
	"net"
	"strconv"
	"strings"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	servicev1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/patcher"
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

// CreateAgentLeaderServiceName returns the name of the service used to access a leader agent.
func CreateAgentLeaderServiceName(deploymentName string) string {
	return deploymentName + "-agent"
}

// CreateExporterService
func CreateExporterService(ctx context.Context, cachedStatus inspector.Inspector,
	deployment meta.Object, ports []core.ServicePort, selectors map[string]string, owner meta.OwnerReference) (string, bool, error) {
	deploymentName := deployment.GetName()
	svcName := CreateExporterClientServiceName(deploymentName)

	if svc, exists := cachedStatus.Service().V1().GetSimple(svcName); exists {
		if changed, err := patcher.ServicePatcher(ctx, cachedStatus.ServicesModInterface().V1(), svc, meta.PatchOptions{}, patcher.PatchServiceSelector(selectors), patcher.PatchServicePorts(ports)); err != nil {
			return "", false, err
		} else {
			return svcName, changed, nil
		}
	}

	svc := &core.Service{
		ObjectMeta: meta.ObjectMeta{
			Name:   svcName,
			Labels: selectors,
		},
		Spec: core.ServiceSpec{
			ClusterIP: core.ClusterIPNone,
			Ports:     ports,
			Selector:  selectors,
		},
	}
	AddOwnerRefToObject(svc.GetObjectMeta(), &owner)

	if _, err := cachedStatus.ServicesModInterface().V1().Create(ctx, svc, meta.CreateOptions{}); kerrors.IsAlreadyExists(err) {
		return svcName, false, nil
	} else if err != nil {
		return svcName, false, errors.WithStack(err)
	}
	return svcName, true, nil
}

func ExporterServiceDetails(deploymentName string) ([]core.ServicePort, map[string]string) {
	return []core.ServicePort{
		{
			Name:       shared.ExporterPortName,
			Protocol:   core.ProtocolTCP,
			Port:       shared.ArangoExporterPort,
			TargetPort: intstr.FromString(shared.ExporterPortName),
		},
	}, LabelsForExporterServiceSelector(deploymentName)
}

// CreateHeadlessService prepares and creates a headless service in k8s, used to provide a stable
// DNS name for all pods.
// If the service already exists, nil is returned.
// If another error occurs, that error is returned.
// The returned bool is true if the service is created, or false when the service already existed.
func CreateHeadlessService(ctx context.Context, svcs servicev1.ModInterface, deployment meta.Object,
	ports []core.ServicePort, selectors map[string]string,
	owner meta.OwnerReference) (string, bool, error) {
	deploymentName := deployment.GetName()
	svcName := CreateHeadlessServiceName(deploymentName)

	newlyCreated, err := createService(ctx, svcs, svcName, shared.ClusterIPNone, core.ServiceTypeClusterIP, ports,
		selectors, "", nil, true, owner)
	if err != nil {
		return "", false, errors.WithStack(err)
	}
	return svcName, newlyCreated, nil
}

func HeadlessServiceDetails(deploymentName string) ([]core.ServicePort, map[string]string) {
	ports := []core.ServicePort{
		{
			Name:       shared.ServerPortName,
			Protocol:   core.ProtocolTCP,
			Port:       shared.ArangoPort,
			TargetPort: intstr.FromString(shared.ServerPortName),
		},
	}
	labels := LabelsForDeployment(deploymentName, "")

	return ports, labels
}

// CreateDatabaseClientService prepares and creates a service in k8s, used by database clients within the k8s cluster.
// If the service already exists, nil is returned.
// If another error occurs, that error is returned.
// The returned bool is true if the service is created, or false when the service already existed.
func CreateDatabaseClientService(ctx context.Context, svcs servicev1.ModInterface, deployment meta.Object,
	ports []core.ServicePort, selectors map[string]string, owner meta.OwnerReference) (string, bool, error) {
	deploymentName := deployment.GetName()
	svcName := CreateDatabaseClientServiceName(deploymentName)

	newlyCreated, err := createService(ctx, svcs, svcName, "", core.ServiceTypeClusterIP, ports, selectors, "", nil,
		false, owner)
	if err != nil {
		return "", false, errors.WithStack(err)
	}
	return svcName, newlyCreated, nil
}

func DatabaseClientDetails(deploymentName string, role string, withLeader bool) ([]core.ServicePort, map[string]string) {
	ports := []core.ServicePort{
		{
			Name:       shared.ServerPortName,
			Protocol:   core.ProtocolTCP,
			Port:       shared.ArangoPort,
			TargetPort: intstr.FromString(shared.ServerPortName),
		},
	}

	labels := LabelsForDeployment(deploymentName, role)
	if withLeader {
		labels[LabelKeyArangoLeader] = "true"
	}

	return ports, labels
}

// CreateExternalAccessService prepares and creates a service in k8s, used to access the database/sync from outside k8s cluster.
// If the service already exists, nil is returned.
// If another error occurs, that error is returned.
// The returned bool is true if the service is created, or false when the service already existed.
func CreateExternalAccessService(ctx context.Context, svcs servicev1.ModInterface, svcName string, serviceType core.ServiceType,
	ports []core.ServicePort, selectors map[string]string, loadBalancerIP string,
	loadBalancerSourceRanges []string, owner meta.OwnerReference) (string, bool, error) {

	newlyCreated, err := createService(ctx, svcs, svcName, "", serviceType, ports, selectors, loadBalancerIP,
		loadBalancerSourceRanges, false, owner)
	if err != nil {
		return "", false, errors.WithStack(err)
	}
	return svcName, newlyCreated, nil
}

func ExternalAccessDetails(port, nodePort int, deploymentName, role string, withLeader bool) ([]core.ServicePort, map[string]string) {
	ports := []core.ServicePort{
		{
			Name:       shared.ServerPortName,
			Protocol:   core.ProtocolTCP,
			Port:       int32(port),
			NodePort:   int32(nodePort),
			TargetPort: intstr.FromString(shared.ServerPortName),
		},
	}

	labels := LabelsForDeployment(deploymentName, role)
	if withLeader {
		labels[LabelKeyArangoLeader] = "true"
	}

	return ports, labels
}

// createService prepares and creates a service in k8s.
// If the service already exists, nil is returned.
// If another error occurs, that error is returned.
// The returned bool is true if the service is created, or false when the service already existed.
func createService(ctx context.Context, svcs servicev1.ModInterface, svcName, clusterIP string,
	serviceType core.ServiceType, ports []core.ServicePort, selectors map[string]string, loadBalancerIP string, loadBalancerSourceRanges []string,
	publishNotReadyAddresses bool, owner meta.OwnerReference) (bool, error) {
	svc := &core.Service{
		ObjectMeta: meta.ObjectMeta{
			Name:        svcName,
			Labels:      selectors,
			Annotations: map[string]string{},
		},
		Spec: core.ServiceSpec{
			Type:                     serviceType,
			Ports:                    ports,
			Selector:                 selectors,
			ClusterIP:                clusterIP,
			PublishNotReadyAddresses: publishNotReadyAddresses,
			LoadBalancerIP:           loadBalancerIP,
			LoadBalancerSourceRanges: loadBalancerSourceRanges,
		},
	}
	AddOwnerRefToObject(svc.GetObjectMeta(), &owner)
	if _, err := svcs.Create(ctx, svc, meta.CreateOptions{}); kerrors.IsAlreadyExists(err) {
		return false, nil
	} else if err != nil {
		return false, errors.WithStack(err)
	}
	return true, nil
}

// CreateServiceURL creates a URL used to reach the given service.
func CreateServiceURL(svc core.Service, scheme string, portPredicate func(core.ServicePort) bool, nodeFetcher func() ([]*core.Node, error)) (string, error) {
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
		if len(nodeList) == 0 {
			return "", errors.WithStack(errors.Newf("No nodes found"))
		}
		node := nodeList[rand.Intn(len(nodeList))]
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

func IsServiceRotationRequired(spec api.DeploymentSpec, svc *core.Service) bool {
	if svc == nil {
		return false
	}

	if svc.Spec.ClusterIP == "" {
		return false
	}

	if p := spec.CommunicationMethod.ServiceClusterIP(); p == core.ClusterIPNone {
		if svc.Spec.ClusterIP != core.ClusterIPNone {
			return true
		}
	} else {
		if svc.Spec.ClusterIP == core.ClusterIPNone {
			return true
		}
	}

	return false
}
