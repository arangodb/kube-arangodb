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

package shared

import (
	"io"

	core "k8s.io/api/core/v1"
	scheme2 "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

// ExecuteCommandInPod executes command in pod with the given pod name and namespace.
func ExecuteCommandInPod(k kclient.Client, container, podName, namespace string,
	command []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {

	req := k.Kubernetes().CoreV1().RESTClient().Post().Resource("pods").Name(podName).
		Namespace(namespace).SubResource("exec")

	option := &core.PodExecOptions{
		Command:   command,
		Container: container,
		Stdin:     stdin != nil,
		Stdout:    true,
		Stderr:    stderr != nil,
		TTY:       false,
	}

	req.VersionedParams(
		option,
		scheme2.ParameterCodec,
	)

	exec, err := remotecommand.NewSPDYExecutor(k.Config(), "POST", req.URL())
	if err != nil {
		return err
	}

	return exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	})
}
