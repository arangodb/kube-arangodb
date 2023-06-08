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

package deployment

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/record"

	"github.com/arangodb/arangosync-client/client"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/acs"
	"github.com/arangodb/kube-arangodb/pkg/deployment/acs/sutil"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/deployment/chaos"
	deploymentClient "github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	memberState "github.com/arangodb/kube-arangodb/pkg/deployment/member"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resilience"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	arangoInformer "github.com/arangodb/kube-arangodb/pkg/generated/informers/externalversions"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/operator/scope"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/conn"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/trigger"
)

// Config holds configuration settings for a Deployment
type Config struct {
	ServiceAccount            string
	AllowChaos                bool
	ScalingIntegrationEnabled bool
	OperatorImage             string
	ArangoImage               string
	Scope                     scope.Scope
}

// Dependencies holds dependent services for a Deployment
type Dependencies struct {
	EventRecorder record.EventRecorder

	Client kclient.Client
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
	maxInspectionInterval    = 10 * util.Interval(time.Second)       // Ensure we inspect the generated resources no less than with this interval
)

// Deployment is the in process state of an ArangoDeployment.
type Deployment struct {
	log logging.Logger

	name      string
	uid       types.UID
	namespace string

	currentObject       *api.ArangoDeployment
	currentObjectStatus *api.DeploymentStatus
	currentObjectLock   sync.RWMutex

	config Config
	deps   Dependencies

	eventCh chan *deploymentEvent
	stopCh  chan struct{}
	stopped int32

	inspectTrigger            trigger.Trigger
	updateDeploymentTrigger   trigger.Trigger
	clientCache               deploymentClient.Cache
	agencyCache               agency.Cache
	recentInspectionErrors    int
	clusterScalingIntegration *clusterScalingIntegration
	reconciler                *reconcile.Reconciler
	resilience                *resilience.Resilience
	resources                 *resources.Resources
	chaosMonkey               *chaos.Monkey
	acs                       sutil.ACS
	syncClientCache           client.ClientCache
	haveServiceMonitorCRD     bool

	memberState memberState.StateInspector

	metrics Metrics
}

func (d *Deployment) IsSyncEnabled() bool {
	d.currentObjectLock.RLock()
	defer d.currentObjectLock.RUnlock()

	if d.currentObject.GetAcceptedSpec().Sync.IsEnabled() {
		return true
	}

	if d.currentObject.Status.Conditions.IsTrue(api.ConditionTypeSyncEnabled) {
		return true
	}

	return false
}

func (d *Deployment) GetMembersState() memberState.StateInspector {
	return d.memberState
}

func (d *Deployment) GetAgencyCache() (state.State, bool) {
	return d.agencyCache.Data()
}

func (d *Deployment) GetAgencyHealth() (agency.Health, bool) {
	return d.agencyCache.Health()
}

// ShardsInSyncMap returns last in sync state of shards. If no state is available, false is returned.
func (d *Deployment) ShardsInSyncMap() (state.ShardsSyncStatus, bool) {
	return d.agencyCache.ShardsInSyncMap()
}

func (d *Deployment) GetAgencyArangoDBCache() (state.DB, bool) {
	return d.agencyCache.DataDB()
}

func (d *Deployment) RefreshAgencyCache(ctx context.Context) (uint64, error) {
	if d.GetSpec().Mode.Get() == api.DeploymentModeSingle {
		return 0, nil
	}

	if info := d.currentObject.Status.Agency; info != nil {
		if size := info.Size; size != nil {
			lCtx, c := globals.GetGlobalTimeouts().Agency().WithTimeout(ctx)
			defer c()

			rsize := int(*size)

			clients := agency.Connections{}
			for _, m := range d.GetStatus().Members.Agents {
				a, err := d.clientCache.GetRaw(api.ServerGroupAgents, m.ID)
				if err != nil {
					return 0, err
				}

				clients[m.ID] = a
			}

			return d.agencyCache.Reload(lCtx, rsize, clients)
		}
	}

	return 0, errors.Newf("Agency not yet established")
}

