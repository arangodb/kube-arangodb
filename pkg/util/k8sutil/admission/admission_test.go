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

package admission

import (
	"testing"

	"github.com/stretchr/testify/require"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	ktls "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tls"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_Webhook(t *testing.T) {
	client := kclient.NewFakeClient()

	_, caData, err := ktls.GetOrCreateTLSCAConfig(t.Context(), client.Kubernetes().CoreV1().Secrets(tests.FakeNamespace), "secret")
	require.NoError(t, err)

	validation, err := client.Kubernetes().AdmissionregistrationV1().ValidatingWebhookConfigurations().Create(t.Context(), &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: meta.ObjectMeta{
			Name: "valid",
		},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{
			{
				Name: "valid",
			},
		},
	}, meta.CreateOptions{})
	require.NoError(t, err)

	require.NoError(t, UpdateValidatingAdmissionHookCA(t.Context(), client, caData, validation.GetName()))

	validation, err = client.Kubernetes().AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(t.Context(), validation.GetName(), meta.GetOptions{})
	require.NoError(t, err)

	require.Equal(t, "valid", validation.Webhooks[0].Name)
	require.NotEmpty(t, validation.Webhooks[0].ClientConfig.CABundle)

	require.Equal(t, caData, validation.Webhooks[0].ClientConfig.CABundle)

	_, caDataNew, err := ktls.GetOrCreateTLSCAConfig(t.Context(), client.Kubernetes().CoreV1().Secrets(tests.FakeNamespace), "secret")
	require.NoError(t, err)

	require.NoError(t, UpdateValidatingAdmissionHookCA(t.Context(), client, caDataNew, validation.GetName()))

	validation, err = client.Kubernetes().AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(t.Context(), validation.GetName(), meta.GetOptions{})
	require.NoError(t, err)

	require.Equal(t, "valid", validation.Webhooks[0].Name)
	require.NotEmpty(t, validation.Webhooks[0].ClientConfig.CABundle)

	require.Equal(t, caData, validation.Webhooks[0].ClientConfig.CABundle)
	require.Equal(t, caDataNew, validation.Webhooks[0].ClientConfig.CABundle)
}
