//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package scheduler

import (
	"context"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func InitCommand(cmd *cobra.Command) error {
	var c cli
	return c.register(cmd)
}

type cli struct {
	Namespace string

	Labels []string
	Envs   []string

	Profiles []string

	Container string

	Image string
}

func (c *cli) asRequest(args ...string) (Request, error) {
	var r = Request{
		Labels: map[string]string{},
		Envs:   map[string]string{},
	}

	for _, l := range c.Labels {
		p := strings.SplitN(l, "=", 2)
		if len(p) == 1 {
			r.Labels[p[0]] = ""
			logger.Debug("Label Discovered: %s", p[0])
		} else {
			r.Labels[p[0]] = p[1]
			logger.Debug("Label Discovered: %s=%s", p[0], p[1])
		}
	}

	for _, l := range c.Envs {
		p := strings.SplitN(l, "=", 2)
		if len(p) == 1 {
			return r, errors.Errorf("Missing value for env: %s", p[0])
		} else {
			r.Envs[p[0]] = p[1]
			logger.Debug("Env Discovered: %s=%s", p[0], p[1])
		}
	}

	if len(c.Profiles) > 0 {
		r.Profiles = c.Profiles
		logger.Debug("Enabling profiles: %s", strings.Join(c.Profiles, ", "))
	}

	r.Container = util.NewType(c.Container)
	if c.Image != "" {
		r.Image = util.NewType(c.Image)
	}

	r.Args = args

	return r, nil
}

func (c *cli) register(cmd *cobra.Command) error {
	if err := logging.Init(cmd); err != nil {
		return err
	}

	cmd.RunE = c.run

	f := cmd.PersistentFlags()

	f.StringVarP(&c.Namespace, "namespace", "n", constants.NamespaceWithDefault("default"), "Kubernetes namespace")
	f.StringSliceVarP(&c.Labels, "label", "l", nil, "Scheduler Render Labels in format <key>=<value>")
	f.StringSliceVarP(&c.Envs, "env", "e", nil, "Scheduler Render Envs in format <key>=<value>")
	f.StringSliceVarP(&c.Profiles, "profile", "p", nil, "Scheduler Render Profiles")
	f.StringVar(&c.Container, "container", DefaultContainerName, "Container Name")
	f.StringVar(&c.Image, "image", "", "Image")

	return nil
}

func (c *cli) run(cmd *cobra.Command, args []string) error {
	if err := logging.Enable(); err != nil {
		return err
	}

	r, err := c.asRequest()
	if err != nil {
		return err
	}

	k, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return errors.Errorf("Unable to create Kubernetes Client")
	}

	s := NewScheduler(k, c.Namespace)

	rendered, profiles, err := s.Render(context.Background(), r)
	if err != nil {
		return err
	}
	logger.Debug("Enabled profiles: %s", strings.Join(profiles, ", "))

	data, err := yaml.Marshal(rendered)
	if err != nil {
		return err
	}

	if _, err := util.WriteAll(os.Stdout, data); err != nil {
		return err
	}

	return nil
}
