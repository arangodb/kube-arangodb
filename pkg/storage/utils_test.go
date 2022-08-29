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

package storage

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/require"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func generateDaemonSet(t *testing.T, podSpec core.PodSpec, lsSpec api.LocalStorageSpec) (*LocalStorage, *apps.DaemonSet) {
	client := kclient.NewFakeClient()

	name := fmt.Sprintf("pod-%s", strings.ToLower(uniuri.NewLen(6)))
	nameLS := fmt.Sprintf("pod-%s", strings.ToLower(uniuri.NewLen(6)))
	ns := fmt.Sprintf("ns-%s", strings.ToLower(uniuri.NewLen(6)))

	pod := core.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: podSpec,
	}

	lg := logging.NewDefaultFactory()

	ls := &LocalStorage{
		log: lg.RegisterAndGetLogger("test", logging.Info),
		apiObject: &api.ArangoLocalStorage{
			ObjectMeta: meta.ObjectMeta{
				Name:      nameLS,
				Namespace: ns,
			},
			Spec: lsSpec,
		},
		deps: Dependencies{
			Client: client,
		},
		config: Config{
			Namespace: ns,
			PodName:   name,
		},
	}

	if _, err := client.Kubernetes().CoreV1().Pods(ns).Create(context.Background(), &pod, meta.CreateOptions{}); err != nil {
		require.NoError(t, err)
	}

	image, pullPolicy, pullSecrets, err := ls.getImage(ns, name, client.Kubernetes())
	require.NoError(t, err)

	ls.image = image
	ls.imagePullPolicy = pullPolicy
	ls.imagePullSecrets = pullSecrets

	err = ls.ensureDaemonSet(ls.apiObject)
	require.NoError(t, err)

	// verify if DaemonSet has been created with correct values
	ds, err := ls.deps.Client.Kubernetes().AppsV1().DaemonSets(ns).Get(context.Background(), nameLS, meta.GetOptions{})
	require.NoError(t, err)

	return ls, ds
}
