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

package resources

import (
	"context"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
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
func (r *Resources) EnsureServices(ctx context.Context, cachedStatus inspectorInterface.Inspector) error {
	log := r.log
	start := time.Now()
	kubecli := r.context.GetKubeCli()
	apiObject := r.context.GetAPIObject()
	status, _ := r.context.GetStatus()
	deploymentName := apiObject.GetName()
	ns := apiObject.GetNamespace()
	owner := apiObject.AsOwner()
	spec := r.context.GetSpec()
	defer metrics.SetDuration(inspectServicesDurationGauges.WithLabelValues(deploymentName), start)
	counterMetric := inspectedServicesCounters.WithLabelValues(deploymentName)

	// Fetch existing services
	svcs := kubecli.CoreV1().Services(ns)

	// Ensure member services
	if err := status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, m := range list {
			memberName := m.ArangoMemberName(r.context.GetAPIObject().GetName(), group)

			member, ok := cachedStatus.ArangoMember(memberName)
			if !ok {
				return errors.Newf("Member %s not found", memberName)
			}

			if s, ok := cachedStatus.Service(member.GetName()); !ok {
				s = &core.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      member.GetName(),
						Namespace: member.GetNamespace(),
						OwnerReferences: []metav1.OwnerReference{
							member.AsOwner(),
						},
					},
					Spec: core.ServiceSpec{
						Type: core.ServiceTypeClusterIP,
						Ports: []core.ServicePort{
							{
								Name:       "server",
								Protocol:   "TCP",
								Port:       k8sutil.ArangoPort,
								TargetPort: intstr.IntOrString{IntVal: k8sutil.ArangoPort},
							},
						},
						PublishNotReadyAddresses: true,
						Selector:                 k8sutil.LabelsForMember(deploymentName, group.AsRole(), m.ID),
					},
				}

				ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
				_, err := svcs.Create(ctxChild, s, metav1.CreateOptions{})
				cancel()
				if err != nil {
					if !k8sutil.IsConflict(err) {
						return err
					}
				}

				return errors.Reconcile()
			} else {
				spec := s.Spec.DeepCopy()

				spec.Type = core.ServiceTypeClusterIP
				spec.Ports = []core.ServicePort{
					{
						Name:       "server",
						Protocol:   "TCP",
						Port:       k8sutil.ArangoPort,
						TargetPort: intstr.IntOrString{IntVal: k8sutil.ArangoPort},
					},
				}
				spec.PublishNotReadyAddresses = true
				spec.Selector = k8sutil.LabelsForMember(deploymentName, group.AsRole(), m.ID)

				if !equality.Semantic.DeepDerivative(*spec, s.Spec) {
					s.Spec = *spec

					ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
					_, err := svcs.Update(ctxChild, s, metav1.UpdateOptions{})
					cancel()
					if err != nil {
						return err
					}

					return errors.Reconcile()
				}
			}
		}

		return nil
	}); err != nil {
		return err
	}

	// Headless service
	counterMetric.Inc()
	if _, exists := cachedStatus.Service(k8sutil.CreateHeadlessServiceName(deploymentName)); !exists {
		ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
		svcName, newlyCreated, err := k8sutil.CreateHeadlessService(ctxChild, svcs, apiObject, owner)
		cancel()
		if err != nil {
			log.Debug().Err(err).Msg("Failed to create headless service")
			return errors.WithStack(err)
		}
		if newlyCreated {
			log.Debug().Str("service", svcName).Msg("Created headless service")
		}
	}

	// Internal database client service
	single := spec.GetMode().HasSingleServers()
	counterMetric.Inc()
	if _, exists := cachedStatus.Service(k8sutil.CreateDatabaseClientServiceName(deploymentName)); !exists {
		ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
		svcName, newlyCreated, err := k8sutil.CreateDatabaseClientService(ctxChild, svcs, apiObject, single, owner)
		cancel()
		if err != nil {
			log.Debug().Err(err).Msg("Failed to create database client service")
			return errors.WithStack(err)
		}
		if newlyCreated {
			log.Debug().Str("service", svcName).Msg("Created database client service")
		}
		{
			status, lastVersion := r.context.GetStatus()
			if status.ServiceName != svcName {
				status.ServiceName = svcName
				if err := r.context.UpdateStatus(ctx, status, lastVersion); err != nil {
					return errors.WithStack(err)
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
	if err := r.ensureExternalAccessServices(ctx, cachedStatus, svcs, eaServiceName, ns, role, "database", k8sutil.ArangoPort, false, spec.ExternalAccess, apiObject, log, counterMetric); err != nil {
		return errors.WithStack(err)
	}

	if spec.Sync.IsEnabled() {
		// External (and internal) Sync master service
		counterMetric.Inc()
		eaServiceName := k8sutil.CreateSyncMasterClientServiceName(deploymentName)
		role := "syncmaster"
		if err := r.ensureExternalAccessServices(ctx, cachedStatus, svcs, eaServiceName, ns, role, "sync", k8sutil.ArangoSyncMasterPort, true, spec.Sync.ExternalAccess.ExternalAccessSpec, apiObject, log, counterMetric); err != nil {
			return errors.WithStack(err)
		}
		status, lastVersion := r.context.GetStatus()
		if status.SyncServiceName != eaServiceName {
			status.SyncServiceName = eaServiceName
			if err := r.context.UpdateStatus(ctx, status, lastVersion); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	if spec.Metrics.IsEnabled() {
		ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
		name, _, err := k8sutil.CreateExporterService(ctxChild, cachedStatus, svcs, apiObject, apiObject.AsOwner())
		cancel()
		if err != nil {
			log.Debug().Err(err).Msgf("Failed to create %s exporter service", name)
			return errors.WithStack(err)
		}
		status, lastVersion := r.context.GetStatus()
		if status.ExporterServiceName != name {
			status.ExporterServiceName = name
			if err := r.context.UpdateStatus(ctx, status, lastVersion); err != nil {
				return errors.WithStack(err)
			}
		}
	}
	return nil
}

// EnsureServices creates all services needed to service the deployment
func (r *Resources) ensureExternalAccessServices(ctx context.Context, cachedStatus inspectorInterface.Inspector, svcs k8sutil.ServiceInterface, eaServiceName, ns, svcRole, title string, port int, noneIsClusterIP bool, spec api.ExternalAccessSpec, apiObject k8sutil.APIObject, log zerolog.Logger, counterMetric prometheus.Counter) error {
	// Database external access service
	createExternalAccessService := false
	deleteExternalAccessService := false
	eaServiceType := spec.GetType().AsServiceType() // Note: Type auto defaults to ServiceTypeLoadBalancer
	if existing, exists := cachedStatus.Service(eaServiceName); exists {
		// External access service exists
		updateExternalAccessService := false
		loadBalancerIP := spec.GetLoadBalancerIP()
		loadBalancerSourceRanges := spec.LoadBalancerSourceRanges
		nodePort := spec.GetNodePort()
		if spec.GetType().IsNone() {
			if noneIsClusterIP {
				eaServiceType = core.ServiceTypeClusterIP
				if existing.Spec.Type != core.ServiceTypeClusterIP {
					deleteExternalAccessService = true // Remove the current and replace with proper one
					createExternalAccessService = true
				}
			} else {
				// Should not be there, remove it
				deleteExternalAccessService = true
			}
		} else if spec.GetType().IsAuto() {
			// Inspect existing service.
			if existing.Spec.Type == core.ServiceTypeLoadBalancer {
				// See if LoadBalancer has been configured & the service is "old enough"
				oldEnoughTimestamp := time.Now().Add(-1 * time.Minute) // How long does the load-balancer provisioner have to act.
				if len(existing.Status.LoadBalancer.Ingress) == 0 && existing.GetObjectMeta().GetCreationTimestamp().Time.Before(oldEnoughTimestamp) {
					log.Info().Str("service", eaServiceName).Msgf("LoadBalancerIP of %s external access service is not set, switching to NodePort", title)
					createExternalAccessService = true
					eaServiceType = core.ServiceTypeNodePort
					deleteExternalAccessService = true // Remove the LoadBalancer ex service, then add the NodePort one
				} else if existing.Spec.Type == core.ServiceTypeLoadBalancer && (loadBalancerIP != "" && existing.Spec.LoadBalancerIP != loadBalancerIP) {
					deleteExternalAccessService = true // LoadBalancerIP is wrong, remove the current and replace with proper one
					createExternalAccessService = true
				} else if existing.Spec.Type == core.ServiceTypeNodePort && len(existing.Spec.Ports) == 1 && (nodePort != 0 && existing.Spec.Ports[0].NodePort != int32(nodePort)) {
					deleteExternalAccessService = true // NodePort is wrong, remove the current and replace with proper one
					createExternalAccessService = true
				}
			}
		} else if spec.GetType().IsLoadBalancer() {
			if existing.Spec.Type != core.ServiceTypeLoadBalancer || (loadBalancerIP != "" && existing.Spec.LoadBalancerIP != loadBalancerIP) {
				deleteExternalAccessService = true // Remove the current and replace with proper one
				createExternalAccessService = true
			}
			if strings.Join(existing.Spec.LoadBalancerSourceRanges, ",") != strings.Join(loadBalancerSourceRanges, ",") {
				updateExternalAccessService = true
				existing.Spec.LoadBalancerSourceRanges = loadBalancerSourceRanges
			}
		} else if spec.GetType().IsNodePort() {
			if existing.Spec.Type != core.ServiceTypeNodePort || len(existing.Spec.Ports) != 1 || (nodePort != 0 && existing.Spec.Ports[0].NodePort != int32(nodePort)) {
				deleteExternalAccessService = true // Remove the current and replace with proper one
				createExternalAccessService = true
			}
		}
		if updateExternalAccessService && !createExternalAccessService && !deleteExternalAccessService {
			ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
			_, err := svcs.Update(ctxChild, existing, metav1.UpdateOptions{})
			cancel()
			if err != nil {
				log.Debug().Err(err).Msgf("Failed to update %s external access service", title)
				return errors.WithStack(err)
			}
		}
	} else {
		// External access service does not exist
		if !spec.GetType().IsNone() || noneIsClusterIP {
			createExternalAccessService = true
		}
	}

	if deleteExternalAccessService {
		log.Info().Str("service", eaServiceName).Msgf("Removing obsolete %s external access service", title)
		ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
		err := svcs.Delete(ctxChild, eaServiceName, metav1.DeleteOptions{})
		cancel()
		if err != nil {
			log.Debug().Err(err).Msgf("Failed to remove %s external access service", title)
			return errors.WithStack(err)
		}
	}
	if createExternalAccessService {
		// Let's create or update the database external access service
		nodePort := spec.GetNodePort()
		loadBalancerIP := spec.GetLoadBalancerIP()
		loadBalancerSourceRanges := spec.LoadBalancerSourceRanges
		ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
		_, newlyCreated, err := k8sutil.CreateExternalAccessService(ctxChild, svcs, eaServiceName, svcRole, apiObject, eaServiceType, port, nodePort, loadBalancerIP, loadBalancerSourceRanges, apiObject.AsOwner())
		cancel()
		if err != nil {
			log.Debug().Err(err).Msgf("Failed to create %s external access service", title)
			return errors.WithStack(err)
		}
		if newlyCreated {
			log.Debug().Str("service", eaServiceName).Msgf("Created %s external access service", title)
		}
	}
	return nil
}
