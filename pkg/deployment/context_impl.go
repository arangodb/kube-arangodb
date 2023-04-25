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
	"crypto/tls"
	"fmt"
	"net"
	nhttp "net/http"
	"strconv"
	"time"

	core "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/arangosync-client/client"
	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/agency"
	"github.com/arangodb/go-driver/http"
	"github.com/arangodb/go-driver/jwt"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/acs/sutil"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconciler"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/operator/scope"
	"github.com/arangodb/kube-arangodb/pkg/replication"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/conn"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	persistentvolumeclaimv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolumeclaim/v1"
	podv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/pod/v1"
	poddisruptionbudgetv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/poddisruptionbudget/v1"
	secretv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret/v1"
	servicev1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service/v1"
	serviceaccountv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/serviceaccount/v1"
	servicemonitorv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/servicemonitor/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
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
	return d.currentObject
}

// GetServerGroupIterator returns the deployment as ServerGroupIterator.
func (d *Deployment) GetServerGroupIterator() reconciler.ServerGroupIterator {
	return d.currentObject
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
	return d.currentObject.Status.Phase
}

// GetSpec returns the current specification
func (d *Deployment) GetSpec() api.DeploymentSpec {
	d.currentObjectLock.RLock()
	defer d.currentObjectLock.RUnlock()

	if s := d.currentObject.Status.AcceptedSpec; s == nil {
		return d.currentObject.Spec
	} else {
		return *s
	}
}

// GetStatus returns the current status of the deployment
// together with the current version of that status.
func (d *Deployment) GetStatus() api.DeploymentStatus {
	d.currentObjectLock.RLock()
	defer d.currentObjectLock.RUnlock()

	if s := d.currentObjectStatus; s == nil {
		return api.DeploymentStatus{}
	} else {
		return *s.DeepCopy()
	}
}

// UpdateStatus replaces the status of the deployment with the given status and
// updates the resources in k8s.
// If the given last version does not match the actual last version of the status object,
// an error is returned.
func (d *Deployment) UpdateStatus(ctx context.Context, status api.DeploymentStatus) error {
	return d.updateCRStatus(ctx, status)
}

// UpdateMember updates the deployment status wrt the given member.
func (d *Deployment) UpdateMember(ctx context.Context, member api.MemberStatus) error {
	status := d.GetStatus()
	_, group, found := status.Members.ElementByID(member.ID)
	if !found {
		return errors.WithStack(errors.Newf("Member %s not found", member.ID))
	}
	if err := status.Members.Update(member, group); err != nil {
		return errors.WithStack(err)
	}
	if err := d.UpdateStatus(ctx, status); err != nil {
		d.log.Err(err).Debug("Updating CR status failed")
		return errors.WithStack(err)
	}
	return nil
}

// GetDatabaseWithWrap wraps client to the database with provided connection.
func (d *Deployment) GetDatabaseWithWrap(wrappers ...conn.ConnectionWrap) (driver.Client, error) {
	c, err := d.GetMembersState().State().GetDatabaseClient()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	dbConn := c.Connection()

	for _, w := range wrappers {
		if w != nil {
			dbConn = w(dbConn)
		}
	}

	return driver.NewClient(driver.ClientConfig{
		Connection: dbConn,
	})
}

