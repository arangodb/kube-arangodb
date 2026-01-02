//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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
	goHttp "net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/exporter"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
	ktls "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tls"
)

var (
	cmdExporter = &cobra.Command{
		Use: "exporter",
		Run: cmdExporterCheck,
	}

	exporterInput struct {
		listenAddress string

		endpoints []string
		jwtFile   string
		timeout   time.Duration

		keyfile string
	}
)

func init() {
	f := cmdExporter.PersistentFlags()

	f.StringVar(&exporterInput.listenAddress, "server.address", ":9101", "Address the exporter will listen on (IP:port)")
	f.StringVar(&exporterInput.keyfile, "ssl.keyfile", "", "File containing TLS certificate used for the metrics server. Format equal to ArangoDB keyfiles")

	f.StringSliceVar(&exporterInput.endpoints, "arangodb.endpoint", []string{"http://127.0.0.1:8529"}, "Endpoints used to reach the ArangoDB server")
	f.StringVar(&exporterInput.jwtFile, "arangodb.jwt-file", "", "File containing the JWT for authentication with ArangoDB server")
	f.DurationVar(&exporterInput.timeout, "arangodb.timeout", time.Second*15, "Timeout of statistics requests for ArangoDB")

	cmdMain.AddCommand(cmdExporter)
}

func cmdExporterCheck(cmd *cobra.Command, args []string) {
	if err := cmdExporterCheckE(); err != nil {
		logger.Err(err).Error("Fatal")
		os.Exit(1)
	}
}

func cmdExporterCheckE() error {
	ctx := util.CreateSignalContext(context.Background())

	if len(exporterInput.endpoints) < 1 {
		return errors.Errorf("Requires at least one ArangoDB Endpoint to be present")
	}

	p, err := exporter.NewPassthru(func() (string, error) {
		if exporterInput.jwtFile == "" {
			return "", nil
		}

		data, err := os.ReadFile(exporterInput.jwtFile)
		if err != nil {
			return "", err
		}

		return string(data), nil
	}, false, 15*time.Second, exporterInput.endpoints...)
	if err != nil {
		return err
	}

	mon := exporter.NewMonitor(exporterInput.endpoints[0], func() (string, error) {
		if exporterInput.jwtFile == "" {
			return "", nil
		}

		data, err := os.ReadFile(exporterInput.jwtFile)
		if err != nil {
			return "", err
		}

		return string(data), nil
	}, false, 15*time.Second)

	go mon.UpdateMonitorStatus(ctx)

	server, err := operatorHTTP.NewServer(ctx,
		operatorHTTP.DefaultHTTPServerSettings,
		operatorHTTP.WithServeMux(func(in *goHttp.ServeMux) {
			in.Handle("/metrics", p)
		}, func(in *goHttp.ServeMux) {
			in.HandleFunc("/", func(w goHttp.ResponseWriter, r *goHttp.Request) {
				w.Write([]byte(`<html>
             <head><title>ArangoDB Exporter</title></head>
             <body>
             <h1>ArangoDB Exporter</h1>
             <p><a href='/metrics'>Metrics</a></p>
             </body>
             </html>`))
			})
		}),
		ktls.WithTLSConfigFetcherGen(func() ktls.TLSConfigFetcher {
			if exporterInput.keyfile != "" {
				return ktls.NewKeyfileTLSConfig(exporterInput.keyfile)
			}

			return nil
		}),
	)
	if err != nil {
		return err
	}

	return server.StartAddr(ctx, exporterInput.listenAddress)
}
