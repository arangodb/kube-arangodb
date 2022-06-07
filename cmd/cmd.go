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

package cmd

import (
	"context"
	goflag "flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	deploymentApi "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/crd"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/scheme"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/operator"
	"github.com/arangodb/kube-arangodb/pkg/operator/scope"
	"github.com/arangodb/kube-arangodb/pkg/server"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	utilsError "github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/probe"
	"github.com/arangodb/kube-arangodb/pkg/util/retry"
	"github.com/arangodb/kube-arangodb/pkg/version"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
)

const (
	defaultServerHost           = "0.0.0.0"
	defaultServerPort           = 8528
	defaultLogLevel             = "debug"
	defaultAdminSecretName      = "arangodb-operator-dashboard"
	defaultAlpineImage          = "alpine:3.7"
	defaultMetricsExporterImage = "arangodb/arangodb-exporter:0.1.6"
	defaultArangoImage          = "arangodb/arangodb:latest"
	defaultShutdownDelay        = 2 * time.Second
	defaultShutdownTimeout      = 30 * time.Second

	UBIImageEnv             util.EnvironmentVariable = "RELATED_IMAGE_UBI"
	ArangoImageEnv          util.EnvironmentVariable = "RELATED_IMAGE_DATABASE"
	MetricsExporterImageEnv util.EnvironmentVariable = "RELATED_IMAGE_METRICSEXPORTER"
)

var (
	cmdMain = cobra.Command{
		Use: "arangodb_operator",
		Run: executeMain,
	}

	logLevels     []string
	cliLog        = logging.NewRootLogger()
	logService    logging.Service
	serverOptions struct {
		host            string
		port            int
		tlsSecretName   string
		adminSecretName string // Name of basic authentication secret containing the admin username+password of the dashboard
		allowAnonymous  bool   // If set, anonymous access to dashboard is allowed
	}
	operatorOptions struct {
		enableDeployment            bool // Run deployment operator
		enableDeploymentReplication bool // Run deployment-replication operator
		enableStorage               bool // Run local-storage operator
		enableBackup                bool // Run backup operator
		enableApps                  bool // Run apps operator
		versionOnly                 bool // Run only version endpoint, explicitly disabled with other
		enableK2KClusterSync        bool // Run k2kClusterSync operator

		scalingIntegrationEnabled bool

		alpineImage, metricsExporterImage, arangoImage string

		singleMode bool
		scope      string
	}
	shutdownOptions struct {
		delay   time.Duration
		timeout time.Duration
	}
	crdOptions struct {
		install bool
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
		k8s            time.Duration
		arangoD        time.Duration
		arangoDCheck   time.Duration
		reconciliation time.Duration
		agency         time.Duration
	}
	chaosOptions struct {
		allowed bool
	}
	livenessProbe              probe.LivenessProbe
	deploymentProbe            probe.ReadyProbe
	deploymentReplicationProbe probe.ReadyProbe
	storageProbe               probe.ReadyProbe
	backupProbe                probe.ReadyProbe
	appsProbe                  probe.ReadyProbe
	k2KClusterSyncProbe        probe.ReadyProbe
)