func (d *Deployment) SetAgencyMaintenanceMode(ctx context.Context, enabled bool) error {
	if !d.GetMode().HasAgents() {
		return nil
	}

	client, err := d.GetMembersState().State().GetDatabaseClient()
	if err != nil {
		return err
	}

	data := "on"
	if !enabled {
		data = "off"
	}

	conn := client.Connection()
	r, err := conn.NewRequest(http.MethodPut, "/_admin/cluster/maintenance")
	if err != nil {
		return err
	}

	if _, err := r.SetBody(data); err != nil {
		return err
	}

	resp, err := conn.Do(ctx, r)
	if err != nil {
		return err
	}

	if err := resp.CheckStatus(http.StatusOK); err != nil {
		return err
	}

	return nil
}

// New creates a new Deployment from the given API object.
func New(config Config, deps Dependencies, apiObject *api.ArangoDeployment) (*Deployment, error) {
	if err := apiObject.Spec.Validate(); err != nil {
		return nil, errors.WithStack(err)
	}

	i := inspector.NewInspector(inspector.NewDefaultThrottle(), deps.Client, apiObject.GetNamespace(), apiObject.GetName())

	d := &Deployment{
		currentObject:       apiObject,
		currentObjectStatus: apiObject.Status.DeepCopy(),
		name:                apiObject.GetName(),
		uid:                 apiObject.GetUID(),
		namespace:           apiObject.GetNamespace(),
		config:              config,
		deps:                deps,
		eventCh:             make(chan *deploymentEvent, deploymentEventQueueSize),
		stopCh:              make(chan struct{}),
		agencyCache:         agency.NewCache(apiObject.GetNamespace(), apiObject.GetName(), apiObject.GetAcceptedSpec().Mode),
		acs:                 acs.NewACS(apiObject.GetUID(), i),
	}

	d.log = logger.WrapObj(d)

	d.memberState = memberState.NewStateInspector(d)

	d.clientCache = deploymentClient.NewClientCache(d, conn.NewFactory(d.getAuth, d.getConnConfig))

	d.reconciler = reconcile.NewReconciler(apiObject.GetNamespace(), apiObject.GetName(), d)
	d.resilience = resilience.NewResilience(apiObject.GetNamespace(), apiObject.GetName(), d)
	d.resources = resources.NewResources(apiObject.GetNamespace(), apiObject.GetName(), d)

	localInventory.Add(d)

	for !d.acs.CurrentClusterCache().Initialised() {
		d.log.Warn("ACS cache not yet initialised")
		err := d.acs.CurrentClusterCache().Refresh(context.Background())
		if err != nil {
			d.log.Err(err).Error("Unable to get resources from ACS")
		}
	}

	aInformer := arangoInformer.NewSharedInformerFactoryWithOptions(deps.Client.Arango(), 0, arangoInformer.WithNamespace(apiObject.GetNamespace()))
	kInformer := informers.NewSharedInformerFactoryWithOptions(deps.Client.Kubernetes(), 0, informers.WithNamespace(apiObject.GetNamespace()))

	i.RegisterInformers(kInformer, aInformer)

	aInformer.Start(d.stopCh)
	kInformer.Start(d.stopCh)

	k8sutil.WaitForInformers(d.stopCh, 5*time.Second, kInformer, aInformer)

	go d.run()
	if apiObject.GetAcceptedSpec().GetMode() == api.DeploymentModeCluster {
		ci := newClusterScalingIntegration(d)
		d.clusterScalingIntegration = ci
		go ci.ListenForClusterEvents(d.stopCh)
	}
	if config.AllowChaos {
		d.chaosMonkey = chaos.NewMonkey(apiObject.GetNamespace(), apiObject.GetName(), d)
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

// Stop the deployment.
// Called when the deployment was deleted by the user.
func (d *Deployment) Stop() {
	d.log.Info("deployment is deleted by user")
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
			d.log.
				Int("used", l).
				Int("capacity", ecap).
				Warn("event queue buffer is almost full")
		}
	case <-d.stopCh:
	}
}

