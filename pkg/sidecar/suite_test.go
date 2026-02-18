//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package sidecar

import (
	tls2 "crypto/tls"
	"fmt"
	goHttp "net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	pbPongV1 "github.com/arangodb/kube-arangodb/integrations/pong/v1/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/http"
	ktls "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tls"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authentication"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
)

type runner interface {
	addArgs(args ...string) runner
	evaluateArgs(in func(t *testing.T) []string) runner

	run(name string, in func(t *testing.T)) runner
}

type runnerImpl struct {
	t    *testing.T
	args []string
}

func (r runnerImpl) evaluateArgs(in func(t *testing.T) []string) runner {
	args := in(r.t)

	return r.addArgs(args...)
}

func (r runnerImpl) addArgs(args ...string) runner {
	q := r
	q.args = append(q.args, args...)
	return q
}

func (r runnerImpl) run(name string, in func(t *testing.T)) runner {
	r.t.Run(name, func(t *testing.T) {
		cmd, err := Register()
		require.NoError(t, err)

		tests.RunWithCLI(t, cmd, r.args...)(func(t *testing.T) {
			tests.WaitForAddress(t, "127.0.0.1", 8109)
			tests.WaitForAddress(t, "127.0.0.1", 8108)

			in(t)

			t.Logf("Test completed")
		})
	})

	return r
}

func runSidecar(t *testing.T) runner {
	return runnerImpl{
		t: t,
	}
}

func executeHttpRequest(t *testing.T, client http.HTTPClient, req *goHttp.Request) (*goHttp.Response, error) {
	resp, err := client.Do(req)
	if resp != nil {
		require.NoError(t, resp.Body.Close())
	}
	return resp, err
}

func renderTLSKeyfileCertificate(t *testing.T) []string {
	dir := t.TempDir()

	caCert, caPriv, err := ktls.CreateTLSCACertificate("testing")
	require.NoError(t, err)

	cert, priv, err := ktls.CreateTLSServerCertificate(caCert, caPriv, ktls.KeyfileInput{
		AltNames: []string{
			"localhost",
			"127.0.0.1",
		},
	})
	require.NoError(t, err)

	keyFile := ktls.AsKeyfile(cert, priv)

	p := path.Join(dir, "tls.keyfile")

	require.NoError(t, os.WriteFile(p, []byte(keyFile), 0644))
	return []string{
		"--sidecar.keyfile",
		p,
	}
}

