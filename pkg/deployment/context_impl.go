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
	"net"
	"strconv"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/arangodb/arangosync-client/client"
	"github.com/arangodb/arangosync-client/tasks"
	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/agency"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	v1 "k8s.io/api/core/v1"
)

// GetBackup receives information about a backup resource
func (d *Deployment) GetBackup(backup string) (*backupApi.ArangoBackup, error) {
	return d.deps.DatabaseCRCli.BackupV1().ArangoBackups(d.Namespace()).Get(backup, metav1.GetOptions{})
}

// GetAPIObject returns the deployment as k8s object.
func (d *Deployment) GetAPIObject() k8sutil.APIObject {
	return d.apiObject
}

// GetServerGroupIterator returns the deployment as ServerGroupIterator.
func (d *Deployment) GetServerGroupIterator() resources.ServerGroupIterator {
	return d.apiObject
}

// GetKubeCli returns the kubernetes client
func (d *Deployment) GetKubeCli() kubernetes.Interface {
	return d.deps.KubeCli
}

// GetLifecycleImage returns the image name containing the lifecycle helper (== name of operator image)
func (d *Deployment) GetLifecycleImage() string {
	return d.config.LifecycleImage
}

// GetOperatorUUIDImage returns the image name containing the uuid helper (== name of operator image)
func (d *Deployment) GetOperatorUUIDImage() string {
	return d.config.OperatorUUIDInitImage
}

// GetNamespSecretsInterfaceace returns the kubernetes namespace that contains
// this deployment.
func (d *Deployment) GetNamespace() string {
	return d.apiObject.GetNamespace()
}

// GetPhase returns the current phase of the deployment
func (d *Deployment) GetPhase() api.DeploymentPhase {
	return d.status.last.Phase
}

// GetSpec returns the current specification
func (d *Deployment) GetSpec() api.DeploymentSpec {
	return d.apiObject.Spec
}

// GetDeploymentHealth returns a copy of the latest known state of cluster health
func (d *Deployment) GetDeploymentHealth() (driver.ClusterHealth, error) {
	return d.resources.GetDeploymentHealth()
}

// GetStatus returns the current status of the deployment
// together with the current version of that status.
func (d *Deployment) GetStatus() (api.DeploymentStatus, int32) {
	d.status.mutex.Lock()
	defer d.status.mutex.Unlock()

	return d.getStatus()
}

func (d *Deployment) getStatus() (api.DeploymentStatus, int32) {
	version := d.status.version
	return *d.status.last.DeepCopy(), version
}

// UpdateStatus replaces the status of the deployment with the given status and
// updates the resources in k8s.
// If the given last version does not match the actual last version of the status object,
// an error is returned.
func (d *Deployment) UpdateStatus(status api.DeploymentStatus, lastVersion int32, force ...bool) error {
	d.status.mutex.Lock()
	defer d.status.mutex.Unlock()

	return d.updateStatus(status, lastVersion, force...)
}

func (d *Deployment) updateStatus(status api.DeploymentStatus, lastVersion int32, force ...bool) error {
	if d.status.version != lastVersion {
		// Status is obsolete
		d.deps.Log.Error().
			Int32("expected-version", lastVersion).
			Int32("actual-version", d.status.version).
			Msg("UpdateStatus version conflict error.")
		return maskAny(fmt.Errorf("Status conflict error. Expected version %d, got %d", lastVersion, d.status.version))
	}
	d.status.version++
	d.status.last = *status.DeepCopy()
	if err := d.updateCRStatus(force...); err != nil {
		return maskAny(err)
	}
	return nil
}

// UpdateMember updates the deployment status wrt the given member.
func (d *Deployment) UpdateMember(member api.MemberStatus) error {
	status, lastVersion := d.GetStatus()
	_, group, found := status.Members.ElementByID(member.ID)
	if !found {
		return maskAny(fmt.Errorf("Member %s not found", member.ID))
	}
	if err := status.Members.Update(member, group); err != nil {
		return maskAny(err)
	}
	if err := d.UpdateStatus(status, lastVersion); err != nil {
		log.Debug().Err(err).Msg("Updating CR status failed")
		return maskAny(err)
	}
	return nil
}

