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

package tests

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/arangodb/arangosync-client/client"
	"github.com/arangodb/arangosync-client/tasks"
	driver "github.com/arangodb/go-driver"
	vst "github.com/arangodb/go-driver/vst"
	vstProtocol "github.com/arangodb/go-driver/vst/protocol"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	rapi "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1alpha"
	cl "github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/retry"
)

const (
	deploymentReadyTimeout   = time.Minute * 4
	deploymentUpgradeTimeout = time.Minute * 20
)

var (
	maskAny                 = errors.WithStack
	syncClientCache         client.ClientCache
	showEnterpriseImageOnce sync.Once
)

// CreateArangodClientForDNSName creates a go-driver client for a given DNS name.
func createArangodVSTClientForDNSName(ctx context.Context, cli corev1.CoreV1Interface, apiObject *api.ArangoDeployment, dnsName string, shortTimeout bool) (driver.Client, error) {
	config := driver.ClientConfig{}
	connConfig, err := createArangodVSTConfigForDNSNames(ctx, cli, apiObject, []string{dnsName}, shortTimeout)
	if err != nil {
		return nil, maskAny(err)
	}
	// TODO deal with TLS with proper CA checking
	conn, err := vst.NewConnection(connConfig)
	if err != nil {
		return nil, maskAny(err)
	}

	// Create client
	config = driver.ClientConfig{
		Connection: conn,
	}

	auth := driver.BasicAuthentication("root", "")
	if err != nil {
		return nil, maskAny(err)
	}
	config.Authentication = auth
	c, err := driver.NewClient(config)
	if err != nil {
		return nil, maskAny(err)
	}
	return c, nil
}

// createArangodVSTConfigForDNSNames creates a go-driver VST connection config for a given DNS names.
func createArangodVSTConfigForDNSNames(ctx context.Context, cli corev1.CoreV1Interface, apiObject *api.ArangoDeployment, dnsNames []string, shortTimeout bool) (vst.ConnectionConfig, error) {
	scheme := "http"
	tlsConfig := &tls.Config{}
	timeout := 90 * time.Second
	if shortTimeout {
		timeout = 100 * time.Millisecond
	}
	if apiObject != nil && apiObject.Spec.IsSecure() {
		scheme = "https"
		tlsConfig = &tls.Config{InsecureSkipVerify: true}
	}
	transport := vstProtocol.TransportConfig{
		IdleConnTimeout: timeout,
		Version:         vstProtocol.Version1_1,
	}
	connConfig := vst.ConnectionConfig{
		TLSConfig: tlsConfig,
		Transport: transport,
	}
	for _, dnsName := range dnsNames {
		connConfig.Endpoints = append(connConfig.Endpoints, scheme+"://"+net.JoinHostPort(dnsName, strconv.Itoa(k8sutil.ArangoPort)))
	}
	return connConfig, nil
}

// CreateArangodDatabaseVSTClient creates a go-driver client for accessing the entire cluster (or single server) via VST
func createArangodDatabaseVSTClient(ctx context.Context, cli corev1.CoreV1Interface, apiObject *api.ArangoDeployment, shortTimeout bool) (driver.Client, error) {
	// Create connection
	dnsName := k8sutil.CreateDatabaseClientServiceDNSName(apiObject)
	c, err := createArangodVSTClientForDNSName(ctx, cli, apiObject, dnsName, shortTimeout)
	if err != nil {
		return nil, maskAny(err)
	}
	return c, nil
}

// longOrSkip checks the short test flag.
// If short is set, the current test is skipped.
// If not, this function returns as normal.
func longOrSkip(t *testing.T) {
	if testing.Short() {
		t.Skip("Test skipped in short test")
	}
}

