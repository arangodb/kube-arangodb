//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package k8sutil

import (
	"context"
	"fmt"
	"math/rand"
	goHttp "net/http"
	"net/url"
	goStrings "strings"
	"sync"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

func PortForwarderServiceDiscovery(ns, name string) PortForwarderPodDiscovery {
	return func(config *restclient.Config) (string, string, error) {
		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			return "", "", err
		}

		svc, err := client.CoreV1().Services(ns).Get(context.Background(), name, meta.GetOptions{})
		if err != nil {
			return "", "", err
		}

		pods, err := client.CoreV1().Pods(ns).List(context.Background(), meta.ListOptions{
			LabelSelector: labels.FormatLabels(svc.Spec.Selector),
		})
		if err != nil {
			return "", "", err
		}

		rand.Shuffle(len(pods.Items), func(i, j int) {

		})

		for _, pod := range pods.Items {
			for _, c := range pod.Status.Conditions {
				if c.Type == core.PodReady && c.Status == core.ConditionTrue {
					return pod.GetNamespace(), pod.GetName(), nil
				}
			}
		}

		return "", "", errors.Errorf("Unable to find proper pod")
	}
}

func PortForwarderPod(ns, name string) PortForwarderPodDiscovery {
	return func(config *restclient.Config) (string, string, error) {
		return ns, name, nil
	}
}

func newDialer(config *restclient.Config, namespace, name string) (httpstream.Dialer, error) {
	roundTripper, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", namespace, name)
	hostIP := goStrings.TrimPrefix(config.Host, "https://")
	serverURL := url.URL{Scheme: "https", Path: path, Host: hostIP}

	dialer := spdy.NewDialer(upgrader, &goHttp.Client{Transport: roundTripper}, goHttp.MethodPost, &serverURL)

	return dialer, nil
}

func NewPortForwarder(config *restclient.Config, discovery PortForwarderPodDiscovery) PortForwarder {
	return portForwarder{
		discovery: discovery,
		config:    config,
	}
}

type PortForwarder interface {
	Start(ctx context.Context, ports ...string) (PortForwarderInstance, error)
}

type PortForwarderPodDiscovery func(config *restclient.Config) (string, string, error)

type portForwarder struct {
	discovery PortForwarderPodDiscovery

	config *restclient.Config
}

func (p portForwarder) Start(ctx context.Context, ports ...string) (PortForwarderInstance, error) {
	ns, n, err := p.discovery(p.config)
	if err != nil {
		return nil, err
	}

	d, err := newDialer(p.config, ns, n)
	if err != nil {
		return nil, err
	}

	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)

	var ps []string

	for id := range ports {
		if p := strings.Split(ports[id], ":"); len(p) == 2 {
			ps = append(ps, ports[id])
		} else {
			ps = append(ps, fmt.Sprintf("0:%s", ports[id]))
		}
	}

	forwarder, err := portforward.New(d, ps, stopChan, readyChan, nil, nil)
	if err != nil {
		return nil, err
	}

	i := &portForwarderInstance{
		pf:         forwarder,
		stoppedCh:  make(chan struct{}),
		stoppedErr: nil,
		stopChan:   stopChan,
		readyChan:  readyChan,
	}

	go func() {
		defer i.stop()
		<-ctx.Done()
	}()

	go i.run()

	return i, nil
}

type PortForwarderInstance interface {
	Wait() error
	Close() error
	Stopped()

	GetPorts() ([]portforward.ForwardedPort, error)
	GetPort(port uint16) (uint16, error)
}

type portForwarderInstance struct {
	lock sync.Mutex

	pf *portforward.PortForwarder

	stoppedCh  chan struct{}
	stoppedErr error

	stopChan, readyChan chan struct{}
}

func (p *portForwarderInstance) GetPorts() ([]portforward.ForwardedPort, error) {
	return p.pf.GetPorts()
}

func (p *portForwarderInstance) GetPort(port uint16) (uint16, error) {
	ports, err := p.GetPorts()
	if err != nil {
		return 0, err
	}

	for _, v := range ports {
		if v.Remote == port {
			return v.Local, nil
		}
	}

	return 0, errors.Errorf("Port %d not forwarded", port)
}

func (p *portForwarderInstance) stop() {
	p.lock.Lock()
	defer p.lock.Unlock()

	select {
	case <-p.stopChan:
	default:
		close(p.stopChan)
	}
}

func (p *portForwarderInstance) Close() error {
	p.stop()

	<-p.stoppedCh

	return p.stoppedErr
}

func (p *portForwarderInstance) Wait() error {
	select {
	case <-p.readyChan:
		return nil
	case <-p.stoppedCh:
		return p.stoppedErr
	}
}

func (p *portForwarderInstance) Stopped() {
	<-p.stoppedCh
}

func (p *portForwarderInstance) run() {
	defer close(p.stoppedCh)

	p.stoppedErr = p.pf.ForwardPorts()
}