// GetDatabaseClient returns a cached client for the entire database (cluster coordinators or single server),
// creating one if needed.
func (d *Deployment) GetDatabaseClient(ctx context.Context) (driver.Client, error) {
	c, err := d.clientCache.GetDatabase(ctx)
	if err != nil {
		return nil, maskAny(err)
	}
	return c, nil
}

// GetServerClient returns a cached client for a specific server.
func (d *Deployment) GetServerClient(ctx context.Context, group api.ServerGroup, id string) (driver.Client, error) {
	c, err := d.clientCache.Get(ctx, group, id)
	if err != nil {
		return nil, maskAny(err)
	}
	return c, nil
}

// GetAgencyClients returns a client connection for every agency member.
// If the given predicate is not nil, only agents are included where the given predicate returns true.
func (d *Deployment) GetAgencyClients(ctx context.Context, predicate func(id string) bool) ([]driver.Connection, error) {
	agencyMembers := d.status.last.Members.Agents
	result := make([]driver.Connection, 0, len(agencyMembers))
	for _, m := range agencyMembers {
		if predicate != nil && !predicate(m.ID) {
			continue
		}
		client, err := d.GetServerClient(ctx, api.ServerGroupAgents, m.ID)
		if err != nil {
			return nil, maskAny(err)
		}
		conn := client.Connection()
		result = append(result, conn)
	}
	return result, nil
}

// GetAgency returns a connection to the entire agency.
func (d *Deployment) GetAgency(ctx context.Context) (agency.Agency, error) {
	result, err := arangod.CreateArangodAgencyClient(ctx, d.deps.KubeCli.CoreV1(), d.apiObject)
	if err != nil {
		return nil, maskAny(err)
	}
	return result, nil
}

// GetSyncServerClient returns a cached client for a specific arangosync server.
func (d *Deployment) GetSyncServerClient(ctx context.Context, group api.ServerGroup, id string) (client.API, error) {
	// Fetch monitoring token
	log := d.deps.Log
	kubecli := d.deps.KubeCli
	ns := d.apiObject.GetNamespace()
	secrets := kubecli.CoreV1().Secrets(ns)
	secretName := d.apiObject.Spec.Sync.Monitoring.GetTokenSecretName()
	monitoringToken, err := k8sutil.GetTokenSecret(secrets, secretName)
	if err != nil {
		log.Debug().Err(err).Str("secret-name", secretName).Msg("Failed to get sync monitoring secret")
		return nil, maskAny(err)
	}

	// Fetch server DNS name
	dnsName := k8sutil.CreatePodDNSName(d.apiObject, group.AsRole(), id)

	// Build client
	port := k8sutil.ArangoSyncMasterPort
	if group == api.ServerGroupSyncWorkers {
		port = k8sutil.ArangoSyncWorkerPort
	}
	source := client.Endpoint{"https://" + net.JoinHostPort(dnsName, strconv.Itoa(port))}
	tlsAuth := tasks.TLSAuthentication{
		TLSClientAuthentication: tasks.TLSClientAuthentication{
			ClientToken: monitoringToken,
		},
	}
	auth := client.NewAuthentication(tlsAuth, "")
	insecureSkipVerify := true
	c, err := d.syncClientCache.GetClient(d.deps.Log, source, auth, insecureSkipVerify)
	if err != nil {
		return nil, maskAny(err)
	}
	return c, nil
}

// CreateMember adds a new member to the given group.
// If ID is non-empty, it will be used, otherwise a new ID is created.
func (d *Deployment) CreateMember(group api.ServerGroup, id string) (string, error) {
	log := d.deps.Log
	status, lastVersion := d.GetStatus()
	id, err := createMember(log, &status, group, id, d.apiObject)
	if err != nil {
		log.Debug().Err(err).Str("group", group.AsRole()).Msg("Failed to create member")
		return "", maskAny(err)
	}
	// Save added member
	if err := d.UpdateStatus(status, lastVersion); err != nil {
		log.Debug().Err(err).Msg("Updating CR status failed")
		return "", maskAny(err)
	}
	// Create event about it
	d.CreateEvent(k8sutil.NewMemberAddEvent(id, group.AsRole(), d.apiObject))

	return id, nil
}