// getEnterpriseImageOrSkip returns the docker image used for enterprise
// tests. If empty, enterprise tests are skipped.
func getEnterpriseImageOrSkip(t *testing.T) string {
	image := strings.TrimSpace(os.Getenv("ENTERPRISEIMAGE"))
	if image == "" {
		t.Skip("Skipping test because ENTERPRISEIMAGE is not set")
	} else {
		showEnterpriseImageOnce.Do(func() {
			t.Logf("Using enterprise image: %s", image)
		})
	}
	return image
}

const testEnterpriseLicenseKeySecretName = "arangodb-jenkins-license-key"

func getEnterpriseLicenseKey() string {
	return strings.TrimSpace(os.Getenv("ENTERPRISELICENSE"))
}

// shouldCleanDeployments returns true when deployments created
// by tests should be removed, even when the test fails.
func shouldCleanDeployments() bool {
	return os.Getenv("CLEANDEPLOYMENTS") != ""
}

// mustNewKubeClient creates a kubernetes client
// failing the test on errors.
func mustNewKubeClient(t *testing.T) kubernetes.Interface {
	c, err := k8sutil.NewKubeClient()
	if err != nil {
		t.Fatalf("Failed to create kube cli: %v", err)
	}
	return c
}

// DatabaseClientOptions contains options for creating an ArangoDB database client.
type DatabaseClientOptions struct {
	ShortTimeout bool // If set, the connection timeout is set very short
	UseVST       bool // If set, a VST connection is created instead of an HTTP connection
}

// mustNewArangodDatabaseClient creates a new database client,
// failing the test on errors.
func mustNewArangodDatabaseClient(ctx context.Context, kubecli kubernetes.Interface, apiObject *api.ArangoDeployment, t *testing.T, options *DatabaseClientOptions) driver.Client {
	var c driver.Client
	var err error
	shortTimeout := options != nil && options.ShortTimeout
	useVST := options != nil && options.UseVST
	if useVST {
		c, err = createArangodDatabaseVSTClient(ctx, kubecli.CoreV1(), apiObject, shortTimeout)
	} else {
		c, err = arangod.CreateArangodDatabaseClient(ctx, kubecli.CoreV1(), apiObject, shortTimeout)
	}
	if err != nil {
		t.Fatalf("Failed to create arango database client: %v", err)
	}
	return c
}

// mustNewArangoSyncClient creates a new arangosync client, with all syncmasters
// as endpoint. It is failing the test on errors.
func mustNewArangoSyncClient(ctx context.Context, kubecli kubernetes.Interface, apiObject *api.ArangoDeployment, t *testing.T) client.API {
	ns := apiObject.GetNamespace()
	secrets := kubecli.CoreV1().Secrets(ns)
	secretName := apiObject.Spec.Sync.Authentication.GetJWTSecretName()
	jwtToken, err := k8sutil.GetTokenSecret(secrets, secretName)
	if err != nil {
		t.Fatalf("Failed to get sync jwt secret '%s': %s", secretName, err)
	}

	// Fetch service DNS name
	dnsName := k8sutil.CreateSyncMasterClientServiceDNSName(apiObject)
	ep := client.Endpoint{"https://" + net.JoinHostPort(dnsName, strconv.Itoa(k8sutil.ArangoSyncMasterPort))}

	// Build client
	log := zerolog.Logger{}
	tlsAuth := tasks.TLSAuthentication{}
	auth := client.NewAuthentication(tlsAuth, jwtToken)
	insecureSkipVerify := true
	c, err := syncClientCache.GetClient(log, ep, auth, insecureSkipVerify)
	if err != nil {
		t.Fatalf("Failed to get sync client: %s", err)
	}
	return c
}

// getNamespace returns the kubernetes namespace in which to run tests.
func getNamespace(t *testing.T) string {
	ns := os.Getenv("TEST_NAMESPACE")
	if ns == "" {
		t.Fatal("Missing environment variable TEST_NAMESPACE")
	}
	return ns
}