func Test_Protocol(t *testing.T) {
	t.Run("HTTP", func(t *testing.T) {
		runSidecar(t).
			run("HTTP Client", func(t *testing.T) {
				req, err := goHttp.NewRequest(goHttp.MethodGet, "http://127.0.0.1:8108/unknown", nil)
				require.NoError(t, err)

				c := http.NewHTTPClient()

				resp, err := executeHttpRequest(t, c, req)

				// Expect Response
				require.NoError(t, err)
				require.Equal(t, goHttp.StatusNotFound, resp.StatusCode)
			}).
			run("HTTPS Client", func(t *testing.T) {
				req, err := goHttp.NewRequest(goHttp.MethodGet, "https://127.0.0.1:8108/unknown", nil)
				require.NoError(t, err)

				c := http.NewHTTPClient()

				_, err = executeHttpRequest(t, c, req)
				// Expect Response
				require.Error(t, err)
			})
	})
	t.Run("HTTPS", func(t *testing.T) {
		runSidecar(t).
			evaluateArgs(renderTLSKeyfileCertificate).
			run("HTTP Client", func(t *testing.T) {
				req, err := goHttp.NewRequest(goHttp.MethodGet, "http://127.0.0.1:8108/unknown", nil)
				require.NoError(t, err)

				c := http.NewHTTPClient()

				resp, err := executeHttpRequest(t, c, req)
				// Expect Response
				require.NoError(t, err)
				require.Equal(t, goHttp.StatusBadRequest, resp.StatusCode)
			}).
			run("HTTPS Client", func(t *testing.T) {
				req, err := goHttp.NewRequest(goHttp.MethodGet, "https://127.0.0.1:8108/unknown", nil)
				require.NoError(t, err)

				c := http.NewHTTPClient(http.WithTransport(http.WithTransportTLS(http.Insecure)))

				resp, err := executeHttpRequest(t, c, req)
				// Expect Response
				require.NoError(t, err)
				require.Equal(t, goHttp.StatusNotFound, resp.StatusCode)
			})
	})

	t.Run("GRPC", func(t *testing.T) {
		runSidecar(t).
			run("Plain", func(t *testing.T) {
				c, err := grpc.NewClient("127.0.0.1:8109", grpc.WithTransportCredentials(insecure.NewCredentials()))
				require.NoError(t, err)
				defer c.Close()

				pong := pbPongV1.NewPongV1Client(c)

				tgrpc.NewExecutor(t, pong.Ping, &pbSharedV1.Empty{}).Get(t)
			}).
			run("Secure", func(t *testing.T) {
				c, err := grpc.NewClient("127.0.0.1:8109", grpc.WithTransportCredentials(credentials.NewTLS(&tls2.Config{InsecureSkipVerify: true})))
				require.NoError(t, err)
				defer c.Close()

				pong := pbPongV1.NewPongV1Client(c)

				tgrpc.NewExecutor(t, pong.Ping, &pbSharedV1.Empty{}).Code(t, codes.Unavailable)
			})
	})

	t.Run("TLS GRPC", func(t *testing.T) {
		runSidecar(t).
			evaluateArgs(renderTLSKeyfileCertificate).
			run("Plain", func(t *testing.T) {
				c, err := grpc.NewClient("127.0.0.1:8109", grpc.WithTransportCredentials(insecure.NewCredentials()))
				require.NoError(t, err)
				defer c.Close()

				pong := pbPongV1.NewPongV1Client(c)

				tgrpc.NewExecutor(t, pong.Ping, &pbSharedV1.Empty{}).Code(t, codes.Unavailable)
			}).
			run("Secure", func(t *testing.T) {
				c, err := grpc.NewClient("127.0.0.1:8109", grpc.WithTransportCredentials(credentials.NewTLS(&tls2.Config{InsecureSkipVerify: true})))
				require.NoError(t, err)
				defer c.Close()

				pong := pbPongV1.NewPongV1Client(c)

				tgrpc.NewExecutor(t, pong.Ping, &pbSharedV1.Empty{}).Get(t)
			})
	})
}

