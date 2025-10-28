//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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
	goHttp "net/http"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb-helper/go-certificates"
	"github.com/arangodb/go-driver/v2/connection"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/token"
)

const (
	ArgDeploymentName = "deployment-name"
	ArgMemberName     = "member-name"
	ArgAcceptedCode   = "accepted-code"
	ArgTimeout        = "timeout"
	ArgTimeoutDefault = time.Minute
)

func init() {
	cmdMain.AddCommand(cmdAdmin)
	cmdAdmin.PersistentFlags().DurationP(ArgTimeout, "t", ArgTimeoutDefault,
		"timeout of the request")

	cmdAdmin.AddCommand(cmdAdminAgency)
	cmdAdmin.AddCommand(cmdAdminMember)

	cmdAdminAgency.AddCommand(cmdAdminAgencyDump)
	cmdAdminAgencyDump.Flags().StringP(ArgDeploymentName, "d", "",
		"necessary when more than one deployment exist within on namespace")

	cmdAdminAgency.AddCommand(cmdAdminAgencyState)
	cmdAdminAgencyState.Flags().StringP(ArgDeploymentName, "d", "",
		"necessary when more than one deployment exist within on namespace")

	cmdAdminMember.AddCommand(cmdAdminMemberRequest)
	cmdAdminMemberRequest.AddCommand(cmdAdminMemberRequestGet)
	cmdAdminMemberRequestGet.Flags().StringP(ArgDeploymentName, "d", "",
		"necessary when more than one deployment exist within on namespace")
	cmdAdminMemberRequestGet.Flags().StringP(ArgMemberName, "m", "",
		"name of the member for the dump")
	cmdAdminMemberRequestGet.Flags().IntP(ArgAcceptedCode, "c", 200,
		"accepted command code")
}

var cmdAdmin = &cobra.Command{
	Use:   "admin",
	Short: "Administration operations",
	RunE:  cli.Usage,
}

var cmdAdminMember = &cobra.Command{
	Use:   "member",
	Short: "Member operations",
	RunE:  cli.Usage,
}

var cmdAdminMemberRequest = &cobra.Command{
	Use:   "request",
	Short: "Runs http request over member and returns object",
	RunE:  cli.Usage,
}

var cmdAdminMemberRequestGet = &cobra.Command{
	Use:   "get",
	Short: "GET Request",
	RunE:  cmdGetAdminMemberRequestGetE,
}

var cmdAdminAgency = &cobra.Command{
	Use:   "agency",
	Short: "Agency operations",
	RunE:  cli.Usage,
}

var cmdAdminAgencyDump = &cobra.Command{
	Use:   "dump",
	Short: "Get agency dump",
	Long:  "Prints the agency history on the stdout",
	RunE:  cmdAdminGetAgencyDumpE,
}

var cmdAdminAgencyState = &cobra.Command{
	Use:   "state",
	Short: "Get agency state",
	Long:  "Prints the agency current state on the stdout",
	RunE:  cmdAdminGetAgencyStateE,
}

func extractTimeout(cmd *cobra.Command) (context.Context, context.CancelFunc) {
	if v, err := cmd.PersistentFlags().GetDuration(ArgTimeout); err == nil {
		return context.WithTimeout(cmd.Context(), v)
	}

	return context.WithTimeout(cmd.Context(), ArgTimeoutDefault)
}

func cmdGetAdminMemberRequestGetE(cmd *cobra.Command, args []string) error {
	deploymentName, err := cmd.Flags().GetString(ArgDeploymentName)
	if err != nil {
		return err
	}
	memberName, err := cmd.Flags().GetString(ArgMemberName)
	if err != nil {
		return err
	}
	acceptedCode, err := cmd.Flags().GetInt(ArgAcceptedCode)
	if err != nil {
		return err
	}

	ctx, c := extractTimeout(cmd)
	defer c()

	d, certCA, auth, err := getDeploymentAndCredentials(ctx, deploymentName)
	if err != nil {
		logger.Err(err).Error("failed to create basic data for the connection")
		return err
	}

	m, g, ok := d.Status.Members.ElementByID(memberName)
	if !ok {
		err := errors.Errorf("Unable to find member with id %s", memberName)
		logger.Err(err).Error("Unable to find member")
		return err
	}

	dnsName := k8sutil.CreatePodDNSName(d.GetObjectMeta(), g.AsRole(), m.ID)
	endpoint := getArangoEndpoint(d.GetAcceptedSpec().IsSecure(), dnsName)
	conn := createClient([]string{endpoint}, certCA, auth, connection.ApplicationJSON)
	body, err := sendStreamRequest(ctx, conn, goHttp.MethodGet, nil, acceptedCode, args...)
	if body != nil {
		defer body.Close()
	}
	if err != nil {
		logger.Err(err).Error("can not get dump")
		return err
	}

	// Print and receive parallely.
	_, err = io.Copy(os.Stdout, body)
	return err
}