func newReplication(name string) *rapi.ArangoDeploymentReplication {
	repl := &rapi.ArangoDeploymentReplication{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rapi.SchemeGroupVersion.String(),
			Kind:       rapi.ArangoDeploymentReplicationResourceKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.ToLower(name),
		},
	}

	return repl
}

// newDeployment creates a basic ArangoDeployment with configured
// type, name and image.
func newDeployment(name string) *api.ArangoDeployment {
	depl := &api.ArangoDeployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: api.SchemeGroupVersion.String(),
			Kind:       api.ArangoDeploymentResourceKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.ToLower(name),
		},
		Spec: api.DeploymentSpec{
			ImagePullPolicy: util.NewPullPolicy(v1.PullAlways),
			License: api.LicenseSpec{
				SecretName: util.NewString(testEnterpriseLicenseKeySecretName),
			},
		},
	}

	// set default image to the value given in env
	// some tests will override this value if they need a specific version
	// like update tests
	// if no value is given, use the operator default, which is arangodb/arangodb:latest
	image := strings.TrimSpace(os.Getenv("ARANGODIMAGE"))
	if image != "" {
		depl.Spec.Image = util.NewString(image)
	}

	disableIPv6 := strings.TrimSpace(os.Getenv("TESTDISABLEIPV6"))
	if disableIPv6 != "" && disableIPv6 != "0" {
		depl.Spec.DisableIPv6 = util.NewBool(true)
	}

	return depl
}

// waitUntilDeployment waits until a deployment with given name in given namespace
// reached a state where the given predicate returns true.
func waitUntilDeployment(cli versioned.Interface, deploymentName, ns string, predicate func(*api.ArangoDeployment) error, timeout ...time.Duration) (*api.ArangoDeployment, error) {
	var result *api.ArangoDeployment
	op := func() error {
		obj, err := cli.DatabaseV1alpha().ArangoDeployments(ns).Get(deploymentName, metav1.GetOptions{})
		if err != nil {
			result = nil
			return maskAny(err)
		}
		result = obj
		if predicate != nil {
			if err := predicate(obj); err != nil {
				return maskAny(err)
			}
		}
		return nil
	}
	actualTimeout := deploymentReadyTimeout
	if len(timeout) > 0 {
		actualTimeout = timeout[0]
	}
	if err := retry.Retry(op, actualTimeout); err != nil {
		return nil, maskAny(err)
	}
	return result, nil
}

// waitUntilSecret waits until a secret with given name in given namespace
// reached a state where the given predicate returns true.
func waitUntilSecret(cli kubernetes.Interface, secretName, ns string, predicate func(*v1.Secret) error, timeout time.Duration) (*v1.Secret, error) {
	var result *v1.Secret
	op := func() error {
		obj, err := cli.CoreV1().Secrets(ns).Get(secretName, metav1.GetOptions{})
		if err != nil {
			result = nil
			return maskAny(err)
		}
		result = obj
		if predicate != nil {
			if err := predicate(obj); err != nil {
				return maskAny(err)
			}
		}
		return nil
	}
	if err := retry.Retry(op, timeout); err != nil {
		return nil, maskAny(err)
	}
	return result, nil
}

// waitUntilService waits until a service with given name in given
// namespace exists and has reached a state where the given predicate
// returns nil.
func waitUntilService(cli kubernetes.Interface, serviceName, ns string, predicate func(*v1.Service) error, timeout time.Duration) (*v1.Service, error) {
	var result *v1.Service
	op := func() error {
		obj, err := cli.CoreV1().Services(ns).Get(serviceName, metav1.GetOptions{})
		if err != nil {
			result = nil
			return maskAny(err)
		}
		result = obj
		if predicate != nil {
			if err := predicate(obj); err != nil {
				return maskAny(err)
			}
		}
		return nil
	}
	if err := retry.Retry(op, timeout); err != nil {
		return nil, maskAny(err)
	}
	return result, nil
}