func Test_Auth_JWT(t *testing.T) {
	tm := tests.NewTokenManager(t)
	tmTemp := tests.NewTokenManager(t)

	token1, token2 := tests.GenerateJWTToken(), tests.GenerateJWTToken()

	tm.Set(t, token1)
	tmTemp.Set(t, token2)

	t.Run("NoAuth", func(t *testing.T) {
		runSidecar(t).
			run("HTTP Client", func(t *testing.T) {
				req, err := goHttp.NewRequest(goHttp.MethodGet, "http://127.0.0.1:8108/unknown", nil)
				require.NoError(t, err)

				c := http.NewHTTPClient()

				resp, err := executeHttpRequest(t, c, req)

				// Expect Response
				require.NoError(t, err)
				require.Equal(t, goHttp.StatusNotFound, resp.StatusCode)
			}).
			run("HTTP Client With auth 1", func(t *testing.T) {
				auth := tm.Sign(t, utilToken.WithRelativeDuration(time.Minute))

				req, err := goHttp.NewRequest(goHttp.MethodGet, "http://127.0.0.1:8108/unknown", nil)
				require.NoError(t, err)

				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth))

				c := http.NewHTTPClient()

				resp, err := executeHttpRequest(t, c, req)

				// Expect Response
				require.NoError(t, err)
				require.Equal(t, goHttp.StatusNotFound, resp.StatusCode)
			}).
			run("HTTP Client With auth 2", func(t *testing.T) {
				auth := tmTemp.Sign(t, utilToken.WithRelativeDuration(time.Minute))

				req, err := goHttp.NewRequest(goHttp.MethodGet, "http://127.0.0.1:8108/unknown", nil)
				require.NoError(t, err)

				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth))

				c := http.NewHTTPClient()

				resp, err := executeHttpRequest(t, c, req)

				// Expect Response
				require.NoError(t, err)
				require.Equal(t, goHttp.StatusNotFound, resp.StatusCode)
			}).
			run("GRPC Plain", func(t *testing.T) {
				c, err := grpc.NewClient("127.0.0.1:8109", grpc.WithTransportCredentials(insecure.NewCredentials()))
				require.NoError(t, err)
				defer c.Close()

				pong := pbPongV1.NewPongV1Client(c)

				tgrpc.NewExecutor(t, pong.Ping, &pbSharedV1.Empty{}).Get(t)
			}).
			run("GRPC Auth 1", func(t *testing.T) {
				auth := tm.Sign(t, utilToken.WithRelativeDuration(time.Minute))

				var opts []grpc.DialOption
				opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
				opts = append(opts, authentication.NewInterceptorClientOptions(authentication.Static(auth))...)
				c, err := grpc.NewClient("127.0.0.1:8109", opts...)
				require.NoError(t, err)
				defer c.Close()

				pong := pbPongV1.NewPongV1Client(c)

				tgrpc.NewExecutor(t, pong.Ping, &pbSharedV1.Empty{}).Get(t)
			}).
			run("GRPC Auth 2", func(t *testing.T) {
				auth := tmTemp.Sign(t, utilToken.WithRelativeDuration(time.Minute))

				var opts []grpc.DialOption
				opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
				opts = append(opts, authentication.NewInterceptorClientOptions(authentication.Static(auth))...)
				c, err := grpc.NewClient("127.0.0.1:8109", opts...)
				require.NoError(t, err)
				defer c.Close()

				pong := pbPongV1.NewPongV1Client(c)

				tgrpc.NewExecutor(t, pong.Ping, &pbSharedV1.Empty{}).Get(t)
			})
	})

	t.Run("Auth", func(t *testing.T) {
		runSidecar(t).
			addArgs("--sidecar.auth", tm.Path()).
			run("HTTP Client", func(t *testing.T) {
				req, err := goHttp.NewRequest(goHttp.MethodGet, "http://127.0.0.1:8108/unknown", nil)
				require.NoError(t, err)

				c := http.NewHTTPClient()

				resp, err := executeHttpRequest(t, c, req)

				// Expect Response
				require.NoError(t, err)
				require.Equal(t, goHttp.StatusNotFound, resp.StatusCode)
			}).
			run("HTTP Client With auth 1", func(t *testing.T) {
				auth := tm.Sign(t, utilToken.WithRelativeDuration(time.Minute))

				req, err := goHttp.NewRequest(goHttp.MethodGet, "http://127.0.0.1:8108/unknown", nil)
				require.NoError(t, err)

				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth))

				c := http.NewHTTPClient()

				resp, err := executeHttpRequest(t, c, req)

				// Expect Response
				require.NoError(t, err)
				defer resp.Body.Close()
				require.Equal(t, goHttp.StatusNotFound, resp.StatusCode)
			}).
			run("HTTP Client With auth 2", func(t *testing.T) {
				auth := tmTemp.Sign(t, utilToken.WithRelativeDuration(time.Minute))

				req, err := goHttp.NewRequest(goHttp.MethodGet, "http://127.0.0.1:8108/unknown", nil)
				require.NoError(t, err)

				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth))

				c := http.NewHTTPClient()

				resp, err := executeHttpRequest(t, c, req)

				// Expect Response
				require.NoError(t, err)
				require.Equal(t, goHttp.StatusNotFound, resp.StatusCode)
			}).
			run("GRPC Plain", func(t *testing.T) {
				c, err := grpc.NewClient("127.0.0.1:8109", grpc.WithTransportCredentials(insecure.NewCredentials()))
				require.NoError(t, err)
				defer c.Close()

				pong := pbPongV1.NewPongV1Client(c)

				tgrpc.NewExecutor(t, pong.Ping, &pbSharedV1.Empty{}).Code(t, codes.Unauthenticated)
			}).
			run("GRPC Auth 1", func(t *testing.T) {
				auth := tm.Sign(t, utilToken.WithRelativeDuration(time.Minute))

				var opts []grpc.DialOption
				opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
				opts = append(opts, authentication.NewInterceptorClientOptions(authentication.Static(auth))...)
				c, err := grpc.NewClient("127.0.0.1:8109", opts...)
				require.NoError(t, err)
				defer c.Close()

				pong := pbPongV1.NewPongV1Client(c)

				tgrpc.NewExecutor(t, pong.Ping, &pbSharedV1.Empty{}).Get(t)
			}).
			run("GRPC Auth 2", func(t *testing.T) {
				auth := tmTemp.Sign(t, utilToken.WithRelativeDuration(time.Minute))

				var opts []grpc.DialOption
				opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
				opts = append(opts, authentication.NewInterceptorClientOptions(authentication.Static(auth))...)
				c, err := grpc.NewClient("127.0.0.1:8109", opts...)
				require.NoError(t, err)
				defer c.Close()

				pong := pbPongV1.NewPongV1Client(c)

				tgrpc.NewExecutor(t, pong.Ping, &pbSharedV1.Empty{}).Code(t, codes.Unauthenticated)
			}).
			run("GRPC Auth with timeout", func(t *testing.T) {
				auth := tm.Sign(t, utilToken.WithRelativeDuration(2*time.Second))

				var opts []grpc.DialOption
				opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
				opts = append(opts, authentication.NewInterceptorClientOptions(authentication.Static(auth))...)
				c, err := grpc.NewClient("127.0.0.1:8109", opts...)
				require.NoError(t, err)
				defer c.Close()

				pong := pbPongV1.NewPongV1Client(c)

				tgrpc.NewExecutor(t, pong.Ping, &pbSharedV1.Empty{}).Get(t)

				time.Sleep(3 * time.Second)

				tgrpc.NewExecutor(t, pong.Ping, &pbSharedV1.Empty{}).Code(t, codes.Unauthenticated)
			})
	})
}