// DeletePod deletes a pod with given name in the namespace
// of the deployment. If the pod does not exist, the error is ignored.
func (d *Deployment) DeletePod(podName string) error {
	log := d.deps.Log
	ns := d.apiObject.GetNamespace()
	if err := d.deps.KubeCli.CoreV1().Pods(ns).Delete(podName, &metav1.DeleteOptions{}); err != nil && !k8sutil.IsNotFound(err) {
		log.Debug().Err(err).Str("pod", podName).Msg("Failed to remove pod")
		return maskAny(err)
	}
	return nil
}

// CleanupPod deletes a given pod with force and explicit UID.
// If the pod does not exist, the error is ignored.
func (d *Deployment) CleanupPod(p *v1.Pod) error {
	log := d.deps.Log
	podName := p.GetName()
	ns := p.GetNamespace()
	options := metav1.NewDeleteOptions(0)
	options.Preconditions = metav1.NewUIDPreconditions(string(p.GetUID()))
	if err := d.deps.KubeCli.CoreV1().Pods(ns).Delete(podName, options); err != nil && !k8sutil.IsNotFound(err) {
		log.Debug().Err(err).Str("pod", podName).Msg("Failed to cleanup pod")
		return maskAny(err)
	}
	return nil
}

// RemovePodFinalizers removes all the finalizers from the Pod with given name in the namespace
// of the deployment. If the pod does not exist, the error is ignored.
func (d *Deployment) RemovePodFinalizers(podName string) error {
	log := d.deps.Log
	ns := d.GetNamespace()
	kubecli := d.deps.KubeCli
	p, err := kubecli.CoreV1().Pods(ns).Get(podName, metav1.GetOptions{})
	if err != nil {
		if k8sutil.IsNotFound(err) {
			return nil
		}
		return maskAny(err)
	}
	if err := k8sutil.RemovePodFinalizers(log, d.deps.KubeCli, p, p.GetFinalizers(), true); err != nil {
		return maskAny(err)
	}
	return nil
}

// DeletePvc deletes a persistent volume claim with given name in the namespace
// of the deployment. If the pvc does not exist, the error is ignored.
func (d *Deployment) DeletePvc(pvcName string) error {
	log := d.deps.Log
	ns := d.apiObject.GetNamespace()
	if err := d.deps.KubeCli.CoreV1().PersistentVolumeClaims(ns).Delete(pvcName, &metav1.DeleteOptions{}); err != nil && !k8sutil.IsNotFound(err) {
		log.Debug().Err(err).Str("pvc", pvcName).Msg("Failed to remove pvc")
		return maskAny(err)
	}
	return nil
}

// UpdatePvc updated a persistent volume claim in the namespace
// of the deployment. If the pvc does not exist, the error is ignored.
func (d *Deployment) UpdatePvc(pvc *v1.PersistentVolumeClaim) error {
	_, err := d.GetKubeCli().CoreV1().PersistentVolumeClaims(d.GetNamespace()).Update(pvc)
	if err == nil {
		return nil
	}

	if errors.IsNotFound(err) {
		return nil
	}

	return maskAny(err)
}

// GetPv returns PV info about PV with given name.
func (d *Deployment) GetPv(pvName string) (*v1.PersistentVolume, error) {
	pv, err := d.GetKubeCli().CoreV1().PersistentVolumes().Get(pvName, metav1.GetOptions{})
	if err == nil {
		return pv, nil
	}

	return nil, maskAny(err)
}

// GetOwnedPVCs returns a list of all PVCs owned by the deployment.
func (d *Deployment) GetOwnedPVCs() ([]v1.PersistentVolumeClaim, error) {
	// Get all current PVCs
	log := d.deps.Log
	pvcs, err := d.deps.KubeCli.CoreV1().PersistentVolumeClaims(d.apiObject.GetNamespace()).List(k8sutil.DeploymentListOpt(d.apiObject.GetName()))
	if err != nil {
		log.Debug().Err(err).Msg("Failed to list PVCs")
		return nil, maskAny(err)
	}
	myPVCs := make([]v1.PersistentVolumeClaim, 0, len(pvcs.Items))
	for _, p := range pvcs.Items {
		if d.isOwnerOf(&p) {
			myPVCs = append(myPVCs, p)
		}
	}
	return myPVCs, nil
}

