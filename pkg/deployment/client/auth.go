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

package client

import (
	"crypto/tls"
	"net"
	nhttp "net/http"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/arangodb/go-driver/jwt"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/conn"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type DeploymentGetter func() *api.ArangoDeployment

func NewAuth(client kubernetes.Interface, g DeploymentGetter) conn.Auth {
	return func() (driver.Authentication, error) {
		d := g()

		if !d.Spec.Authentication.IsAuthenticated() {
			return nil, nil
		}

		secrets := client.CoreV1().Secrets(d.GetNamespace())

		var secret string
		if i := d.Status.CurrentImage; i == nil || !features.JWTRotation().Supported(i.ArangoDBVersion, i.Enterprise) {
			s, err := secrets.Get(d.Spec.Authentication.GetJWTSecretName(), meta.GetOptions{})
			if err != nil {
				return nil, errors.Newf("JWT Secret is missing")
			}

			jwt, ok := s.Data[constants.SecretKeyToken]
			if !ok {
				return nil, errors.Newf("JWT Secret is invalid")
			}

			secret = string(jwt)
		} else {
			s, err := secrets.Get(pod.JWTSecretFolder(d.GetName()), meta.GetOptions{})
			if err != nil {
				return nil, errors.Newf("JWT Folder Secret is missing")
			}

			if len(s.Data) == 0 {
				return nil, errors.Newf("JWT Folder Secret is empty")
			}

			if q, ok := s.Data[pod.ActiveJWTKey]; ok {
				secret = string(q)
			} else {
				for _, q := range s.Data {
					secret = string(q)
					break
				}
			}
		}

		jwt, err := jwt.CreateArangodJwtAuthorizationHeader(secret, "kube-arangodb")
		if err != nil {
			return nil, errors.WithStack(err)
		}

		return driver.RawAuthentication(jwt), nil
	}
}

func NewConfig(g DeploymentGetter) conn.Config {
	return func() (http.ConnectionConfig, error) {
		d := g()

		transport := &nhttp.Transport{
			Proxy: nhttp.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 100 * time.Millisecond,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       100 * time.Millisecond,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}

		if d.Spec.TLS.IsSecure() {
			transport.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}

		connConfig := http.ConnectionConfig{
			Transport:          transport,
			DontFollowRedirect: true,
		}

		return connConfig, nil
	}
}
