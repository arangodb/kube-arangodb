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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb-helper/go-certificates"
	"github.com/arangodb/go-driver/jwt"
	"github.com/arangodb/go-driver/v2/connection"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	secretv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

const ArgDeploymentName = "deployment-name"

func init() {
	var deploymentName string

	cmdMain.AddCommand(cmdAdmin)
	cmdAdmin.AddCommand(cmdAgency)

	cmdAgency.AddCommand(cmdAgencyDump)
	cmdAgencyDump.Flags().StringVarP(&deploymentName, ArgDeploymentName, "d", "",
		"necessary when more than one deployment exist within on namespace")

	cmdAgency.AddCommand(cmdAgencyState)
	cmdAgencyState.Flags().StringVarP(&deploymentName, ArgDeploymentName, "d", "",
		"necessary when more than one deployment exist within on namespace")
}

var cmdAdmin = &cobra.Command{
	Use:   "admin",
	Short: "Administration operations",
	Run:   adminShowUsage,
}

var cmdAgency = &cobra.Command{
	Use:   "agency",
	Short: "Agency operations",
	Run:   agencyShowUsage,
}

var cmdAgencyDump = &cobra.Command{
	Use:   "dump",
	Short: "Get agency dump",
	Long:  "It prints the agency history on the stdout",
	Run:   cmdGetAgencyDump,
}

var cmdAgencyState = &cobra.Command{
	Use:   "state",
	Short: "Get agency state",
	Long:  "It prints the agency current state on the stdout",
	Run:   cmdGetAgencyState,
}

func agencyShowUsage(cmd *cobra.Command, _ []string) {
	cmd.Usage()
}

func adminShowUsage(cmd *cobra.Command, _ []string) {
	cmd.Usage()
}

func cmdGetAgencyState(cmd *cobra.Command, _ []string) {
	deploymentName, _ := cmd.Flags().GetString(ArgDeploymentName)
	ctx := getInterruptionContext()
	d, certCA, auth, err := getDeploymentAndCredentials(ctx, deploymentName)
	if err != nil {
		logger.Err(err).Fatal("failed to create basic data for the connection")
	}

	if d.GetAcceptedSpec().GetMode() != api.DeploymentModeCluster {
		logger.Fatal("agency state does not work for the \"%s\" deployment \"%s\"", d.GetAcceptedSpec().GetMode(),
			d.GetName())
	}

	dnsName := k8sutil.CreatePodDNSName(d.GetObjectMeta(), api.ServerGroupAgents.AsRole(), d.Status.Members.Agents[0].ID)
	endpoint := getArangoEndpoint(d.GetAcceptedSpec().IsSecure(), dnsName)
	conn := createClient([]string{endpoint}, certCA, auth, connection.ApplicationJSON)
	leaderID, err := getAgencyLeader(ctx, conn)
	if err != nil {
		logger.Err(err).Fatal("failed to get leader ID")
	}

	dnsLeaderName := k8sutil.CreatePodDNSName(d.GetObjectMeta(), api.ServerGroupAgents.AsRole(), leaderID)
	leaderEndpoint := getArangoEndpoint(d.GetAcceptedSpec().IsSecure(), dnsLeaderName)
	conn = createClient([]string{leaderEndpoint}, certCA, auth, connection.PlainText)
	body, err := getAgencyState(ctx, conn)
	if body != nil {
		defer body.Close()
	}
	if err != nil {
		logger.Err(err).Fatal("can not get state of the agency")
	}

	// Print and receive parallelly.
	io.Copy(os.Stdout, body)
}

