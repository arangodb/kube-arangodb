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

package deployment

import (
	"context"
	"crypto/tls"
	"net"
	nhttp "net/http"
	"strconv"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/globals"

	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangomember"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolumeclaim"
	podMod "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/poddisruptionbudget"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/serviceaccount"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/servicemonitor"

	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile"

	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/arangodb/kube-arangodb/pkg/util/arangod/conn"

	"github.com/arangodb/kube-arangodb/pkg/operator/scope"

	"github.com/arangodb/kube-arangodb/pkg/deployment/features"

	"github.com/arangodb/go-driver/http"
	"github.com/arangodb/go-driver/jwt"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/arangodb/arangosync-client/client"
	"github.com/arangodb/arangosync-client/tasks"
	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/agency"
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconciler"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/rs/zerolog/log"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ resources.Context = &Deployment{}

// GetBackup receives information about a backup resource
func (d *Deployment) GetBackup(ctx context.Context, backup string) (*backupApi.ArangoBackup, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()

	return d.deps.Client.Arango().BackupV1().ArangoBackups(d.Namespace()).Get(ctxChild, backup, meta.GetOptions{})
}

// GetAPIObject returns the deployment as k8s object.
func (d *Deployment) GetAPIObject() k8sutil.APIObject {
	return d.apiObject
}

// GetServerGroupIterator returns the deployment as ServerGroupIterator.
func (d *Deployment) GetServerGroupIterator() reconciler.ServerGroupIterator {
	return d.apiObject
}

func (d *Deployment) GetScope() scope.Scope {
	return d.config.Scope
}

func (d *Deployment) GetOperatorImage() string {
	return d.config.OperatorImage
}

// GetNamespace returns the kubernetes namespace that contains
// this deployment.
func (d *Deployment) GetNamespace() string {
	return d.namespace
}

// GetPhase returns the current phase of the deployment
func (d *Deployment) GetPhase() api.DeploymentPhase {
	return d.status.last.Phase
}

// GetSpec returns the current specification
func (d *Deployment) GetSpec() api.DeploymentSpec {
	return d.apiObject.Spec
}

// GetStatus returns the current status of the deployment
// together with the current version of that status.
func (d *Deployment) GetStatus() (api.DeploymentStatus, int32) {
	return d.getStatus()
}

func (d *Deployment) getStatus() (api.DeploymentStatus, int32) {
	obj := d.status.deploymentStatusObject
	return *obj.last.DeepCopy(), obj.version
}

// UpdateStatus replaces the status of the deployment with the given status and
// updates the resources in k8s.
// If the given last version does not match the actual last version of the status object,
// an error is returned.
func (d *Deployment) UpdateStatus(ctx context.Context, status api.DeploymentStatus, lastVersion int32, force ...bool) error {
	d.status.mutex.Lock()
	defer d.status.mutex.Unlock()

	return d.updateStatus(ctx, status, lastVersion, force...)
}

func (d *Deployment) updateStatus(ctx context.Context, status api.DeploymentStatus, lastVersion int32, force ...bool) error {
	if d.status.version != lastVersion {
		// Status is obsolete
		d.deps.Log.Error().
			Int32("expected-version", lastVersion).
			Int32("actual-version", d.status.version).
			Msg("UpdateStatus version conflict error.")
		return errors.WithStack(errors.Newf("Status conflict error. Expected version %d, got %d", lastVersion, d.status.version))
	}

	d.status.deploymentStatusObject = deploymentStatusObject{
		version: d.status.deploymentStatusObject.version + 1,
		last:    *status.DeepCopy(),
	}
	if err := d.updateCRStatus(ctx, force...); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// UpdateMember updates the deployment status wrt the given member.
func (d *Deployment) UpdateMember(ctx context.Context, member api.MemberStatus) error {
	status, lastVersion := d.GetStatus()
	_, group, found := status.Members.ElementByID(member.ID)
	if !found {
		return errors.WithStack(errors.Newf("Member %s not found", member.ID))
	}
	if err := status.Members.Update(member, group); err != nil {
		return errors.WithStack(err)
	}
	if err := d.UpdateStatus(ctx, status, lastVersion); err != nil {
		d.deps.Log.Debug().Err(err).Msg("Updating CR status failed")
		return errors.WithStack(err)
	}
	return nil
}

// GetDatabaseClient returns a cached client for the entire database (cluster coordinators or single server),
// creating one if needed.
func (d *Deployment) GetDatabaseClient(ctx context.Context) (driver.Client, error) {
	c, err := d.clientCache.GetDatabase(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// GetServerClient returns a cached client for a specific server.
func (d *Deployment) GetServerClient(ctx context.Context, group api.ServerGroup, id string) (driver.Client, error) {
	c, err := d.clientCache.Get(ctx, group, id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// GetAuthentication return authentication for members
func (d *Deployment) GetAuthentication() conn.Auth {
	return d.clientCache.GetAuth()
}

// GetAgencyClients returns a client connection for every agency member.
func (d *Deployment) GetAgencyClients(ctx context.Context) ([]driver.Connection, error) {
	return d.GetAgencyClientsWithPredicate(ctx, nil)
}

// GetAgencyClientsWithPredicate returns a client connection for every agency member.
// If the given predicate is not nil, only agents are included where the given predicate returns true.
func (d *Deployment) GetAgencyClientsWithPredicate(ctx context.Context, predicate func(id string) bool) ([]driver.Connection, error) {
	agencyMembers := d.status.last.Members.Agents
	result := make([]driver.Connection, 0, len(agencyMembers))
	for _, m := range agencyMembers {
		if predicate != nil && !predicate(m.ID) {
			continue
		}
		client, err := d.GetServerClient(ctx, api.ServerGroupAgents, m.ID)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		conn := client.Connection()
		result = append(result, conn)
	}
	return result, nil
}

// GetAgency returns a connection to the entire agency.
func (d *Deployment) GetAgency(ctx context.Context) (agency.Agency, error) {
	return d.clientCache.GetAgency(ctx)
}

func (d *Deployment) getConnConfig() (http.ConnectionConfig, error) {
	transport := &nhttp.Transport{
		Proxy: nhttp.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 100 * time.Millisecond,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       100 * time.Millisecond,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if d.apiObject.Spec.TLS.IsSecure() {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	connConfig := http.ConnectionConfig{
		Transport:          transport,
		DontFollowRedirect: true,
	}

	return connConfig, nil
}

func (d *Deployment) getAuth() (driver.Authentication, error) {
	if !d.apiObject.Spec.Authentication.IsAuthenticated() {
		return nil, nil
	}

	var secret string
	var found bool

	// Check if we can find token in folder
	if i := d.apiObject.Status.CurrentImage; i == nil || features.JWTRotation().Supported(i.ArangoDBVersion, i.Enterprise) {
		secret, found = d.getJWTFolderToken()
	}

	// Fallback to token
	if !found {
		secret, found = d.getJWTToken()
	}

	if !found {
		return nil, errors.Newf("JWT Secret is invalid")
	}

	jwt, err := jwt.CreateArangodJwtAuthorizationHeader(secret, "kube-arangodb")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return driver.RawAuthentication(jwt), nil
}

func (d *Deployment) getJWTFolderToken() (string, bool) {
	if i := d.apiObject.Status.CurrentImage; i == nil || features.JWTRotation().Supported(i.ArangoDBVersion, i.Enterprise) {
		s, err := d.GetCachedStatus().SecretReadInterface().Get(context.Background(), pod.JWTSecretFolder(d.GetName()), meta.GetOptions{})
		if err != nil {
			d.deps.Log.Error().Err(err).Msgf("Unable to get secret")
			return "", false
		}

		if len(s.Data) == 0 {
			return "", false
		}

		if q, ok := s.Data[pod.ActiveJWTKey]; ok {
			return string(q), true
		} else {
			for _, q := range s.Data {
				return string(q), true
			}
		}
	}

	return "", false
}

func (d *Deployment) getJWTToken() (string, bool) {
	s, err := d.GetCachedStatus().SecretReadInterface().Get(context.Background(), d.apiObject.Spec.Authentication.GetJWTSecretName(), meta.GetOptions{})
	if err != nil {
		return "", false
	}

	jwt, ok := s.Data[constants.SecretKeyToken]
	if !ok {
		return "", false
	}

	return string(jwt), true
}

// GetSyncServerClient returns a cached client for a specific arangosync server.
func (d *Deployment) GetSyncServerClient(ctx context.Context, group api.ServerGroup, id string) (client.API, error) {
	// Fetch monitoring token
	log := d.deps.Log
	secretName := d.apiObject.Spec.Sync.Monitoring.GetTokenSecretName()
	monitoringToken, err := k8sutil.GetTokenSecret(ctx, d.GetCachedStatus().SecretReadInterface(), secretName)
	if err != nil {
		log.Debug().Err(err).Str("secret-name", secretName).Msg("Failed to get sync monitoring secret")
		return nil, errors.WithStack(err)
	}

	// Fetch server DNS name
	dnsName := k8sutil.CreatePodDNSNameWithDomain(d.apiObject, d.apiObject.Spec.ClusterDomain, group.AsRole(), id)

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
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// CreateMember adds a new member to the given group.
// If ID is non-empty, it will be used, otherwise a new ID is created.
func (d *Deployment) CreateMember(ctx context.Context, group api.ServerGroup, id string, mods ...reconcile.CreateMemberMod) (string, error) {
	log := d.deps.Log
	if err := d.WithStatusUpdateErr(ctx, func(s *api.DeploymentStatus) (bool, error) {
		nid, err := createMember(log, s, group, id, d.apiObject, mods...)
		if err != nil {
			log.Debug().Err(err).Str("group", group.AsRole()).Msg("Failed to create member")
			return false, errors.WithStack(err)
		}

		id = nid

		return true, nil
	}); err != nil {
		return "", err
	}

	// Create event about it
	d.CreateEvent(k8sutil.NewMemberAddEvent(id, group.AsRole(), d.apiObject))

	return id, nil
}

// GetPod returns pod.
func (d *Deployment) GetPod(ctx context.Context, podName string) (*core.Pod, error) {
	return d.GetCachedStatus().PodReadInterface().Get(ctx, podName, meta.GetOptions{})
}

// DeletePod deletes a pod with given name in the namespace
// of the deployment. If the pod does not exist, the error is ignored.
func (d *Deployment) DeletePod(ctx context.Context, podName string, options meta.DeleteOptions) error {
	log := d.deps.Log
	err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return d.PodsModInterface().Delete(ctxChild, podName, options)
	})
	if err != nil && !k8sutil.IsNotFound(err) {
		log.Debug().Err(err).Str("pod", podName).Msg("Failed to remove pod")
		return errors.WithStack(err)
	}
	return nil
}

// CleanupPod deletes a given pod with force and explicit UID.
// If the pod does not exist, the error is ignored.
func (d *Deployment) CleanupPod(ctx context.Context, p *core.Pod) error {
	log := d.deps.Log
	podName := p.GetName()
	options := meta.NewDeleteOptions(0)
	options.Preconditions = meta.NewUIDPreconditions(string(p.GetUID()))
	err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return d.PodsModInterface().Delete(ctxChild, podName, *options)
	})
	if err != nil && !k8sutil.IsNotFound(err) {
		log.Debug().Err(err).Str("pod", podName).Msg("Failed to cleanup pod")
		return errors.WithStack(err)
	}
	return nil
}

// RemovePodFinalizers removes all the finalizers from the Pod with given name in the namespace
// of the deployment. If the pod does not exist, the error is ignored.
func (d *Deployment) RemovePodFinalizers(ctx context.Context, podName string) error {
	log := d.deps.Log

	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	p, err := d.GetCachedStatus().PodReadInterface().Get(ctxChild, podName, meta.GetOptions{})
	if err != nil {
		if k8sutil.IsNotFound(err) {
			return nil
		}
		return errors.WithStack(err)
	}

	_, err = k8sutil.RemovePodFinalizers(ctx, d.GetCachedStatus(), log, d.PodsModInterface(), p, p.GetFinalizers(), true)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// DeletePvc deletes a persistent volume claim with given name in the namespace
// of the deployment. If the pvc does not exist, the error is ignored.
func (d *Deployment) DeletePvc(ctx context.Context, pvcName string) error {
	log := d.deps.Log
	err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return d.PersistentVolumeClaimsModInterface().Delete(ctxChild, pvcName, meta.DeleteOptions{})
	})
	if err != nil && !k8sutil.IsNotFound(err) {
		log.Debug().Err(err).Str("pvc", pvcName).Msg("Failed to remove pvc")
		return errors.WithStack(err)
	}
	return nil
}

// UpdatePvc updated a persistent volume claim in the namespace
// of the deployment. If the pvc does not exist, the error is ignored.
func (d *Deployment) UpdatePvc(ctx context.Context, pvc *core.PersistentVolumeClaim) error {
	err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		_, err := d.PersistentVolumeClaimsModInterface().Update(ctxChild, pvc, meta.UpdateOptions{})
		return err
	})
	if err == nil {
		return nil
	}

	if apiErrors.IsNotFound(err) {
		return nil
	}

	return errors.WithStack(err)
}

// GetOwnedPVCs returns a list of all PVCs owned by the deployment.
func (d *Deployment) GetOwnedPVCs() ([]core.PersistentVolumeClaim, error) {
	// Get all current PVCs
	pvcs := d.GetCachedStatus().PersistentVolumeClaims()
	myPVCs := make([]core.PersistentVolumeClaim, 0, len(pvcs))
	for _, p := range pvcs {
		if d.isOwnerOf(p) {
			myPVCs = append(myPVCs, *p)
		}
	}
	return myPVCs, nil
}

// GetPvc gets a PVC by the given name, in the samespace of the deployment.
func (d *Deployment) GetPvc(ctx context.Context, pvcName string) (*core.PersistentVolumeClaim, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()

	pvc, err := d.GetCachedStatus().PersistentVolumeClaimReadInterface().Get(ctxChild, pvcName, meta.GetOptions{})
	if err != nil {
		log.Debug().Err(err).Str("pvc-name", pvcName).Msg("Failed to get PVC")
		return nil, errors.WithStack(err)
	}
	return pvc, nil
}

// GetTLSKeyfile returns the keyfile encoded TLS certificate+key for
// the given member.
func (d *Deployment) GetTLSKeyfile(group api.ServerGroup, member api.MemberStatus) (string, error) {
	secretName := k8sutil.CreateTLSKeyfileSecretName(d.GetName(), group.AsRole(), member.ID)
	result, err := k8sutil.GetTLSKeyfileSecret(d.GetCachedStatus().SecretReadInterface(), secretName)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return result, nil
}

// DeleteTLSKeyfile removes the Secret containing the TLS keyfile for the given member.
// If the secret does not exist, the error is ignored.
func (d *Deployment) DeleteTLSKeyfile(ctx context.Context, group api.ServerGroup, member api.MemberStatus) error {
	secretName := k8sutil.CreateTLSKeyfileSecretName(d.GetName(), group.AsRole(), member.ID)
	err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return d.SecretsModInterface().Delete(ctxChild, secretName, meta.DeleteOptions{})
	})
	if err != nil && !k8sutil.IsNotFound(err) {
		return errors.WithStack(err)
	}
	return nil
}