// run is the core the core worker.
// It processes the event queue and polls the state of generated
// resource on a regular basis.
func (d *Deployment) run() {
	log := d.log

	// Create agency mapping
	if err := d.createAgencyMapping(context.TODO()); err != nil {
		d.CreateEvent(k8sutil.NewErrorEvent("Failed to create agency mapping members", err, d.GetAPIObject()))
	}

	if d.GetPhase() == api.DeploymentPhaseNone {
		// Create initial topology
		if err := d.createInitialTopology(context.TODO()); err != nil {
			d.CreateEvent(k8sutil.NewErrorEvent("Failed to create initial topology", err, d.GetAPIObject()))
		}

		status := d.GetStatus()
		status.Phase = api.DeploymentPhaseRunning
		if err := d.UpdateStatus(context.TODO(), status); err != nil {
			log.Err(err).Warn("update initial CR status failed")
		}
		log.Info("start running...")
	}

	d.lookForServiceMonitorCRD()

	// Execute inspection for first time without delay of 10s
	log.Debug("Initially inspect deployment...")
	inspectionInterval := d.inspectDeployment(minInspectionInterval)
	log.Str("interval", inspectionInterval.String()).Debug("...deployment inspect started")

	d.sendCIUpdate()

	for {
		select {
		case <-d.stopCh:
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
			log.Trace("Inspect deployment...")
			inspectionInterval = d.inspectDeployment(inspectionInterval)
			log.Str("interval", inspectionInterval.String()).Trace("...inspected deployment")

		case <-d.updateDeploymentTrigger.Done():
			inspectionInterval = minInspectionInterval
			d.handleArangoDeploymentUpdatedEvent()
		case <-inspectionInterval.After():
			// Trigger inspection
			d.inspectTrigger.Trigger()
			// Backoff with next interval
			inspectionInterval = inspectionInterval.Backoff(1.5, maxInspectionInterval)
		}
	}
}

// validateNewSpec returns (canProceed, changed, error)
func (d *Deployment) acceptNewSpec(ctx context.Context, depl *api.ArangoDeployment) (bool, bool, error) {
	spec := depl.Spec.DeepCopy()

	origChecksum, err := spec.Checksum()
	if err != nil {
		return false, false, err
	}

	// Set defaults to the spec
	spec.SetDefaults(d.name)

	if features.DeploymentSpecDefaultsRestore().Enabled() {
		if accepted := depl.Status.AcceptedSpec; accepted != nil {
			spec.SetDefaultsFrom(*accepted)
		}
	}

	checksum, err := spec.Checksum()
	if err != nil {
		return false, false, err
	}

	if features.DeploymentSpecDefaultsRestore().Enabled() {
		if origChecksum != checksum {
			// Set defaults in deployment
			if err := d.updateCRSpec(ctx, *spec); err != nil {
				return false, false, err
			}

			return false, true, nil
		}
	}

	if accepted := depl.Status.AcceptedSpec; accepted == nil {
		// There is no last status, it is fresh one
		if err := spec.Validate(); err != nil {
			d.metrics.Errors.DeploymentValidationErrors++
			return false, false, err
		}

		// Update accepted spec
		if err := d.patchAcceptedSpec(ctx, spec, origChecksum); err != nil {
			return false, false, err
		}

		// Reconcile with new accepted spec
		return false, true, nil
	} else {
		// If we are equal then proceed
		acceptedChecksum, err := accepted.Checksum()
		if err != nil {
			return false, false, err
		}

		if v := depl.Status.AcceptedSpecVersion; acceptedChecksum == checksum && (v != nil && *v == origChecksum) {
			return true, false, nil
		}

		// We have already accepted spec, verify immutable part
		if fields := accepted.ResetImmutableFields(spec); len(fields) > 0 {
			d.metrics.Errors.DeploymentImmutableErrors += uint64(len(fields))
			if features.DeploymentSpecDefaultsRestore().Enabled() {
				d.log.Error("Restoring immutable fields: %s", strings.Join(fields, ", "))

				// In case of enabled, do restore
				if err := d.updateCRSpec(ctx, *spec); err != nil {
					return false, false, err
				}

				return false, true, nil
			}

			// We have immutable fields, throw an error and proceed
			return true, false, errors.Newf("Immutable fields cannot be changed: %s", strings.Join(fields, ", "))
		}

		// Update accepted spec
		if err := d.patchAcceptedSpec(ctx, spec, origChecksum); err != nil {
			return false, false, err
		}

		// Reconcile with new accepted spec
		return false, true, nil
	}
}

func (d *Deployment) patchAcceptedSpec(ctx context.Context, spec *api.DeploymentSpec, checksum string) error {
	return d.ApplyPatch(ctx, patch.ItemReplace(patch.NewPath("status", "accepted-spec"), spec),
		patch.ItemReplace(patch.NewPath("status", "acceptedSpecVersion"), checksum))
}

