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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/jwt"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
)

const (
	ProbePort util.EnvironmentVariable = "ARANGODB_SERVER_PORT"
)

var (
	cmdLifecycleProbe = &cobra.Command{
		Use: "probe",
		Run: cmdLifecycleProbeRun,
	}
	cmdLifecycleProbeLiveness = &cobra.Command{
		Use: "liveness",
		Run: cmdLifecycleProbeRun,
	}
	cmdLifecycleProbeReadiness = &cobra.Command{
		Use: "readiness",
		Run: cmdLifecycleProbeRun,
	}
	cmdLifecycleProbeStartUp = &cobra.Command{
		Use: "startup",
		Run: cmdLifecycleProbeRun,
	}

	probeInput struct {
		Endpoint        string
		JWTPath         string
		ArangoDBVersion string
		ServerGroup     string
		DeploymentMode  string
		SSL             bool
		Auth            bool
		Enterprise      bool
	}
)

func init() {
	f := cmdLifecycleProbe.PersistentFlags()

	cmdLifecycleProbe.AddCommand(cmdLifecycleProbeLiveness)
	cmdLifecycleProbe.AddCommand(cmdLifecycleProbeReadiness)
	cmdLifecycleProbe.AddCommand(cmdLifecycleProbeStartUp)

	f.BoolVarP(&probeInput.SSL, "ssl", "", false, "Determines if SSL is enabled")
	f.BoolVarP(&probeInput.Auth, "auth", "", false, "Determines if authentication is enabled")
	f.StringVarP(&probeInput.Endpoint, "endpoint", "", client.ServerApiVersionEndpoint, "Endpoint (path) to call for lifecycle probe")
	f.MarkDeprecated("endpoint", "Endpoint is chosen automatically by the lifecycle process")
	f.StringVarP(&probeInput.JWTPath, "jwt", "", shared.ClusterJWTSecretVolumeMountDir, "Path to the JWT tokens")
	f.StringVar(&probeInput.ArangoDBVersion, "arangodb-version", os.Getenv(resources.ArangoDBOverrideVersionEnv),
		"Version of the ArangoDB")
	f.StringVar(&probeInput.ServerGroup, "serverGroup", os.Getenv(resources.ArangoDBOverrideServerGroupEnv),
		"Name of the group where a server belongs to")
	f.StringVar(&probeInput.DeploymentMode, "deploymentMode", os.Getenv(resources.ArangoDBOverrideDeploymentModeEnv),
		"A deployment mode (Cluster, Single, ActiveFailover)")
	enterprise, _ := strconv.ParseBool(os.Getenv(resources.ArangoDBOverrideEnterpriseEnv))
	f.BoolVar(&probeInput.Enterprise, "enterprise", enterprise, "Determines if ArangoDB is enterprise")
}

func probeClient() *http.Client {
	tr := &http.Transport{}

	if probeInput.SSL {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{
		Transport: tr,
	}

	return client
}

func probeEndpoint(endpoint string) string {
	proto := "http"
	if probeInput.SSL {
		proto = "https"
	}

	port := ProbePort.GetOrDefault(fmt.Sprintf("%d", shared.ArangoPort))

	return fmt.Sprintf("%s://%s:%s%s", proto, "127.0.0.1", port, endpoint)
}

func readJWTFile(file string) ([]byte, error) {
	p := path.Join(probeInput.JWTPath, file)
	log.Info().Str("path", p).Msgf("Try to use file")

	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func getJWTToken() ([]byte, error) {
	// Try read default one
	if token, err := readJWTFile(constants.SecretKeyToken); err == nil {
		log.Info().Str("token", constants.SecretKeyToken).Msgf("Using JWT Token")
		return token, nil
	}

	// Try read active one
	if token, err := readJWTFile(pod.ActiveJWTKey); err == nil {
		log.Info().Str("token", pod.ActiveJWTKey).Msgf("Using JWT Token")
		return token, nil
	}

	if files, err := os.ReadDir(probeInput.JWTPath); err == nil {
		for _, file := range files {
			if token, err := readJWTFile(file.Name()); err == nil {
				log.Info().Str("token", file.Name()).Msgf("Using JWT Token")
				return token, nil
			}
		}
	}

	return nil, errors.Errorf("Unable to find any token")
}

func addAuthHeader(req *http.Request) error {
	if !probeInput.Auth {
		return nil
	}

	token, err := getJWTToken()
	if err != nil {
		return err
	}

	header, err := jwt.CreateArangodJwtAuthorizationHeader(string(token), "probe")
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", header)
	return nil
}

func doRequest(endpoint string) (*http.Response, error) {
	client := probeClient()

	req, err := http.NewRequest(http.MethodGet, probeEndpoint(endpoint), nil)
	if err != nil {
		return nil, err
	}

	if err := addAuthHeader(req); err != nil {
		return nil, err
	}

	return client.Do(req)
}

func cmdLifecycleProbeRun(cmd *cobra.Command, _ []string) {
	if err := cmdLifecycleProbeRunE(cmd); err != nil {
		log.Error().Err(err).Msgf("Fatal")
		os.Exit(1)
	}
}

func cmdLifecycleProbeRunE(cmd *cobra.Command) error {
	endpoint := getEndpoint(api.ProbeType(cmd.Use))
	resp, err := doRequest(endpoint)
	if err != nil {
		return err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		if resp.Body != nil {
			if data, err := io.ReadAll(resp.Body); err == nil {
				return errors.Errorf("Unexpected code: %d - %s", resp.StatusCode, string(data))
			}
		}

		return errors.Errorf("Unexpected code: %d", resp.StatusCode)
	}

	if endpoint == client.ServerStatusEndpoint {
		// When server status endpoint is used then HTTP status code 200 is not enough.
		// The progress should be also checked.
		if resp.Body == nil {
			return errors.Errorf("Expected body from the \"%s\" endpoint", endpoint)
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Errorf("Failed to read body from the \"%s\" endpoint", endpoint)
		}

		status := client.ServerStatus{}
		if err = json.Unmarshal(data, &status); err != nil {
			return errors.Errorf("Failed to unmarshal %s into server status", string(data))
		}

		if progress, ok := status.GetProgress(); !ok {
			return errors.Errorf("server not ready: %s", progress)
		}
	}

	log.Info().Msgf("Check passed")

	return nil
}

// getEndpoint returns endpoint to the ArangoDB instance where readiness should be checked.
func getEndpoint(probeType api.ProbeType) string {
	if probeType == api.ProbeTypeReadiness {
		if probeInput.DeploymentMode == string(api.DeploymentModeActiveFailover) {
			v := driver.Version(probeInput.ArangoDBVersion)
			if features.FailoverLeadership().Supported(v, probeInput.Enterprise) {
				return client.ServerApiVersionEndpoint
			}
		}

		return client.ServerAvailabilityEndpoint
	}

	if probeInput.ServerGroup == api.ServerGroupDBServersString {
		v := driver.Version(probeInput.ArangoDBVersion)
		if features.Version310().Supported(v, probeInput.Enterprise) {
			return client.ServerStatusEndpoint
		}
	}

	return client.ServerApiVersionEndpoint
}