// waitUntilEndpoints waits until an endpoints resource with given name
// in given namespace exists and has reached a state where the given
// predicate returns nil.
func waitUntilEndpoints(cli kubernetes.Interface, serviceName, ns string, predicate func(*v1.Endpoints) error, timeout time.Duration) (*v1.Endpoints, error) {
	var result *v1.Endpoints
	op := func() error {
		obj, err := cli.CoreV1().Endpoints(ns).Get(serviceName, metav1.GetOptions{})
		if err != nil {
			result = nil
			return maskAny(err)
		}
		result = obj
		if predicate != nil {
			if err := predicate(obj); err != nil {
				return maskAny(err)
			}
		}
		return nil
	}
	if err := retry.Retry(op, timeout); err != nil {
		return nil, maskAny(err)
	}
	return result, nil
}

// waitUntilSecretNotFound waits until a secret with given name in given namespace
// is no longer found.
func waitUntilSecretNotFound(cli kubernetes.Interface, secretName, ns string, timeout time.Duration) error {
	op := func() error {
		if _, err := cli.CoreV1().Secrets(ns).Get(secretName, metav1.GetOptions{}); k8sutil.IsNotFound(err) {
			return nil
		} else if err != nil {
			return maskAny(err)
		}
		return maskAny(fmt.Errorf("Secret %s still there", secretName))
	}
	if err := retry.Retry(op, timeout); err != nil {
		return maskAny(err)
	}
	return nil
}

// waitUntilClusterHealth waits until an arango cluster
// reached a state where the given predicate returns nil.
func waitUntilClusterHealth(cli driver.Client, predicate func(driver.ClusterHealth) error, timeout ...time.Duration) error {
	ctx := context.Background()
	op := func() error {
		cluster, err := cli.Cluster(ctx)
		if err != nil {
			return maskAny(err)
		}
		h, err := cluster.Health(ctx)
		if err != nil {
			return maskAny(err)
		}
		if predicate != nil {
			if err := predicate(h); err != nil {
				return maskAny(err)
			}
		}
		return nil
	}
	actualTimeout := deploymentReadyTimeout
	if len(timeout) > 0 {
		actualTimeout = timeout[0]
	}
	if err := retry.Retry(op, actualTimeout); err != nil {
		return maskAny(err)
	}
	return nil
}

// waitUntilClusterVersionUp waits until an arango cluster is healthy and
// all servers are running the given version.
func waitUntilClusterVersionUp(cli driver.Client, version driver.Version) error {
	return waitUntilClusterHealth(cli, func(h driver.ClusterHealth) error {
		for s, r := range h.Health {
			if cmp := r.Version.CompareTo(version); cmp != 0 {
				return maskAny(fmt.Errorf("Member %s has version %s, expecting %s", s, r.Version, version))
			}
		}

		return nil
	}, deploymentUpgradeTimeout)
}

// waitUntilVersionUp waits until the arango database responds to
// an `/_api/version` request without an error. An additional Predicate
// can do a check on the VersionInfo object returned by the server.
func waitUntilVersionUp(cli driver.Client, predicate func(driver.VersionInfo) error, allowNoLeaderResponse ...bool) error {
	var noLeaderErr error
	allowNoLead := len(allowNoLeaderResponse) > 0 && allowNoLeaderResponse[0]
	ctx := context.Background()

	op := func() error {
		if version, err := cli.Version(ctx); allowNoLead && driver.IsNoLeader(err) {
			noLeaderErr = err
			return nil //return nil to make the retry below pass
		} else if err != nil {
			return maskAny(err)
		} else if predicate != nil {
			return predicate(version)
		}
		return nil
	}

	if err := retry.Retry(op, deploymentReadyTimeout); err != nil {
		return maskAny(err)
	}

	// noLeadErr updated in op
	if noLeaderErr != nil {
		return maskAny(noLeaderErr)
	}

	return nil
}

