//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

package cmd

import (
	"context"
	goflag "flag"
	"fmt"
	"net"
	goHttp "net/http"
	"os"
	"reflect"
	"strconv"
	goStrings "strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/arangodb/kube-arangodb/pkg/api"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/crd"
	agencyConfig "github.com/arangodb/kube-arangodb/pkg/deployment/agency/config"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/scheme"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/metrics/collector"
	"github.com/arangodb/kube-arangodb/pkg/operator"
	"github.com/arangodb/kube-arangodb/pkg/server"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util/probe"
	"github.com/arangodb/kube-arangodb/pkg/util/retry"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
	"github.com/arangodb/kube-arangodb/pkg/version"
)

const (
	defaultServerHost          = "0.0.0.0"
	defaultServerPort          = 8528
	defaultAPIHTTPPort         = 8628
	defaultAPIGRPCPort         = 8728
	defaultAdminSecretName     = "arangodb-operator-dashboard"
	defaultAPIJWTSecretName    = "arangodb-operator-api-jwt"
	defaultAPIJWTKeySecretName = "arangodb-operator-api-jwt-key"
	defaultShutdownDelay       = 2 * time.Second
	defaultShutdownTimeout     = 30 * time.Second
)

var (
	cmdMain = cobra.Command{
		Use: "arangodb_operator",
		Run: executeMain,
	}

	memoryLimit struct {
		hardLimit uint64
	}

	serverOptions struct {
		host            string
		port            int
		tlsSecretName   string
		adminSecretName string // Name of basic authentication secret containing the admin username+password of the dashboard
		allowAnonymous  bool   // If set, anonymous access to dashboard is allowed
	}
	apiOptions struct {
		enabled          bool
		httpPort         int
		grpcPort         int
		jwtSecretName    string
		jwtKeySecretName string
		tlsSecretName    string
	}
	operatorOptions struct {
		enableDeployment            bool // Run deployment operator
		enableDeploymentReplication bool // Run deployment-replication operator
		enableStorage               bool // Run local-storage operator
		enableBackup                bool // Run backup operator
		enableApps                  bool // Run apps operator
		enableML                    bool // Run ml operator
		enableAnalytics             bool // Run analytics operator
		enableNetworking            bool // Run networking operator
		enablePlatform              bool // Run platform operator
		enableScheduler             bool // Run scheduler operator
		versionOnly                 bool // Run only version endpoint, explicitly disabled with other
		enableK2KClusterSync        bool // Run k2kClusterSync operator

		operatorFeatureConfigMap string // ConfigMap name

		scalingIntegrationEnabled bool

		reconciliationDelay time.Duration

		singleMode, skipLeaderLabel bool
	}
	shutdownOptions struct {
		delay   time.Duration
		timeout time.Duration
	}
	crdOptions struct {
		install               bool
		preserveUnknownFields []string
		validationSchema      []string
	}
	operatorKubernetesOptions struct {
		maxBatchSize int64

		qps   float32
		burst int
	}
	operatorBackup struct {
		concurrentUploads int
	}
	operatorTimeouts struct {
		k8s                         time.Duration
		arangoD                     time.Duration
		arangoDCheck                time.Duration
		reconciliation              time.Duration
		agency                      time.Duration
		shardRebuild                time.Duration
		shardRebuildRetry           time.Duration
		backupArangoD               time.Duration
		backupUploadArangoD         time.Duration
		forcePodDeletionGracePeriod time.Duration
		podSchedulingGracePeriod    time.Duration
	}
	operatorImageDiscovery struct {
		timeout                time.Duration
		defaultStatusDiscovery bool
	}
	operatorReconciliationRetry struct {
		delay time.Duration
		count int
	}
	chaosOptions struct {
		allowed bool
	}
	metricsOptions struct {
		excludedMetricPrefixes []string
	}
	livenessProbe              probe.LivenessProbe
	deploymentProbe            probe.ReadyProbe
	deploymentReplicationProbe probe.ReadyProbe
	storageProbe               probe.ReadyProbe
	backupProbe                probe.ReadyProbe
	appsProbe                  probe.ReadyProbe
	mlProbe                    probe.ReadyProbe
	analyticsProbe             probe.ReadyProbe
	networkingProbe            probe.ReadyProbe
	platformProbe              probe.ReadyProbe
	schedulerProbe             probe.ReadyProbe
	k2KClusterSyncProbe        probe.ReadyProbe
)

