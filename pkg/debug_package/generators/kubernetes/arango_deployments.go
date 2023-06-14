//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package kubernetes

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/cli"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	arangoClient "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func Deployments() shared.Factory {
	return shared.NewFactory("deployments", true, deployments)
}

func listArangoDeployments(client arangoClient.Interface) func() ([]*api.ArangoDeployment, error) {
	return func() ([]*api.ArangoDeployment, error) {
		return ListObjects[*api.ArangoDeploymentList, *api.ArangoDeployment](context.Background(), client.DatabaseV1().ArangoDeployments(cli.GetInput().Namespace), func(result *api.ArangoDeploymentList) []*api.ArangoDeployment {
			q := make([]*api.ArangoDeployment, len(result.Items))

			for id, e := range result.Items {
				q[id] = e.DeepCopy()
			}

			return q
		})
	}
}

func deployments(logger zerolog.Logger, files chan<- shared.File) error {
	k, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return errors.Newf("Client is not initialised")
	}

	deploymentList, err := listArangoDeployments(k.Arango())()
	if err != nil {
		return err
	}

	errDeployments := make([]error, len(deploymentList))

	for id := range deploymentList {
		errDeployments[id] = deployment(k, deploymentList[id], files)
	}

	if err := errors.Errors(errDeployments...); err != nil {
		logger.Err(err).Msgf("Error while collecting arango deployments")
		return err
	}

	return nil
}

func deployment(client kclient.Client, depl *api.ArangoDeployment, files chan<- shared.File) error {
	files <- shared.NewYAMLFile(fmt.Sprintf("kubernetes/arango/deployments/%s.yaml", depl.GetName()), func() ([]interface{}, error) {
		return []interface{}{depl}, nil
	})

	if err := deploymentMembers(client, depl, files); err != nil {
		return err
	}

	return nil
}

func deploymentMembers(client kclient.Client, depl *api.ArangoDeployment, files chan<- shared.File) error {
	for _, member := range depl.Status.Members.AsList() {
		mName := member.Member.ArangoMemberName(depl.GetName(), member.Group)

		arangoMember, err := client.Arango().DatabaseV1().ArangoMembers(cli.GetInput().Namespace).Get(context.Background(), mName, meta.GetOptions{})
		if err != nil {
			return err
		}
		files <- shared.NewYAMLFile(fmt.Sprintf("kubernetes/arango/deployments/%s/members/%s.yaml", depl.GetName(), arangoMember.GetName()), func() ([]interface{}, error) {
			return []interface{}{arangoMember}, nil
		})
	}

	return nil
}
