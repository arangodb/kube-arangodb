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

package resources

import (
	"context"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	servicev1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service/v1"
)

var (
	inspectedServicesCounters     = metrics.MustRegisterCounterVec(metricsComponent, "inspected_services", "Number of Service inspections per deployment", metrics.DeploymentName)
	inspectServicesDurationGauges = metrics.MustRegisterGaugeVec(metricsComponent, "inspect_services_duration", "Amount of time taken by a single inspection of all Services for a deployment (in sec)", metrics.DeploymentName)
)

// createService returns service's object.
func (r *Resources) createService(name, namespace string, owner meta.OwnerReference, targetPort int32,
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
			Type: core.ServiceTypeClusterIP,
			Ports: []core.ServicePort{
				{
					Name:       "server",
					Protocol:   "TCP",
					Port:       shared.ArangoPort,
					TargetPort: intstr.IntOrString{IntVal: targetPort},
				},
			},
			PublishNotReadyAddresses: true,
			Selector:                 selector,
		},
	}
}

// adjustService checks whether service contains is valid and if not than it reconciles service.
// Returns true if service is adjusted.
func (r *Resources) adjustService(ctx context.Context, s *core.Service, targetPort int32, selector map[string]string) (error, bool) {
	services := r.context.ACS().CurrentClusterCache().ServicesModInterface().V1()
	spec := s.Spec.DeepCopy()

	spec.Type = core.ServiceTypeClusterIP
	spec.Ports = []core.ServicePort{
		{
			Name:       "server",
			Protocol:   "TCP",
			Port:       shared.ArangoPort,
			TargetPort: intstr.IntOrString{IntVal: targetPort},
		},
	}
	spec.PublishNotReadyAddresses = true
	spec.Selector = selector
	if equality.Semantic.DeepDerivative(*spec, s.Spec) {
		// The service has not changed, so nothing should be changed.
		return nil, false
	}

	s.Spec = *spec
	err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		_, err := services.Update(ctxChild, s, meta.UpdateOptions{})
		return err
	})
	if err != nil {
		return err, false
	}

	// The service has been changed.
	return nil, true

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
	defer metrics.SetDuration(inspectServicesDurationGauges.WithLabelValues(deploymentName), start)
	counterMetric := inspectedServicesCounters.WithLabelValues(deploymentName)

	// Fetch existing services
	svcs := cachedStatus.ServicesModInterface().V1()

	reconcileRequired := k8sutil.NewReconcile(cachedStatus)

	// Ensure member services
	for _, e := range status.Members.AsList() {
		var targetPort int32 = shared.ArangoPort

		switch e.Group {
		case api.ServerGroupSyncMasters:
			targetPort = shared.ArangoSyncMasterPort
		case api.ServerGroupSyncWorkers:
			targetPort = shared.ArangoSyncWorkerPort
		}

		memberName := e.Member.ArangoMemberName(r.context.GetAPIObject().GetName(), e.Group)

		member, ok := cachedStatus.ArangoMember().V1().GetSimple(memberName)
		if !ok {
			return errors.Newf("Member %s not found", memberName)
		}

		selector := k8sutil.LabelsForMember(deploymentName, e.Group.AsRole(), e.Member.ID)
		if s, ok := cachedStatus.Service().V1().GetSimple(member.GetName()); !ok {
			s := r.createService(member.GetName(), member.GetNamespace(), member.AsOwner(), targetPort, selector)

			err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
				_, err := svcs.Create(ctxChild, s, meta.CreateOptions{})
				return err
			})
			if err != nil {
				if !k8sutil.IsConflict(err) {
					return err
				}
			}

			reconcileRequired.Required()
			continue
		} else {
			if err, adjusted := r.adjustService(ctx, s, targetPort, selector); err == nil {
				if adjusted {
					reconcileRequired.Required()
				}
				// Continue the loop.
			} else {
				return err
			}
		}
	}

	// Headless service
	counterMetric.Inc()
	if _, exists := cachedStatus.Service().V1().GetSimple(k8sutil.CreateHeadlessServiceName(deploymentName)); !exists {
		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()
		svcName, newlyCreated, err := k8sutil.CreateHeadlessService(ctxChild, svcs, apiObject, owner)
		if err != nil {
			log.Err(err).Debug("Failed to create headless service")
			return errors.WithStack(err)
		}
		if newlyCreated {
			log.Str("service", svcName).Debug("Created headless service")
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
	if _, exists := cachedStatus.Service().V1().GetSimple(k8sutil.CreateDatabaseClientServiceName(deploymentName)); !exists {
		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()
		svcName, newlyCreated, err := k8sutil.CreateDatabaseClientService(ctxChild, svcs, apiObject, single, withLeader, owner)
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
	}

	// Database external access service
	eaServiceName := k8sutil.CreateDatabaseExternalAccessServiceName(deploymentName)
	role := "coordinator"
	if single {
		role = "single"
	}
	if err := r.ensureExternalAccessServices(ctx, cachedStatus, svcs, eaServiceName, role, shared.ArangoPort,
		false, withLeader, spec.ExternalAccess, apiObject); err != nil {
		return errors.WithStack(err)
	}

	if spec.Sync.IsEnabled() {
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
		name, _, err := k8sutil.CreateExporterService(ctxChild, cachedStatus, svcs, apiObject, apiObject.AsOwner())
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
	svcs servicev1.ModInterface, eaServiceName, svcRole string, port int, noneIsClusterIP bool, withLeader bool,
	spec api.ExternalAccessSpec, apiObject k8sutil.APIObject) error {

	if spec.GetType().IsManaged() {
		// Managed services should not be created or removed by the operator.
		return r.ensureExternalAccessManagedServices(ctx, cachedStatus, svcs, eaServiceName, svcRole, spec, apiObject,
			withLeader)
	}

	log := r.log.Str("section", "service-ea").Str("role", svcRole).Str("service", eaServiceName)
	createExternalAccessService := false
	deleteExternalAccessService := false
	eaServiceType := spec.GetType().AsServiceType() // Note: Type auto defaults to ServiceTypeLoadBalancer
	if existing, exists := cachedStatus.Service().V1().GetSimple(eaServiceName); exists {
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
					log.Info("LoadBalancerIP of external access service is not set, switching to NodePort")
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
			err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
				_, err := svcs.Update(ctxChild, existing, meta.UpdateOptions{})
				return err
			})
			if err != nil {
				log.Err(err).Debug("Failed to update external access service")
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
		log.Info("Removing obsolete external access service")
		err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			return svcs.Delete(ctxChild, eaServiceName, meta.DeleteOptions{})
		})
		if err != nil {
			log.Err(err).Debug("Failed to remove external access service")
			return errors.WithStack(err)
		}
	}
	if createExternalAccessService {
		// Let's create or update the database external access service
		nodePort := spec.GetNodePort()
		loadBalancerIP := spec.GetLoadBalancerIP()
		loadBalancerSourceRanges := spec.LoadBalancerSourceRanges
		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()
		_, newlyCreated, err := k8sutil.CreateExternalAccessService(ctxChild, svcs, eaServiceName, svcRole, apiObject,
			eaServiceType, port, nodePort, loadBalancerIP, loadBalancerSourceRanges, apiObject.AsOwner(), withLeader)
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
func (r *Resources) ensureExternalAccessManagedServices(ctx context.Context, cachedStatus inspectorInterface.Inspector,
	services servicev1.ModInterface, eaServiceName, svcRole string, spec api.ExternalAccessSpec,
	apiObject k8sutil.APIObject, withLeader bool) error {

	log := r.log.Str("section", "service-ea").Str("role", svcRole).Str("service", eaServiceName)
	managedServiceNames := spec.GetManagedServiceNames()
	deploymentName := apiObject.GetName()
	var selector map[string]string
	if withLeader {
		selector = k8sutil.LabelsForLeaderMember(deploymentName, svcRole, "")
	} else {
		selector = k8sutil.LabelsForDeployment(deploymentName, svcRole)
	}

	// Check if hardcoded service has correct selector.
	if svc, ok := cachedStatus.Service().V1().GetSimple(eaServiceName); !ok {
		// Hardcoded service (e.g. <deplname>-ea or <deplname>-sync) is not mandatory in `managed` type.
		if len(managedServiceNames) == 0 {
			log.Warn("the field \"spec.externalAccess.managedServiceNames\" should be provided for \"managed\" service type")
			return nil
		}
	} else if changed, err := ensureManagedServiceSelector(ctx, selector, svc, services); err != nil {
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

		if changed, err := ensureManagedServiceSelector(ctx, selector, svc, services); err != nil {
			return errors.WithMessage(err, "failed to ensure service selector")
		} else if changed {
			log.Info("selector applied to the managed service \"%s\"", svcName)
		}
	}

	return nil
}

// ensureManagedServiceSelector ensures if there is correct selector on a service.
func ensureManagedServiceSelector(ctx context.Context, selector map[string]string, svc *core.Service,
	services servicev1.ModInterface) (bool, error) {
	for key, value := range selector {
		if currentValue, ok := svc.Spec.Selector[key]; ok && value == currentValue {
			continue
		}

		p := patch.NewPatch()
		p.ItemReplace(patch.NewPath("spec", "selector"), selector)
		data, err := p.Marshal()
		if err != nil {
			return false, errors.WithMessage(err, "failed to marshal service selector")
		}

		if _, err = services.Patch(ctx, svc.GetName(), types.JSONPatchType, data, meta.PatchOptions{}); err != nil {
			return false, errors.WithMessage(err, "failed to patch service selector")
		}

		return true, nil
	}

	return false, nil
}