func init() {
	var deprecatedStr string

	f := cmdMain.Flags()

	f.StringVar(&serverOptions.host, "server.host", defaultServerHost, "Host to listen on")
	f.IntVar(&serverOptions.port, "server.port", defaultServerPort, "Port to listen on")
	f.StringVar(&serverOptions.tlsSecretName, "server.tls-secret-name", "", "Name of secret containing tls.crt & tls.key for HTTPS server (if empty, self-signed certificate is used)")
	f.StringVar(&serverOptions.adminSecretName, "server.admin-secret-name", defaultAdminSecretName, "Name of secret containing username + password for login to the dashboard")
	f.BoolVar(&serverOptions.allowAnonymous, "server.allow-anonymous-access", false, "Allow anonymous access to the dashboard")
	f.BoolVar(&apiOptions.enabled, "api.enabled", true, "Enable operator HTTP and gRPC API")
	f.IntVar(&apiOptions.httpPort, "api.http-port", defaultAPIHTTPPort, "HTTP API port to listen on")
	f.IntVar(&apiOptions.grpcPort, "api.grpc-port", defaultAPIGRPCPort, "gRPC API port to listen on")
	f.StringVar(&apiOptions.tlsSecretName, "api.tls-secret-name", "", "Name of secret containing tls.crt & tls.key for HTTPS API (if empty, self-signed certificate is used)")
	f.StringVar(&apiOptions.jwtSecretName, "api.jwt-secret-name", defaultAPIJWTSecretName, "Name of secret which will contain JWT to authenticate API requests.")
	f.StringVar(&apiOptions.jwtKeySecretName, "api.jwt-key-secret-name", defaultAPIJWTKeySecretName, "Name of secret containing key used to sign JWT. If there is no such secret present, value will be saved here")
	f.BoolVar(&operatorOptions.enableDeployment, "operator.deployment", false, "Enable to run the ArangoDeployment operator")
	f.BoolVar(&operatorOptions.enableDeploymentReplication, "operator.deployment-replication", false, "Enable to run the ArangoDeploymentReplication operator")
	f.BoolVar(&operatorOptions.enableStorage, "operator.storage", false, "Enable to run the ArangoLocalStorage operator")
	f.BoolVar(&operatorOptions.enableBackup, "operator.backup", false, "Enable to run the ArangoBackup operator")
	f.BoolVar(&operatorOptions.enableApps, "operator.apps", false, "Enable to run the ArangoApps operator")
	f.BoolVar(&operatorOptions.enableML, "operator.ml", false, "Enable to run the ArangoML operator")
	f.BoolVar(&operatorOptions.enableAnalytics, "operator.analytics", false, "Enable to run the Analytics operator")
	f.BoolVar(&operatorOptions.enableNetworking, "operator.networking", false, "Enable to run the Networking operator")
	f.BoolVar(&operatorOptions.enablePlatform, "operator.platform", false, "Enable to run the Platform operator")
	f.BoolVar(&operatorOptions.enableScheduler, "operator.scheduler", false, "Enable to run the Scheduler operator")
	f.BoolVar(&operatorOptions.enableK2KClusterSync, "operator.k2k-cluster-sync", false, "Enable to run the ListSimple operator")
	f.MarkDeprecated("operator.k2k-cluster-sync", "Enabled within deployment operator")
	f.BoolVar(&operatorOptions.versionOnly, "operator.version", false, "Enable only version endpoint in Operator")
	f.StringVar(&deprecatedStr, "operator.alpine-image", "alpine:3.7", "Docker image used for alpine containers")
	f.MarkDeprecated("operator.alpine-image", "Value is not used anymore")
	f.StringVar(&deprecatedStr, "operator.metrics-exporter-image", "arangodb/arangodb-exporter:0.1.6", "Docker image used for metrics containers by default")
	f.MarkDeprecated("operator.metrics-exporter-image", "Value is not used anymore")
	f.StringVar(&deprecatedStr, "operator.arango-image", "arangodb/arangodb:latest", "Docker image used for arango by default")
	f.MarkDeprecated("operator.arango-image", "Value is not used anymore")
	f.BoolVar(&chaosOptions.allowed, "chaos.allowed", false, "Set to allow chaos in deployments. Only activated when allowed and enabled in deployment")
	f.BoolVar(&operatorOptions.skipLeaderLabel, "leader.label.skip", false, "Skips Leader Label for the Pod")
	f.BoolVar(&operatorOptions.singleMode, "mode.single", false, "Enable single mode in Operator. WARNING: There should be only one replica of Operator, otherwise Operator can take unexpected actions")
	f.String("scope", "", "Define scope on which Operator works. Legacy - pre 1.1.0 scope with limited cluster access")
	f.MarkDeprecated("scope", "Value is not used anymore")
	f.MarkHidden("scope")
	f.DurationVar(&operatorTimeouts.k8s, "timeout.k8s", globals.DefaultKubernetesTimeout, "The request timeout to the kubernetes")
	f.DurationVar(&operatorTimeouts.arangoD, "timeout.arangod", globals.DefaultArangoDTimeout, "The request timeout to the ArangoDB")
	f.DurationVar(&operatorTimeouts.arangoDCheck, "timeout.arangod-check", globals.DefaultArangoDCheckTimeout, "The version check request timeout to the ArangoDB")
	f.DurationVar(&operatorTimeouts.agency, "timeout.agency", globals.DefaultArangoDAgencyTimeout, "The Agency read timeout")
	f.DurationVar(&operatorTimeouts.reconciliation, "timeout.reconciliation", globals.DefaultReconciliationTimeout, "The reconciliation timeout to the ArangoDB CR")
	f.DurationVar(&operatorTimeouts.shardRebuild, "timeout.shard-rebuild", globals.DefaultOutSyncedShardRebuildTimeout, "Timeout after which particular out-synced shard is considered as failed and rebuild is triggered")
	f.DurationVar(&operatorTimeouts.shardRebuildRetry, "timeout.shard-rebuild-retry", globals.DefaultOutSyncedShardRebuildRetryTimeout, "Timeout after which rebuild shards retry flow is triggered")
	f.DurationVar(&operatorTimeouts.backupArangoD, "timeout.backup-arangod", globals.BackupDefaultArangoClientTimeout, "The request timeout to the ArangoDB during backup calls")
	f.DurationVar(&operatorTimeouts.backupUploadArangoD, "timeout.backup-upload", globals.BackupUploadArangoClientTimeout, "The request timeout to the ArangoDB during uploading files")
	f.DurationVar(&operatorTimeouts.forcePodDeletionGracePeriod, "timeout.force-delete-pod-grace-period", globals.DefaultForcePodDeletionGracePeriodTimeout, "Default period when ArangoDB Pod should be forcefully removed after all containers were stopped - set to 0 to disable forceful removals")
	f.DurationVar(&operatorTimeouts.podSchedulingGracePeriod, "timeout.pod-scheduling-grace-period", globals.DefaultPodSchedulingGracePeriod, "Default period when ArangoDB Pod should be deleted in case of scheduling info change - set to 0 to disable")
	f.DurationVar(&shutdownOptions.delay, "shutdown.delay", defaultShutdownDelay, "The delay before running shutdown handlers")
	f.DurationVar(&shutdownOptions.timeout, "shutdown.timeout", defaultShutdownTimeout, "Timeout for shutdown handlers")
	f.DurationVar(&operatorReconciliationRetry.delay, "operator.reconciliation.retry.delay", globals.DefaultOperatorUpdateRetryDelay, "Delay between Object Update operations in the Reconciliation loop")
	f.IntVar(&operatorReconciliationRetry.count, "operator.reconciliation.retry.count", globals.DefaultOperatorUpdateRetryCount, "Count of retries during Object Update operations in the Reconciliation loop")
	f.BoolVar(&operatorOptions.scalingIntegrationEnabled, "internal.scaling-integration", false, "Enable Scaling Integration")
	f.DurationVar(&operatorOptions.reconciliationDelay, "reconciliation.delay", 0, "Delay between reconciliation loops (<= 0 -> Disabled)")
	f.Int64Var(&operatorKubernetesOptions.maxBatchSize, "kubernetes.max-batch-size", globals.DefaultKubernetesRequestBatchSize, "Size of batch during objects read")
	f.Float32Var(&operatorKubernetesOptions.qps, "kubernetes.qps", kclient.DefaultQPS, "Number of queries per second for k8s API")
	f.IntVar(&operatorKubernetesOptions.burst, "kubernetes.burst", kclient.DefaultBurst, "Burst for the k8s API")
	f.BoolVar(&crdOptions.install, "crd.install", true, "Install missing CRD if access is possible")
	f.StringArrayVar(&crdOptions.preserveUnknownFields, "crd.preserve-unknown-fields", nil, "Controls which CRD should have enabled preserve unknown fields in validation schema <crd-name>=<true/false>. To apply for all, use crd-name 'all'.")
	f.StringArrayVar(&crdOptions.validationSchema, "crd.validation-schema", defaultValidationSchemaEnabled, "Overrides default set of CRDs which should have validation schema enabled <crd-name>=<true/false>. To apply for all, use crd-name 'all'.")
	f.IntVar(&operatorBackup.concurrentUploads, "backup-concurrent-uploads", globals.DefaultBackupConcurrentUploads, "Number of concurrent uploads per deployment")
	f.Uint64Var(&memoryLimit.hardLimit, "memory-limit", 0, "Define memory limit for hard shutdown and the dump of goroutines. Used for testing")
	f.StringArrayVar(&metricsOptions.excludedMetricPrefixes, "metrics.excluded-prefixes", nil, "List of the excluded metrics prefixes")
	f.BoolVar(&operatorImageDiscovery.defaultStatusDiscovery, "image.discovery.status", true, "Discover Operator Image from Pod Status by default. When disabled Pod Spec is used.")
	f.DurationVar(&operatorImageDiscovery.timeout, "image.discovery.timeout", time.Minute, "Timeout for image discovery process")
	if err := logging.Init(&cmdMain); err != nil {
		panic(err.Error())
	}
	if err := features.Init(&cmdMain); err != nil {
		panic(err.Error())
	}
	if err := agencyConfig.Init(&cmdMain); err != nil {
		panic(err.Error())
	}
	if err := reconcile.ActionsConfigGlobal.Init(&cmdMain); err != nil {
		panic(err.Error())
	}
	if err := operatorHTTP.InitConfiguration(&cmdMain); err != nil {
		panic(err.Error())
	}
}

