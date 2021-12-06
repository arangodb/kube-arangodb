package main

import (
	"errors"
	"fmt"
	arangoUtil "github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/spf13/cobra"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
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

	proxyService   string
	proxyNamespace string
	proxyPorts     string
)

func init() {
	cmdProxy.Flags().StringVarP(&proxyService, "proxy.service", "s", ProxyServiceEnv.GetOrDefault(defaultProxyService),
		"Name of the Service on remote k8s cluster to connect to")
	cmdProxy.Flags().StringVarP(&proxyNamespace, "proxy.namespace", "n", ProxyNamespaceEnv.GetOrDefault(defaultNamespace),
		"Name of the Namespace on remote k8s cluster to use")
	cmdProxy.Flags().StringVarP(&proxyPorts, "proxy.ports", "p", ProxyPortsEnv.GetOrDefault(defaultPorts),
		"lList of ports forwarding in form of: {from}:{to},{from}:{to}")
}

type PortForwardToPodRequest struct {
	// RestConfig is the kubernetes config
	RestConfig *rest.Config
	// Pod is the selected pod for this port forwarding
	Pod *v1.Pod
	// Ports is a list of {localPort}:{podPort} that will be selected to expose the podPort on localPort
	Ports []string
	// Steams configures where to write or read input from
	Streams genericclioptions.IOStreams
	// StopCh is the channel used to manage the port forward lifecycle
	StopCh <-chan struct{}
	// ReadyCh communicates when the tunnel is ready to receive traffic
	ReadyCh chan struct{}
}

func cmdForwardPorts(_ *cobra.Command, _ []string) {
	var wg sync.WaitGroup
	wg.Add(1)

	cliLog.Info().Msg(fmt.Sprintf("Starting proxy server on ports %s for %s service in %s namespace", proxyPorts, proxyService, proxyNamespace))

	config, err := k8sutil.NewKubeConfig()
	if err != nil {
		cliLog.Panic().Err(err).Msg("cannot load kubeconfig file")
	}

	// stopCh control the port forwarding lifecycle. When it gets closed the
	// port forward will terminate
	stopCh := make(chan struct{}, 1)
	// readyCh communicate when the port forward is ready to get traffic
	readyCh := make(chan struct{})
	// stream is used to tell the port forwarder where to place its output or
	// where to expect input if needed. For the port forwarding we just need
	// the output eventually
	stream := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	// managing termination signal from the terminal. As you can see the stopCh
	// gets closed to gracefully handle its termination.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cliLog.Info().Msg("Exit...")
		close(stopCh)
		wg.Done()
	}()

	pod, svc, err := getFirstPodForSvc(proxyService, proxyNamespace, config)
	if err != nil {
		cliLog.Panic().Err(err).Msg("error on looking service pod")
	}

	podPorts, err := translateServicePortToTargetPort(strings.Split(proxyPorts, ","), svc, pod)
	if err != nil {
		cliLog.Panic().Err(err).Msg("cannot translate Service Port to Pod Port")
	}

	go func() {
		err := PortForwardToPod(PortForwardToPodRequest{
			RestConfig: config,
			Pod:        pod,
			Ports:      podPorts,
			Streams:    stream,
			StopCh:     stopCh,
			ReadyCh:    readyCh,
		})
		if err != nil {
			cliLog.Panic().Err(err).Msg("pod forward request failed")
		}
	}()
	select {
	case <-readyCh:
		break
	}
	cliLog.Info().Msg("Port forwarding is ready to get traffic")

	wg.Wait()
}

func PortForwardToPod(req PortForwardToPodRequest) error {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", req.Pod.Namespace, req.Pod.Name)
	hostIP := strings.TrimLeft(req.RestConfig.Host, "https://")

	transport, upgrader, err := spdy.RoundTripperFor(req.RestConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url.URL{Scheme: "https", Path: path, Host: hostIP})
	fw, err := portforward.New(dialer, req.Ports, req.StopCh, req.ReadyCh, req.Streams.Out, req.Streams.ErrOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

func getFirstPodForSvc(proxyService, proxyNamespace string, config *restclient.Config) (*corev1.Pod, *corev1.Service, error) {
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	ctx := arangoUtil.CreateSignalContext(context.Background())

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