func init() {

	f := cmdMain.Flags()
	f.StringVar(&serverOptions.host, "server.host", defaultServerHost, "Host to listen on")
	f.IntVar(&serverOptions.port, "server.port", defaultServerPort, "Port to listen on")
	f.StringVar(&serverOptions.tlsSecretName, "server.tls-secret-name", "", "Name of secret containing tls.crt & tls.key for HTTPS server (if empty, self-signed certificate is used)")
	f.StringVar(&serverOptions.adminSecretName, "server.admin-secret-name", defaultAdminSecretName, "Name of secret containing username + password for login to the dashboard")
	f.BoolVar(&serverOptions.allowAnonymous, "server.allow-anonymous-access", false, "Allow anonymous access to the dashboard")
	f.StringArrayVar(&logLevels, "log.level", []string{defaultLogLevel}, fmt.Sprintf("Set log levels in format <level> or <logger>=<level>. Possible loggers: %s", strings.Join(logging.LoggerNames(), ", ")))
	f.BoolVar(&operatorOptions.enableDeployment, "operator.deployment", false, "Enable to run the ArangoDeployment operator")
	f.BoolVar(&operatorOptions.enableDeploymentReplication, "operator.deployment-replication", false, "Enable to run the ArangoDeploymentReplication operator")
	f.BoolVar(&operatorOptions.enableStorage, "operator.storage", false, "Enable to run the ArangoLocalStorage operator")
	f.BoolVar(&operatorOptions.enableBackup, "operator.backup", false, "Enable to run the ArangoBackup operator")
	f.BoolVar(&operatorOptions.enableApps, "operator.apps", false, "Enable to run the ArangoApps operator")
	f.BoolVar(&operatorOptions.enableK2KClusterSync, "operator.k2k-cluster-sync", false, "Enable to run the ListSimple operator")
	f.BoolVar(&operatorOptions.versionOnly, "operator.version", false, "Enable only version endpoint in Operator")
	f.StringVar(&operatorOptions.alpineImage, "operator.alpine-image", UBIImageEnv.GetOrDefault(defaultAlpineImage), "Docker image used for alpine containers")
	f.MarkDeprecated("operator.alpine-image", "Value is not used anymore")
	f.StringVar(&operatorOptions.metricsExporterImage, "operator.metrics-exporter-image", MetricsExporterImageEnv.GetOrDefault(defaultMetricsExporterImage), "Docker image used for metrics containers by default")
	f.StringVar(&operatorOptions.arangoImage, "operator.arango-image", ArangoImageEnv.GetOrDefault(defaultArangoImage), "Docker image used for arango by default")
	f.BoolVar(&chaosOptions.allowed, "chaos.allowed", false, "Set to allow chaos in deployments. Only activated when allowed and enabled in deployment")
	f.BoolVar(&operatorOptions.singleMode, "mode.single", false, "Enable single mode in Operator. WARNING: There should be only one replica of Operator, otherwise Operator can take unexpected actions")
	f.StringVar(&operatorOptions.scope, "scope", scope.DefaultScope.String(), "Define scope on which Operator works. Legacy - pre 1.1.0 scope with limited cluster access")
	f.DurationVar(&operatorTimeouts.k8s, "timeout.k8s", globals.DefaultKubernetesTimeout, "The request timeout to the kubernetes")
	f.DurationVar(&operatorTimeouts.arangoD, "timeout.arangod", globals.DefaultArangoDTimeout, "The request timeout to the ArangoDB")
	f.DurationVar(&operatorTimeouts.arangoDCheck, "timeout.arangod-check", globals.DefaultArangoDCheckTimeout, "The version check request timeout to the ArangoDB")
	f.DurationVar(&operatorTimeouts.agency, "timeout.agency", globals.DefaultArangoDAgencyTimeout, "The Agency read timeout")
	f.DurationVar(&operatorTimeouts.reconciliation, "timeout.reconciliation", globals.DefaultReconciliationTimeout, "The reconciliation timeout to the ArangoDB CR")
	f.DurationVar(&shutdownOptions.delay, "shutdown.delay", defaultShutdownDelay, "The delay before running shutdown handlers")
	f.DurationVar(&shutdownOptions.timeout, "shutdown.timeout", defaultShutdownTimeout, "Timeout for shutdown handlers")
	f.BoolVar(&operatorOptions.scalingIntegrationEnabled, "internal.scaling-integration", true, "Enable Scaling Integration")
	f.Int64Var(&operatorKubernetesOptions.maxBatchSize, "kubernetes.max-batch-size", globals.DefaultKubernetesRequestBatchSize, "Size of batch during objects read")
	f.Float32Var(&operatorKubernetesOptions.qps, "kubernetes.qps", kclient.DefaultQPS, "Number of queries per second for k8s API")
	f.IntVar(&operatorKubernetesOptions.burst, "kubernetes.burst", kclient.DefaultBurst, "Burst for the k8s API")
	f.BoolVar(&crdOptions.install, "crd.install", true, "Install missing CRD if access is possible")
	f.IntVar(&operatorBackup.concurrentUploads, "backup-concurrent-uploads", globals.DefaultBackupConcurrentUploads, "Number of concurrent uploads per deployment")
	features.Init(&cmdMain)
}

func Execute() int {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	if err := cmdMain.Execute(); err != nil {
		return 1
	}

	return 0
}

// Show usage
func executeUsage(cmd *cobra.Command, args []string) {
	cmd.Usage()
}