func Command() *cobra.Command {
	return &cmdMain
}

func Execute() int {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	if err := cmdMain.ExecuteContext(shutdown.Context()); err != nil {
		var v cli.CommandExitCode
		if errors.As(err, &v) {
			return v.ExitCode
		}

		return 1
	}

	return 0
}

// Run the operator
func executeMain(cmd *cobra.Command, args []string) {
	// Get environment
	namespace := os.Getenv(utilConstants.EnvOperatorPodNamespace)
	name := os.Getenv(utilConstants.EnvOperatorPodName)
	ip := os.Getenv(utilConstants.EnvOperatorPodIP)

	go monitorMemoryLimit()

	globals.GetGlobalTimeouts().Kubernetes().Set(operatorTimeouts.k8s)
	globals.GetGlobalTimeouts().ArangoD().Set(operatorTimeouts.arangoD)
	globals.GetGlobalTimeouts().Agency().Set(operatorTimeouts.agency)
	globals.GetGlobalTimeouts().ArangoDCheck().Set(operatorTimeouts.arangoDCheck)
	globals.GetGlobalTimeouts().Reconciliation().Set(operatorTimeouts.reconciliation)
	globals.GetGlobalTimeouts().ShardRebuild().Set(operatorTimeouts.shardRebuild)
	globals.GetGlobalTimeouts().ShardRebuildRetry().Set(operatorTimeouts.shardRebuildRetry)
	globals.GetGlobalTimeouts().BackupArangoClientTimeout().Set(operatorTimeouts.backupArangoD)
	globals.GetGlobalTimeouts().BackupArangoClientUploadTimeout().Set(operatorTimeouts.backupUploadArangoD)
	globals.GetGlobalTimeouts().ForcePodDeletionGracePeriodTimeout().Set(operatorTimeouts.forcePodDeletionGracePeriod)
	globals.GetGlobalTimeouts().PodSchedulingGracePeriod().Set(operatorTimeouts.podSchedulingGracePeriod)

	globals.GetGlobals().Retry().OperatorUpdateRetryDelay().Set(operatorReconciliationRetry.delay)
	globals.GetGlobals().Retry().OperatorUpdateRetryCount().Set(operatorReconciliationRetry.count)

	globals.GetGlobals().Kubernetes().RequestBatchSize().Set(operatorKubernetesOptions.maxBatchSize)
	globals.GetGlobals().Backup().ConcurrentUploads().Set(operatorBackup.concurrentUploads)

	collector.GetCollector().SetFilter(metrics.NegateMetricPushFilter(metrics.NewPrefixMetricPushFilter(metricsOptions.excludedMetricPrefixes...)))

	kclient.SetDefaultQPS(operatorKubernetesOptions.qps)
	kclient.SetDefaultBurst(operatorKubernetesOptions.burst)

	// Prepare log service
	if err := logging.Enable(); err != nil {
		logger.Err(err).Fatal("Unable to enable logger")
	}

	podNameParts := goStrings.Split(name, "-")
	operatorID := podNameParts[len(podNameParts)-1]

	if operatorID != "" {
		logging.Global().RegisterWrappers(func(in *zerolog.Event) *zerolog.Event {
			return in.Str("operator-id", operatorID)
		})
	}

	logger.Info("nice to meet you")

	// Print all enabled featured
	features.Iterate(func(name string, feature features.Feature) {
		logger.Info("Operator Feature %s (%s) is %s.", name, features.GetFeatureArgName(name), util.BoolSwitch(feature.Enabled(), "enabled", "disabled"))
	})

	// Check operating mode
	if !operatorOptions.enableDeployment &&
		!operatorOptions.enableDeploymentReplication &&
		!operatorOptions.enableStorage &&
		!operatorOptions.enableBackup &&
		!operatorOptions.enableApps &&
		!operatorOptions.enableK2KClusterSync &&
		!operatorOptions.enableML &&
		!operatorOptions.enableAnalytics &&
		!operatorOptions.enableNetworking &&
		!operatorOptions.enableScheduler &&
		!operatorOptions.enablePlatform {
		if !operatorOptions.versionOnly {
			if version.GetVersionV1().IsEnterprise() {
				logger.Fatal("Turn on --operator.deployment, --operator.deployment-replication, --operator.storage, --operator.backup, --operator.apps, --operator.k2k-cluster-sync, --operator.ml, --operator.analytics, --operator.networking, --operator.scheduler, --operator.platform or any combination of these")
			} else {
				logger.Fatal("Turn on --operator.deployment, --operator.deployment-replication, --operator.storage, --operator.backup, --operator.apps, --operator.k2k-cluster-sync, --operator.networking, --operator.scheduler, --operator.platform or any combination of these")
			}
		}
	} else if operatorOptions.versionOnly {
		logger.Fatal("Options --operator.deployment, --operator.deployment-replication, --operator.storage, --operator.backup, --operator.apps, --operator.k2k-cluster-sync, --operator.ml, --operator.analytics, --operator.networking cannot be enabled together with --operator.version")
	} else if !version.GetVersionV1().IsEnterprise() {
		if operatorOptions.enableML || operatorOptions.enableAnalytics {
			logger.Fatal("Options --operator.ml, --operator.analytics can be enabled only on the Enterprise Operator")
		}
	}

	// Log version
	logger.
		Str("pod-name", name).
		Str("pod-namespace", namespace).
		Info("Starting arangodb-operator (%s), version %s build %s", version.GetVersionV1().Edition.Title(), version.GetVersionV1().Version, version.GetVersionV1().Build)

	// Check environment
	if !operatorOptions.versionOnly {
		if len(namespace) == 0 {
			logger.Fatal("%s environment variable missing", utilConstants.EnvOperatorPodNamespace)
		}
		if len(name) == 0 {
			logger.Fatal("%s environment variable missing", utilConstants.EnvOperatorPodName)
		}
		if len(ip) == 0 {
			logger.Fatal("%s environment variable missing", utilConstants.EnvOperatorPodIP)
		}

		// Get host name
		id, err := os.Hostname()
		if err != nil {
			logger.Err(err).Fatal("Failed to get hostname")
		}

		client, ok := kclient.GetDefaultFactory().Client()
		if !ok {
			logger.Fatal("Failed to get client")
		}

		if crdOptions.install {
			ctx, cancel := context.WithTimeout(shutdown.Context(), time.Minute)
			defer cancel()

			crdOpts, err := prepareCRDOptions(crdOptions.validationSchema, crdOptions.preserveUnknownFields)
			if err != nil {
				logger.Fatal("Invalid --crd.validation-schema args: %s", err)
			}

			_ = crd.EnsureCRDWithOptions(ctx, client, crd.EnsureCRDOptions{IgnoreErrors: true, CRDOptions: crdOpts})
		}

		secrets := client.Kubernetes().CoreV1().Secrets(namespace)

		// Create operator
		cfg, deps, err := newOperatorConfigAndDeps(id+"-"+name, namespace, name)
		if err != nil {
			logger.Err(err).Fatal("Failed to create operator config & deps")
		}
		if err := ensureFeaturesConfigMap(context.Background(), client.Kubernetes().CoreV1().ConfigMaps(namespace), cfg); err != nil {
			logger.Err(err).Error("Failed to create features config map")
		}
		o, err := operator.NewOperator(cfg, deps)
		if err != nil {
			logger.Err(err).Fatal("Failed to create operator")
		}

		if apiOptions.enabled {
			apiServerCfg := api.ServerConfig{
				Namespace:        namespace,
				ServerName:       name,
				ServerAltNames:   []string{ip},
				HTTPAddress:      net.JoinHostPort("0.0.0.0", strconv.Itoa(apiOptions.httpPort)),
				GRPCAddress:      net.JoinHostPort("0.0.0.0", strconv.Itoa(apiOptions.grpcPort)),
				TLSSecretName:    apiOptions.tlsSecretName,
				JWTSecretName:    apiOptions.jwtSecretName,
				JWTKeySecretName: apiOptions.jwtKeySecretName,
				LivelinessProbe:  &livenessProbe,
				ProbeDeployment: api.ReadinessProbeConfig{
					Enabled: cfg.EnableDeployment,
					Probe:   &deploymentProbe,
				},
				ProbeDeploymentReplication: api.ReadinessProbeConfig{
					Enabled: cfg.EnableDeploymentReplication,
					Probe:   &deploymentReplicationProbe,
				},
				ProbeStorage: api.ReadinessProbeConfig{
					Enabled: cfg.EnableStorage,
					Probe:   &storageProbe,
				},
			}
			apiServer, err := api.NewServer(client.Kubernetes().CoreV1(), apiServerCfg)
			if err != nil {
				logger.Err(err).Fatal("Failed to create API server")
			}
			go func() {
				if err := apiServer.Run(); err != nil {
					logger.Err(err).Error("while running API server")
				}
			}()
		}

		listenAddr := net.JoinHostPort(serverOptions.host, strconv.Itoa(serverOptions.port))
		if svr, err := server.NewServer(client.Kubernetes().CoreV1(), server.Config{
			Namespace:          namespace,
			Address:            listenAddr,
			TLSSecretName:      serverOptions.tlsSecretName,
			TLSSecretNamespace: namespace,
			PodName:            name,
			PodIP:              ip,
			AdminSecretName:    serverOptions.adminSecretName,
			AllowAnonymous:     serverOptions.allowAnonymous,
		}, server.Dependencies{
			LivenessProbe: &livenessProbe,
			Deployment: server.OperatorDependency{
				Enabled: cfg.EnableDeployment,
				Probe:   &deploymentProbe,
			},
			DeploymentReplication: server.OperatorDependency{
				Enabled: cfg.EnableDeploymentReplication,
				Probe:   &deploymentReplicationProbe,
			},
			Storage: server.OperatorDependency{
				Enabled: cfg.EnableStorage,
				Probe:   &storageProbe,
			},
			Backup: server.OperatorDependency{
				Enabled: cfg.EnableBackup,
				Probe:   &backupProbe,
			},
			Apps: server.OperatorDependency{
				Enabled: cfg.EnableApps,
				Probe:   &appsProbe,
			},
			ML: server.OperatorDependency{
				Enabled: cfg.EnableML,
				Probe:   &mlProbe,
			},
			Analytics: server.OperatorDependency{
				Enabled: cfg.EnableAnalytics,
				Probe:   &analyticsProbe,
			},
			Networking: server.OperatorDependency{
				Enabled: cfg.EnableNetworking,
				Probe:   &networkingProbe,
			},
			Platform: server.OperatorDependency{
				Enabled: cfg.EnablePlatform,
				Probe:   &platformProbe,
			},
			Scheduler: server.OperatorDependency{
				Enabled: cfg.EnableScheduler,
				Probe:   &schedulerProbe,
			},
			ClusterSync: server.OperatorDependency{
				Enabled: cfg.EnableK2KClusterSync,
				Probe:   &k2KClusterSyncProbe,
			},
			Operators: o,

			Secrets: secrets,
		}); err != nil {
			logger.Err(err).Fatal("Failed to create HTTP server")
		} else {
			go func() {
				if err := svr.Run(); err != nil {
					logger.Err(err).Error("error while starting server")
				}
			}()
		}

		//	startChaos(context.Background(), cfg.KubeCli, cfg.Namespace, chaosLevel)

		// Start operator
		o.Run()
	} else {
		if err := startVersionProcess(); err != nil {
			logger.Err(err).Fatal("Failed to create HTTP server")
		}
	}
}

