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

package deployment

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/arangod"

	"github.com/arangodb/arangosync-client/client"
	"github.com/rs/zerolog"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/chaos"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resilience"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/trigger"
)

// Config holds configuration settings for a Deployment
type Config struct {
	ServiceAccount        string
	AllowChaos            bool
	LifecycleImage        string
	OperatorUUIDInitImage string
	MetricsExporterImage  string
	ArangoImage           string
}

// Dependencies holds dependent services for a Deployment
type Dependencies struct {
	Log           zerolog.Logger
	KubeCli       kubernetes.Interface
	KubeExtCli    apiextensionsclient.Interface
	DatabaseCRCli versioned.Interface
	EventRecorder record.EventRecorder
}

// deploymentEventType strongly typed type of event
type deploymentEventType string

const (
	eventArangoDeploymentUpdated deploymentEventType = "ArangoDeploymentUpdated"
)

// deploymentEvent holds an event passed from the controller to the deployment.
type deploymentEvent struct {
	Type       deploymentEventType
	Deployment *api.ArangoDeployment
}

const (
	deploymentEventQueueSize = 256
	minInspectionInterval    = 250 * util.Interval(time.Millisecond) // Ensure we inspect the generated resources no less than with this interval
	maxInspectionInterval    = 30 * util.Interval(time.Second) // Ensure we inspect the generated resources no less than with this interval
)

// Deployment is the in process state of an ArangoDeployment.
type Deployment struct {
	apiObject *api.ArangoDeployment // API object
	status    struct {
		mutex   sync.Mutex
		version int32
		last    api.DeploymentStatus // Internal status copy of the CR
	}
	config Config
	deps   Dependencies

	eventCh chan *deploymentEvent
	stopCh  chan struct{}
	stopped int32

	inspectTrigger            trigger.Trigger
	inspectCRDTrigger         trigger.Trigger
	updateDeploymentTrigger   trigger.Trigger
	clientCache               *clientCache
	recentInspectionErrors    int
	clusterScalingIntegration *clusterScalingIntegration
	reconciler                *reconcile.Reconciler
	resilience                *resilience.Resilience
	resources                 *resources.Resources
	chaosMonkey               *chaos.Monkey
	syncClientCache           client.ClientCache
	haveServiceMonitorCRD     bool
}

// New creates a new Deployment from the given API object.
func New(config Config, deps Dependencies, apiObject *api.ArangoDeployment) (*Deployment, error) {
	if err := apiObject.Spec.Validate(); err != nil {
		return nil, maskAny(err)
	}
	d := &Deployment{
		apiObject:   apiObject,
		config:      config,
		deps:        deps,
		eventCh:     make(chan *deploymentEvent, deploymentEventQueueSize),
		stopCh:      make(chan struct{}),
		clientCache: newClientCache(deps.KubeCli, apiObject),
	}
	d.status.last = *(apiObject.Status.DeepCopy())
	d.reconciler = reconcile.NewReconciler(deps.Log, d)
	d.resilience = resilience.NewResilience(deps.Log, d)
	d.resources = resources.NewResources(deps.Log, d)
	if d.status.last.AcceptedSpec == nil {
		// We've validated the spec, so let's use it from now.
		d.status.last.AcceptedSpec = apiObject.Spec.DeepCopy()
	}

	go d.run()
	go d.listenForPodEvents(d.stopCh)
	go d.listenForPVCEvents(d.stopCh)
	go d.listenForSecretEvents(d.stopCh)
	go d.listenForServiceEvents(d.stopCh)
	go d.listenForCRDEvents(d.stopCh)
	if apiObject.Spec.GetMode() == api.DeploymentModeCluster {
		ci := newClusterScalingIntegration(d)
		d.clusterScalingIntegration = ci
		go ci.ListenForClusterEvents(d.stopCh)
		go d.resources.RunDeploymentHealthLoop(d.stopCh)
		go d.resources.RunDeploymentShardSyncLoop(d.stopCh)
	}
	if config.AllowChaos {
		d.chaosMonkey = chaos.NewMonkey(deps.Log, d)
		go d.chaosMonkey.Run(d.stopCh)
	}

	return d, nil
}

// Update the deployment.
// This sends an update event in the deployment event queue.
func (d *Deployment) Update(apiObject *api.ArangoDeployment) {
	d.send(&deploymentEvent{
		Type:       eventArangoDeploymentUpdated,
		Deployment: apiObject,
	})
}