// GetDatabaseAsyncClient returns asynchronous client to the database.
func (d *Deployment) GetDatabaseAsyncClient(ctx context.Context) (driver.Client, error) {
	c, err := d.GetDatabaseWithWrap(conn.NewAsyncConnection)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// GetServerAsyncClient returns an async client for a specific server.
func (d *Deployment) GetServerAsyncClient(id string) (driver.Client, error) {
	c, err := d.GetMembersState().GetMemberClient(id)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return driver.NewClient(driver.ClientConfig{
		Connection: conn.NewAsyncConnection(c.Connection()),
	})
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

// GetAgency returns a connection to the agency.
func (d *Deployment) GetAgency(ctx context.Context, agencyIDs ...string) (agency.Agency, error) {
	return d.clientCache.GetAgency(ctx, agencyIDs...)
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

	if d.GetSpec().TLS.IsSecure() {
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
	if !d.GetSpec().Authentication.IsAuthenticated() {
		return nil, nil
	}

	if !d.GetCachedStatus().Initialised() {
		return nil, errors.Newf("Cache is not yet started")
	}

	var secret string
	var found bool

	// Check if we can find token in folder
	if i := d.currentObject.Status.CurrentImage; i == nil || features.JWTRotation().Supported(i.ArangoDBVersion, i.Enterprise) {
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
	if i := d.currentObject.Status.CurrentImage; i == nil || features.JWTRotation().Supported(i.ArangoDBVersion, i.Enterprise) {
		s, err := d.GetCachedStatus().Secret().V1().Read().Get(context.Background(), pod.JWTSecretFolder(d.GetName()), meta.GetOptions{})
		if err != nil {
			d.log.Err(err).Error("Unable to get secret")
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
	s, err := d.GetCachedStatus().Secret().V1().Read().Get(context.Background(), d.GetSpec().Authentication.GetJWTSecretName(), meta.GetOptions{})
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
	secretName := d.GetSpec().Sync.Monitoring.GetTokenSecretName()
	monitoringToken, err := k8sutil.GetTokenSecret(ctx, d.GetCachedStatus().Secret().V1().Read(), secretName)
	if err != nil {
		d.log.Err(err).Str("secret-name", secretName).Debug("Failed to get sync monitoring secret")
		return nil, errors.WithStack(err)
	}

	// Fetch server DNS name
	dnsName := k8sutil.CreatePodDNSNameWithDomain(d.currentObject, d.GetSpec().ClusterDomain, group.AsRole(), id)

	// Build client
	port := shared.ArangoSyncMasterPort
	if group == api.ServerGroupSyncWorkers {
		port = shared.ArangoSyncWorkerPort
	}
	source := client.Endpoint{"https://" + net.JoinHostPort(dnsName, strconv.Itoa(port))}

	c, err := replication.GetSyncServerClient(&d.syncClientCache, monitoringToken, source)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// CreateMember adds a new member to the given group.
// If ID is non-empty, it will be used, otherwise a new ID is created.
func (d *Deployment) CreateMember(ctx context.Context, group api.ServerGroup, id string, mods ...reconcile.CreateMemberMod) (string, error) {
	if err := d.WithStatusUpdateErr(ctx, func(s *api.DeploymentStatus) (bool, error) {
		nid, err := d.createMember(d.GetSpec(), s, group, id, d.currentObject, mods...)
		if err != nil {
			d.log.Err(err).Str("group", group.AsRole()).Debug("Failed to create member")
			return false, errors.WithStack(err)
		}

		id = nid

		return true, nil
	}); err != nil {
		return "", err
	}

	// Create event about it
	d.CreateEvent(k8sutil.NewMemberAddEvent(id, group.AsRole(), d.currentObject))

	return id, nil
}

// GetPod returns pod.
func (d *Deployment) GetPod(ctx context.Context, podName string) (*core.Pod, error) {
	return d.GetCachedStatus().Pod().V1().Read().Get(ctx, podName, meta.GetOptions{})
}

// DeletePod deletes a pod with given name in the namespace
// of the deployment. If the pod does not exist, the error is ignored.
func (d *Deployment) DeletePod(ctx context.Context, podName string, options meta.DeleteOptions) error {
	log := d.log
	err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return d.PodsModInterface().Delete(ctxChild, podName, options)
	})
	if err != nil && !kerrors.IsNotFound(err) {
		log.Err(err).Str("pod", podName).Debug("Failed to remove pod")
		return errors.WithStack(err)
	}
	return nil
}

// CleanupPod deletes a given pod with force and explicit UID.
// If the pod does not exist, the error is ignored.
func (d *Deployment) CleanupPod(ctx context.Context, p *core.Pod) error {
	log := d.log
	podName := p.GetName()
	options := meta.NewDeleteOptions(0)
	options.Preconditions = meta.NewUIDPreconditions(string(p.GetUID()))
	err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return d.PodsModInterface().Delete(ctxChild, podName, *options)
	})
	if err != nil && !kerrors.IsNotFound(err) {
		log.Err(err).Str("pod", podName).Debug("Failed to cleanup pod")
		return errors.WithStack(err)
	}
	return nil
}

// RemovePodFinalizers removes all the finalizers from the Pod with given name in the namespace
// of the deployment. If the pod does not exist, the error is ignored.
func (d *Deployment) RemovePodFinalizers(ctx context.Context, podName string) error {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	p, err := d.GetCachedStatus().Pod().V1().Read().Get(ctxChild, podName, meta.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return errors.WithStack(err)
	}

	_, err = k8sutil.RemovePodFinalizers(ctx, d.GetCachedStatus(), d.PodsModInterface(), p, p.GetFinalizers(), true)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// DeletePvc deletes a persistent volume claim with given name in the namespace
// of the deployment. If the pvc does not exist, the error is ignored.
func (d *Deployment) DeletePvc(ctx context.Context, pvcName string) error {
	log := d.log
	err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return d.PersistentVolumeClaimsModInterface().Delete(ctxChild, pvcName, meta.DeleteOptions{})
	})
	if err != nil && !kerrors.IsNotFound(err) {
		log.Err(err).Str("pvc", pvcName).Debug("Failed to remove pvc")
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
	pvcs := d.GetCachedStatus().PersistentVolumeClaim().V1().ListSimple()
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

	pvc, err := d.GetCachedStatus().PersistentVolumeClaim().V1().Read().Get(ctxChild, pvcName, meta.GetOptions{})
	if err != nil {
		d.log.Err(err).Str("pvc-name", pvcName).Debug("Failed to get PVC")
		return nil, errors.WithStack(err)
	}
	return pvc, nil
}

// GetTLSKeyfile returns the keyfile encoded TLS certificate+key for
// the given member.
func (d *Deployment) GetTLSKeyfile(group api.ServerGroup, member api.MemberStatus) (string, error) {
	secretName := k8sutil.CreateTLSKeyfileSecretName(d.GetName(), group.AsRole(), member.ID)
	result, err := k8sutil.GetTLSKeyfileSecret(d.GetCachedStatus().Secret().V1().Read(), secretName)
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
	if err != nil && !kerrors.IsNotFound(err) {
		return errors.WithStack(err)
	}
	return nil
}

// DeleteSecret removes the Secret with given name.
// If the secret does not exist, the error is ignored.
func (d *Deployment) DeleteSecret(secretName string) error {
	if err := d.SecretsModInterface().Delete(context.Background(), secretName, meta.DeleteOptions{}); err != nil && !kerrors.IsNotFound(err) {
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

// GetAgencyData returns agency plan.
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

func (d *Deployment) RenderPodForMember(ctx context.Context, acs sutil.ACS, spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*core.Pod, error) {
	return d.resources.RenderPodForMember(ctx, acs, spec, status, memberID, imageInfo)
}

func (d *Deployment) RenderPodTemplateForMember(ctx context.Context, acs sutil.ACS, spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*core.PodTemplateSpec, error) {
	return d.resources.RenderPodTemplateForMember(ctx, acs, spec, status, memberID, imageInfo)
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

func (d *Deployment) WithStatusUpdateErr(ctx context.Context, action reconciler.DeploymentStatusUpdateErrFunc) error {
	status := d.GetStatus()

	changed, err := action(&status)

	if err != nil {
		return err
	}

	if !changed {
		return nil
	}

	return d.updateCRStatus(ctx, status)
}

func (d *Deployment) WithStatusUpdate(ctx context.Context, action reconciler.DeploymentStatusUpdateFunc) error {
	return d.WithStatusUpdateErr(ctx, func(s *api.DeploymentStatus) (bool, error) {
		return action(s), nil
	})
}

func (d *Deployment) SecretsModInterface() secretv1.ModInterface {
	d.acs.CurrentClusterCache().GetThrottles().Secret().Invalidate()
	return kclient.NewModInterface(d.deps.Client, d.namespace).Secrets()
}

func (d *Deployment) PodsModInterface() podv1.ModInterface {
	d.acs.CurrentClusterCache().GetThrottles().Pod().Invalidate()
	return kclient.NewModInterface(d.deps.Client, d.namespace).Pods()
}

func (d *Deployment) ServiceAccountsModInterface() serviceaccountv1.ModInterface {
	d.acs.CurrentClusterCache().GetThrottles().ServiceAccount().Invalidate()
	return kclient.NewModInterface(d.deps.Client, d.namespace).ServiceAccounts()
}

func (d *Deployment) ServicesModInterface() servicev1.ModInterface {
	d.acs.CurrentClusterCache().GetThrottles().Service().Invalidate()
	return kclient.NewModInterface(d.deps.Client, d.namespace).Services()
}

func (d *Deployment) PersistentVolumeClaimsModInterface() persistentvolumeclaimv1.ModInterface {
	d.acs.CurrentClusterCache().GetThrottles().PersistentVolumeClaim().Invalidate()
	return kclient.NewModInterface(d.deps.Client, d.namespace).PersistentVolumeClaims()
}

func (d *Deployment) PodDisruptionBudgetsModInterface() poddisruptionbudgetv1.ModInterface {
	d.acs.CurrentClusterCache().GetThrottles().PodDisruptionBudget().Invalidate()
	return kclient.NewModInterface(d.deps.Client, d.namespace).PodDisruptionBudgets()
}

func (d *Deployment) ServiceMonitorsModInterface() servicemonitorv1.ModInterface {
	d.acs.CurrentClusterCache().GetThrottles().ServiceMonitor().Invalidate()
	return kclient.NewModInterface(d.deps.Client, d.namespace).ServiceMonitors()
}

func (d *Deployment) GetName() string {
	return d.name
}

func (d *Deployment) GetOwnedPods(ctx context.Context) ([]core.Pod, error) {
	pods := d.GetCachedStatus().Pod().V1().ListSimple()

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
	return d.acs.CurrentClusterCache()
}

func (d *Deployment) ApplyPatchOnPod(ctx context.Context, pod *core.Pod, p ...patch.Item) error {
	parser := patch.Patch(p)

	data, err := parser.Marshal()
	if err != nil {
		return err
	}

	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	_, err = d.PodsModInterface().Patch(ctxChild, pod.GetName(), types.JSONPatchType, data, meta.PatchOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (d *Deployment) GenerateMemberEndpoint(group api.ServerGroup, member api.MemberStatus) (string, error) {
	cache := d.GetCachedStatus()

	return pod.GenerateMemberEndpoint(cache, d.GetAPIObject(), d.GetSpec(), group, member)
}

func (d *Deployment) ACS() sutil.ACS {
	return d.acs
}

func (d *Deployment) CreateOperatorEngineOpsAlertEvent(message string, args ...interface{}) {
	if d == nil {
		return
	}

	d.metrics.ArangodbOperatorEngineOpsAlerts++

	d.CreateEvent(k8sutil.NewOperatorEngineOpsAlertEvent(fmt.Sprintf(message, args...), d.GetAPIObject()))
}

func (d *Deployment) WithMemberStatusUpdateErr(ctx context.Context, id string, group api.ServerGroup, action reconciler.DeploymentMemberStatusUpdateErrFunc) error {
	return d.WithStatusUpdateErr(ctx, func(s *api.DeploymentStatus) (bool, error) {
		m, g, ok := s.Members.ElementByID(id)
		if !ok {
			return false, errors.Newf("Member %s not found", id)
		}

		if g != group {
			return false, errors.Newf("Invalid group for %s. Wanted %s, found %s", id, group.AsRole(), g.AsRole())
		}

		changed, err := action(&m)
		if err != nil {
			return false, err
		}

		if !changed {
			return false, nil
		}

		if err := s.Members.Update(m, g); err != nil {
			return false, err
		}

		return true, nil
	})
}

func (d *Deployment) WithMemberStatusUpdate(ctx context.Context, id string, group api.ServerGroup, action reconciler.DeploymentMemberStatusUpdateFunc) error {
	return d.WithMemberStatusUpdateErr(ctx, id, group, func(s *api.MemberStatus) (bool, error) {
		return action(s), nil
	})
}
