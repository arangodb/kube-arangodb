//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package scheduler

import (
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/handlers/scheduler/webhooks/policies"
	inspectorConstants "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/webhook"
)

func WebhookAdmissions(client kclient.Client) webhook.Admissions {
	return webhook.Admissions{
		webhook.NewAdmissionHandler[*core.Pod](
			"policies",
			inspectorConstants.PodGroup,
			inspectorConstants.PodVersionV1,
			inspectorConstants.PodKind,
			inspectorConstants.PodResource,
			policies.NewPoliciesPodHandler(client),
		),
	}
}
