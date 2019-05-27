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

package resources

import (
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

var (
	inspectedServicesCounters     = metrics.MustRegisterCounterVec(metricsComponent, "inspected_services", "Number of Service inspections per deployment", metrics.DeploymentName)
	inspectServicesDurationGauges = metrics.MustRegisterGaugeVec(metricsComponent, "inspect_services_duration", "Amount of time taken by a single inspection of all Services for a deployment (in sec)", metrics.DeploymentName)
)

// EnsureServices creates all services needed to service the deployment
func (r *Resources) EnsureServices() error {
	log := r.log
	start := time.Now()
	kubecli := r.context.GetKubeCli()
	apiObject := r.context.GetAPIObject()
	deploymentName := apiObject.GetName()
	ns := apiObject.GetNamespace()
	owner := apiObject.AsOwner()
	spec := r.context.GetSpec()
	defer metrics.SetDuration(inspectServicesDurationGauges.WithLabelValues(deploymentName), start)
	counterMetric := inspectedServicesCounters.WithLabelValues(deploymentName)

	// Fetch existing services
	svcs := k8sutil.NewServiceCache(kubecli.CoreV1().Services(ns))
	// Headless service
	counterMetric.Inc()
	if _, err := svcs.Get(k8sutil.CreateHeadlessServiceName(deploymentName), metav1.GetOptions{}); err != nil {
		svcName, newlyCreated, err := k8sutil.CreateHeadlessService(svcs, apiObject, owner)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to create headless service")
			return maskAny(err)
		}
		if newlyCreated {
			log.Debug().Str("service", svcName).Msg("Created headless service")
		}
	}

	// Internal database client service
	single := spec.GetMode().HasSingleServers()
	counterMetric.Inc()
	if _, err := svcs.Get(k8sutil.CreateDatabaseClientServiceName(deploymentName), metav1.GetOptions{}); err != nil {
		svcName, newlyCreated, err := k8sutil.CreateDatabaseClientService(svcs, apiObject, single, owner)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to create database client service")
			return maskAny(err)
		}
		if newlyCreated {
			log.Debug().Str("service", svcName).Msg("Created database client service")
		}
		{
			status, lastVersion := r.context.GetStatus()
			if status.ServiceName != svcName {
				status.ServiceName = svcName
				if err := r.context.UpdateStatus(status, lastVersion); err != nil {
					return maskAny(err)
				}
			}
		}
	}

	// Database external access service
	eaServiceName := k8sutil.CreateDatabaseExternalAccessServiceName(deploymentName)
	role := "coordinator"
	if single {
		role = "single"
	}
	if err := r.ensureExternalAccessServices(svcs, eaServiceName, ns, role, "database", k8sutil.ArangoPort, false, spec.ExternalAccess, apiObject, log, counterMetric); err != nil {
		return maskAny(err)
	}

	if spec.Sync.IsEnabled() {
		// External (and internal) Sync master service
		counterMetric.Inc()
		eaServiceName := k8sutil.CreateSyncMasterClientServiceName(deploymentName)
		role := "syncmaster"
		if err := r.ensureExternalAccessServices(svcs, eaServiceName, ns, role, "sync", k8sutil.ArangoSyncMasterPort, true, spec.Sync.ExternalAccess.ExternalAccessSpec, apiObject, log, counterMetric); err != nil {
			return maskAny(err)
		}
		status, lastVersion := r.context.GetStatus()
		if status.SyncServiceName != eaServiceName {
			status.SyncServiceName = eaServiceName
			if err := r.context.UpdateStatus(status, lastVersion); err != nil {
				return maskAny(err)
			}
		}
	}

	if spec.Metrics.IsEnabled() {
		name, _, err := k8sutil.CreateExporterService(svcs, apiObject, apiObject.AsOwner())
		if err != nil {
			log.Debug().Err(err).Msgf("Failed to create %s exporter service", name)
			return maskAny(err)
		}
		status, lastVersion := r.context.GetStatus()
		if status.ExporterServiceName != name {
			status.ExporterServiceName = name
			if err := r.context.UpdateStatus(status, lastVersion); err != nil {
				return maskAny(err)
			}
		}
	}
	return nil
}