// Delete the deployment.
// Called when the deployment was deleted by the user.
func (d *Deployment) Delete() {
	d.deps.Log.Info().Msg("deployment is deleted by user")
	if atomic.CompareAndSwapInt32(&d.stopped, 0, 1) {
		close(d.stopCh)
	}
}

// send given event into the deployment event queue.
func (d *Deployment) send(ev *deploymentEvent) {
	select {
	case d.eventCh <- ev:
		l, ecap := len(d.eventCh), cap(d.eventCh)
		if l > int(float64(ecap)*0.8) {
			d.deps.Log.Warn().
				Int("used", l).
				Int("capacity", ecap).
				Msg("event queue buffer is almost full")
		}
	case <-d.stopCh:
	}
}

// run is the core the core worker.
// It processes the event queue and polls the state of generated
// resource on a regular basis.
func (d *Deployment) run() {
	log := d.deps.Log

	if d.GetPhase() == api.DeploymentPhaseNone {
		// Create secrets
		if err := d.resources.EnsureSecrets(); err != nil {
			d.CreateEvent(k8sutil.NewErrorEvent("Failed to create secrets", err, d.GetAPIObject()))
		}

		// Create services
		if err := d.resources.EnsureServices(); err != nil {
			d.CreateEvent(k8sutil.NewErrorEvent("Failed to create services", err, d.GetAPIObject()))
		}

		// Create service monitor
		if d.haveServiceMonitorCRD {
			if err := d.resources.EnsureServiceMonitor(); err != nil {
				d.CreateEvent(k8sutil.NewErrorEvent("Failed to create service monitor", err, d.GetAPIObject()))
			}
		}

		// Create members
		if err := d.createInitialMembers(d.apiObject); err != nil {
			d.CreateEvent(k8sutil.NewErrorEvent("Failed to create initial members", err, d.GetAPIObject()))
		}

		// Create PVCs
		if err := d.resources.EnsurePVCs(); err != nil {
			d.CreateEvent(k8sutil.NewErrorEvent("Failed to create persistent volume claims", err, d.GetAPIObject()))
		}

		// Create pods
		if err := d.resources.EnsurePods(); err != nil {
			d.CreateEvent(k8sutil.NewErrorEvent("Failed to create pods", err, d.GetAPIObject()))
		}

		// Create Pod Disruption Budgets
		if err := d.resources.EnsurePDBs(); err != nil {
			d.CreateEvent(k8sutil.NewErrorEvent("Failed to create pdbs", err, d.GetAPIObject()))
		}

		status, lastVersion := d.GetStatus()
		status.Phase = api.DeploymentPhaseRunning
		if err := d.UpdateStatus(status, lastVersion); err != nil {
			log.Warn().Err(err).Msg("update initial CR status failed")
		}
		log.Info().Msg("start running...")
	}

	if err := d.resources.EnsureAnnotations(); err != nil {
		log.Warn().Err(err).Msg("unable to update annotations")
	}

	d.lookForServiceMonitorCRD()

	inspectionInterval := maxInspectionInterval
	for {
		select {
		case <-d.stopCh:
			// Remove finalizers from created resources
			log.Info().Msg("Deployment removed, removing finalizers to prevent orphaned resources")
			if err := d.removePodFinalizers(); err != nil {
				log.Warn().Err(err).Msg("Failed to remove Pod finalizers")
			}
			if err := d.removePVCFinalizers(); err != nil {
				log.Warn().Err(err).Msg("Failed to remove PVC finalizers")
			}
			// We're being stopped.
			return

		case event := <-d.eventCh:
			// Got event from event queue
			switch event.Type {
			case eventArangoDeploymentUpdated:
				d.updateDeploymentTrigger.Trigger()
			default:
				panic("unknown event type" + event.Type)
			}

		case <-d.inspectTrigger.Done():
			log.Debug().Msg("Inspect deployment...")
			inspectionInterval = d.inspectDeployment(inspectionInterval)
			log.Debug().Str("interval", inspectionInterval.String()).Msg("...inspected deployment")

		case <-d.inspectCRDTrigger.Done():
			d.lookForServiceMonitorCRD()
		case <-d.updateDeploymentTrigger.Done():
			inspectionInterval = minInspectionInterval
			if err := d.handleArangoDeploymentUpdatedEvent(); err != nil {
				d.CreateEvent(k8sutil.NewErrorEvent("Failed to handle deployment update", err, d.GetAPIObject()))
			}

		case <-inspectionInterval.After():
			// Trigger inspection
			d.inspectTrigger.Trigger()
			// Backoff with next interval
			inspectionInterval = inspectionInterval.Backoff(1.5, maxInspectionInterval)
		}
	}
}

