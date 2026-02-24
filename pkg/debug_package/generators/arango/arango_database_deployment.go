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

package arango

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func arangoDatabaseDeploymentAgencyDump(ctx context.Context, logger zerolog.Logger, client kclient.Client, files chan<- shared.File, item *api.ArangoDeployment) error {
	if !item.Spec.GetMode().HasAgents() {
		return nil
	}

	handler, err := shared.DiscoverExecFunc()
	if err != nil {
		return err
	}

	files, c := shared.WithPrefix(files, "agency/")
	defer c()

	files <- shared.NewFile("dump.json", func() ([]byte, error) {
		out, _, err := handler(logger, "admin", "agency", "dump", "-d", item.GetName())

		if err != nil {
			return nil, err
		}

		return out, nil
	})

	files <- shared.NewFile("state.json", func() ([]byte, error) {
		out, _, err := handler(logger, "admin", "agency", "state", "-d", item.GetName())

		if err != nil {
			return nil, err
		}

		return out, nil
	})

	return nil
}

func arangoDatabaseDeploymentMembers(ctx context.Context, logger zerolog.Logger, client kclient.Client, files chan<- shared.File, item *api.ArangoDeployment) error {
	files, c := shared.WithPrefix(files, "members/")
	defer c()

	for _, member := range item.Status.Members.AsList() {
		mName := member.Member.ArangoMemberName(item.GetName(), member.Group)

		arangoMember, err := client.Arango().DatabaseV1().ArangoMembers(item.GetNamespace()).Get(ctx, mName, meta.GetOptions{})
		if err != nil {
			logger.Err(err).Msgf("Unable to get member")
			continue
		}

		files <- shared.NewYAMLFile(fmt.Sprintf("%s.yaml", arangoMember.GetName()), func() ([]interface{}, error) {
			return []interface{}{arangoMember}, nil
		})

		switch member.Group.Type() {
		case api.ServerGroupTypeArangoD:
			if err := arangoDeploymentMemberArangoD(logger, files, item, member); err != nil {
				return err
			}
		}
	}

	return nil
}

func arangoDeploymentMemberArangoD(logger zerolog.Logger, files chan<- shared.File, item *api.ArangoDeployment, member api.DeploymentStatusMemberElement) error {
	files, c := shared.WithPrefix(files, "%s/", member.Member.ID)
	defer c()

	handler, err := shared.DiscoverExecFunc()
	if err != nil {
		return err
	}

	files <- shared.NewFile("activities.json", shared.GenerateDataFuncP2(func(depl, m string) ([]byte, error) {
		h, err := shared.DiscoverExecFunc()
		if err != nil {
			return nil, err
		}
		out, _, err := h(logger, "admin", "member", "request", "get", "-d", depl, "-m", m, "_admin", "activities")
		if err != nil {
			return nil, err
		}
		return out, nil
	}, item.GetName(), member.Member.ID))

	files <- shared.NewFile("async-registry.json", shared.GenerateDataFuncP2(func(depl, m string) ([]byte, error) {
		out, _, err := handler(logger, "admin", "member", "request", "get", "-d", depl, "-m", m, "_admin", "async-registry")
		if err != nil {
			return nil, err
		}
		return out, nil
	}, item.GetName(), member.Member.ID))

	listOut, _, listErr := handler(logger, "admin", "member", "request", "get", "-d", item.GetName(), "-m", member.Member.ID, "_admin", "server", "crashes")
	if listErr != nil {
		logger.Debug().Err(listErr).Msg("Failed to get crash list, skipping crash dump files")
	} else {
		files <- shared.NewFile("crashes.json", func() ([]byte, error) { return listOut, nil })
		var wrap struct {
			Result []string `json:"result"`
		}
		if json.Unmarshal(listOut, &wrap) == nil && len(wrap.Result) > 0 {
			for _, id := range wrap.Result {
				crashID := id
				files <- shared.NewFile("crashes/"+crashID+".json", shared.GenerateDataFuncP2(func(depl, m string) ([]byte, error) {
					h, err := shared.DiscoverExecFunc()
					if err != nil {
						return nil, err
					}
					out, _, err := h(logger, "admin", "member", "request", "get", "-d", depl, "-m", m, "_admin", "server", "crashes", crashID)
					if err != nil {
						return nil, err
					}
					return out, nil
				}, item.GetName(), member.Member.ID))
			}
		}
	}

	files <- shared.NewFile("stack", shared.GenerateDataFuncP2(func(depl, pod string) ([]byte, error) {
		handler, err := shared.RemoteExecFunc("/usr/bin/eu-stack", pod, "server")
		if err != nil {
			return nil, err
		}

		out, _, err := handler(logger, "-n32", "-p1")

		if err != nil {
			return nil, err
		}

		return out, nil
	}, item.GetName(), member.Member.PodName))

	return nil
}

func arangoDatabaseDeploymentPlatform(ctx context.Context, logger zerolog.Logger, client kclient.Client, files chan<- shared.File, item *api.ArangoDeployment) error {
	files, c := shared.WithPrefix(files, "platform/")
	defer c()

	files <- shared.NewFile("package.yaml", func() ([]byte, error) {
		p, err := helm.NewPackage(ctx, client, item.GetNamespace(), item.GetName())
		if err != nil {
			return nil, err
		}

		return yaml.Marshal(p)
	})

	return nil
}