func cmdGetAgencyDump(cmd *cobra.Command, _ []string) {
	deploymentName, _ := cmd.Flags().GetString(ArgDeploymentName)
	ctx := getInterruptionContext()
	d, certCA, auth, err := getDeploymentAndCredentials(ctx, deploymentName)
	if err != nil {
		logger.Err(err).Fatal("failed to create basic data for the connection")
	}

	if d.GetAcceptedSpec().GetMode() != api.DeploymentModeCluster {
		logger.Fatal("agency dump does not work for the \"%s\" deployment \"%s\"", d.GetAcceptedSpec().GetMode(),
			d.GetName())
	}

	endpoint := getArangoEndpoint(d.GetAcceptedSpec().IsSecure(), k8sutil.CreateDatabaseClientServiceDNSName(d.GetObjectMeta()))
	conn := createClient([]string{endpoint}, certCA, auth, connection.ApplicationJSON)
	body, err := getAgencyDump(ctx, conn)
	if body != nil {
		defer body.Close()
	}
	if err != nil {
		logger.Err(err).Fatal("can not get dump")
	}

	// Print and receive parallelly.
	io.Copy(os.Stdout, body)
}

// getAgencyState returns the current state in the agency.
func getAgencyState(ctx context.Context, conn connection.Connection) (io.ReadCloser, error) {
	url := connection.NewUrl("_api", "agency", "read")
	data := []byte(`[["/"]]`)
	resp, body, err := connection.CallStream(ctx, conn, http.MethodPost, url, connection.WithBody(data))
	if err != nil {
		return nil, err
	}
	if resp.Code() != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("unexpected HTTP status from \"%s\" endpoint", url))
	}

	return body, nil
}

// getDeploymentAndCredentials returns deployment and necessary credentials to communicate with ArangoDB pods.
func getDeploymentAndCredentials(ctx context.Context,
	deploymentName string) (d api.ArangoDeployment, certCA *x509.CertPool, auth connection.Authentication, err error) {

	namespace := os.Getenv(constants.EnvOperatorPodNamespace)
	if len(namespace) == 0 {
		err = errors.New(fmt.Sprintf("\"%s\" environment variable missing", constants.EnvOperatorPodNamespace))
		return
	}

	client, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		err = errors.Newf("Client not initialised")
		return
	}

	kubeCli := client.Kubernetes()

	d, err = getDeployment(ctx, namespace, deploymentName)
	if err != nil {
		err = errors.WithMessage(err, "failed to get deployment")
		return
	}

	var secrets = kubeCli.CoreV1().Secrets(d.GetNamespace())
	certCA, err = getCACertificate(ctx, secrets, d.GetAcceptedSpec().TLS.GetCASecretName())
	if err != nil {
		err = errors.WithMessage(err, "failed to get CA certificate")
		return
	}

	if d.GetAcceptedSpec().IsAuthenticated() {
		auth, err = getJWTTokenFromSecrets(ctx, secrets, d.GetAcceptedSpec().Authentication.GetJWTSecretName())
		if err != nil {
			err = errors.WithMessage(err, "failed to get JWT token")
			return
		}
	}

	return
}

// getArangoEndpoint returns ArangoDB endpoint with scheme and port for the given dnsName.
func getArangoEndpoint(secure bool, dnsName string) string {
	if secure {
		return "https://" + net.JoinHostPort(dnsName, strconv.Itoa(shared.ArangoPort))
	}

	return "http://" + net.JoinHostPort(dnsName, strconv.Itoa(shared.ArangoPort))
}

// getAgencyLeader returns the leader ID of the agency.
func getAgencyLeader(ctx context.Context, conn connection.Connection) (string, error) {
	url := connection.NewUrl("_api", "agency", "config")
	output := make(map[string]interface{})
	resp, err := connection.CallGet(ctx, conn, url, &output)
	if err != nil {
		return "", err
	}
	if resp.Code() != http.StatusOK {
		return "", errors.New("unexpected HTTP status from agency-dump endpoint")
	}

	if leaderID, ok := output["leaderId"]; ok {
		if id, ok := leaderID.(string); ok {
			return id, nil
		}
	}

	return "", errors.New("failed get agency leader ID")
}

