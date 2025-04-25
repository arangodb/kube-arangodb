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

package helm

import (
	"github.com/rs/zerolog"
	"helm.sh/helm/v3/pkg/action"

	"github.com/arangodb/kube-arangodb/pkg/debug_package/cli"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

func Register(f shared.FactoryGen) {
	f.AddSection("helm").
		Register("releases", true, helmReleases)
}

func helmReleases(logger zerolog.Logger, files chan<- shared.File) error {
	client, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return errors.Errorf("Client is not initialised")
	}

	hclient, err := helm.NewClient(helm.Configuration{
		Namespace: cli.GetInput().Namespace,
		Config:    client.Config(),
		Driver:    nil,
	})
	if err != nil {
		return err
	}

	files, c := shared.WithPrefix(files, "helm/")
	defer c()

	existingReleases, err := hclient.List(shutdown.Context(), func(in *action.List) {
		in.All = true
	})
	if err != nil {
		return err
	}

	for _, release := range existingReleases {
		if err := helmRelease(files, release); err != nil {
			return err
		}
	}

	return nil
}

func helmRelease(files chan<- shared.File, release helm.Release) error {
	files, c := shared.WithPrefix(files, "%s/", release.Name)
	defer c()

	files <- shared.NewYAMLFile("release.yaml", func() ([]helm.Release, error) {
		return []helm.Release{release}, nil
	})

	return nil
}