// waitUntilSyncVersionUp waits until the syncmasters responds to
// an `/_api/version` request without an error. An additional Predicate
// can do a check on the VersionInfo object returned by the server.
func waitUntilSyncVersionUp(cli client.API, predicate func(client.VersionInfo) error) error {
	ctx := context.Background()

	op := func() error {
		if version, err := cli.Version(ctx); err != nil {
			return maskAny(err)
		} else if predicate != nil {
			return predicate(version)
		}
		return nil
	}

	if err := retry.Retry(op, deploymentReadyTimeout); err != nil {
		return maskAny(err)
	}

	return nil
}

// waitUntilSyncMasterCountReached waits until the number of syncmasters
// is equal to the given number.
func waitUntilSyncMasterCountReached(cli client.API, expectedSyncMasters int) error {
	ctx := context.Background()

	op := func() error {
		if list, err := cli.Master().Masters(ctx); err != nil {
			return maskAny(err)
		} else if len(list) != expectedSyncMasters {
			return maskAny(fmt.Errorf("Expected %d syncmasters, got %d", expectedSyncMasters, len(list)))
		}
		return nil
	}

	if err := retry.Retry(op, deploymentReadyTimeout); err != nil {
		return maskAny(err)
	}

	return nil
}

// waitUntilSyncWorkerCountReached waits until the number of syncworkers
// is equal to the given number.
func waitUntilSyncWorkerCountReached(cli client.API, expectedSyncWorkers int) error {
	ctx := context.Background()

	op := func() error {
		if list, err := cli.Master().RegisteredWorkers(ctx); err != nil {
			return maskAny(err)
		} else if len(list) != expectedSyncWorkers {
			return maskAny(fmt.Errorf("Expected %d syncworkers, got %d", expectedSyncWorkers, len(list)))
		}
		return nil
	}

	if err := retry.Retry(op, deploymentReadyTimeout); err != nil {
		return maskAny(err)
	}

	return nil
}

// creates predicate to be used in waitUntilVersionUp
func createEqualVersionsPredicate(version driver.Version) func(driver.VersionInfo) error {
	return func(infoFromServer driver.VersionInfo) error {
		if version.CompareTo(infoFromServer.Version) != 0 {
			return maskAny(fmt.Errorf("given version %v and version from server %v do not match", version, infoFromServer.Version))
		}
		return nil
	}
}

// clusterSidecarsEqualSpec returns nil if sidecars from spec and cluster match
func waitUntilClusterSidecarsEqualSpec(t *testing.T, spec api.DeploymentMode, depl api.ArangoDeployment) error {

	c := cl.MustNewInCluster()
	ns := getNamespace(t)

	var noGood int
	for start := time.Now(); time.Since(start) < 300*time.Second; {

		// Fetch latest status so we know all member details
		apiObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Get(depl.GetName(), metav1.GetOptions{})
		if err != nil {
			t.Fatalf("Failed to get deployment: %v", err)
		}

		// How many pods not matching
		noGood = 0

		// Check member after another
		apiObject.ForeachServerGroup(func(group api.ServerGroup, spec api.ServerGroupSpec, status *api.MemberStatusList) error {
			for _, m := range *status {
				for _, scar := range spec.GetSidecars() {
					mcar, found := m.SideCarSpecs[scar.Name]
					if found {
						if !reflect.DeepEqual(mcar, scar) {
							noGood++
						}
					} else {
						noGood++
					}
				}
			}
			return nil
		}, &apiObject.Status)

		if noGood == 0 {
			return nil
		}

		time.Sleep(2 * time.Second)
	}

	return maskAny(fmt.Errorf("%d pods with unmatched sidecars", noGood))
}

