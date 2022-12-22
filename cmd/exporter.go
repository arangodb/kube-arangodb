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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/exporter"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

var (
	cmdExporter = &cobra.Command{
		Use: "exporter",
		Run: cmdExporterCheck,
	}

	exporterInput struct {
		listenAddress string

		endpoint string
		jwtFile  string
		timeout  time.Duration

		keyfile string
	}
)

func init() {
	f := cmdExporter.PersistentFlags()

	f.StringVar(&exporterInput.listenAddress, "server.address", ":9101", "Address the exporter will listen on (IP:port)")
	f.StringVar(&exporterInput.keyfile, "ssl.keyfile", "", "File containing TLS certificate used for the metrics server. Format equal to ArangoDB keyfiles")

	f.StringVar(&exporterInput.endpoint, "arangodb.endpoint", "http://127.0.0.1:8529", "Endpoint used to reach the ArangoDB server")
	f.StringVar(&exporterInput.jwtFile, "arangodb.jwt-file", "", "File containing the JWT for authentication with ArangoDB server")
	f.DurationVar(&exporterInput.timeout, "arangodb.timeout", time.Second*15, "Timeout of statistics requests for ArangoDB")

	cmdMain.AddCommand(cmdExporter)
}

func cmdExporterCheck(cmd *cobra.Command, args []string) {
	if err := cmdExporterCheckE(); err != nil {
		log.Error().Err(err).Msgf("Fatal")
		os.Exit(1)
	}
}

func onSigterm(f func()) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer f()
		<-sigs
	}()
}

func cmdExporterCheckE() error {
	p, err := exporter.NewPassthru(exporterInput.endpoint, func() (string, error) {
		if exporterInput.jwtFile == "" {
			return "", nil
		}

		data, err := os.ReadFile(exporterInput.jwtFile)
		if err != nil {
			return "", err
		}

		return string(data), nil
	}, false, 15*time.Second)
	if err != nil {
		return err
	}

	mon := exporter.NewMonitor(exporterInput.endpoint, func() (string, error) {
		if exporterInput.jwtFile == "" {
			return "", nil
		}

		data, err := os.ReadFile(exporterInput.jwtFile)
		if err != nil {
			return "", err
		}

		return string(data), nil
	}, false, 15*time.Second)

	go mon.UpdateMonitorStatus(util.CreateSignalContext(context.Background()))

	exporter := exporter.NewExporter(exporterInput.listenAddress, "/metrics", p)
	if exporterInput.keyfile != "" {
		if e, err := exporter.WithKeyfile(exporterInput.keyfile); err != nil {
			return err
		} else {
			if r, err := e.Start(); err != nil {
				return err
			} else {
				onSigterm(r.Stop)
				return r.Wait()
			}
		}
	} else {
		if r, err := exporter.Start(); err != nil {
			return err
		} else {
			onSigterm(r.Stop)
			return r.Wait()
		}
	}
}
