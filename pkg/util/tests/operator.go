//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/informers"

	arangoInformer "github.com/arangodb/kube-arangodb/pkg/generated/informers/externalversions"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

type TestingOperatorRegisterer func(operator operator.Operator, recorder event.Recorder, client kclient.Client, informer arangoInformer.SharedInformerFactory, kubeInformer informers.SharedInformerFactory) error

func NewTestingOperator(ctx context.Context, t *testing.T, ns string, image util.Image, client kclient.Client, registers ...TestingOperatorRegisterer) context.CancelFunc {
	nctx, c := context.WithCancel(ctx)

	operator := operator.NewOperator("test", ns, image)

	eventRecorder := event.NewEventRecorder("test", client.Kubernetes())

	arangoInformer := arangoInformer.NewSharedInformerFactoryWithOptions(client.Arango(), 10*time.Second, arangoInformer.WithNamespace(ns))

	kubeInformer := informers.NewSharedInformerFactoryWithOptions(client.Kubernetes(), 15*time.Second, informers.WithNamespace(ns))

	for _, reg := range registers {
		require.NoError(t, reg(operator, eventRecorder, client, arangoInformer, kubeInformer))
	}

	require.NoError(t, operator.RegisterStarter(arangoInformer))

	require.NoError(t, operator.RegisterStarter(kubeInformer))

	go func() {
		require.NoError(t, operator.Start(8, nctx.Done()))
	}()

	return c
}