// clusterHealthEqualsSpec returns nil when the given health matches
// with the given deployment spec.
func clusterHealthEqualsSpec(h driver.ClusterHealth, spec api.DeploymentSpec) error {
	agents := 0
	goodDBServers := 0
	goodCoordinators := 0
	for _, s := range h.Health {
		if s.Role == driver.ServerRoleAgent {
			agents++
		} else if s.Status == driver.ServerStatusGood {
			switch s.Role {
			case driver.ServerRoleDBServer:
				goodDBServers++
			case driver.ServerRoleCoordinator:
				goodCoordinators++
			}
		}
	}
	if spec.Agents.GetCount() == agents &&
		spec.DBServers.GetCount() == goodDBServers &&
		spec.Coordinators.GetCount() == goodCoordinators {
		return nil
	}
	return fmt.Errorf("Expected %d,%d,%d got %d,%d,%d",
		spec.Agents.GetCount(), spec.DBServers.GetCount(), spec.Coordinators.GetCount(),
		agents, goodDBServers, goodCoordinators,
	)
}

// updateDeployment updates a deployment
func updateDeployment(cli versioned.Interface, deploymentName, ns string, update func(*api.DeploymentSpec)) (*api.ArangoDeployment, error) {
	for {
		// Get current version
		current, err := cli.DatabaseV1alpha().ArangoDeployments(ns).Get(deploymentName, metav1.GetOptions{})
		if err != nil {
			return nil, maskAny(err)
		}
		update(&current.Spec)
		current, err = cli.DatabaseV1alpha().ArangoDeployments(ns).Update(current)
		if k8sutil.IsConflict(err) {
			// Retry
		} else if err != nil {
			return nil, maskAny(err)
		}
		return current, nil
	}
}

// removeDeployment removes a deployment
func removeDeployment(cli versioned.Interface, deploymentName, ns string) error {
	if err := cli.DatabaseV1alpha().ArangoDeployments(ns).Delete(deploymentName, nil); err != nil && k8sutil.IsNotFound(err) {
		return maskAny(err)
	}
	return nil
}

// removeReplication removes a deployment
func removeReplication(cli versioned.Interface, replicationName, ns string) error {
	if err := cli.ReplicationV1alpha().ArangoDeploymentReplications(ns).Delete(replicationName, nil); err != nil && k8sutil.IsNotFound(err) {
		return maskAny(err)
	}
	return nil
}

// deferedCleanupDeployment removes a deployment when shouldCleanDeployments return true.
// This function is intended to be used in a defer statement.
func deferedCleanupDeployment(cli versioned.Interface, deploymentName, ns string) error {
	if !shouldCleanDeployments() {
		return nil
	}
	if err := removeDeployment(cli, deploymentName, ns); err != nil {
		return maskAny(err)
	}
	return nil
}

// deferedCleanupReplication removes a replication when shouldCleanDeployments return true.
// This function is intended to be used in a defer statement.
func deferedCleanupReplication(cli versioned.Interface, replicationName, ns string) error {
	if !shouldCleanDeployments() {
		return nil
	}
	if err := removeReplication(cli, replicationName, ns); err != nil {
		return maskAny(err)
	}
	return nil
}

// removeSecret removes a secret
func removeSecret(cli kubernetes.Interface, secretName, ns string) error {
	if err := cli.CoreV1().Secrets(ns).Delete(secretName, nil); err != nil && k8sutil.IsNotFound(err) {
		return maskAny(err)
	}
	return nil
}

