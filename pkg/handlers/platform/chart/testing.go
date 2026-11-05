//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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

//go:build testing

package chart

import (
	platform "github.com/arangodb/kube-arangodb/pkg/handlers/platform"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func Handler(operator operator.Operator, recorder event.Recorder, client kclient.Client) operator.Handler {
	return &handler{
		client:     client.Arango(),
		kubeClient: client.Kubernetes(),

		eventRecorder: recorder.NewInstance(Group(), Version(), Kind()),

		operator: operator,

		Cache: platform.NewCache(0),
	}
}
