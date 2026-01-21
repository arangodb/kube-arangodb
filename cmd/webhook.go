//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/handlers/permission"
	"github.com/arangodb/kube-arangodb/pkg/handlers/scheduler"
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/http"
	ktls "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tls"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/webhook"
)

var (
	cmdWebhook = &cobra.Command{
		Use: "webhook",
		Run: cmdWebhookCheck,
	}

	webhookInput struct {
		listenAddress string

		secretName, secretNamespace string
	}
)

func init() {
	f := cmdWebhook.PersistentFlags()

	f.StringVar(&webhookInput.listenAddress, "server.address", "0.0.0.0:8828", "Address the webhook will listen on (IP:port)")
	f.StringVar(&webhookInput.secretName, "ssl.secret.name", "", "Secret Name containing TLS certificate used for the metrics server")
	f.StringVar(&webhookInput.secretNamespace, "ssl.secret.namespace", os.Getenv(utilConstants.EnvOperatorPodNamespace), "Secret Name containing TLS certificate used for the metrics server")

	cmdMain.AddCommand(cmdWebhook)
}

func cmdWebhookCheck(cmd *cobra.Command, args []string) {
	if err := cmdWebhookCheckE(); err != nil {
		logger.Err(err).Error("Fatal")
		os.Exit(1)
	}
}

func cmdWebhookCheckE() error {
	ctx := util.CreateSignalContext(context.Background())

	client, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return errors.Errorf("Unable to get client")
	}

	var admissions webhook.Admissions

	admissions = append(admissions, scheduler.WebhookAdmissions(client)...)
	admissions = append(admissions, permission.WebhookAdmissions(client)...)

	server, err := webhookServer(ctx, client, admissions...)
	if err != nil {
		return err
	}

	logger.Str("addr", webhookInput.listenAddress).Info("Starting Webhook Server")

	return server.StartAddr(ctx, webhookInput.listenAddress)
}

func webhookServer(ctx context.Context, client kclient.Client, admissions ...webhook.Admission) (http.Server, error) {
	return http.NewServer(ctx,
		http.DefaultHTTPServerSettings,
		ktls.WithTLSConfigFetcherGen(func() ktls.TLSConfigFetcher {
			if webhookInput.secretName != "" && webhookInput.secretNamespace != "" {
				return ktls.NewSecretTLSConfig(client.Kubernetes().CoreV1().Secrets(webhookInput.secretNamespace), webhookInput.secretName)
			}

			return ktls.NewSelfSignedTLSConfig("operator")
		}),
		http.WithServeMux(
			func(in *goHttp.ServeMux) {
				in.HandleFunc("/ready", func(writer goHttp.ResponseWriter, request *goHttp.Request) {
					writer.WriteHeader(goHttp.StatusOK)
				})
				in.HandleFunc("/health", func(writer goHttp.ResponseWriter, request *goHttp.Request) {
					writer.WriteHeader(goHttp.StatusOK)
				})
			},
			webhook.Admissions(admissions).Register(),
		),
	)
}
