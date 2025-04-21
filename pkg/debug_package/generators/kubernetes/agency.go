//
// DISCLAIMER
//
// Copyright 2023-2025 ArangoDB GmbH, Cologne, Germany
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
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/rs/zerolog"

	"github.com/arangodb/kube-arangodb/pkg/debug_package/cli"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

const LocalBinDir = "/usr/bin/arangodb_operator"

func AgencyDump() shared.Factory {
	return shared.NewFactory("agency-dump", true, agencyDump)
}

func agencyDump(logger zerolog.Logger, files chan<- shared.File) error {
	ef, err := discoverExecFunc()
	if err != nil {
		return err
	}

	k, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return errors.Errorf("Client is not initialised")
	}

	deployments, err := listArangoDeployments(k.Arango())()
	if err != nil {
		return err
	}

	for _, deployment := range deployments {
		if !deployment.GetAcceptedSpec().Mode.HasAgents() {
			continue
		}

		NewDeploymentAgencyInfo(logger, files, deployment.GetName(), ef)
	}

	return nil
}

func NewDeploymentAgencyInfo(logger zerolog.Logger, out chan<- shared.File, name string, handler ArangoOperatorExecFunc) {
	out <- shared.NewFile(fmt.Sprintf("kubernetes/arango/deployments/%s/agency/dump.json", name), func() ([]byte, error) {
		out, _, err := handler(logger, "admin", "agency", "dump", "-d", name)

		if err != nil {
			return nil, err
		}

		return out, nil
	})

	out <- shared.NewFile(fmt.Sprintf("kubernetes/arango/deployments/%s/agency/state.json", name), func() ([]byte, error) {
		out, _, err := handler(logger, "admin", "agency", "state", "-d", name)

		if err != nil {
			return nil, err
		}

		return out, nil
	})
}

type ArangoOperatorExecFunc func(logger zerolog.Logger, args ...string) ([]byte, []byte, error)

func discoverExecFunc() (ArangoOperatorExecFunc, error) {
	if _, err := os.Stat(LocalBinDir); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		return remoteOperatorExecFunc(LocalBinDir)
	} else {
		return localExecFunc(LocalBinDir)
	}
}

func localExecFunc(binary string) (ArangoOperatorExecFunc, error) {
	return func(logger zerolog.Logger, args ...string) ([]byte, []byte, error) {
		logger.Debug().Str("binary", binary).Strs("args", args).Msgf("Executing remote command")

		cmd := exec.Command(binary, args...)
		var stderr, stdout bytes.Buffer

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()

		return stdout.Bytes(), stderr.Bytes(), err
	}, nil
}

func remoteOperatorExecFunc(binary string) (ArangoOperatorExecFunc, error) {
	id, err := discoverOperatorPod(binary)
	if err != nil {
		return nil, err
	}

	return remoteExecFunc(binary, id, "operator")
}

func discoverOperatorPod(binary string) (string, error) {
	k, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return "", errors.Errorf("Client is not initialised")
	}

	pods, err := listPods(k.Kubernetes())()
	if err != nil {
		return "", err
	}

	var id string

	for _, v := range pods {
		if id != "" {
			break
		}

		for _, container := range v.Spec.Containers {
			if container.Name != "operator" {
				continue
			}

			var stderr, stdout bytes.Buffer

			if err := shared.ExecuteCommandInPod(shutdown.Context(), k, "operator", v.GetName(), v.GetNamespace(), []string{binary, "version"}, nil, &stdout, &stderr); err != nil {
				continue
			}

			id = v.GetName()
		}
	}

	if id == "" {
		return "", errors.Errorf("Unable to find Operator pod")
	}

	return id, nil
}

func remoteExecFunc(binary, pod, container string) (ArangoOperatorExecFunc, error) {
	k, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return nil, errors.Errorf("Client is not initialised")
	}

	return func(logger zerolog.Logger, args ...string) ([]byte, []byte, error) {
		var stderr, stdout bytes.Buffer

		in := make([]string, len(args)+1)
		in[0] = binary
		for id := range args {
			in[id+1] = args[id]
		}

		logger.Debug().Str("binary", binary).Strs("args", args).Str("namespace", cli.GetInput().Namespace).Str("container", container).Str("pod", pod).Msgf("Executing remote command")

		err := shared.ExecuteCommandInPod(shutdown.Context(), k, container, pod, cli.GetInput().Namespace, in, nil, &stdout, &stderr)

		return stdout.Bytes(), stderr.Bytes(), err
	}, nil
}