func cmdAdminGetAgencyStateE(cmd *cobra.Command, _ []string) error {
	deploymentName, err := cmd.Flags().GetString(ArgDeploymentName)
	if err != nil {
		return err
	}

	ctx, c := extractTimeout(cmd)
	defer c()

	d, certCA, auth, err := getDeploymentAndCredentials(ctx, deploymentName)
	if err != nil {
		logger.Err(err).Error("failed to create basic data for the connection")
		return err
	}

	if d.GetAcceptedSpec().GetMode() != api.DeploymentModeCluster {
		err = errors.Errorf("agency state does not work for the \"%s\" deployment \"%s\"", d.GetAcceptedSpec().GetMode(),
			d.GetName())
		logger.Err(err).Error("Invalid deployment type")
		return err
	}

	dnsName := k8sutil.CreatePodDNSName(d.GetObjectMeta(), api.ServerGroupAgents.AsRole(), d.Status.Members.Agents[0].ID)
	endpoint := getArangoEndpoint(d.GetAcceptedSpec().IsSecure(), dnsName)
	conn := createClient([]string{endpoint}, certCA, auth, connection.ApplicationJSON)
	leaderID, err := getAgencyLeader(ctx, conn)
	if err != nil {
		logger.Err(err).Error("failed to get leader ID")
		return err
	}

	dnsLeaderName := k8sutil.CreatePodDNSName(d.GetObjectMeta(), api.ServerGroupAgents.AsRole(), leaderID)
	leaderEndpoint := getArangoEndpoint(d.GetAcceptedSpec().IsSecure(), dnsLeaderName)
	conn = createClient([]string{leaderEndpoint}, certCA, auth, connection.PlainText)
	body, err := getAgencyState(ctx, conn)
	if body != nil {
		defer body.Close()
	}
	if err != nil {
		logger.Err(err).Error("can not get state of the agency")
		return err
	}

	// Print and receive parallely.
	_, err = io.Copy(os.Stdout, body)
	return err
}

func cmdAdminGetAgencyDumpE(cmd *cobra.Command, _ []string) error {
	deploymentName, err := cmd.Flags().GetString(ArgDeploymentName)
	if err != nil {
		return err
	}

	ctx, c := extractTimeout(cmd)
	defer c()

	d, certCA, auth, err := getDeploymentAndCredentials(ctx, deploymentName)
	if err != nil {
		logger.Err(err).Error("failed to create basic data for the connection")
		return err
	}

	if d.GetAcceptedSpec().GetMode() != api.DeploymentModeCluster {
		err = errors.Errorf("agency state does not work for the \"%s\" deployment \"%s\"", d.GetAcceptedSpec().GetMode(),
			d.GetName())
		logger.Err(err).Error("Invalid deployment type")
		return err
	}

	endpoint := getArangoEndpoint(d.GetAcceptedSpec().IsSecure(), k8sutil.CreateDatabaseClientServiceDNSName(d.GetObjectMeta()))
	conn := createClient([]string{endpoint}, certCA, auth, connection.ApplicationJSON)
	body, err := getAgencyDump(ctx, conn)
	if body != nil {
		defer body.Close()
	}
	if err != nil {
		logger.Err(err).Error("can not get dump")
		return err
	}

	// Print and receive parallely.
	_, err = io.Copy(os.Stdout, body)
	return err
}

