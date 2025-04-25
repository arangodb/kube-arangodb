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

package shared

import (
	"bytes"
	"os"
	"os/exec"

	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/debug_package/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/list"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

const LocalBinDir = "/usr/bin/arangodb_operator"

type ArangoOperatorExecFunc func(logger zerolog.Logger, args ...string) ([]byte, []byte, error)

func DiscoverExecFunc() (ArangoOperatorExecFunc, error) {
	if _, err := os.Stat(LocalBinDir); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		return RemoteOperatorExecFunc(LocalBinDir)
	} else {
		return LocalExecFunc(LocalBinDir)
	}
}

func LocalExecFunc(binary string) (ArangoOperatorExecFunc, error) {
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

func RemoteOperatorExecFunc(binary string) (ArangoOperatorExecFunc, error) {
	id, err := DiscoverOperatorPod(binary)
	if err != nil {
		return nil, err
	}

	return RemoteExecFunc(binary, id, "operator")
}

func DiscoverOperatorPod(binary string) (string, error) {
	k, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return "", errors.Errorf("Client is not initialised")
	}

	pods, err := list.ListObjects[*core.PodList, *core.Pod](shutdown.Context(), k.Kubernetes().CoreV1().Pods(cli.GetInput().Namespace), func(result *core.PodList) []*core.Pod {
		q := make([]*core.Pod, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
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

			if err := ExecuteCommandInPod(shutdown.Context(), k, "operator", v.GetName(), v.GetNamespace(), []string{binary, "version"}, nil, &stdout, &stderr); err != nil {
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

func RemoteExecFunc(binary, pod, container string) (ArangoOperatorExecFunc, error) {
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

		err := ExecuteCommandInPod(shutdown.Context(), k, container, pod, cli.GetInput().Namespace, in, nil, &stdout, &stderr)

		return stdout.Bytes(), stderr.Bytes(), err
	}, nil
}
