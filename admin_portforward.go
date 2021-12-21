//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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

package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	arangoUtil "github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	kubectlUtil "k8s.io/kubectl/pkg/util"
)

const (
	defaultProxyService = "arango-arango-deployment-operator"
	defaultNamespace    = "default"
	defaultPorts        = "9001:8528"

	ProxyServiceEnv   arangoUtil.EnvironmentVariable = "PROXY_SERVICE"
	ProxyNamespaceEnv arangoUtil.EnvironmentVariable = "PROXY_NAMESPACE"
	ProxyPortsEnv     arangoUtil.EnvironmentVariable = "PROXY_PORTS"
)

var (
	cmdProxy = &cobra.Command{
		Use:   "proxy",
		Short: "Proxy server to another k8s cluster",
		Long:  "Proxy server to another k8s cluster with kubeconfig file defined by KUBECONFIG env",
		Run:   cmdForwardPorts,
	}

	podForwardInput struct {
		proxyService   string
		proxyNamespace string
		proxyPorts     string

		restConfig *rest.Config
		// Pod is the selected pod for this port forwarding
		pod *v1.Pod
		// ports is a list of {localPort}:{podPort} that will be selected to expose the podPort on localPort
		ports []string
		// Steams configures where to write or read input from
		streams genericclioptions.IOStreams
		// stopCh is the channel used to manage the port forward lifecycle
		stopCh <-chan struct{}
		// readyCh communicates when the tunnel is ready to receive traffic
		readyCh chan struct{}
	}
)

func init() {
	cmdProxy.Flags().StringVarP(&podForwardInput.proxyService, "proxy.service", "s", ProxyServiceEnv.GetOrDefault(defaultProxyService),
		"Name of the Service on remote k8s cluster to connect to")
	cmdProxy.Flags().StringVarP(&podForwardInput.proxyNamespace, "proxy.namespace", "n", ProxyNamespaceEnv.GetOrDefault(defaultNamespace),
		"Name of the Namespace on remote k8s cluster to use")
	cmdProxy.Flags().StringVarP(&podForwardInput.proxyPorts, "proxy.ports", "p", ProxyPortsEnv.GetOrDefault(defaultPorts),
		"List of ports forwarding in form of: {from}:{to},{from}:{to}")
}

func handleForwarderErrors() {
	runtime.ErrorHandlers = append(runtime.ErrorHandlers, func(err error) {

		// trigger container restart on failure to reattach to the new pod
		if strings.Contains(err.Error(), "container not running") {
			cliLog.Error().Msg("target container not exist - quiting")
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}
	})
}

func cmdForwardPorts(_ *cobra.Command, _ []string) {
	handleForwarderErrors()
	ctx := arangoUtil.CreateSignalContext(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	cliLog.Info().Msg(fmt.Sprintf("Starting proxy server on ports %s for %s service in %s namespace",
		podForwardInput.proxyPorts, podForwardInput.proxyService, podForwardInput.proxyNamespace))

	config, err := k8sutil.NewKubeConfig()
	if err != nil {
		cliLog.Panic().Err(err).Msg("cannot load kubeconfig file")
	}
	podForwardInput.restConfig = config

	stopCh := make(chan struct{}, 1)
	podForwardInput.stopCh = stopCh
	podForwardInput.readyCh = make(chan struct{})
	podForwardInput.streams = genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cliLog.Info().Msg("Exit...")
		close(stopCh)
		wg.Done()
	}()

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	pod, svc, err := getFirstPodForSvc(ctx, podForwardInput.proxyService, podForwardInput.proxyNamespace, *k8sClient)
	if err != nil {
		cliLog.Panic().Err(err).Msg("error on looking service pod")
	}
	podForwardInput.pod = pod

	go watchPodForRestart(ctx, pod, podForwardInput.proxyNamespace, *k8sClient)

	podForwardInput.ports, err = translateServicePortToTargetPort(strings.Split(podForwardInput.proxyPorts, ","), svc, pod)
	if err != nil {
		cliLog.Panic().Err(err).Msg("cannot translate Service Port to Pod Port")
	}

	go func() {
		err := PortForwardToPod()
		if err != nil {
			cliLog.Panic().Err(err).Msg("pod forward request failed")
		}
	}()

	// wait till portForward connection will be ready
	select {
	case <-podForwardInput.readyCh:
		break
	}
	cliLog.Info().Msg("Port forwarding is ready to get traffic")

	wg.Wait()
}

