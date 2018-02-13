// Copyright (c) 2016 Pulcy.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package docker

import (
	"path"

	log "github.com/op/go-logging"

	"github.com/pulcy/pulsar/util"
)

// Push a docker image to the arvika-ssh registry
func Push(log *log.Logger, image, dockerRegistry, dockerNamespace string) error {
	localTag := path.Join(dockerNamespace, image)
	registryTag := path.Join(dockerRegistry, dockerNamespace, image)
	if err := util.ExecPrintError(log, "docker", "tag", localTag, registryTag); err != nil {
		return err
	}
	// Push
	if err := util.ExecPrintError(log, "docker", "push", registryTag); err != nil {
		return err
	}
	// Remove registry tag
	if err := util.ExecPrintError(log, "docker", "rmi", registryTag); err != nil {
		return err
	}
	return nil
}

// Pull a docker image from the arvika-ssh registry
func Pull(log *log.Logger, image, dockerRegistry, dockerNamespace string) error {
	localTag := path.Join(dockerNamespace, image)
	registryTag := path.Join(dockerRegistry, dockerNamespace, image)
	// Pull
	if err := util.ExecPrintError(log, "docker", "pull", registryTag); err != nil {
		return err
	}
	if err := util.ExecPrintError(log, "docker", "tag", registryTag, localTag); err != nil {
		return err
	}
	// Remove registry tag
	if err := util.ExecPrintError(log, "docker", "rmi", registryTag); err != nil {
		return err
	}
	return nil
}