func startVersionProcess() error {
	// Just expose version
	listenAddr := net.JoinHostPort(serverOptions.host, strconv.Itoa(serverOptions.port))
	logger.Str("addr", listenAddr).Info("Starting version endpoint")

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	versionV1Responser, err := operatorHTTP.NewSimpleJSONResponse(version.GetVersionV1())
	if err != nil {
		return errors.WithStack(err)
	}
	r.GET("/_api/version", gin.WrapF(versionV1Responser.ServeHTTP))
	r.GET("/api/v1/version", gin.WrapF(versionV1Responser.ServeHTTP))

	s := goHttp.Server{
		Addr:    listenAddr,
		Handler: r,
	}

	return s.ListenAndServe()
}

// newOperatorConfigAndDeps creates operator config & dependencies.
func newOperatorConfigAndDeps(id, namespace, name string) (operator.Config, operator.Dependencies, error) {
	client, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return operator.Config{}, operator.Dependencies{}, errors.Errorf("Failed to get client")
	}

	image, serviceAccount, err := getMyPodInfo(client.Kubernetes(), namespace, name)
	if err != nil {
		return operator.Config{}, operator.Dependencies{}, errors.WithStack(fmt.Errorf("Failed to get my pod's service account: %s", err))
	}

	eventRecorder := createRecorder(client.Kubernetes(), name, namespace)

	cfg := operator.Config{
		ID:                          id,
		Namespace:                   namespace,
		PodName:                     name,
		ServiceAccount:              serviceAccount,
		OperatorImage:               image,
		SkipLeaderLabel:             operatorOptions.skipLeaderLabel,
		EnableDeployment:            operatorOptions.enableDeployment,
		EnableDeploymentReplication: operatorOptions.enableDeploymentReplication,
		EnableStorage:               operatorOptions.enableStorage,
		EnableBackup:                operatorOptions.enableBackup,
		EnableApps:                  operatorOptions.enableApps,
		EnableML:                    operatorOptions.enableML,
		EnableAnalytics:             operatorOptions.enableAnalytics,
		EnableNetworking:            operatorOptions.enableNetworking,
		EnablePlatform:              operatorOptions.enablePlatform,
		EnableScheduler:             operatorOptions.enableScheduler,
		EnableK2KClusterSync:        operatorOptions.enableK2KClusterSync,
		AllowChaos:                  chaosOptions.allowed,
		ScalingIntegrationEnabled:   operatorOptions.scalingIntegrationEnabled,
		SingleMode:                  operatorOptions.singleMode,
		ReconciliationDelay:         operatorOptions.reconciliationDelay,
		ShutdownDelay:               shutdownOptions.delay,
		ShutdownTimeout:             shutdownOptions.timeout,
	}
	deps := operator.Dependencies{
		Client:                     client,
		EventRecorder:              eventRecorder,
		LivenessProbe:              &livenessProbe,
		DeploymentProbe:            &deploymentProbe,
		DeploymentReplicationProbe: &deploymentReplicationProbe,
		StorageProbe:               &storageProbe,
		BackupProbe:                &backupProbe,
		AppsProbe:                  &appsProbe,
		MlProbe:                    &mlProbe,
		AnalyticsProbe:             &analyticsProbe,
		NetworkingProbe:            &networkingProbe,
		PlatformProbe:              &platformProbe,
		SchedulerProbe:             &schedulerProbe,
		K2KClusterSyncProbe:        &k2KClusterSyncProbe,
	}

	return cfg, deps, nil
}