// handleArangoDeploymentUpdatedEvent is called when the deployment is updated by the user.
func (d *Deployment) handleArangoDeploymentUpdatedEvent() error {
	log := d.deps.Log.With().Str("deployment", d.apiObject.GetName()).Logger()

	// Get the most recent version of the deployment from the API server
	current, err := d.deps.DatabaseCRCli.DatabaseV1().ArangoDeployments(d.apiObject.GetNamespace()).Get(d.apiObject.GetName(), metav1.GetOptions{})
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get current version of deployment from API server")
		if k8sutil.IsNotFound(err) {
			return nil
		}
		return maskAny(err)
	}

	specBefore := d.apiObject.Spec
	status := d.status.last
	if d.status.last.AcceptedSpec != nil {
		specBefore = *status.AcceptedSpec.DeepCopy()
	}
	newAPIObject := current.DeepCopy()
	newAPIObject.Spec.SetDefaultsFrom(specBefore)
	newAPIObject.Spec.SetDefaults(d.apiObject.GetName())

	resetFields := specBefore.ResetImmutableFields(&newAPIObject.Spec)
	if len(resetFields) > 0 {
		log.Debug().Strs("fields", resetFields).Msg("Found modified immutable fields")
		newAPIObject.Spec.SetDefaults(d.apiObject.GetName())
	}
	if err := newAPIObject.Spec.Validate(); err != nil {
		d.CreateEvent(k8sutil.NewErrorEvent("Validation failed", err, d.apiObject))
		// Try to reset object
		if err := d.updateCRSpec(d.apiObject.Spec, true); err != nil {
			log.Error().Err(err).Msg("Restore original spec failed")
			d.CreateEvent(k8sutil.NewErrorEvent("Restore original failed", err, d.apiObject))
		}
		return nil
	}
	if len(resetFields) > 0 {
		for _, fieldName := range resetFields {
			log.Debug().Str("field", fieldName).Msg("Reset immutable field")
			d.CreateEvent(k8sutil.NewImmutableFieldEvent(fieldName, d.apiObject))
		}
	}

	// Save updated spec
	if err := d.updateCRSpec(newAPIObject.Spec, true); err != nil {
		return maskAny(fmt.Errorf("failed to update ArangoDeployment spec: %v", err))
	}
	// Save updated accepted spec
	{
		status, lastVersion := d.GetStatus()
		if newAPIObject.Status.IsForceReload() {
			log.Warn().Msg("Forced status reload!")
			status = newAPIObject.Status
			status.ForceStatusReload = nil
		}
		status.AcceptedSpec = newAPIObject.Spec.DeepCopy()
		if err := d.UpdateStatus(status, lastVersion); err != nil {
			return maskAny(fmt.Errorf("failed to update ArangoDeployment status: %v", err))
		}
	}

	// Notify cluster of desired server count
	if ci := d.clusterScalingIntegration; ci != nil {
		ci.SendUpdateToCluster(d.apiObject.Spec)
	}

	// Trigger inspect
	d.inspectTrigger.Trigger()

	return nil
}

// CreateEvent creates a given event.
// On error, the error is logged.
func (d *Deployment) CreateEvent(evt *k8sutil.Event) {
	d.deps.EventRecorder.Event(evt.InvolvedObject, evt.Type, evt.Reason, evt.Message)
}