// GetPvc gets a PVC by the given name, in the samespace of the deployment.
func (d *Deployment) GetPvc(pvcName string) (*v1.PersistentVolumeClaim, error) {
	pvc, err := d.deps.KubeCli.CoreV1().PersistentVolumeClaims(d.apiObject.GetNamespace()).Get(pvcName, metav1.GetOptions{})
	if err != nil {
		log.Debug().Err(err).Str("pvc-name", pvcName).Msg("Failed to get PVC")
		return nil, maskAny(err)
	}
	return pvc, nil
}

// GetTLSKeyfile returns the keyfile encoded TLS certificate+key for
// the given member.
func (d *Deployment) GetTLSKeyfile(group api.ServerGroup, member api.MemberStatus) (string, error) {
	secretName := k8sutil.CreateTLSKeyfileSecretName(d.apiObject.GetName(), group.AsRole(), member.ID)
	ns := d.apiObject.GetNamespace()
	secrets := d.deps.KubeCli.CoreV1().Secrets(ns)
	result, err := k8sutil.GetTLSKeyfileSecret(secrets, secretName)
	if err != nil {
		return "", maskAny(err)
	}
	return result, nil
}

// DeleteTLSKeyfile removes the Secret containing the TLS keyfile for the given member.
// If the secret does not exist, the error is ignored.
func (d *Deployment) DeleteTLSKeyfile(group api.ServerGroup, member api.MemberStatus) error {
	secretName := k8sutil.CreateTLSKeyfileSecretName(d.apiObject.GetName(), group.AsRole(), member.ID)
	ns := d.apiObject.GetNamespace()
	if err := d.deps.KubeCli.CoreV1().Secrets(ns).Delete(secretName, &metav1.DeleteOptions{}); err != nil && !k8sutil.IsNotFound(err) {
		return maskAny(err)
	}
	return nil
}

// DeleteSecret removes the Secret with given name.
// If the secret does not exist, the error is ignored.
func (d *Deployment) DeleteSecret(secretName string) error {
	ns := d.apiObject.GetNamespace()
	if err := d.deps.KubeCli.CoreV1().Secrets(ns).Delete(secretName, &metav1.DeleteOptions{}); err != nil && !k8sutil.IsNotFound(err) {
		return maskAny(err)
	}
	return nil
}

// GetShardSyncStatus returns true if all shards are in sync
func (d *Deployment) GetShardSyncStatus() bool {
	return d.resources.GetShardSyncStatus()
}

// InvalidateSyncStatus resets the sync state to false and triggers an inspection
func (d *Deployment) InvalidateSyncStatus() {
	d.resources.InvalidateSyncStatus()
}

func (d *Deployment) DisableScalingCluster() error {
	return d.clusterScalingIntegration.DisableScalingCluster()
}

func (d *Deployment) EnableScalingCluster() error {
	return d.clusterScalingIntegration.EnableScalingCluster()
}

// GetAgencyPlan returns agency plan
func (d *Deployment) GetAgencyData(ctx context.Context, i interface{}, keyParts ...string) error {
	a, err := d.GetAgency(ctx)
	if err != nil {
		return err
	}

	if err = a.ReadKey(ctx, keyParts, i); err != nil {
		return err
	}

	return err
}

func (d *Deployment) RenderPodForMember(cachedStatus inspector.Inspector, spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*v1.Pod, error) {
	return d.resources.RenderPodForMember(cachedStatus, spec, status, memberID, imageInfo)
}

func (d *Deployment) SelectImage(spec api.DeploymentSpec, status api.DeploymentStatus) (api.ImageInfo, bool) {
	return d.resources.SelectImage(spec, status)
}

func (d *Deployment) GetMetricsExporterImage() string {
	return d.config.MetricsExporterImage
}

func (d *Deployment) GetArangoImage() string {
	return d.config.ArangoImage
}

func (d *Deployment) WithStatusUpdate(action func(s *api.DeploymentStatus) bool, force ...bool) error {
	d.status.mutex.Lock()
	defer d.status.mutex.Unlock()

	status, version := d.getStatus()

	changed := action(&status)

	if !changed {
		return nil
	}

	return d.updateStatus(status, version, force...)
}

func (d *Deployment) SecretsInterface() k8sutil.SecretInterface {
	return d.GetKubeCli().CoreV1().Secrets(d.GetNamespace())
}

func (d *Deployment) GetName() string {
	return d.apiObject.GetName()
}