// check if a deployment is up and has reached a state where it is able to answer to /_api/version requests.
// Optionally the returned version can be checked against a user provided version
func waitUntilArangoDeploymentHealthy(deployment *api.ArangoDeployment, DBClient driver.Client, k8sClient kubernetes.Interface, versionString driver.Version) error {
	// deployment checks
	var checkVersionPredicate func(driver.VersionInfo) error
	if len(versionString) > 0 {
		checkVersionPredicate = createEqualVersionsPredicate(versionString)
	}
	switch mode := deployment.Spec.GetMode(); mode {
	case api.DeploymentModeCluster:
		// Wait for cluster to be completely ready
		if err := waitUntilClusterHealth(DBClient, func(h driver.ClusterHealth) error {
			if len(versionString) > 0 {
				for s, r := range h.Health {
					if r.Version == "" { // Older versions of arangodb do not export the version string in cluster health
						continue
					}
					if cmp := r.Version.CompareTo(versionString); cmp != 0 {
						return maskAny(fmt.Errorf("Member %s has version %s, expecting %s", s, r.Version, versionString))
					}
				}
			}

			return clusterHealthEqualsSpec(h, deployment.Spec)
		}); err != nil {
			return maskAny(fmt.Errorf("Cluster not running in expected health in time: %s", err))
		}
	case api.DeploymentModeSingle:
		if err := waitUntilVersionUp(DBClient, checkVersionPredicate); err != nil {
			return maskAny(fmt.Errorf("Single Server not running in time: %s", err))
		}
	case api.DeploymentModeActiveFailover:
		if err := waitUntilVersionUp(DBClient, checkVersionPredicate); err != nil {
			return maskAny(fmt.Errorf("Single Server not running in time: %s", err))
		}

		members := deployment.Status.Members
		singles := members.Single
		agents := members.Agents

		if len(singles) != *deployment.Spec.Single.Count || len(agents) != *deployment.Spec.Agents.Count {
			return maskAny(fmt.Errorf("Wrong number of servers: single %d - agents %d", len(singles), len(agents)))
		}

		ctx := context.Background()

		//check agents
		for _, agent := range agents {
			dbclient, err := arangod.CreateArangodClient(ctx, k8sClient.CoreV1(), deployment, api.ServerGroupAgents, agent.ID)
			if err != nil {
				return maskAny(fmt.Errorf("Unable to create connection to: %s", agent.ID))
			}

			if err := waitUntilVersionUp(dbclient, checkVersionPredicate); err != nil {
				return maskAny(fmt.Errorf("Version check failed for: %s", agent.ID))
			}
		}
		//check single servers
		{
			var goodResults, noLeaderResults int
			for _, single := range singles {
				dbclient, err := arangod.CreateArangodClient(ctx, k8sClient.CoreV1(), deployment, api.ServerGroupSingle, single.ID)
				if err != nil {
					return maskAny(fmt.Errorf("Unable to create connection to: %s", single.ID))
				}

				if err := waitUntilVersionUp(dbclient, checkVersionPredicate, true); err == nil {
					goodResults++
				} else if driver.IsNoLeader(err) {
					noLeaderResults++
				} else {
					return maskAny(fmt.Errorf("Version check failed for: %s", single.ID))
				}
			}

			expectedGood := *deployment.Spec.Single.Count
			expectedNoLeader := 0
			if goodResults != expectedGood || noLeaderResults != expectedNoLeader {
				return maskAny(fmt.Errorf("Wrong number of results: good %d (expected: %d)- noleader %d (expected %d)", goodResults, expectedGood, noLeaderResults, expectedNoLeader))
			}
		}
	default:
		return maskAny(fmt.Errorf("DeploymentMode %s is not supported", mode))
	}
	return nil
}

// testServerRole performs a synchronize endpoints and then requests the server role.
// On success, the role is compared with the given expected role.
// When the requests fail or the role is not equal to the expected role, an error is returned.
func testServerRole(ctx context.Context, client driver.Client, expectedRole driver.ServerRole) error {
	op := func(ctx context.Context) error {
		if err := client.SynchronizeEndpoints(ctx); err != nil {
			return maskAny(err)
		}
		role, err := client.ServerRole(ctx)
		if err != nil {
			return maskAny(err)
		}
		if role != expectedRole {
			return retry.Permanent(fmt.Errorf("Unexpected server role: Expected '%s', got '%s'", expectedRole, role))
		}
		return nil
	}
	if err := retry.RetryWithContext(ctx, op, time.Second*20); err != nil {
		return maskAny(err)
	}
	return nil
}
