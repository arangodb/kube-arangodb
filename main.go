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

package main

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

	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
	"github.com/gin-gonic/gin"

	"github.com/arangodb/kube-arangodb/pkg/version"

	"github.com/arangodb/kube-arangodb/pkg/util/arangod"

	"github.com/arangodb/kube-arangodb/pkg/operator/scope"

	"github.com/arangodb/kube-arangodb/pkg/deployment/features"

	"github.com/rs/zerolog/log"

	deploymentApi "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"

	utilsError "github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/scheme"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/operator"
	"github.com/arangodb/kube-arangodb/pkg/server"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/probe"
	"github.com/arangodb/kube-arangodb/pkg/util/retry"
	v1 "k8s.io/api/core/v1"
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

	UBIImageEnv             util.EnvironmentVariable = "RELATED_IMAGE_UBI"
	ArangoImageEnv          util.EnvironmentVariable = "RELATED_IMAGE_DATABASE"
	MetricsExporterImageEnv util.EnvironmentVariable = "RELATED_IMAGE_METRICSEXPORTER"
)

var (
	maskAny = errors.WithStack

	cmdMain = cobra.Command{
		Use: "arangodb_operator",
		Run: cmdMainRun,
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
		versionOnly                 bool // Run only version endpoint, explicitly disabled with other

		scalingIntegrationEnabled bool

		alpineImage, metricsExporterImage, arangoImage string

		singleMode bool
		scope      string
	}
	timeouts struct {
		k8s     time.Duration
		arangoD time.Duration
	}
	chaosOptions struct {
		allowed bool
	}
	livenessProbe              probe.LivenessProbe
	deploymentProbe            probe.ReadyProbe
	deploymentReplicationProbe probe.ReadyProbe
	storageProbe               probe.ReadyProbe
	backupProbe                probe.ReadyProbe
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
	f.BoolVar(&operatorOptions.versionOnly, "operator.version", false, "Enable only version endpoint in Operator")
	f.StringVar(&operatorOptions.alpineImage, "operator.alpine-image", UBIImageEnv.GetOrDefault(defaultAlpineImage), "Docker image used for alpine containers")
	f.MarkDeprecated("operator.alpine-image", "Value is not used anymore")
	f.StringVar(&operatorOptions.metricsExporterImage, "operator.metrics-exporter-image", MetricsExporterImageEnv.GetOrDefault(defaultMetricsExporterImage), "Docker image used for metrics containers by default")
	f.StringVar(&operatorOptions.arangoImage, "operator.arango-image", ArangoImageEnv.GetOrDefault(defaultArangoImage), "Docker image used for arango by default")
	f.BoolVar(&chaosOptions.allowed, "chaos.allowed", false, "Set to allow chaos in deployments. Only activated when allowed and enabled in deployment")
	f.BoolVar(&operatorOptions.singleMode, "mode.single", false, "Enable single mode in Operator. WARNING: There should be only one replica of Operator, otherwise Operator can take unexpected actions")
	f.StringVar(&operatorOptions.scope, "scope", scope.DefaultScope.String(), "Define scope on which Operator works. Legacy - pre 1.1.0 scope with limited cluster access")
	f.DurationVar(&timeouts.k8s, "timeout.k8s", time.Second*3, "The request timeout to the kubernetes")
	f.DurationVar(&timeouts.arangoD, "timeout.arangod", time.Second*10, "The request timeout to the ArangoDB")
	f.BoolVar(&operatorOptions.scalingIntegrationEnabled, "scaling-integration", false, "Enable Scaling Integration")
	features.Init(&cmdMain)
}

func main() {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	cmdMain.Execute()
}

// Show usage
func cmdUsage(cmd *cobra.Command, args []string) {
	cmd.Usage()
}

