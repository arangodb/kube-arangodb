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

package deployment

import (
	"context"
	"crypto/tls"
	"net"
	nhttp "net/http"
	"strconv"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/arangodb/kube-arangodb/pkg/util/arangod/conn"

	"github.com/arangodb/kube-arangodb/pkg/operator/scope"

	monitoringClient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"

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
	"github.com/rs/zerolog/log"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	core "k8s.io/api/core/v1"
)

var _ resources.Context = &Deployment{}

// GetBackup receives information about a backup resource
func (d *Deployment) GetBackup(ctx context.Context, backup string) (*backupApi.ArangoBackup, error) {
	ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
	defer cancel()

	return d.deps.DatabaseCRCli.BackupV1().ArangoBackups(d.Namespace()).Get(ctxChild, backup, meta.GetOptions{})
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

func (d *Deployment) GetMonitoringV1Cli() monitoringClient.MonitoringV1Interface {
	return d.deps.KubeMonitoringCli
}

func (d *Deployment) GetArangoCli() versioned.Interface {
	return d.deps.DatabaseCRCli
}

func (d *Deployment) GetScope() scope.Scope {
	return d.config.Scope
}

// GetLifecycleImage returns the image name containing the lifecycle helper (== name of operator image)
func (d *Deployment) GetLifecycleImage() string {
	return d.config.LifecycleImage
}

// GetOperatorUUIDImage returns the image name containing the uuid helper (== name of operator image)
func (d *Deployment) GetOperatorUUIDImage() string {
	return d.config.OperatorUUIDInitImage
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
	d.status.version++
	d.status.last = *status.DeepCopy()
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
			Timeout:   30 * time.Second,
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

	var secrets secret.ReadInterface = d.GetKubeCli().CoreV1().Secrets(d.GetNamespace())
	if currentState := d.currentState; currentState != nil {
		secrets = currentState.SecretReadInterface()
	}

	var secret string
	if i := d.apiObject.Status.CurrentImage; i == nil || !features.JWTRotation().Supported(i.ArangoDBVersion, i.Enterprise) {
		s, err := secrets.Get(context.Background(), d.apiObject.Spec.Authentication.GetJWTSecretName(), meta.GetOptions{})
		if err != nil {
			return nil, errors.Newf("JWT Secret is missing")
		}

		jwt, ok := s.Data[constants.SecretKeyToken]
		if !ok {
			return nil, errors.Newf("JWT Secret is invalid")
		}

		secret = string(jwt)
	} else {
		s, err := secrets.Get(context.Background(), pod.JWTSecretFolder(d.GetName()), meta.GetOptions{})
		if err != nil {
			d.deps.Log.Error().Err(err).Msgf("Unable to get secret")
			return nil, errors.Newf("JWT Folder Secret is missing")
		}

		if len(s.Data) == 0 {
			return nil, errors.Newf("JWT Folder Secret is empty")
		}

		if q, ok := s.Data[pod.ActiveJWTKey]; ok {
			secret = string(q)
		} else {
			for _, q := range s.Data {
				secret = string(q)
				break
			}
		}
	}

	jwt, err := jwt.CreateArangodJwtAuthorizationHeader(secret, "kube-arangodb")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return driver.RawAuthentication(jwt), nil
}

// GetSyncServerClient returns a cached client for a specific arangosync server.
func (d *Deployment) GetSyncServerClient(ctx context.Context, group api.ServerGroup, id string) (client.API, error) {
	// Fetch monitoring token
	log := d.deps.Log
	kubecli := d.deps.KubeCli
	ns := d.GetNamespace()
	secrets := kubecli.CoreV1().Secrets(ns)
	secretName := d.apiObject.Spec.Sync.Monitoring.GetTokenSecretName()
	monitoringToken, err := k8sutil.GetTokenSecret(ctx, secrets, secretName)
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
func (d *Deployment) CreateMember(ctx context.Context, group api.ServerGroup, id string) (string, error) {
	log := d.deps.Log
	status, lastVersion := d.GetStatus()
	id, err := createMember(log, &status, group, id, d.apiObject)
	if err != nil {
		log.Debug().Err(err).Str("group", group.AsRole()).Msg("Failed to create member")
		return "", errors.WithStack(err)
	}
	// Save added member
	if err := d.UpdateStatus(ctx, status, lastVersion); err != nil {
		log.Debug().Err(err).Msg("Updating CR status failed")
		return "", errors.WithStack(err)
	}
	// Create event about it
	d.CreateEvent(k8sutil.NewMemberAddEvent(id, group.AsRole(), d.apiObject))

	return id, nil
}

// GetPod returns pod.
func (d *Deployment) GetPod(ctx context.Context, podName string) (*core.Pod, error) {
	ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
	defer cancel()

	return d.deps.KubeCli.CoreV1().Pods(d.GetNamespace()).Get(ctxChild, podName, meta.GetOptions{})
}

// DeletePod deletes a pod with given name in the namespace
// of the deployment. If the pod does not exist, the error is ignored.
func (d *Deployment) DeletePod(ctx context.Context, podName string) error {
	log := d.deps.Log
	ns := d.GetNamespace()
	err := k8sutil.RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return d.deps.KubeCli.CoreV1().Pods(ns).Delete(ctxChild, podName, meta.DeleteOptions{})
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
	ns := p.GetNamespace()
	options := meta.NewDeleteOptions(0)
	options.Preconditions = meta.NewUIDPreconditions(string(p.GetUID()))
	err := k8sutil.RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return d.deps.KubeCli.CoreV1().Pods(ns).Delete(ctxChild, podName, *options)
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
	ns := d.GetNamespace()
	kubecli := d.deps.KubeCli

	ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
	defer cancel()
	p, err := kubecli.CoreV1().Pods(ns).Get(ctxChild, podName, meta.GetOptions{})
	if err != nil {
		if k8sutil.IsNotFound(err) {
			return nil
		}
		return errors.WithStack(err)
	}

	err = k8sutil.RemovePodFinalizers(ctx, log, d.deps.KubeCli, p, p.GetFinalizers(), true)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// DeletePvc deletes a persistent volume claim with given name in the namespace
// of the deployment. If the pvc does not exist, the error is ignored.
func (d *Deployment) DeletePvc(ctx context.Context, pvcName string) error {
	log := d.deps.Log
	ns := d.GetNamespace()
	err := k8sutil.RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return d.deps.KubeCli.CoreV1().PersistentVolumeClaims(ns).Delete(ctxChild, pvcName, meta.DeleteOptions{})
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
	err := k8sutil.RunWithTimeout(ctx, func(ctxChild context.Context) error {
		_, err := d.GetKubeCli().CoreV1().PersistentVolumeClaims(d.GetNamespace()).Update(ctxChild, pvc, meta.UpdateOptions{})
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
	log := d.deps.Log
	pvcs, err := d.deps.KubeCli.CoreV1().PersistentVolumeClaims(d.GetNamespace()).List(context.Background(), k8sutil.DeploymentListOpt(d.GetName()))
	if err != nil {
		log.Debug().Err(err).Msg("Failed to list PVCs")
		return nil, errors.WithStack(err)
	}
	myPVCs := make([]core.PersistentVolumeClaim, 0, len(pvcs.Items))
	for _, p := range pvcs.Items {
		if d.isOwnerOf(&p) {
			myPVCs = append(myPVCs, p)
		}
	}
	return myPVCs, nil
}

// GetPvc gets a PVC by the given name, in the samespace of the deployment.
func (d *Deployment) GetPvc(ctx context.Context, pvcName string) (*core.PersistentVolumeClaim, error) {
	ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
	defer cancel()

	pvc, err := d.deps.KubeCli.CoreV1().PersistentVolumeClaims(d.GetNamespace()).Get(ctxChild, pvcName, meta.GetOptions{})
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
	secrets := d.deps.KubeCli.CoreV1().Secrets(d.GetNamespace())
	result, err := k8sutil.GetTLSKeyfileSecret(secrets, secretName)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return result, nil
}

// DeleteTLSKeyfile removes the Secret containing the TLS keyfile for the given member.
// If the secret does not exist, the error is ignored.
func (d *Deployment) DeleteTLSKeyfile(ctx context.Context, group api.ServerGroup, member api.MemberStatus) error {
	secretName := k8sutil.CreateTLSKeyfileSecretName(d.GetName(), group.AsRole(), member.ID)
	ns := d.GetNamespace()
	err := k8sutil.RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return d.deps.KubeCli.CoreV1().Secrets(ns).Delete(ctxChild, secretName, meta.DeleteOptions{})
	})
	if err != nil && !k8sutil.IsNotFound(err) {
		return errors.WithStack(err)
	}
	return nil
}

// DeleteSecret removes the Secret with given name.
// If the secret does not exist, the error is ignored.
func (d *Deployment) DeleteSecret(secretName string) error {
	ns := d.GetNamespace()
	if err := d.deps.KubeCli.CoreV1().Secrets(ns).Delete(context.Background(), secretName, meta.DeleteOptions{}); err != nil && !k8sutil.IsNotFound(err) {
		return errors.WithStack(err)
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

func (d *Deployment) RenderPodTemplateForMember(ctx context.Context, cachedStatus inspectorInterface.Inspector, spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*core.PodTemplateSpec, error) {
	return d.resources.RenderPodTemplateForMember(ctx, cachedStatus, spec, status, memberID, imageInfo)
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

func (d *Deployment) WithStatusUpdate(ctx context.Context, action resources.DeploymentStatusUpdateFunc, force ...bool) error {
	d.status.mutex.Lock()
	defer d.status.mutex.Unlock()

	status, version := d.getStatus()

	changed := action(&status)

	if !changed {
		return nil
	}

	return d.updateStatus(ctx, status, version, force...)
}

func (d *Deployment) SecretsInterface() k8sutil.SecretInterface {
	return d.GetKubeCli().CoreV1().Secrets(d.GetNamespace())
}

func (d *Deployment) GetName() string {
	return d.name
}

func (d *Deployment) GetOwnedPods(ctx context.Context) ([]core.Pod, error) {
	ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
	defer cancel()

	pods, err := d.GetKubeCli().CoreV1().Pods(d.GetNamespace()).List(ctxChild, meta.ListOptions{})
	if err != nil {
		return nil, err
	}

	podList := make([]core.Pod, 0, len(pods.Items))

	for _, p := range pods.Items {
		if !d.isOwnerOf(&p) {
			continue
		}
		c := p.DeepCopy()
		podList = append(podList, *c)
	}

	return podList, nil
}

func (d *Deployment) GetCachedStatus() inspectorInterface.Inspector {
	return d.currentState
}

func (d *Deployment) SetCachedStatus(i inspectorInterface.Inspector) {
	d.currentState = i
}

func (d *Deployment) WithArangoMemberUpdate(ctx context.Context, namespace, name string, action resources.ArangoMemberUpdateFunc) error {
	o, err := d.deps.DatabaseCRCli.DatabaseV1().ArangoMembers(namespace).Get(ctx, name, meta.GetOptions{})
	if err != nil {
		return err
	}

	if action(o) {
		if _, err := d.deps.DatabaseCRCli.DatabaseV1().ArangoMembers(namespace).Update(ctx, o, meta.UpdateOptions{}); err != nil {
			return err
		}
	}

	return nil
}

func (d *Deployment) WithArangoMemberStatusUpdate(ctx context.Context, namespace, name string, action resources.ArangoMemberStatusUpdateFunc) error {
	o, err := d.deps.DatabaseCRCli.DatabaseV1().ArangoMembers(namespace).Get(ctx, name, meta.GetOptions{})
	if err != nil {
		return err
	}

	status := o.Status.DeepCopy()

	if action(o, status) {
		o.Status = *status
		if _, err := d.deps.DatabaseCRCli.DatabaseV1().ArangoMembers(namespace).UpdateStatus(ctx, o, meta.UpdateOptions{}); err != nil {
			return err
		}
	}

	return nil
}
