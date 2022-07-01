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

package job

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/apps"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
)

var _ operator.LifecyclePreStart = &handler{}

// LifecyclePreStart is executed before operator starts to work, additional checks can be placed here
// Wait for CR to be present
func (h *handler) LifecyclePreStart() error {
	log.Info().Msgf("Starting Lifecycle PreStart for %s", h.Name())

	defer func() {
		log.Info().Msgf("Lifecycle PreStart for %s completed", h.Name())
	}()

	for {
		_, err := h.client.AppsV1().ArangoJobs(h.operator.Namespace()).List(context.Background(), meta.ListOptions{})

		if err != nil {
			log.Warn().Err(err).Msgf("CR for %s not found", apps.ArangoJobResourceKind)

			time.Sleep(250 * time.Millisecond)
			continue
		}

		return nil
	}
}
