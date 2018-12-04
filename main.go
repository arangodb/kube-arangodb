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

package main

import (
	goflag "flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/arangodb/kube-arangodb/pkg/client"
	scheme "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/scheme"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/operator"
	"github.com/arangodb/kube-arangodb/pkg/server"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/probe"
	"github.com/arangodb/kube-arangodb/pkg/util/retry"
)

const (
	defaultServerHost      = "0.0.0.0"
	defaultServerPort      = 8528
	defaultLogLevel        = "debug"
	defaultAdminSecretName = "arangodb-operator-dashboard"
	defaultAlpineImage     = "alpine:3.7"
)

var (
	projectVersion = "dev"
	projectBuild   = "dev"

	maskAny = errors.WithStack

	cmdMain = cobra.Command{
		Use: "arangodb_operator",
		Run: cmdMainRun,
	}

	logLevel      string
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
		alpineImage                 string
	}
	chaosOptions struct {
		allowed bool
	}
	livenessProbe              probe.LivenessProbe
	deploymentProbe            probe.ReadyProbe
	deploymentReplicationProbe probe.ReadyProbe
	storageProbe               probe.ReadyProbe
)

func init() {
	f := cmdMain.Flags()
	f.StringVar(&serverOptions.host, "server.host", defaultServerHost, "Host to listen on")
	f.IntVar(&serverOptions.port, "server.port", defaultServerPort, "Port to listen on")
	f.StringVar(&serverOptions.tlsSecretName, "server.tls-secret-name", "", "Name of secret containing tls.crt & tls.key for HTTPS server (if empty, self-signed certificate is used)")
	f.StringVar(&serverOptions.adminSecretName, "server.admin-secret-name", defaultAdminSecretName, "Name of secret containing username + password for login to the dashboard")
	f.BoolVar(&serverOptions.allowAnonymous, "server.allow-anonymous-access", false, "Allow anonymous access to the dashboard")
	f.StringVar(&logLevel, "log.level", defaultLogLevel, "Set initial log level")
	f.BoolVar(&operatorOptions.enableDeployment, "operator.deployment", false, "Enable to run the ArangoDeployment operator")
	f.BoolVar(&operatorOptions.enableDeploymentReplication, "operator.deployment-replication", false, "Enable to run the ArangoDeploymentReplication operator")
	f.BoolVar(&operatorOptions.enableStorage, "operator.storage", false, "Enable to run the ArangoLocalStorage operator")
	f.StringVar(&operatorOptions.alpineImage, "operator.alpine-image", defaultAlpineImage, "Docker image used for alpine containers")
	f.BoolVar(&chaosOptions.allowed, "chaos.allowed", false, "Set to allow chaos in deployments. Only activated when allowed and enabled in deployment")
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
	// Get environment
	namespace := os.Getenv(constants.EnvOperatorPodNamespace)
	name := os.Getenv(constants.EnvOperatorPodName)
	ip := os.Getenv(constants.EnvOperatorPodIP)

	// Prepare log service
	goflag.CommandLine.Parse([]string{"-logtostderr"})
	var err error
	logService, err = logging.NewService(logLevel)
	if err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to initialize log service")
	}
	logService.ConfigureRootLogger(func(log zerolog.Logger) zerolog.Logger {
		podNameParts := strings.Split(name, "-")
		operatorID := podNameParts[len(podNameParts)-1]
		cliLog = cliLog.With().Str("operator-id", operatorID).Logger()
		return log.With().Str("operator-id", operatorID).Logger()
	})
	logService.CaptureGLog(logService.MustGetLogger("glog"))

	// Check operating mode
	if !operatorOptions.enableDeployment && !operatorOptions.enableDeploymentReplication && !operatorOptions.enableStorage {
		cliLog.Fatal().Err(err).Msg("Turn on --operator.deployment, --operator.deployment-replication, --operator.storage or any combination of these")
	}

	// Log version
	cliLog.Info().
		Str("pod-name", name).
		Str("pod-namespace", namespace).
		Msgf("Starting arangodb-operator, version %s build %s", projectVersion, projectBuild)

	// Check environment
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
		Log:                        logService.MustGetLogger("server"),
		LivenessProbe:              &livenessProbe,
		DeploymentProbe:            &deploymentProbe,
		DeploymentReplicationProbe: &deploymentReplicationProbe,
		StorageProbe:               &storageProbe,
		Operators:                  o,
		Secrets:                    secrets,
	}); err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to create HTTP server")
	} else {
		go svr.Run()
	}

	//	startChaos(context.Background(), cfg.KubeCli, cfg.Namespace, chaosLevel)

	// Start operator
	o.Run()
}

// newOperatorConfigAndDeps creates operator config & dependencies.
func newOperatorConfigAndDeps(id, namespace, name string) (operator.Config, operator.Dependencies, error) {
	kubecli, err := k8sutil.NewKubeClient()
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
	crCli, err := client.NewInCluster()
	if err != nil {
		return operator.Config{}, operator.Dependencies{}, maskAny(fmt.Errorf("Failed to created versioned client: %s", err))
	}
	eventRecorder := createRecorder(cliLog, kubecli, name, namespace)

	cfg := operator.Config{
		ID:                          id,
		Namespace:                   namespace,
		PodName:                     name,
		ServiceAccount:              serviceAccount,
		LifecycleImage:              image,
		EnableDeployment:            operatorOptions.enableDeployment,
		EnableDeploymentReplication: operatorOptions.enableDeploymentReplication,
		EnableStorage:               operatorOptions.enableStorage,
		AllowChaos:                  chaosOptions.allowed,
		AlpineImage:                 operatorOptions.alpineImage,
	}
	deps := operator.Dependencies{
		LogService:                 logService,
		KubeCli:                    kubecli,
		KubeExtCli:                 kubeExtCli,
		CRCli:                      crCli,
		EventRecorder:              eventRecorder,
		LivenessProbe:              &livenessProbe,
		DeploymentProbe:            &deploymentProbe,
		DeploymentReplicationProbe: &deploymentReplicationProbe,
		StorageProbe:               &storageProbe,
	}

	return cfg, deps, nil
}

// getMyPodInfo looks up the image & service account of the pod with given name in given namespace
// Returns image, serviceAccount, error.
func getMyPodInfo(kubecli kubernetes.Interface, namespace, name string) (string, string, error) {
	var image, sa string
	op := func() error {
		pod, err := kubecli.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
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
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: v1core.New(kubecli.Core().RESTClient()).Events(namespace)})
	combinedScheme := runtime.NewScheme()
	scheme.AddToScheme(combinedScheme)
	v1.AddToScheme(combinedScheme)
	appsv1beta2.AddToScheme(combinedScheme)
	return eventBroadcaster.NewRecorder(combinedScheme, v1.EventSource{Component: name})
}