// Run the operator
func executeMain(cmd *cobra.Command, args []string) {
	// Set global logger
	log.Logger = logging.NewRootLogger()

	// Get environment
	namespace := os.Getenv(constants.EnvOperatorPodNamespace)
	name := os.Getenv(constants.EnvOperatorPodName)
	ip := os.Getenv(constants.EnvOperatorPodIP)

	deploymentApi.DefaultImage = operatorOptions.arangoImage

	globals.GetGlobalTimeouts().Kubernetes().Set(operatorTimeouts.k8s)
	globals.GetGlobalTimeouts().ArangoD().Set(operatorTimeouts.arangoD)
	globals.GetGlobalTimeouts().Agency().Set(operatorTimeouts.agency)
	globals.GetGlobalTimeouts().ArangoDCheck().Set(operatorTimeouts.arangoDCheck)
	globals.GetGlobalTimeouts().Reconciliation().Set(operatorTimeouts.reconciliation)
	globals.GetGlobals().Kubernetes().RequestBatchSize().Set(operatorKubernetesOptions.maxBatchSize)
	globals.GetGlobals().Backup().ConcurrentUploads().Set(operatorBackup.concurrentUploads)

	kclient.SetDefaultQPS(operatorKubernetesOptions.qps)
	kclient.SetDefaultBurst(operatorKubernetesOptions.burst)

	// Prepare log service
	var err error
	if err := logging.InitGlobalLogger(defaultLogLevel, logLevels); err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to initialize log service")
	}

	logService = logging.GlobalLogger()

	logService.ConfigureRootLogger(func(log zerolog.Logger) zerolog.Logger {
		podNameParts := strings.Split(name, "-")
		operatorID := podNameParts[len(podNameParts)-1]
		cliLog = cliLog.With().Str("operator-id", operatorID).Logger()
		return log.With().Str("operator-id", operatorID).Logger()
	})

	klog.SetOutput(logService.MustGetLogger(logging.LoggerNameKLog))
	klog.Info("nice to meet you")
	klog.Flush()

	// Check operating mode
	if !operatorOptions.enableDeployment && !operatorOptions.enableDeploymentReplication && !operatorOptions.enableStorage &&
		!operatorOptions.enableBackup && !operatorOptions.enableApps && !operatorOptions.enableK2KClusterSync {
		if !operatorOptions.versionOnly {
			cliLog.Fatal().Err(err).Msg("Turn on --operator.deployment, --operator.deployment-replication, --operator.storage, --operator.backup, --operator.apps, --operator.k2k-cluster-sync or any combination of these")
		}
	} else if operatorOptions.versionOnly {
		cliLog.Fatal().Err(err).Msg("Options --operator.deployment, --operator.deployment-replication, --operator.storage, --operator.backup, --operator.apps, --operator.k2k-cluster-sync cannot be enabled together with --operator.version")
	}

	// Log version
	cliLog.Info().
		Str("pod-name", name).
		Str("pod-namespace", namespace).
		Msgf("Starting arangodb-operator (%s), version %s build %s", version.GetVersionV1().Edition.Title(), version.GetVersionV1().Version, version.GetVersionV1().Build)

	// Check environment
	if !operatorOptions.versionOnly {
		if len(namespace) == 0 {
			cliLog.Fatal().Msgf("%s environment variable missing", constants.EnvOperatorPodNamespace)
		}
		if len(name) == 0 {
			cliLog.Fatal().Msgf("%s environment variable missing", constants.EnvOperatorPodName)
		}
		if len(ip) == 0 {
			cliLog.Fatal().Msgf("%s environment variable missing", constants.EnvOperatorPodIP)
		}

		// Get host name
		id, err := os.Hostname()
		if err != nil {
			cliLog.Fatal().Err(err).Msg("Failed to get hostname")
		}

		client, ok := kclient.GetDefaultFactory().Client()
		if !ok {
			cliLog.Fatal().Msg("Failed to get client")
		}

		if crdOptions.install {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			crd.EnsureCRD(ctx, logService.MustGetLogger("crd"), client)
		}

		secrets := client.Kubernetes().CoreV1().Secrets(namespace)

		// Create operator
		cfg, deps, err := newOperatorConfigAndDeps(id+"-"+name, namespace, name)
		if err != nil {
			cliLog.Fatal().Err(err).Msg("Failed to create operator config & deps")
		}
		o, err := operator.NewOperator(cfg, deps)
		if err != nil {
			cliLog.Fatal().Err(err).Msg("Failed to create operator")
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
			Log:           logService.MustGetLogger(logging.LoggerNameServer),
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
			ClusterSync: server.OperatorDependency{
				Enabled: cfg.EnableK2KClusterSync,
				Probe:   &k2KClusterSyncProbe,
			},
			Operators: o,

			Secrets: secrets,
		}); err != nil {
			cliLog.Fatal().Err(err).Msg("Failed to create HTTP server")
		} else {
			go utilsError.LogError(cliLog, "error while starting service", svr.Run)
		}

		//	startChaos(context.Background(), cfg.KubeCli, cfg.Namespace, chaosLevel)

		// Start operator
		o.Run()
	} else {
		if err := startVersionProcess(); err != nil {
			cliLog.Fatal().Err(err).Msg("Failed to create HTTP server")
		}
	}
}

