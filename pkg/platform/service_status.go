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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/pretty"
)

type ServiceTable struct {
	Name    string `table:"Name" table_align:"center" table_header_align:"center"`
	Status  string `table:"Status" table_align:"center" table_header_align:"center"`
	Version string `table:"Version" table_align:"center" table_header_align:"center"`
}

func serviceStatus() (*cobra.Command, error) {
	var cmd cobra.Command

	cmd.Use = "status [flags] deployment"
	cmd.Short = "Shows Service Status"

	if err := cli.RegisterFlags(&cmd, flagPlatformStage, flagPlatformEndpoint, flagPlatformName, flagOutput); err != nil {
		return nil, err
	}

	cmd.RunE = getRunner().With(serviceStatusRun).Run

	return &cmd, nil
}

func serviceStatusRun(cmd *cobra.Command, args []string) error {
	client, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return errors.Errorf("Unable to get client")
	}

	hclient, err := getHelmClient(cmd)
	if err != nil {
		return errors.Wrapf(err, "Unable to get helm client")
	}

	if len(args) < 1 {
		return errors.Errorf("Invalid arguments")
	}

	ns, err := flagNamespace.Get(cmd)
	if err != nil {
		return err
	}

	deployment := args[0]

	logger := logger.Str("deployment", deployment)

	logger.Debug("Fetch ArangoDeployment")
	deploymentObject, err := client.Arango().DatabaseV1().ArangoDeployments(ns).Get(cmd.Context(), deployment, meta.GetOptions{})
	if err != nil {
		logger.Err(err).Error("Unable to get deployment")
		return err
	}
	logger.Str("uid", string(deploymentObject.GetUID())).Debug("ArangoDeployment Found")

	releases, err := hclient.List(cmd.Context(), func(in *action.List) {
		in.Selector = meta.FormatLabelSelector(&meta.LabelSelector{
			MatchLabels: map[string]string{
				constants.HelmLabelArangoDBManaged:    "true",
				constants.HelmLabelArangoDBDeployment: deployment,
				constants.HelmLabelArangoDBType:       "platform",
			},
		})
	})
	if err != nil {
		logger.Err(err).Error("Unable to list releases")
		return err
	}

	t := pretty.NewTable[ServiceTable]()

	for _, rel := range releases {
		t.Add(ServiceTable{
			Name:    rel.Name,
			Status:  rel.Info.Status.String(),
			Version: rel.GetChart().GetMetadata().GetVersion(),
		})
	}

	return renderOutput(cmd, t)
}