func Test_Auth_ECDSA(t *testing.T) {
	tm := tests.NewTokenManager(t)
	tmTemp := tests.NewTokenManager(t)

	token1, token2 := tests.GenerateECDSAP256Token(t), tests.GenerateECDSAP256Token(t)

	tm.Set(t, token1)
	tmTemp.Set(t, token2)

	t.Run("NoAuth", func(t *testing.T) {
		runSidecar(t).
			run("HTTP Client", func(t *testing.T) {
				req, err := goHttp.NewRequest(goHttp.MethodGet, "http://127.0.0.1:8108/unknown", nil)
				require.NoError(t, err)

				c := http.NewHTTPClient()

				resp, err := executeHttpRequest(t, c, req)

				// Expect Response
				require.NoError(t, err)
				require.Equal(t, goHttp.StatusNotFound, resp.StatusCode)
			}).
			run("HTTP Client With auth 1", func(t *testing.T) {
				auth := tm.Sign(t, utilToken.WithRelativeDuration(time.Minute))

				req, err := goHttp.NewRequest(goHttp.MethodGet, "http://127.0.0.1:8108/unknown", nil)
				require.NoError(t, err)

				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth))

				c := http.NewHTTPClient()

				resp, err := executeHttpRequest(t, c, req)

				// Expect Response
				require.NoError(t, err)
				require.Equal(t, goHttp.StatusNotFound, resp.StatusCode)
			}).
			run("HTTP Client With auth 2", func(t *testing.T) {
				auth := tmTemp.Sign(t, utilToken.WithRelativeDuration(time.Minute))

				req, err := goHttp.NewRequest(goHttp.MethodGet, "http://127.0.0.1:8108/unknown", nil)
				require.NoError(t, err)

				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth))

				c := http.NewHTTPClient()

				resp, err := executeHttpRequest(t, c, req)

				// Expect Response
				require.NoError(t, err)
				require.Equal(t, goHttp.StatusNotFound, resp.StatusCode)
			}).
			run("GRPC Plain", func(t *testing.T) {
				c, err := grpc.NewClient("127.0.0.1:8109", grpc.WithTransportCredentials(insecure.NewCredentials()))
				require.NoError(t, err)
				defer c.Close()

				pong := pbPongV1.NewPongV1Client(c)

				tgrpc.NewExecutor(t, pong.Ping, &pbSharedV1.Empty{}).Get(t)
			}).
			run("GRPC Auth 1", func(t *testing.T) {
				auth := tm.Sign(t, utilToken.WithRelativeDuration(time.Minute))

				var opts []grpc.DialOption
				opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
				opts = append(opts, authentication.NewInterceptorClientOptions(authentication.Static(auth))...)
				c, err := grpc.NewClient("127.0.0.1:8109", opts...)
				require.NoError(t, err)
				defer c.Close()

				pong := pbPongV1.NewPongV1Client(c)

				tgrpc.NewExecutor(t, pong.Ping, &pbSharedV1.Empty{}).Get(t)
			}).
			run("GRPC Auth 2", func(t *testing.T) {
				auth := tmTemp.Sign(t, utilToken.WithRelativeDuration(time.Minute))

				var opts []grpc.DialOption
				opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
				opts = append(opts, authentication.NewInterceptorClientOptions(authentication.Static(auth))...)
				c, err := grpc.NewClient("127.0.0.1:8109", opts...)
				require.NoError(t, err)
				defer c.Close()

				pong := pbPongV1.NewPongV1Client(c)

				tgrpc.NewExecutor(t, pong.Ping, &pbSharedV1.Empty{}).Get(t)
			})
	})

	t.Run("Auth", func(t *testing.T) {
		runSidecar(t).
			addArgs("--sidecar.auth", tm.Path()).
			run("HTTP Client", func(t *testing.T) {
				req, err := goHttp.NewRequest(goHttp.MethodGet, "http://127.0.0.1:8108/unknown", nil)
				require.NoError(t, err)

				c := http.NewHTTPClient()

				resp, err := executeHttpRequest(t, c, req)

				// Expect Response
				require.NoError(t, err)
				require.Equal(t, goHttp.StatusNotFound, resp.StatusCode)
			}).
			run("HTTP Client With auth 1", func(t *testing.T) {
				auth := tm.Sign(t, utilToken.WithRelativeDuration(time.Minute))

				req, err := goHttp.NewRequest(goHttp.MethodGet, "http://127.0.0.1:8108/unknown", nil)
				require.NoError(t, err)

				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth))

				c := http.NewHTTPClient()

				resp, err := executeHttpRequest(t, c, req)

				// Expect Response
				require.NoError(t, err)
				defer resp.Body.Close()
				require.Equal(t, goHttp.StatusNotFound, resp.StatusCode)
			}).
			run("HTTP Client With auth 2", func(t *testing.T) {
				auth := tmTemp.Sign(t, utilToken.WithRelativeDuration(time.Minute))

				req, err := goHttp.NewRequest(goHttp.MethodGet, "http://127.0.0.1:8108/unknown", nil)
				require.NoError(t, err)

				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth))

				c := http.NewHTTPClient()

				resp, err := executeHttpRequest(t, c, req)

				// Expect Response
				require.NoError(t, err)
				require.Equal(t, goHttp.StatusNotFound, resp.StatusCode)
			}).
			run("GRPC Plain", func(t *testing.T) {
				c, err := grpc.NewClient("127.0.0.1:8109", grpc.WithTransportCredentials(insecure.NewCredentials()))
				require.NoError(t, err)
				defer c.Close()

				pong := pbPongV1.NewPongV1Client(c)

				tgrpc.NewExecutor(t, pong.Ping, &pbSharedV1.Empty{}).Code(t, codes.Unauthenticated)
			}).
			run("GRPC Auth 1", func(t *testing.T) {
				auth := tm.Sign(t, utilToken.WithRelativeDuration(time.Minute))

				var opts []grpc.DialOption
				opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
				opts = append(opts, authentication.NewInterceptorClientOptions(authentication.Static(auth))...)
				c, err := grpc.NewClient("127.0.0.1:8109", opts...)
				require.NoError(t, err)
				defer c.Close()

				pong := pbPongV1.NewPongV1Client(c)

				tgrpc.NewExecutor(t, pong.Ping, &pbSharedV1.Empty{}).Get(t)
			}).
			run("GRPC Auth 2", func(t *testing.T) {
				auth := tmTemp.Sign(t, utilToken.WithRelativeDuration(time.Minute))

				var opts []grpc.DialOption
				opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
				opts = append(opts, authentication.NewInterceptorClientOptions(authentication.Static(auth))...)
				c, err := grpc.NewClient("127.0.0.1:8109", opts...)
				require.NoError(t, err)
				defer c.Close()

				pong := pbPongV1.NewPongV1Client(c)

				tgrpc.NewExecutor(t, pong.Ping, &pbSharedV1.Empty{}).Code(t, codes.Unauthenticated)
			}).
			run("GRPC Auth with timeout", func(t *testing.T) {
				auth := tm.Sign(t, utilToken.WithRelativeDuration(2*time.Second))

				var opts []grpc.DialOption
				opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
				opts = append(opts, authentication.NewInterceptorClientOptions(authentication.Static(auth))...)
				c, err := grpc.NewClient("127.0.0.1:8109", opts...)
				require.NoError(t, err)
				defer c.Close()

				pong := pbPongV1.NewPongV1Client(c)

				tgrpc.NewExecutor(t, pong.Ping, &pbSharedV1.Empty{}).Get(t)

				time.Sleep(3 * time.Second)

				tgrpc.NewExecutor(t, pong.Ping, &pbSharedV1.Empty{}).Code(t, codes.Unauthenticated)
			})
	})
}