func startVersionProcess() error {
	// Just expose version
	listenAddr := net.JoinHostPort(serverOptions.host, strconv.Itoa(serverOptions.port))
	cliLog.Info().Str("addr", listenAddr).Msgf("Starting version endpoint")

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	versionV1Responser, err := operatorHTTP.NewSimpleJSONResponse(version.GetVersionV1())
	if err != nil {
		return errors.WithStack(err)
	}
	r.GET("/_api/version", gin.WrapF(versionV1Responser.ServeHTTP))
	r.GET("/api/v1/version", gin.WrapF(versionV1Responser.ServeHTTP))

	s := http.Server{
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

	eventRecorder := createRecorder(cliLog, client.Kubernetes(), name, namespace)

	scope, ok := scope.AsScope(operatorOptions.scope)
	if !ok {
		return operator.Config{}, operator.Dependencies{}, errors.WithStack(fmt.Errorf("Scope %s is not known by Operator", operatorOptions.scope))
	}

	cfg := operator.Config{
		ID:                          id,
		Namespace:                   namespace,
		PodName:                     name,
		ServiceAccount:              serviceAccount,
		OperatorImage:               image,
		EnableDeployment:            operatorOptions.enableDeployment,
		EnableDeploymentReplication: operatorOptions.enableDeploymentReplication,
		EnableStorage:               operatorOptions.enableStorage,
		EnableBackup:                operatorOptions.enableBackup,
		EnableApps:                  operatorOptions.enableApps,
		EnableK2KClusterSync:        operatorOptions.enableK2KClusterSync,
		AllowChaos:                  chaosOptions.allowed,
		ScalingIntegrationEnabled:   operatorOptions.scalingIntegrationEnabled,
		ArangoImage:                 operatorOptions.arangoImage,
		SingleMode:                  operatorOptions.singleMode,
		Scope:                       scope,
		ShutdownDelay:               shutdownOptions.delay,
		ShutdownTimeout:             shutdownOptions.timeout,
	}
	deps := operator.Dependencies{
		LogService:                 logService,
		Client:                     client,
		EventRecorder:              eventRecorder,
		LivenessProbe:              &livenessProbe,
		DeploymentProbe:            &deploymentProbe,
		DeploymentReplicationProbe: &deploymentReplicationProbe,
		StorageProbe:               &storageProbe,
		BackupProbe:                &backupProbe,
		AppsProbe:                  &appsProbe,
		K2KClusterSyncProbe:        &k2KClusterSyncProbe,
	}

	return cfg, deps, nil
}

// getMyPodInfo looks up the image & service account of the pod with given name in given namespace
// Returns image, serviceAccount, error.
func getMyPodInfo(kubecli kubernetes.Interface, namespace, name string) (string, string, error) {
	var image, sa string
	op := func() error {
		pod, err := kubecli.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			cliLog.Error().
				Err(err).
				Str("name", name).
				Msg("Failed to get operator pod")
			return errors.WithStack(err)
		}
		sa = pod.Spec.ServiceAccountName
		if image, err = k8sutil.GetArangoDBImageIDFromPod(pod); err != nil {
			return errors.Wrap(err, "failed to get image ID from pod")
		}
		if image == "" {
			// Fallback in case we don't know the id.
			image = pod.Spec.Containers[0].Image
		}
		return nil
	}
	if err := retry.Retry(op, time.Minute*5); err != nil {
		return "", "", errors.WithStack(err)
	}
	return image, sa, nil
}

func createRecorder(log zerolog.Logger, kubecli kubernetes.Interface, name, namespace string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(func(format string, args ...interface{}) {
		log.Info().Msgf(format, args...)
	})
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: v1core.New(kubecli.CoreV1().RESTClient()).Events(namespace)})
	combinedScheme := runtime.NewScheme()
	scheme.AddToScheme(combinedScheme)
	v1.AddToScheme(combinedScheme)
	appsv1.AddToScheme(combinedScheme)
	return eventBroadcaster.NewRecorder(combinedScheme, v1.EventSource{Component: name})
}
