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
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"

	"github.com/arangodb/k8s-operator/pkg/client"
	"github.com/arangodb/k8s-operator/pkg/controller"
	"github.com/arangodb/k8s-operator/pkg/logging"
	"github.com/arangodb/k8s-operator/pkg/util/constants"
	"github.com/arangodb/k8s-operator/pkg/util/k8sutil"
	"github.com/arangodb/k8s-operator/pkg/util/retry"
)

const (
	defaultServerHost = "0.0.0.0"
	defaultServerPort = 8528
	defaultLogLevel   = "debug"
)

var (
	projectVersion = "dev"
	projectBuild   = "dev"

	maskAny = errors.WithStack

	cmdMain = cobra.Command{
		Use: "arangodb_operator",
		Run: cmdMainRun,
	}

	logLevel   string
	cliLog     = logging.NewRootLogger()
	logService logging.Service
	server     struct {
		host string
		port int
	}
	createCRD bool
)

func init() {
	f := cmdMain.Flags()
	f.StringVar(&server.host, "server.host", defaultServerHost, "Host to listen on")
	f.IntVar(&server.port, "server.port", defaultServerPort, "Port to listen on")
	f.StringVar(&logLevel, "log.level", defaultLogLevel, "Set initial log level")
	f.BoolVar(&createCRD, "operator.create-crd", true, "Disable to avoid create the custom resource definition")
}

func main() {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	cmdMain.Execute()
}

// Run the operator
func cmdMainRun(cmd *cobra.Command, args []string) {
	var err error
	logService, err = logging.NewService(logLevel)
	if err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to initialize log service")
	}

	// Log version
	cliLog.Info().Msgf("Starting arangodb-operator, version %s build %s", projectVersion, projectBuild)

	// Get environment
	namespace := os.Getenv(constants.EnvOperatorPodNamespace)
	if len(namespace) == 0 {
		cliLog.Fatal().Msgf("%s environment variable missing", constants.EnvOperatorPodNamespace)
	}
	name := os.Getenv(constants.EnvOperatorPodName)
	if len(name) == 0 {
		cliLog.Fatal().Msgf("%s environment variable missing", constants.EnvOperatorPodName)
	}

	// Get host name
	id, err := os.Hostname()
	if err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to get hostname")
	}

	// Create k8s client
	kubecli, err := k8sutil.NewKubeClient()
	if err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to create kubernetes client")
	}

	//http.HandleFunc(probe.HTTPReadyzEndpoint, probe.ReadyzHandler)
	http.Handle("/metrics", prometheus.Handler())
	listenAddr := net.JoinHostPort(server.host, strconv.Itoa(server.port))
	go http.ListenAndServe(listenAddr, nil)

	rl, err := resourcelock.New(resourcelock.EndpointsResourceLock,
		namespace,
		"arangodb-operator",
		kubecli.CoreV1(),
		resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: createRecorder(cliLog, kubecli, name, namespace),
		})
	if err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to create resource lock")
	}

	leaderelection.RunOrDie(leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: 15 * time.Second,
		RenewDeadline: 10 * time.Second,
		RetryPeriod:   2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(stop <-chan struct{}) {
				run(stop, namespace, name)
			},
			OnStoppedLeading: func() {
				cliLog.Fatal().Msg("Leader election lost")
			},
		},
	})
}

// run the operator
func run(stop <-chan struct{}, namespace, name string) {
	cfg, deps, err := newControllerConfigAndDeps(namespace, name)
	if err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to create controller config & deps")
	}

	//	startChaos(context.Background(), cfg.KubeCli, cfg.Namespace, chaosLevel)

	c, err := controller.NewController(cfg, deps)
	if err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to create controller")
	}
	if err := c.Start(); err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to start controller")
	}
}

// newControllerConfigAndDeps creates controller config & dependencies.
func newControllerConfigAndDeps(namespace, name string) (controller.Config, controller.Dependencies, error) {
	kubecli, err := k8sutil.NewKubeClient()
	if err != nil {
		return controller.Config{}, controller.Dependencies{}, maskAny(err)
	}

	serviceAccount, err := getMyPodServiceAccount(kubecli, namespace, name)
	if err != nil {
		return controller.Config{}, controller.Dependencies{}, maskAny(fmt.Errorf("Failed to get my pod's service account: %s", err))
	}

	kubeExtCli, err := k8sutil.NewKubeExtClient()
	if err != nil {
		return controller.Config{}, controller.Dependencies{}, maskAny(fmt.Errorf("Failed to create k8b api extensions client: %s", err))
	}
	databaseCRCli, err := client.NewInCluster()
	if err != nil {
		return controller.Config{}, controller.Dependencies{}, maskAny(fmt.Errorf("Failed to created versioned client: %s", err))
	}

	cfg := controller.Config{
		Namespace:      namespace,
		ServiceAccount: serviceAccount,
		CreateCRD:      createCRD,
	}
	deps := controller.Dependencies{
		Log:           logService.MustGetLogger("controller"),
		KubeCli:       kubecli,
		KubeExtCli:    kubeExtCli,
		DatabaseCRCli: databaseCRCli,
	}

	return cfg, deps, nil
}

// getMyPodServiceAccount looks up the service account of the pod with given name in given namespace
func getMyPodServiceAccount(kubecli kubernetes.Interface, namespace, name string) (string, error) {
	var sa string
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
		return nil
	}
	if err := retry.Retry(op, time.Minute*5); err != nil {
		return "", maskAny(err)
	}
	return sa, nil
}

func createRecorder(log zerolog.Logger, kubecli kubernetes.Interface, name, namespace string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(func(format string, args ...interface{}) {
		log.Info().Msgf(format, args...)
	})
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: v1core.New(kubecli.Core().RESTClient()).Events(namespace)})
	return eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: name})
}
