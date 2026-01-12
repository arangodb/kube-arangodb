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
	"bytes"
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func UpdateValidatingAdmissionHookCA(ctx context.Context, client kclient.Client, caBundle []byte, names ...string) error {
	for _, n := range names {
		logger := logger.Str("type", "validating").Str("name", n)

		logger.Debug("Reading current webhook")
		v, err := client.Kubernetes().AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(ctx, n, meta.GetOptions{})
		if err != nil {
			logger.Err(err).Warn("Failed to get validating webhook")
			return err
		}

		changed := false

		for id := range v.Webhooks {
			w := v.Webhooks[id].DeepCopy()

			if !bytes.Equal(caBundle, w.ClientConfig.CABundle) {
				w.ClientConfig.CABundle = caBundle

				logger.Str("webhook", w.Name).Debug("Updating ca")

				changed = true

				w.DeepCopyInto(&v.Webhooks[id])
			}
		}

		if !changed {
			continue
		}

		logger.Info("Updating webhook")

		_, err = client.Kubernetes().AdmissionregistrationV1().ValidatingWebhookConfigurations().Update(ctx, v, meta.UpdateOptions{})
		if err != nil {
			logger.Err(err).Warn("Failed to update validating webhook")
			return err
		}
	}
	return nil
}

func UpdateMutatingAdmissionHookCA(ctx context.Context, client kclient.Client, caBundle []byte, names ...string) error {
	for _, n := range names {
		logger := logger.Str("type", "mutating").Str("name", n)

		logger.Debug("Reading current webhook")
		v, err := client.Kubernetes().AdmissionregistrationV1().MutatingWebhookConfigurations().Get(ctx, n, meta.GetOptions{})
		if err != nil {
			logger.Err(err).Warn("Failed to get validating webhook")
			return err
		}

		changed := false

		for id := range v.Webhooks {
			w := v.Webhooks[id].DeepCopy()

			if !bytes.Equal(caBundle, w.ClientConfig.CABundle) {
				w.ClientConfig.CABundle = caBundle

				logger.Str("webhook", w.Name).Debug("Updating ca")

				changed = true

				w.DeepCopyInto(&v.Webhooks[id])
			}
		}

		if !changed {
			continue
		}

		logger.Info("Updating webhook")

		_, err = client.Kubernetes().AdmissionregistrationV1().MutatingWebhookConfigurations().Update(ctx, v, meta.UpdateOptions{})
		if err != nil {
			logger.Err(err).Warn("Failed to update mutating webhook")
			return err
		}
	}

	return nil
}