// handleArangoDeploymentUpdatedEvent is called when the deployment is updated by the user.
func (d *Deployment) handleArangoDeploymentUpdatedEvent() {
	// Trigger inspect
	d.inspectTrigger.Trigger()
}

// CreateEvent creates a given event.
// On error, the error is logged.
func (d *Deployment) CreateEvent(evt *k8sutil.Event) {
	d.deps.EventRecorder.Event(evt.InvolvedObject, evt.Type, evt.Reason, evt.Message)
}

func (d *Deployment) updateCRStatus(ctx context.Context, status api.DeploymentStatus) error {
	return d.ApplyPatch(ctx, patch.ItemReplace(patch.NewPath("status"), status))
}

func (d *Deployment) updateCRSpec(ctx context.Context, spec api.DeploymentSpec) error {
	return d.ApplyPatch(ctx, patch.ItemReplace(patch.NewPath("spec"), spec))
}

// isOwnerOf returns true if the given object belong to this deployment.
func (d *Deployment) isOwnerOf(obj meta.Object) bool {
	ownerRefs := obj.GetOwnerReferences()
	if len(ownerRefs) < 1 {
		return false
	}
	return ownerRefs[0].UID == d.currentObject.UID
}

// lookForServiceMonitorCRD checks if there is a CRD for the ServiceMonitor
// CR and sets the flag haveServiceMonitorCRD accordingly. This is called
// once at creation time of the deployment.
func (d *Deployment) lookForServiceMonitorCRD() {
	if d.haveServiceMonitorCRD {
		return
	}

	var err error
	if d.GetScope().IsNamespaced() {
		_, err = d.acs.CurrentClusterCache().ServiceMonitor().V1()
		if kerrors.IsForbiddenOrNotFound(err) {
			return
		}
	} else {
		_, err = d.deps.Client.KubernetesExtensions().ApiextensionsV1().CustomResourceDefinitions().Get(context.Background(), "servicemonitors.monitoring.coreos.com", meta.GetOptions{})
	}
	log := d.log
	log.Debug("Looking for ServiceMonitor CRD...")
	if err == nil {
		if !d.haveServiceMonitorCRD {
			log.Info("...have discovered ServiceMonitor CRD")
		}
		d.haveServiceMonitorCRD = true
		d.triggerInspection()
		return
	} else if kerrors.IsNotFound(err) {
		if d.haveServiceMonitorCRD {
			log.Info("...ServiceMonitor CRD no longer there")
		}
		d.haveServiceMonitorCRD = false
		return
	}
	log.Err(err).Warn("Error when looking for ServiceMonitor CRD")
}

// SetNumberOfServers adjust number of DBservers and coordinators in arangod
func (d *Deployment) SetNumberOfServers(ctx context.Context, noCoordinators, noDBServers *int) error {
	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	c, err := d.clientCache.GetDatabase(ctxChild)
	if err != nil {
		return errors.WithStack(err)
	}

	err = globals.GetGlobalTimeouts().ArangoD().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return arangod.SetNumberOfServers(ctxChild, c.Connection(), noCoordinators, noDBServers)
	})

	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (d *Deployment) ApplyPatch(ctx context.Context, p ...patch.Item) error {
	if len(p) == 0 {
		return nil
	}

	d.currentObjectLock.Lock()
	defer d.currentObjectLock.Unlock()

	return d.applyPatch(ctx, p...)
}

func (d *Deployment) applyPatch(ctx context.Context, p ...patch.Item) error {
	depls := d.deps.Client.Arango().DatabaseV1().ArangoDeployments(d.GetNamespace())

	if len(p) == 0 {
		return nil
	}

	pd, err := patch.NewPatch(p...).Marshal()
	if err != nil {
		return err
	}

	attempt := 0
	for {
		attempt++

		var newAPIObject *api.ArangoDeployment
		err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			newAPIObject, err = depls.Patch(ctxChild, d.GetName(), types.JSONPatchType, pd, meta.PatchOptions{})

			return err
		})
		if err == nil {
			// Update internal object

			d.currentObject = newAPIObject.DeepCopy()
			d.currentObjectStatus = newAPIObject.Status.DeepCopy()

			return nil
		}
		if attempt < 10 {
			continue
		}
		if err != nil {
			d.log.Err(err).Debug("failed to patch ArangoDeployment")
			return errors.WithStack(errors.Newf("failed to patch ArangoDeployment: %v", err))
		}
	}
}