// EnsureServices creates all services needed to service the deployment
func (r *Resources) ensureExternalAccessServices(svcs k8sutil.ServiceInterface, eaServiceName, ns, svcRole, title string, port int, noneIsClusterIP bool, spec api.ExternalAccessSpec, apiObject k8sutil.APIObject, log zerolog.Logger, counterMetric prometheus.Counter) error {
	// Database external access service
	createExternalAccessService := false
	deleteExternalAccessService := false
	eaServiceType := spec.GetType().AsServiceType() // Note: Type auto defaults to ServiceTypeLoadBalancer
	if existing, err := svcs.Get(eaServiceName, metav1.GetOptions{}); err == nil {
		// External access service exists
		loadBalancerIP := spec.GetLoadBalancerIP()
		nodePort := spec.GetNodePort()
		if spec.GetType().IsNone() {
			if noneIsClusterIP {
				eaServiceType = v1.ServiceTypeClusterIP
				if existing.Spec.Type != v1.ServiceTypeClusterIP {
					deleteExternalAccessService = true // Remove the current and replace with proper one
					createExternalAccessService = true
				}
			} else {
				// Should not be there, remove it
				deleteExternalAccessService = true
			}
		} else if spec.GetType().IsAuto() {
			// Inspect existing service.
			if existing.Spec.Type == v1.ServiceTypeLoadBalancer {
				// See if LoadBalancer has been configured & the service is "old enough"
				oldEnoughTimestamp := time.Now().Add(-1 * time.Minute) // How long does the load-balancer provisioner have to act.
				if len(existing.Status.LoadBalancer.Ingress) == 0 && existing.GetObjectMeta().GetCreationTimestamp().Time.Before(oldEnoughTimestamp) {
					log.Info().Str("service", eaServiceName).Msgf("LoadBalancerIP of %s external access service is not set, switching to NodePort", title)
					createExternalAccessService = true
					eaServiceType = v1.ServiceTypeNodePort
					deleteExternalAccessService = true // Remove the LoadBalancer ex service, then add the NodePort one
				} else if existing.Spec.Type == v1.ServiceTypeLoadBalancer && (loadBalancerIP != "" && existing.Spec.LoadBalancerIP != loadBalancerIP) {
					deleteExternalAccessService = true // LoadBalancerIP is wrong, remove the current and replace with proper one
					createExternalAccessService = true
				} else if existing.Spec.Type == v1.ServiceTypeNodePort && len(existing.Spec.Ports) == 1 && (nodePort != 0 && existing.Spec.Ports[0].NodePort != int32(nodePort)) {
					deleteExternalAccessService = true // NodePort is wrong, remove the current and replace with proper one
					createExternalAccessService = true
				}
			}
		} else if spec.GetType().IsLoadBalancer() {
			if existing.Spec.Type != v1.ServiceTypeLoadBalancer || (loadBalancerIP != "" && existing.Spec.LoadBalancerIP != loadBalancerIP) {
				deleteExternalAccessService = true // Remove the current and replace with proper one
				createExternalAccessService = true
			}
		} else if spec.GetType().IsNodePort() {
			if existing.Spec.Type != v1.ServiceTypeNodePort || len(existing.Spec.Ports) != 1 || (nodePort != 0 && existing.Spec.Ports[0].NodePort != int32(nodePort)) {
				deleteExternalAccessService = true // Remove the current and replace with proper one
				createExternalAccessService = true
			}
		}
	} else if k8sutil.IsNotFound(err) {
		// External access service does not exist
		if !spec.GetType().IsNone() || noneIsClusterIP {
			createExternalAccessService = true
		}
	}
	if deleteExternalAccessService {
		log.Info().Str("service", eaServiceName).Msgf("Removing obsolete %s external access service", title)
		if err := svcs.Delete(eaServiceName, &metav1.DeleteOptions{}); err != nil {
			log.Debug().Err(err).Msgf("Failed to remove %s external access service", title)
			return maskAny(err)
		}
	}
	if createExternalAccessService {
		// Let's create or update the database external access service
		nodePort := spec.GetNodePort()
		loadBalancerIP := spec.GetLoadBalancerIP()
		loadBalancerSourceRanges := spec.LoadBalancerSourceRanges
		_, newlyCreated, err := k8sutil.CreateExternalAccessService(svcs, eaServiceName, svcRole, apiObject, eaServiceType, port, nodePort, loadBalancerIP, loadBalancerSourceRanges, apiObject.AsOwner())
		if err != nil {
			log.Debug().Err(err).Msgf("Failed to create %s external access service", title)
			return maskAny(err)
		}
		if newlyCreated {
			log.Debug().Str("service", eaServiceName).Msgf("Created %s external access service", title)
		}
	}
	return nil
}