// getAgencyDump returns dump of the agency.
func getAgencyDump(ctx context.Context, conn connection.Connection) (io.ReadCloser, error) {
	url := connection.NewUrl("_api", "cluster", "agency-dump")
	resp, body, err := connection.CallStream(ctx, conn, http.MethodGet, url)
	if err != nil {
		return nil, err
	}
	if resp.Code() != http.StatusOK {
		return nil, errors.New("unexpected HTTP status from agency-dump endpoint")
	}

	return body, nil
}

type JWTAuthentication struct {
	key, value string
}

func (j JWTAuthentication) RequestModifier(r connection.Request) error {
	r.AddHeader(j.key, j.value)
	return nil
}

// createClient creates client for the provided credentials.
func createClient(endpoints []string, certCA *x509.CertPool, auth connection.Authentication,
	contentType string) connection.Connection {

	conf := connection.HttpConfiguration{
		Authentication: auth,
		ContentType:    contentType,
		Endpoint:       connection.NewEndpoints(endpoints...),
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certCA,
			},
		},
	}

	return connection.NewHttpConnection(conf)
}

// getJWTTokenFromSecrets returns token from the secret.
func getJWTTokenFromSecrets(ctx context.Context, secrets secretv1.ReadInterface, name string) (connection.Authentication, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()

	token, err := k8sutil.GetTokenSecret(ctxChild, secrets, name)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("failed to get secret \"%s\"", name))
	}

	bearerToken, err := jwt.CreateArangodJwtAuthorizationHeader(token, "kube-arangodb")
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("failed to create bearer token from secret \"%s\"", name))
	}

	return JWTAuthentication{key: "Authorization", value: bearerToken}, nil
}

// getCACertificate returns CA certificate from the secret.
func getCACertificate(ctx context.Context, secrets secretv1.ReadInterface, name string) (*x509.CertPool, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()

	s, err := secrets.Get(ctxChild, name, meta.GetOptions{})
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("failed to get secret \"%s\"", name))
	}

	if data, ok := s.Data[core.ServiceAccountRootCAKey]; ok {
		return certificates.LoadCertPool(string(data))
	}

	return nil, errors.New(fmt.Sprintf("the \"%s\" does not exist in the secret \"%s\"", core.ServiceAccountRootCAKey,
		name))
}

// getDeployment returns ArangoDeployment within the provided namespace.
// If there are more than two deployments within one namespace then
// deployment name must be provided, otherwise error is returned.
func getDeployment(ctx context.Context, namespace, deplName string) (api.ArangoDeployment, error) {
	client, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return api.ArangoDeployment{}, errors.Newf("Client not initialised")
	}

	extCli := client.Arango()

	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()

	deployments, err := extCli.DatabaseV1().ArangoDeployments(namespace).List(ctxChild, meta.ListOptions{})
	if err != nil {
		if api.IsNotFound(err) {
			return api.ArangoDeployment{}, errors.WithMessage(err, "there are no deployments")
		}
		return api.ArangoDeployment{}, errors.WithMessage(err, "failed to get deployments")
	}

	if len(deployments.Items) == 0 {
		return api.ArangoDeployment{}, errors.WithMessage(err, "there are no deployments")
	}

	if len(deplName) > 0 {
		// The specific deployment is requested.
		for _, d := range deployments.Items {
			if d.GetName() == deplName {
				return d, nil
			}
		}

		return api.ArangoDeployment{}, errors.New(
			fmt.Sprintf("the deployment \"%s\" does not exist in the namespace \"%s\"", deplName, namespace))
	}

	if len(deployments.Items) == 1 {
		// The specific deployment is not requested and the only one deployment exist in the namespace.
		return deployments.Items[0], nil
	}

	message := fmt.Sprintf("more than one deployment exist in the namespace \"%s\":", namespace)
	for _, item := range deployments.Items {
		message += fmt.Sprintf(" %s", item.GetName())
	}

	return api.ArangoDeployment{}, errors.New(message)
}

// getInterruptionContext returns context which will be cancelled when the process is interrupted.
func getInterruptionContext() context.Context {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// Block until SIGTERM or SIGINT occurs.
		<-c
		cancel()
	}()

	return ctx
}
