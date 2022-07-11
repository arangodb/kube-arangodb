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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/arangodb/go-driver/jwt"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
)

var (
	cmdLifecycleProbe = &cobra.Command{
		Use: "probe",
		Run: cmdLifecycleProbeCheck,
	}

	probeInput struct {
		SSL      bool
		Auth     bool
		Endpoint string
		JWTPath  string
	}
)

func init() {
	f := cmdLifecycleProbe.PersistentFlags()

	f.BoolVarP(&probeInput.SSL, "ssl", "", false, "Determines if SSL is enabled")
	f.BoolVarP(&probeInput.Auth, "auth", "", false, "Determines if authentication is enabled")
	f.StringVarP(&probeInput.Endpoint, "endpoint", "", "/_api/version", "Endpoint (path) to call for lifecycle probe")
	f.StringVarP(&probeInput.JWTPath, "jwt", "", shared.ClusterJWTSecretVolumeMountDir, "Path to the JWT tokens")
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

	return fmt.Sprintf("%s://%s:%d%s", proto, "127.0.0.1", shared.ArangoPort, endpoint)
}

func readJWTFile(file string) ([]byte, error) {
	p := path.Join(probeInput.JWTPath, file)
	log.Info().Str("path", p).Msgf("Try to use file")

	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	data, err := ioutil.ReadAll(f)
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

	if files, err := ioutil.ReadDir(probeInput.JWTPath); err == nil {
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

func doRequest() (*http.Response, error) {
	client := probeClient()

	req, err := http.NewRequest(http.MethodGet, probeEndpoint(probeInput.Endpoint), nil)
	if err != nil {
		return nil, err
	}

	if err := addAuthHeader(req); err != nil {
		return nil, err
	}

	return client.Do(req)
}

func cmdLifecycleProbeCheck(cmd *cobra.Command, args []string) {
	if err := cmdLifecycleProbeCheckE(); err != nil {
		log.Error().Err(err).Msgf("Fatal")
		os.Exit(1)
	}
}

func cmdLifecycleProbeCheckE() error {
	resp, err := doRequest()
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.Body != nil {
			defer resp.Body.Close()
			if data, err := ioutil.ReadAll(resp.Body); err == nil {
				return errors.Errorf("Unexpected code: %d - %s", resp.StatusCode, string(data))
			}
		}

		return errors.Errorf("Unexpected code: %d", resp.StatusCode)
	}

	log.Info().Msgf("Check passed")

	return nil
}
