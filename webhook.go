//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package main

import (
	"net/http"

	"github.com/arangodb/kube-arangodb/pkg/webhook"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	cmdWebhook = &cobra.Command{
		Use:    "webhook",
		RunE:   webhookCommand,
		Hidden: true,
	}
)

var (
	webhookConfig struct {
		host, cert, key string
		port            uint16
	}
)

func init() {
	f := cmdWebhook.Flags()

	f.StringVar(&webhookConfig.host, "host", "0.0.0.0", "Webhook Server host")
	f.Uint16Var(&webhookConfig.port, "port", 443, "Webhook Server port")
	f.StringVar(&webhookConfig.cert, "cert", "./cert.pem", "Webhook Server cert in PEM format")
	f.StringVar(&webhookConfig.key, "key", "./key.pem", "Webhook Server key in PEM format")

	cmdMain.AddCommand(cmdWebhook)
}

func webhookCommand(cmd *cobra.Command, args []string) error {
	log := log.Logger.Level(zerolog.DebugLevel)

	log.Info().Msg("Creating server")

	server, err := webhook.NewWebServer(log, webhookConfig.host, webhookConfig.port, webhookConfig.key, webhookConfig.cert, "")
	if err != nil {
		return err
	}

	log.Info().Msg("Starting server")

	if err := server.ListenAndServeTLS("", ""); err != nil {
		if err != http.ErrServerClosed {
			return err
		}
	}

	return nil
}