// Update the status of the API object from the internal status
func (d *Deployment) updateCRStatus(force ...bool) error {
	if len(force) == 0 || !force[0] {
		if d.apiObject.Status.Equal(d.status.last) {
			// Nothing has changed
			return nil
		}
	}

	// Send update to API server
	ns := d.apiObject.GetNamespace()
	depls := d.deps.DatabaseCRCli.DatabaseV1().ArangoDeployments(ns)
	update := d.apiObject.DeepCopy()
	attempt := 0
	for {
		attempt++
		update.Status = d.status.last
		if update.GetDeletionTimestamp() == nil {
			ensureFinalizers(update)
		}
		newAPIObject, err := depls.Update(update)
		if err == nil {
			// Update internal object
			d.apiObject = newAPIObject
			return nil
		}
		if attempt < 10 && k8sutil.IsConflict(err) {
			// API object may have been changed already,
			// Reload api object and try again
			var current *api.ArangoDeployment
			current, err = depls.Get(update.GetName(), metav1.GetOptions{})
			if err == nil {
				update = current.DeepCopy()
				continue
			}
		}
		if err != nil {
			d.deps.Log.Debug().Err(err).Msg("failed to patch ArangoDeployment status")
			return maskAny(fmt.Errorf("failed to patch ArangoDeployment status: %v", err))
		}
	}
}

// Update the spec part of the API object (d.apiObject)
// to the given object, while preserving the status.
// On success, d.apiObject is updated.
func (d *Deployment) updateCRSpec(newSpec api.DeploymentSpec, force ...bool) error {

	if len(force) == 0 || !force[0] {
		if d.apiObject.Spec.Equal(&newSpec) {
			d.deps.Log.Debug().Msg("Nothing to update in updateCRSpec")
			// Nothing to update
			return nil
		}
	}

	// Send update to API server
	update := d.apiObject.DeepCopy()
	attempt := 0
	for {
		attempt++
		update.Spec = newSpec
		update.Status = d.status.last
		ns := d.apiObject.GetNamespace()
		newAPIObject, err := d.deps.DatabaseCRCli.DatabaseV1().ArangoDeployments(ns).Update(update)
		if err == nil {
			// Update internal object
			d.apiObject = newAPIObject
			return nil
		}
		if attempt < 10 && k8sutil.IsConflict(err) {
			// API object may have been changed already,
			// Reload api object and try again
			var current *api.ArangoDeployment
			current, err = d.deps.DatabaseCRCli.DatabaseV1().ArangoDeployments(ns).Get(update.GetName(), metav1.GetOptions{})
			if err == nil {
				update = current.DeepCopy()
				continue
			}
		}
		if err != nil {
			d.deps.Log.Debug().Err(err).Msg("failed to patch ArangoDeployment spec")
			return maskAny(fmt.Errorf("failed to patch ArangoDeployment spec: %v", err))
		}
	}
}

// isOwnerOf returns true if the given object belong to this deployment.
func (d *Deployment) isOwnerOf(obj metav1.Object) bool {
	ownerRefs := obj.GetOwnerReferences()
	if len(ownerRefs) < 1 {
		return false
	}
	return ownerRefs[0].UID == d.apiObject.UID
}

// lookForServiceMonitorCRD checks if there is a CRD for the ServiceMonitor
// CR and sets the flag haveServiceMonitorCRD accordingly. This is called
// once at creation time of the deployment and then always if the CRD
// informer is triggered.
func (d *Deployment) lookForServiceMonitorCRD() {
	_, err := d.deps.KubeExtCli.ApiextensionsV1beta1().CustomResourceDefinitions().Get("servicemonitors.monitoring.coreos.com", metav1.GetOptions{})
	log := d.deps.Log
	log.Debug().Msgf("Looking for ServiceMonitor CRD...")
	if err == nil {
		if !d.haveServiceMonitorCRD {
			log.Info().Msgf("...have discovered ServiceMonitor CRD")
		}
		d.haveServiceMonitorCRD = true
		d.triggerInspection()
		return
	} else if k8sutil.IsNotFound(err) {
		if d.haveServiceMonitorCRD {
			log.Info().Msgf("...ServiceMonitor CRD no longer there")
		}
		d.haveServiceMonitorCRD = false
		return
	}
	log.Warn().Err(err).Msgf("Error when looking for ServiceMonitor CRD")
}

// SetNumberOfServers adjust number of DBservers and coordinators in arangod
func (d *Deployment) SetNumberOfServers(ctx context.Context, noCoordinators, noDBServers *int) error {
	c, err := d.clientCache.GetDatabase(ctx)
	if err != nil {
		return maskAny(err)
	}

	err = arangod.SetNumberOfServers(ctx, c.Connection(), noCoordinators, noDBServers)
	if err != nil {
		return maskAny(err)
	}
	return nil
}