// Run the operator
func cmdMainRun(cmd *cobra.Command, args []string) {
	// Set global logger
	log.Logger = logging.NewRootLogger()

	// Get environment
	namespace := os.Getenv(constants.EnvOperatorPodNamespace)
	name := os.Getenv(constants.EnvOperatorPodName)
	ip := os.Getenv(constants.EnvOperatorPodIP)

	deploymentApi.DefaultImage = operatorOptions.arangoImage
	k8sutil.SetRequestTimeout(timeouts.k8s)
	arangod.SetRequestTimeout(timeouts.arangoD)

	// Prepare log service
	var err error
	logService, err = logging.NewService(defaultLogLevel, logLevels)
	if err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to initialize log service")
	}
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
	if !operatorOptions.enableDeployment && !operatorOptions.enableDeploymentReplication && !operatorOptions.enableStorage && !operatorOptions.enableBackup {
		if !operatorOptions.versionOnly {
			cliLog.Fatal().Err(err).Msg("Turn on --operator.deployment, --operator.deployment-replication, --operator.storage, --operator.backup or any combination of these")
		}
	} else if operatorOptions.versionOnly {
		cliLog.Fatal().Err(err).Msg("Options --operator.deployment, --operator.deployment-replication, --operator.storage, --operator.backup cannot be enabled together with --operator.version")
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

		// Create kubernetes client
		kubecli, err := k8sutil.NewKubeClient()
		if err != nil {
			cliLog.Fatal().Err(err).Msg("Failed to create Kubernetes client")
		}
		secrets := kubecli.CoreV1().Secrets(namespace)

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
		if svr, err := server.NewServer(kubecli.CoreV1(), server.Config{
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
	kubecli, err := k8sutil.NewKubeClient()
	if err != nil {
		return operator.Config{}, operator.Dependencies{}, maskAny(err)
	}

	kubeMonCli, err := k8sutil.NewKubeMonitoringV1Client()
	if err != nil {
		return operator.Config{}, operator.Dependencies{}, maskAny(err)
	}

	image, serviceAccount, err := getMyPodInfo(kubecli, namespace, name)
	if err != nil {
		return operator.Config{}, operator.Dependencies{}, maskAny(fmt.Errorf("Failed to get my pod's service account: %s", err))
	}

	kubeExtCli, err := k8sutil.NewKubeExtClient()
	if err != nil {
		return operator.Config{}, operator.Dependencies{}, maskAny(fmt.Errorf("Failed to create k8b api extensions client: %s", err))
	}
	crCli, err := client.NewClient()
	if err != nil {
		return operator.Config{}, operator.Dependencies{}, maskAny(fmt.Errorf("Failed to created versioned client: %s", err))
	}
	eventRecorder := createRecorder(cliLog, kubecli, name, namespace)

	scope, ok := scope.AsScope(operatorOptions.scope)
	if !ok {
		return operator.Config{}, operator.Dependencies{}, maskAny(fmt.Errorf("Scope %s is not known by Operator", operatorOptions.scope))
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
		AllowChaos:                  chaosOptions.allowed,
		ScalingIntegrationEnabled:   operatorOptions.scalingIntegrationEnabled,
		ArangoImage:                 operatorOptions.arangoImage,
		SingleMode:                  operatorOptions.singleMode,
		Scope:                       scope,
	}
	deps := operator.Dependencies{
		LogService:                 logService,
		KubeCli:                    kubecli,
		KubeExtCli:                 kubeExtCli,
		KubeMonitoringCli:          kubeMonCli,
		CRCli:                      crCli,
		EventRecorder:              eventRecorder,
		LivenessProbe:              &livenessProbe,
		DeploymentProbe:            &deploymentProbe,
		DeploymentReplicationProbe: &deploymentReplicationProbe,
		StorageProbe:               &storageProbe,
		BackupProbe:                &backupProbe,
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
			return maskAny(err)
		}
		sa = pod.Spec.ServiceAccountName
		image = k8sutil.GetArangoDBImageIDFromPod(pod)
		if image == "" {
			// Fallback in case we don't know the id.
			image = pod.Spec.Containers[0].Image
		}
		return nil
	}
	if err := retry.Retry(op, time.Minute*5); err != nil {
		return "", "", maskAny(err)
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