func PortForwardToPod() error {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", podForwardInput.pod.Namespace, podForwardInput.pod.Name)
	hostIP := strings.TrimLeft(podForwardInput.restConfig.Host, "htps:/")

	transport, upgrader, err := spdy.RoundTripperFor(podForwardInput.restConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url.URL{Scheme: "https", Path: path, Host: hostIP})
	fw, err := portforward.New(dialer, podForwardInput.ports, podForwardInput.stopCh, podForwardInput.readyCh, podForwardInput.streams.Out, podForwardInput.streams.ErrOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

func getFirstPodForSvc(ctx context.Context, proxyService, proxyNamespace string, k8sClient kubernetes.Clientset) (*corev1.Pod, *corev1.Service, error) {
	svc, err := k8sClient.CoreV1().Services(proxyNamespace).Get(ctx, proxyService, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}

	set := labels.Set(svc.Spec.Selector)
	listOptions := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
	pods, err := k8sClient.CoreV1().Pods(proxyNamespace).List(context.Background(), listOptions)
	if err != nil {
		return nil, nil, err
	}

	for _, pod := range pods.Items {
		return &pod, svc, nil
	}
	return nil, nil, errors.New("no pod found")
}

func watchPodForRestart(ctx context.Context, pod *corev1.Pod, proxyNamespace string, k8sClient kubernetes.Clientset) {
	podWatch, err := k8sClient.CoreV1().Pods(proxyNamespace).Watch(ctx, metav1.ListOptions{FieldSelector: fields.OneTermEqualSelector("metadata.name", pod.Name).String()})
	if err != nil {
		cliLog.Panic().Err(err).Msg("cannot watch pod")
	}
	defer podWatch.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case obj, ok := <-podWatch.ResultChan():
			if !ok {
				cliLog.Error().Err(err).Msg("result channel bad - quiting")
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
				return
			}

			podRefreshed, ok := obj.Object.(*corev1.Pod)
			if !ok {
				cliLog.Error().Err(err).Msg("failed to get pod - quiting")
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
				return
			}

			if podRefreshed.DeletionTimestamp != nil {
				cliLog.Error().Msg("pod has been restarted - quiting")
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
				return
			}

			if podRefreshed.UID != pod.UID {
				cliLog.Error().Msg("pod UID has been changed - quiting")
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
				return
			}
		case <-time.After(5 * time.Second):
			// manual refresh
			podRefreshed, err := k8sClient.CoreV1().Pods(proxyNamespace).Get(ctx, pod.Name, metav1.GetOptions{})
			if err != nil {
				cliLog.Error().Err(err).Msg("failed to get pod - quiting")
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
				return
			}
			if podRefreshed.DeletionTimestamp != nil {
				cliLog.Error().Msg("pod has been restarted - quiting")
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
				return
			}

			if podRefreshed.UID != pod.UID {
				cliLog.Error().Msg("pod UID has been changed - quiting")
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
				return
			}
		}
	}
}

// Translates service port to target port
// It rewrites ports as needed if the Service port declares targetPort.
// It returns an error when a named targetPort can't find a match in the pod, or the Service did not declare
// the port.
func translateServicePortToTargetPort(ports []string, svc *corev1.Service, pod *corev1.Pod) ([]string, error) {
	var translated []string
	for _, port := range ports {
		localPort, remotePort := splitPort(port)

		portnum, err := strconv.Atoi(remotePort)
		if err != nil {
			svcPort, err := kubectlUtil.LookupServicePortNumberByName(*svc, remotePort)
			if err != nil {
				return nil, err
			}
			portnum = int(svcPort)

			if localPort == remotePort {
				localPort = strconv.Itoa(portnum)
			}
		}
		containerPort, err := kubectlUtil.LookupContainerPortNumberByServicePort(*svc, *pod, int32(portnum))
		if err != nil {
			// can't resolve a named port, or Service did not declare this port, return an error
			return nil, err
		}

		// convert the resolved target port back to a string
		remotePort = strconv.Itoa(int(containerPort))

		translated = append(translated, fmt.Sprintf("%s:%s", localPort, remotePort))
	}
	return translated, nil
}

// splitPort splits port string which is in form of LOCAL_PORT:REMOTE_PORT,LOCAL_PORT:REMOTE_PORT
func splitPort(port string) (local, remote string) {
	parts := strings.Split(port, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return parts[0], parts[0]
}
