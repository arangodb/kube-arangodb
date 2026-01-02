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

package webhook

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	goHttp "net/http"
	"testing"

	"github.com/stretchr/testify/require"
	admission "k8s.io/api/admission/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/http"
	ktls "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tls"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	addr := startHTTPServer(t, ctx, newPodAdmission("test", newPodHandler(func(ctx context.Context, log logging.Logger, t AdmissionRequestType, request *admission.AdmissionRequest, old, new *core.Pod) bool {
		return true
	}, nil, func(ctx context.Context, log logging.Logger, at AdmissionRequestType, request *admission.AdmissionRequest, old, new *core.Pod) (ValidationResponse, error) {
		require.Nil(t, old)
		require.NotNil(t, new)

		require.EqualValues(t, AdmissionRequestValidate, at)

		return ValidationResponse{
			Allowed: true,
		}, nil
	})))

	resp := requestPod(t, addr, "test", AdmissionRequestValidate,
		newPodAdmissionRequest(t, "test", tests.FakeNamespace, admission.Create, nil, &core.Pod{
			ObjectMeta: meta.ObjectMeta{
				Name:      "test",
				Namespace: tests.FakeNamespace,
			},
		}),
	)
	require.NotNil(t, resp)
	require.True(t, resp.Allowed)
}

func requestPod(t *testing.T, addr string, name string, mode AdmissionRequestType, req *admission.AdmissionRequest) *admission.AdmissionResponse {
	return request(t, addr, meta.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}, name, mode, req)
}

func request(t *testing.T, addr string, gvs meta.GroupVersionResource, name string, mode AdmissionRequestType, request *admission.AdmissionRequest) *admission.AdmissionResponse {
	if request == nil {
		request = &admission.AdmissionRequest{}
	}

	request.UID = uuid.NewUUID()

	data, err := json.Marshal(admission.AdmissionReview{
		Request: request,
	})
	require.NoError(t, err)

	var c = goHttp.Client{
		Transport: &goHttp.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	url := fmt.Sprintf("https://%s/webhook/%s/%s/%s", addr, gvsAsPath(gvs), name, util.BoolSwitch(mode == AdmissionRequestValidate, "validate", "mutate"))

	t.Logf("Request send to: %s", url)

	resp, err := c.Post(url, "application/json", bytes.NewReader(data))
	require.NoError(t, err)

	require.EqualValues(t, 200, resp.StatusCode)

	var rv admission.AdmissionReview

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&rv))
	require.NoError(t, resp.Body.Close())

	require.NotNil(t, rv.Response)
	require.EqualValues(t, request.UID, rv.Response.UID)

	return rv.Response
}

func startHTTPServer(t *testing.T, ctx context.Context, admissions ...Admission) string {
	return tests.NewHTTPServer(ctx, t,
		http.DefaultHTTPServerSettings,
		ktls.WithTLSConfigFetcher(ktls.NewSelfSignedTLSConfig("localhost", "127.0.0.1")),
		http.WithServeMux(
			Admissions(admissions).Register(),
		),
	)
}