// getMyPodInfo looks up the image & service account of the pod with given name in given namespace
// Returns image, serviceAccount, error.
func getMyPodInfo(kubecli kubernetes.Interface, namespace, name string) (string, string, error) {
	if image, sa, ok := getMyPodInfoWrap(kubecli, namespace, name, getMyImageInfoFunc(operatorImageDiscovery.defaultStatusDiscovery)); ok {
		return image, sa, nil
	}

	logger.Warn("Unable to discover image, fallback to second method")

	if image, sa, ok := getMyPodInfoWrap(kubecli, namespace, name, getMyImageInfoFunc(!operatorImageDiscovery.defaultStatusDiscovery)); ok {
		return image, sa, nil
	}

	return "", "", errors.Errorf("Unable to discover image")
}

func getMyPodInfoWrap(kubecli kubernetes.Interface, namespace, name string, imageFunc func(in *core.Pod) (string, bool)) (string, string, bool) {
	var image, sa string
	op := func() error {
		pod, err := kubecli.CoreV1().Pods(namespace).Get(context.Background(), name, meta.GetOptions{})
		if err != nil {
			logger.
				Err(err).
				Str("name", name).
				Error("Failed to get operator pod")
			return errors.WithStack(err)
		}
		sa = pod.Spec.ServiceAccountName
		if i, ok := imageFunc(pod); !ok {
			return errors.Errorf("failed to get image ID from pod")
		} else {
			image = i
		}
		return nil
	}
	if err := retry.Retry(op, operatorImageDiscovery.timeout/2); err == nil {
		return image, sa, true
	}
	return "", "", false
}

