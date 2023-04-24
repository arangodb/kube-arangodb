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

package resources

import (
	"context"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	v1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangomember/v1"
	servicev1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/patcher"
)

var (
	inspectedServicesCounters     = metrics.MustRegisterCounterVec(metricsComponent, "inspected_services", "Number of Service inspections per deployment", metrics.DeploymentName)
	inspectServicesDurationGauges = metrics.MustRegisterGaugeVec(metricsComponent, "inspect_services_duration", "Amount of time taken by a single inspection of all Services for a deployment (in sec)", metrics.DeploymentName)
)

// createService returns service's object.
func (r *Resources) createService(name, namespace, clusterIP string, serviceType core.ServiceType, owner meta.OwnerReference, ports []core.ServicePort,
	selector map[string]string) *core.Service {

	return &core.Service{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			OwnerReferences: []meta.OwnerReference{
				owner,
			},
		},
		Spec: core.ServiceSpec{
			Type:                     serviceType,
			ClusterIP:                clusterIP,
			Ports:                    ports,
			PublishNotReadyAddresses: true,
			Selector:                 selector,
		},
	}
}

// EnsureServices creates all services needed to service the deployment
func (r *Resources) EnsureServices(ctx context.Context, cachedStatus inspectorInterface.Inspector) error {
	log := r.log.Str("section", "service")
	start := time.Now()
	apiObject := r.context.GetAPIObject()
	status := r.context.GetStatus()
	deploymentName := apiObject.GetName()
	owner := apiObject.AsOwner()
	spec := r.context.GetSpec()
	role := spec.Mode.Get().ServingGroup().AsRole()
	defer metrics.SetDuration(inspectServicesDurationGauges.WithLabelValues(deploymentName), start)
	counterMetric := inspectedServicesCounters.WithLabelValues(deploymentName)

	// Fetch existing services
	svcs := cachedStatus.ServicesModInterface().V1()
	amInspector := cachedStatus.ArangoMember().V1()

	reconcileRequired := k8sutil.NewReconcile(cachedStatus)

	// Ensure member services
	for _, e := range status.Members.AsList() {
		memberName := e.Member.ArangoMemberName(r.context.GetAPIObject().GetName(), e.Group)

		member, ok := cachedStatus.ArangoMember().V1().GetSimple(memberName)
		if !ok {
			return errors.Newf("Member %s not found", memberName)
		}

		ports := CreateServerServicePortsWithSidecars(amInspector, e.Member.ArangoMemberName(deploymentName, e.Group))
		selector := k8sutil.LabelsForActiveMember(deploymentName, e.Group.AsRole(), e.Member.ID)
		if s, ok := cachedStatus.Service().V1().GetSimple(member.GetName()); !ok {
			s := r.createService(member.GetName(), member.GetNamespace(), spec.CommunicationMethod.ServiceClusterIP(), spec.CommunicationMethod.ServiceType(), member.AsOwner(), ports, selector)

			err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
				_, err := svcs.Create(ctxChild, s, meta.CreateOptions{})
				return err
			})
			if err != nil {
				if !kerrors.IsConflict(err) {
					return err
				}
			}

			reconcileRequired.Required()
			continue
		} else {

			if changed, err := patcher.ServicePatcher(ctx, svcs, s, meta.PatchOptions{},
				patcher.PatchServicePorts(ports),
				patcher.PatchServiceSelector(selector),
				patcher.PatchServicePublishNotReadyAddresses(true),
				patcher.PatchServiceType(spec.CommunicationMethod.ServiceType())); err != nil {
				return err
			} else if changed {
				reconcileRequired.Required()
			}
		}
	}

	// Headless service
	counterMetric.Inc()
	headlessPorts, headlessSelector := k8sutil.HeadlessServiceDetails(deploymentName)

	if s, exists := cachedStatus.Service().V1().GetSimple(k8sutil.CreateHeadlessServiceName(deploymentName)); !exists {
		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()
		svcName, newlyCreated, err := k8sutil.CreateHeadlessService(ctxChild, svcs, apiObject, headlessPorts, headlessSelector, owner)
		if err != nil {
			log.Err(err).Debug("Failed to create headless service")
			return errors.WithStack(err)
		}
		if newlyCreated {
			log.Str("service", svcName).Debug("Created headless service")
		}
	} else {
		if changed, err := patcher.ServicePatcher(ctx, svcs, s, meta.PatchOptions{}, patcher.PatchServicePorts(headlessPorts), patcher.PatchServiceSelector(headlessSelector)); err != nil {
			log.Err(err).Debug("Failed to patch headless service")
			return errors.WithStack(err)
		} else if changed {
			log.Str("service", s.GetName()).Debug("Updated headless service")
		}
	}

	// Internal database client service
	var single, withLeader bool
	if single = spec.GetMode().HasSingleServers(); single {
		if spec.GetMode() == api.DeploymentModeActiveFailover && features.FailoverLeadership().Enabled() {
			withLeader = true
		}
	}
	counterMetric.Inc()

	clientServicePorts, clientServiceSelectors := k8sutil.DatabaseClientDetails(deploymentName, role, withLeader)

	if s, exists := cachedStatus.Service().V1().GetSimple(k8sutil.CreateDatabaseClientServiceName(deploymentName)); !exists {
		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()
		svcName, newlyCreated, err := k8sutil.CreateDatabaseClientService(ctxChild, svcs, apiObject, clientServicePorts, clientServiceSelectors, owner)
		if err != nil {
			log.Err(err).Debug("Failed to create database client service")
			return errors.WithStack(err)
		}
		if newlyCreated {
			log.Str("service", svcName).Debug("Created database client service")
		}
		{
			status := r.context.GetStatus()
			if status.ServiceName != svcName {
				status.ServiceName = svcName
				if err := r.context.UpdateStatus(ctx, status); err != nil {
					return errors.WithStack(err)
				}
			}
		}
	} else {
		if changed, err := patcher.ServicePatcher(ctx, svcs, s, meta.PatchOptions{}, patcher.PatchServiceOnlyPorts(clientServicePorts...), patcher.PatchServiceSelector(clientServiceSelectors)); err != nil {
			log.Err(err).Debug("Failed to patch database client service")
			return errors.WithStack(err)
		} else if changed {
			log.Str("service", s.GetName()).Debug("Updated database client service")
		}
	}

	// Database external access service
	eaServiceName := k8sutil.CreateDatabaseExternalAccessServiceName(deploymentName)
	if err := r.ensureExternalAccessServices(ctx, cachedStatus, svcs, eaServiceName, role, shared.ArangoPort,
		false, withLeader, spec.ExternalAccess, apiObject); err != nil {
		return errors.WithStack(err)
	}

	if r.context.IsSyncEnabled() {
		// External (and internal) Sync master service
		counterMetric.Inc()
		eaServiceName := k8sutil.CreateSyncMasterClientServiceName(deploymentName)
		if err := r.ensureExternalAccessServices(ctx, cachedStatus, svcs, eaServiceName, api.ServerGroupSyncMastersString,
			shared.ArangoSyncMasterPort, true, false, spec.Sync.ExternalAccess.ExternalAccessSpec, apiObject); err != nil {
			return errors.WithStack(err)
		}
		status := r.context.GetStatus()
		if status.SyncServiceName != eaServiceName {
			status.SyncServiceName = eaServiceName
			if err := r.context.UpdateStatus(ctx, status); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	if spec.Metrics.IsEnabled() {
		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()

		ports, selectors := k8sutil.ExporterServiceDetails(deploymentName)

		name, _, err := k8sutil.CreateExporterService(ctxChild, cachedStatus, apiObject, ports, selectors, apiObject.AsOwner())
		if err != nil {
			log.Err(err).Debug("Failed to create %s exporter service", name)
			return errors.WithStack(err)
		}
		status := r.context.GetStatus()
		if status.ExporterServiceName != name {
			status.ExporterServiceName = name
			if err := r.context.UpdateStatus(ctx, status); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return reconcileRequired.Reconcile(ctx)
}

// ensureExternalAccessServices ensures all services needed for a deployment.
func (r *Resources) ensureExternalAccessServices(ctx context.Context, cachedStatus inspectorInterface.Inspector,
	svcs servicev1.ModInterface, eaServiceName, role string, port int, noneIsClusterIP bool, withLeader bool,
	spec api.ExternalAccessSpec, apiObject k8sutil.APIObject) error {

	eaPorts, eaSelector := k8sutil.ExternalAccessDetails(port, spec.GetNodePort(), apiObject.GetName(), role, withLeader)

	if spec.GetType().IsManaged() {
		// Managed services should not be created or removed by the operator.
		return r.ensureExternalAccessManagedServices(ctx, cachedStatus, eaServiceName, eaSelector, spec)
	}

	log := r.log.Str("section", "service-ea").Str("role", role).Str("service", eaServiceName)
	createExternalAccessService := false
	deleteExternalAccessService := false
	owned := false
	eaServiceType := spec.GetType().AsServiceType() // Note: Type auto defaults to ServiceTypeLoadBalancer
	if existing, exists := cachedStatus.Service().V1().GetSimple(eaServiceName); exists {
		// External access service exists
		owned = apiObject.OwnerOf(existing)

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
					log.Info("LoadBalancerIP of external access service is not set, switching to NodePort")
					createExternalAccessService = true
					eaServiceType = core.ServiceTypeNodePort
					deleteExternalAccessService = true // Remove the LoadBalancer ex service, then add the NodePort one
				} else if existing.Spec.Type == core.ServiceTypeLoadBalancer && (loadBalancerIP != "" && existing.Spec.LoadBalancerIP != loadBalancerIP) {
					deleteExternalAccessService = true // LoadBalancerIP is wrong, remove the current and replace with proper one
					createExternalAccessService = true
				} else if existing.Spec.Type == core.ServiceTypeNodePort && len(existing.Spec.Ports) < 1 || existing.Spec.Ports[0].Name != shared.ServerPortName && (nodePort != 0 && existing.Spec.Ports[0].NodePort != int32(nodePort)) {
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
			if existing.Spec.Type != core.ServiceTypeNodePort || len(existing.Spec.Ports) < 1 || existing.Spec.Ports[0].Name != shared.ServerPortName || (nodePort != 0 && existing.Spec.Ports[0].NodePort != int32(nodePort)) {
				deleteExternalAccessService = true // Remove the current and replace with proper one
				createExternalAccessService = true
			}
		}
		if updateExternalAccessService && !createExternalAccessService && !deleteExternalAccessService {
			err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
				_, err := svcs.Update(ctxChild, existing, meta.UpdateOptions{})
				return err
			})
			if err != nil {
				log.Err(err).Debug("Failed to update external access service")
				return errors.WithStack(err)
			}
		}
		if !createExternalAccessService && !deleteExternalAccessService {
			if changed, err := patcher.ServicePatcher(ctx, svcs, existing, meta.PatchOptions{},
				patcher.PatchServiceSelector(eaSelector),
				patcher.Optional(patcher.PatchServiceOnlyPorts(eaPorts...), owned)); err != nil {
				log.Err(err).Debug("Failed to patch database client service")
				return errors.WithStack(err)
			} else if changed {
				log.Str("service", existing.GetName()).Debug("Updated database client service")
			}
		}
	} else {
		// External access service does not exist
		if !spec.GetType().IsNone() || noneIsClusterIP {
			createExternalAccessService = true
		}
	}

	if deleteExternalAccessService {
		if owned {
			log.Info("Removing obsolete external access service")
			err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
				return svcs.Delete(ctxChild, eaServiceName, meta.DeleteOptions{})
			})
			if err != nil {
				log.Err(err).Debug("Failed to remove external access service")
				return errors.WithStack(err)
			}
		}
	}
	if createExternalAccessService {
		// Let's create or update the database external access service
		loadBalancerIP := spec.GetLoadBalancerIP()
		loadBalancerSourceRanges := spec.LoadBalancerSourceRanges
		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()
		_, newlyCreated, err := k8sutil.CreateExternalAccessService(ctxChild, svcs, eaServiceName, eaServiceType, eaPorts, eaSelector, loadBalancerIP, loadBalancerSourceRanges, apiObject.AsOwner())
		if err != nil {
			log.Err(err).Debug("Failed to create external access service")
			return errors.WithStack(err)
		}
		if newlyCreated {
			log.Debug("Created %s external access service")
		}
	}
	return nil
}