// DeleteSecret removes the Secret with given name.
// If the secret does not exist, the error is ignored.
func (d *Deployment) DeleteSecret(secretName string) error {
	if err := d.SecretsModInterface().Delete(context.Background(), secretName, meta.DeleteOptions{}); err != nil && !k8sutil.IsNotFound(err) {
		return errors.WithStack(err)
	}
	return nil
}

func (d *Deployment) DisableScalingCluster(ctx context.Context) error {
	return d.clusterScalingIntegration.DisableScalingCluster(ctx)
}

func (d *Deployment) EnableScalingCluster(ctx context.Context) error {
	return d.clusterScalingIntegration.EnableScalingCluster(ctx)
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

func (d *Deployment) RenderPodForMember(ctx context.Context, cachedStatus inspectorInterface.Inspector, spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*core.Pod, error) {
	return d.resources.RenderPodForMember(ctx, cachedStatus, spec, status, memberID, imageInfo)
}

func (d *Deployment) RenderPodForMemberFromCurrent(ctx context.Context, cachedStatus inspectorInterface.Inspector, memberID string) (*core.Pod, error) {
	return d.resources.RenderPodForMemberFromCurrent(ctx, cachedStatus, memberID)
}

func (d *Deployment) RenderPodTemplateForMember(ctx context.Context, cachedStatus inspectorInterface.Inspector, spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*core.PodTemplateSpec, error) {
	return d.resources.RenderPodTemplateForMember(ctx, cachedStatus, spec, status, memberID, imageInfo)
}

func (d *Deployment) RenderPodTemplateForMemberFromCurrent(ctx context.Context, cachedStatus inspectorInterface.Inspector, memberID string) (*core.PodTemplateSpec, error) {
	return d.resources.RenderPodTemplateForMemberFromCurrent(ctx, cachedStatus, memberID)
}

func (d *Deployment) SelectImage(spec api.DeploymentSpec, status api.DeploymentStatus) (api.ImageInfo, bool) {
	return d.resources.SelectImage(spec, status)
}

func (d *Deployment) SelectImageForMember(spec api.DeploymentSpec, status api.DeploymentStatus, member api.MemberStatus) (api.ImageInfo, bool) {
	return d.resources.SelectImageForMember(spec, status, member)
}

func (d *Deployment) GetArangoImage() string {
	return d.config.ArangoImage
}

func (d *Deployment) WithStatusUpdateErr(ctx context.Context, action reconciler.DeploymentStatusUpdateErrFunc, force ...bool) error {
	d.status.mutex.Lock()
	defer d.status.mutex.Unlock()

	status, version := d.getStatus()

	changed, err := action(&status)

	if err != nil {
		return err
	}

	if !changed {
		return nil
	}

	return d.updateStatus(ctx, status, version, force...)
}

func (d *Deployment) WithStatusUpdate(ctx context.Context, action reconciler.DeploymentStatusUpdateFunc, force ...bool) error {
	return d.WithStatusUpdateErr(ctx, func(s *api.DeploymentStatus) (bool, error) {
		return action(s), nil
	}, force...)
}

func (d *Deployment) SecretsModInterface() secret.ModInterface {
	return kclient.NewModInterface(d.deps.Client, d.namespace).Secrets()
}

func (d *Deployment) PodsModInterface() podMod.ModInterface {
	return kclient.NewModInterface(d.deps.Client, d.namespace).Pods()
}

func (d *Deployment) ServiceAccountsModInterface() serviceaccount.ModInterface {
	return kclient.NewModInterface(d.deps.Client, d.namespace).ServiceAccounts()
}

func (d *Deployment) ServicesModInterface() service.ModInterface {
	return kclient.NewModInterface(d.deps.Client, d.namespace).Services()
}

func (d *Deployment) PersistentVolumeClaimsModInterface() persistentvolumeclaim.ModInterface {
	return kclient.NewModInterface(d.deps.Client, d.namespace).PersistentVolumeClaims()
}

func (d *Deployment) PodDisruptionBudgetsModInterface() poddisruptionbudget.ModInterface {
	return kclient.NewModInterface(d.deps.Client, d.namespace).PodDisruptionBudgets()
}

func (d *Deployment) ServiceMonitorsModInterface() servicemonitor.ModInterface {
	return kclient.NewModInterface(d.deps.Client, d.namespace).ServiceMonitors()
}

func (d *Deployment) ArangoMembersModInterface() arangomember.ModInterface {
	return kclient.NewModInterface(d.deps.Client, d.namespace).ArangoMembers()
}

func (d *Deployment) GetName() string {
	return d.name
}

func (d *Deployment) GetOwnedPods(ctx context.Context) ([]core.Pod, error) {
	pods := d.GetCachedStatus().Pods()

	podList := make([]core.Pod, 0, len(pods))

	for _, p := range pods {
		if !d.isOwnerOf(p) {
			continue
		}
		c := p.DeepCopy()
		podList = append(podList, *c)
	}

	return podList, nil
}

func (d *Deployment) GetCachedStatus() inspectorInterface.Inspector {
	if c := d.currentState; c != nil {
		return c
	}

	return inspector.NewEmptyInspector()
}

func (d *Deployment) SetCachedStatus(i inspectorInterface.Inspector) {
	d.currentState = i
}

func (d *Deployment) WithArangoMemberUpdate(ctx context.Context, namespace, name string, action reconciler.ArangoMemberUpdateFunc) error {
	o, err := d.deps.Client.Arango().DatabaseV1().ArangoMembers(namespace).Get(ctx, name, meta.GetOptions{})
	if err != nil {
		return err
	}

	if action(o) {
		if _, err := d.deps.Client.Arango().DatabaseV1().ArangoMembers(namespace).Update(ctx, o, meta.UpdateOptions{}); err != nil {
			return err
		}
	}

	return nil
}

func (d *Deployment) WithArangoMemberStatusUpdate(ctx context.Context, namespace, name string, action reconciler.ArangoMemberStatusUpdateFunc) error {
	o, err := d.deps.Client.Arango().DatabaseV1().ArangoMembers(namespace).Get(ctx, name, meta.GetOptions{})
	if err != nil {
		return err
	}

	status := o.Status.DeepCopy()

	if action(o, status) {
		o.Status = *status
		if _, err := d.deps.Client.Arango().DatabaseV1().ArangoMembers(namespace).UpdateStatus(ctx, o, meta.UpdateOptions{}); err != nil {
			return err
		}
	}

	return nil
}

func (d *Deployment) ApplyPatchOnPod(ctx context.Context, pod *core.Pod, p ...patch.Item) error {
	parser := patch.Patch(p)

	data, err := parser.Marshal()
	if err != nil {
		return err
	}

	c := d.deps.Client.Kubernetes().CoreV1().Pods(pod.GetNamespace())

	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	_, err = c.Patch(ctxChild, pod.GetName(), types.JSONPatchType, data, meta.PatchOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (d *Deployment) GenerateMemberEndpoint(group api.ServerGroup, member api.MemberStatus) (string, error) {
	cache := d.GetCachedStatus()

	return pod.GenerateMemberEndpoint(cache, d.GetAPIObject(), d.GetSpec(), group, member)
}

func (d *Deployment) GetStatusSnapshot() api.DeploymentStatus {
	s, _ := d.GetStatus()

	z := s.DeepCopy()

	return *z
}
