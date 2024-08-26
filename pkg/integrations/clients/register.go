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

package clients

import (
	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

var registerer = util.NewRegisterer[string, Factory]()

type Factory func(c *Config) Client

type Config struct {
	Address string
	Token   string
}

func (c *Config) Register(cmd *cobra.Command) error {
	f := cmd.PersistentFlags()

	f.StringVar(&c.Address, "address", "127.0.0.1:8080", "GRPC Service Address")
	f.StringVar(&c.Token, "token", "", "GRPC Token")

	return nil
}

type Client interface {
	Name() string
	Version() string

	Register(cmd *cobra.Command) error
}

func Register(cmd *cobra.Command) error {
	client := &cobra.Command{Use: "client"}
	cmd.AddCommand(client)

	var cfg config

	return cfg.Register(client)
}

type config struct {
	cfg Config
}

func (c *config) Register(cmd *cobra.Command) error {
	if err := c.cfg.Register(cmd); err != nil {
		return err
	}

	cmds := map[string]*cobra.Command{}

	for _, command := range registerer.Items() {
		r := command.V(&c.cfg)

		v, ok := cmds[r.Name()]
		if !ok {
			v = &cobra.Command{
				Use: r.Name(),
			}
			cmd.AddCommand(v)
			cmds[r.Name()] = v
		}

		p := &cobra.Command{
			Use: r.Version(),
		}

		if err := r.Register(p); err != nil {
			return err
		}

		v.AddCommand(p)
	}

	return nil
}