// ensureExternalAccessServices ensures if there are correct selectors on a managed services.
// If hardcoded external service names are not on the list of managed services then it will be checked additionally.
func (r *Resources) ensureExternalAccessManagedServices(ctx context.Context, cachedStatus inspectorInterface.Inspector, eaServiceName string,
	selectors map[string]string, spec api.ExternalAccessSpec) error {

	log := r.log.Str("section", "service-ea").Str("service", eaServiceName)
	managedServiceNames := spec.GetManagedServiceNames()

	apply := func(svc *core.Service) (bool, error) {
		return patcher.ServicePatcher(ctx, cachedStatus.ServicesModInterface().V1(), svc, meta.PatchOptions{},
			patcher.PatchServiceSelector(selectors))
	}

	// Check if hardcoded service has correct selector.
	if svc, ok := cachedStatus.Service().V1().GetSimple(eaServiceName); !ok {
		// Hardcoded service (e.g. <deplname>-ea or <deplname>-sync) is not mandatory in `managed` type.
		if len(managedServiceNames) == 0 {
			log.Warn("the field \"spec.externalAccess.managedServiceNames\" should be provided for \"managed\" service type")
			return nil
		}
	} else if changed, err := apply(svc); err != nil {
		return errors.WithMessage(err, "failed to ensure service selector")
	} else if changed {
		log.Info("selector applied to the managed service \"%s\"", svc.GetName())
	}

	for _, svcName := range managedServiceNames {
		if svcName == eaServiceName {
			// Hardcoded service has been applied before this loop.
			continue
		}

		svc, ok := cachedStatus.Service().V1().GetSimple(svcName)
		if !ok {
			log.Warn("managed service \"%s\" should have existed", svcName)
			continue
		}

		if changed, err := apply(svc); err != nil {
			return errors.WithMessage(err, "failed to ensure service selector")
		} else if changed {
			log.Info("selector applied to the managed service \"%s\"", svcName)
		}
	}

	return nil
}

// CreateServerServicePortsWithSidecars returns ports for the service.
func CreateServerServicePortsWithSidecars(amInspector v1.Inspector, am string) []core.ServicePort {
	// Create service port for the `server` container.
	ports := []core.ServicePort{CreateServerServicePort()}

	if amInspector == nil {
		return ports
	}

	if am, ok := amInspector.GetSimple(am); ok {
		if t := am.Status.Template; t != nil {
			if p := t.PodSpec; p != nil {
				for _, c := range p.Spec.Containers {
					if c.Name == api.ServerGroupReservedContainerNameServer {
						// It is already added.
						continue
					}
					for _, port := range c.Ports {
						ports = append(ports, core.ServicePort{
							Name:       port.Name,
							Protocol:   core.ProtocolTCP,
							Port:       port.ContainerPort,
							TargetPort: intstr.FromString(port.Name),
						})
					}
				}
			}
		}
	}

	return ports
}

// CreateServerServicePort creates main server service port.
func CreateServerServicePort() core.ServicePort {
	return core.ServicePort{
		Name:       shared.ServerPortName,
		Protocol:   core.ProtocolTCP,
		Port:       shared.ArangoPort,
		TargetPort: intstr.FromString(shared.ServerPortName),
	}
}
