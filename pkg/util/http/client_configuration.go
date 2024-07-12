//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package http

import (
	"net"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

func InitConfiguration(cmd *cobra.Command) error {
	return configuration.Init(cmd)
}

const (
	defaultTransportKeepAlive         bool = true
	defaultTransportForceAttemptHTTP2 bool = false

	defaultTransportMaxIdleConns int = 100

	defaultTransportDialTimeout           = 30 * time.Second
	defaultTransportKeepAliveTimeout      = 90 * time.Second
	defaultTransportKeepAliveTimeoutShort = 100 * time.Millisecond
	defaultTransportIdleConnTimeout       = 90 * time.Second
	defaultTransportIdleConnTimeoutShort  = 100 * time.Millisecond
	defaultTransportTLSHandshakeTimeout   = 10 * time.Second
	defaultTransportExpectContinueTimeout = 1 * time.Second
)

var configuration = configurationObject{
	TransportKeepAlive:             defaultTransportKeepAlive,
	TransportForceAttemptHTTP2:     defaultTransportForceAttemptHTTP2,
	TransportMaxIdleConns:          defaultTransportMaxIdleConns,
	TransportDialTimeout:           defaultTransportDialTimeout,
	TransportKeepAliveTimeout:      defaultTransportKeepAliveTimeout,
	TransportKeepAliveTimeoutShort: defaultTransportKeepAliveTimeoutShort,
	TransportIdleConnTimeout:       defaultTransportIdleConnTimeout,
	TransportIdleConnTimeoutShort:  defaultTransportIdleConnTimeoutShort,
	TransportTLSHandshakeTimeout:   defaultTransportTLSHandshakeTimeout,
	TransportExpectContinueTimeout: defaultTransportExpectContinueTimeout,
}

type configurationObject struct {
	TransportKeepAlive         bool
	TransportForceAttemptHTTP2 bool

	TransportMaxIdleConns int

	TransportDialTimeout           time.Duration
	TransportKeepAliveTimeout      time.Duration
	TransportKeepAliveTimeoutShort time.Duration
	TransportIdleConnTimeout       time.Duration
	TransportIdleConnTimeoutShort  time.Duration
	TransportTLSHandshakeTimeout   time.Duration
	TransportExpectContinueTimeout time.Duration
}

func (c *configurationObject) Init(cmd *cobra.Command) error {
	if c == nil {
		return nil
	}

	f := cmd.PersistentFlags()

	f.BoolVar(&configuration.TransportKeepAlive, "http1.keep-alive", defaultTransportKeepAlive, "If false, disables HTTP keep-alives and will only use the connection to the server for a single HTTP request")
	f.BoolVar(&configuration.TransportForceAttemptHTTP2, "http1.force-attempt-http2", defaultTransportForceAttemptHTTP2, "controls whether HTTP/2 is enabled")

	f.IntVar(&configuration.TransportMaxIdleConns, "http1.transport.max-idle-conns", defaultTransportMaxIdleConns, "Maximum number of idle (keep-alive) connections across all hosts. Zero means no limit")

	f.DurationVar(&configuration.TransportDialTimeout, "http1.transport.dial-timeout", defaultTransportDialTimeout, "Maximum amount of time a dial will wait for a connect to complete")
	f.DurationVar(&configuration.TransportKeepAliveTimeout, "http1.transport.keep-alive-timeout", defaultTransportKeepAliveTimeout, "Interval between keep-alive probes for an active network connection")
	f.DurationVar(&configuration.TransportKeepAliveTimeoutShort, "http1.transport.keep-alive-timeout-short", defaultTransportKeepAliveTimeoutShort, "Interval between keep-alive probes for an active network connection")

	f.DurationVar(&configuration.TransportIdleConnTimeout, "http1.transport.idle-conn-timeout", defaultTransportIdleConnTimeout, "Maximum amount of time an idle (keep-alive) connection will remain idle before closing itself. Zero means no limit")
	f.DurationVar(&configuration.TransportIdleConnTimeoutShort, "http1.transport.idle-conn-timeout-short", defaultTransportIdleConnTimeoutShort, "Maximum amount of time an idle (keep-alive) connection will remain idle before closing itself. Zero means no limit")
	f.DurationVar(&configuration.TransportTLSHandshakeTimeout, "http1.transport.tls-handshake-timeout", defaultTransportTLSHandshakeTimeout, "Maximum amount of time to wait for a TLS handshake. Zero means no timeout")
	f.DurationVar(&configuration.TransportExpectContinueTimeout, "http1.transport.except-continue-timeout", defaultTransportExpectContinueTimeout, "")

	if err := f.MarkHidden("http1.transport.except-continue-timeout"); err != nil {
		return err
	}

	if err := f.MarkHidden("http1.force-attempt-http2"); err != nil {
		return err
	}

	return nil
}

func (c *configurationObject) DefaultTransport(in *http.Transport) {
	if c == nil {
		return
	}
	in.Proxy = http.ProxyFromEnvironment
	in.ForceAttemptHTTP2 = c.TransportForceAttemptHTTP2
	in.DialContext = (&net.Dialer{
		Timeout:   c.TransportDialTimeout,
		KeepAlive: c.TransportKeepAliveTimeout,
		DualStack: true,
	}).DialContext

	in.MaxIdleConns = c.TransportMaxIdleConns
	in.IdleConnTimeout = c.TransportIdleConnTimeout
	in.TLSHandshakeTimeout = c.TransportTLSHandshakeTimeout
	in.ExpectContinueTimeout = c.TransportExpectContinueTimeout
	in.DisableKeepAlives = !c.TransportKeepAlive
}

func (c *configurationObject) ShortTransport(in *http.Transport) {
	if c == nil {
		return
	}
	in.DialContext = (&net.Dialer{
		Timeout:   c.TransportDialTimeout,
		KeepAlive: c.TransportKeepAliveTimeoutShort,
		DualStack: true,
	}).DialContext

	in.IdleConnTimeout = c.TransportIdleConnTimeoutShort
}
