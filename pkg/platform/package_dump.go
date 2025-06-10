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

package platform

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func packageDump() (*cobra.Command, error) {
	var cmd cobra.Command

	cmd.Use = "dump [flags]"
	cmd.Short = "Dumps the current setup of the platform"

	if err := cli.RegisterFlags(&cmd); err != nil {
		return nil, err
	}

	cmd.RunE = getRunner().With(packageDumpRun).Run

	return &cmd, nil
}

func packageDumpRun(cmd *cobra.Command, args []string) error {
	client, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return errors.Errorf("Unable to get client")
	}

	ns, err := flagNamespace.Get(cmd)
	if err != nil {
		return err
	}

	deployment, err := flagPlatformName.Get(cmd)
	if err != nil {
		return err
	}

	out, err := helm.NewPackage(cmd.Context(), client, ns, deployment)
	if err != nil {
		return err
	}

	d, err := json.Marshal(out)
	if err != nil {
		return err
	}

	d, err = yaml.JSONToYAML(d)
	if err != nil {
		return err
	}

	return render(cmd, "---\n\n%s", string(d))
}