func getMyImageInfoFunc(status bool) func(pod *core.Pod) (string, bool) {
	return func(pod *core.Pod) (string, bool) {
		if status {
			return k8sutil.GetArangoDBImageIDFromContainerStatuses(pod.Status.ContainerStatuses, shared.ServerContainerName, shared.OperatorContainerName, utilConstants.MyContainerNameEnv.GetOrDefault(shared.OperatorContainerName))
		}
		return k8sutil.GetArangoDBImageFromContainers(pod.Spec.Containers, shared.ServerContainerName, shared.OperatorContainerName, utilConstants.MyContainerNameEnv.GetOrDefault(shared.OperatorContainerName))
	}
}

func createRecorder(kubecli kubernetes.Interface, name, namespace string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(func(format string, args ...interface{}) {
		eventRecorder.Info(format, args...)
	})
	eventBroadcaster.StartRecordingToSink(&typedCore.EventSinkImpl{Interface: typedCore.New(kubecli.CoreV1().RESTClient()).Events(namespace)})
	combinedScheme := runtime.NewScheme()
	scheme.AddToScheme(combinedScheme)
	core.AddToScheme(combinedScheme)
	apps.AddToScheme(combinedScheme)
	return eventBroadcaster.NewRecorder(combinedScheme, core.EventSource{Component: name})
}

// ensureFeaturesConfigMap creates or updates config map with enabled features.
func ensureFeaturesConfigMap(ctx context.Context, client typedCore.ConfigMapInterface, cfg operator.Config) error {
	ft := features.GetFeatureMap()

	featuresCM := make(map[string]string, len(ft))

	for k, v := range ft {
		if v {
			featuresCM[k] = features.Enabled
		} else {
			featuresCM[k] = features.Disabled
		}
	}

	nctx, c := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer c()
	if cm, err := client.Get(nctx, features.ConfigMapName(), meta.GetOptions{}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return err
		}

		nctx, c := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer c()
		if _, err := client.Create(nctx, &core.ConfigMap{
			ObjectMeta: meta.ObjectMeta{
				Name:      features.ConfigMapName(),
				Namespace: cfg.Namespace,
			},
			Data: featuresCM,
		}, meta.CreateOptions{}); err != nil {
			return err
		}

		return nil
	} else if !reflect.DeepEqual(cm.Data, featuresCM) {
		q := cm.DeepCopy()
		q.Data = featuresCM

		nctx, c := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer c()
		if _, err := client.Update(nctx, q, meta.UpdateOptions{}); err != nil {
			return err
		}

		return nil
	}

	return nil
}