// sendStreamRequest sends the request to a member
func sendStreamRequest(ctx context.Context, conn connection.Connection, method string, body []byte, code int, parts ...string) (io.ReadCloser, error) {
	url := connection.NewUrl(parts...)

	var mods []connection.RequestModifier

	if body != nil {
		mods = append(mods, connection.WithBody(body))
	}

	resp, output, err := connection.CallStream(ctx, conn, method, url, mods...)
	if err != nil {
		return nil, err
	}
	if resp.Code() != code {
		return nil, errors.New(fmt.Sprintf("unexpected HTTP status from \"%s\" endpoint. Expected: '%d', got '%d'", url, code, resp.Code()))
	}

	return output, nil
}

// sendRequest sends the request to a member and returns object
func sendRequest[OUT any](ctx context.Context, conn connection.Connection, method string, body []byte, code int, parts ...string) (OUT, error) {
	url := connection.NewUrl(parts...)

	var mods []connection.RequestModifier

	if body != nil {
		mods = append(mods, connection.WithBody(body))
	}

	var out OUT

	resp, err := connection.Call(ctx, conn, method, url, &out, mods...)
	if err != nil {
		return util.Default[OUT](), err
	}
	if resp.Code() != code {
		return util.Default[OUT](), errors.New(fmt.Sprintf("unexpected HTTP status from \"%s\" endpoint. Expected: '%d', got '%d'", url, code, resp.Code()))
	}

	return out, nil
}

// getAgencyState returns the current state in the agency.
func getAgencyState(ctx context.Context, conn connection.Connection) (io.ReadCloser, error) {
	return sendStreamRequest(ctx, conn, goHttp.MethodPost, []byte(`[["/"]]`), goHttp.StatusOK, "_api", "agency", "read")
}

// getDeploymentAndCredentials returns deployment and necessary credentials to communicate with ArangoDB pods.
func getDeploymentAndCredentials(ctx context.Context,
	deploymentName string) (d api.ArangoDeployment, certCA *x509.CertPool, auth connection.Authentication, err error) {

	namespace := os.Getenv(utilConstants.EnvOperatorPodNamespace)
	if len(namespace) == 0 {
		err = errors.New(fmt.Sprintf("\"%s\" environment variable missing", utilConstants.EnvOperatorPodNamespace))
		return
	}

	client, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		err = errors.Errorf("Client not initialised")
		return
	}

	kubeCli := client.Kubernetes()

	d, err = getDeployment(ctx, namespace, deploymentName)
	if err != nil {
		err = errors.WithMessage(err, "failed to get deployment")
		return
	}

	var secrets = kubeCli.CoreV1().Secrets(d.GetNamespace())
	if d.GetAcceptedSpec().TLS.IsSecure() {
		certCA, err = getCACertificate(ctx, secrets, d.GetAcceptedSpec().TLS.GetCASecretName())
		if err != nil {
			err = errors.WithMessage(err, "failed to get CA certificate")
			return
		}
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
	output, err := sendRequest[map[string]interface{}](ctx, conn, goHttp.MethodGet, nil, goHttp.StatusOK, "_api", "agency", "config")
	if err != nil {
		return "", err
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
	return sendStreamRequest(ctx, conn, goHttp.MethodGet, nil, goHttp.StatusOK, "_api", "cluster", "agency-dump")
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
		Endpoint:       connection.NewRoundRobinEndpoints(endpoints),
		Transport: &goHttp.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certCA,
			},
		},
	}

	return connection.NewHttpConnection(conf)
}

// getJWTTokenFromSecrets returns token from the secret.
func getJWTTokenFromSecrets(ctx context.Context, secrets generic.ReadClient[*core.Secret], name string, paths ...string) (connection.Authentication, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()

	secret, err := k8sutil.GetTokenSecret(ctxChild, secrets, name)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("failed to get secret \"%s\"", name))
	}

	claims := token.NewClaims().With(
		token.WithDefaultClaims(),
		token.WithServerID("kube-arangodb"),
	)

	if len(paths) > 0 {
		claims = claims.With(token.WithAllowedPaths(paths...))
	}

	authz, err := claims.Sign(secret)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("failed to create bearer token from secret \"%s\"", name))
	}

	bearerToken := fmt.Sprintf("bearer %s", authz)

	return JWTAuthentication{key: "Authorization", value: bearerToken}, nil
}

// getCACertificate returns CA certificate from the secret.
func getCACertificate(ctx context.Context, secrets generic.ReadClient[*core.Secret], name string) (*x509.CertPool, error) {
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
		return api.ArangoDeployment{}, errors.Errorf("Client not initialised")
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
