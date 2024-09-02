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
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

type commandRun[T any] interface {
	Register(name, desc string, in func(ctx context.Context, client T) error) commandRun[T]
}

type commandRunImpl[T any] struct {
	cmd *cobra.Command
	cfg *Config
	in  func(cc grpc.ClientConnInterface) T
}

func (c commandRunImpl[T]) Register(name, desc string, in func(ctx context.Context, client T) error) commandRun[T] {
	c.cmd.AddCommand(&cobra.Command{
		Use:   name,
		Short: desc,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, closer, err := client(shutdown.Context(), c.cfg, c.in)
			if err != nil {
				return err
			}

			defer closer.Close()

			return in(shutdown.Context(), client)
		},
	})
	return c
}

func withCommandRun[T any](cmd *cobra.Command, cfg *Config, in func(cc grpc.ClientConnInterface) T) commandRun[T] {
	return &commandRunImpl[T]{
		cmd: cmd,
		cfg: cfg,
		in:  in,
	}
}

func client[T any](ctx context.Context, cfg *Config, in func(cc grpc.ClientConnInterface) T) (T, io.Closer, error) {
	var opts []grpc.DialOption

	if token := cfg.Token; token != "" {
		opts = append(opts, util.TokenAuthInterceptors(token)...)
	}

	if cfg.TLS.Enabled {
		config := &tls.Config{}

		if ca := cfg.TLS.CA; ca != "" {
			pemServerCA, err := os.ReadFile(ca)
			if err != nil {
				return util.Default[T](), nil, err
			}

			certPool := x509.NewCertPool()
			if !certPool.AppendCertsFromPEM(pemServerCA) {
				return util.Default[T](), nil, err
			}

			config.RootCAs = certPool
		}

		if cfg.TLS.Insecure {
			config.InsecureSkipVerify = true
		}

		if cfg.TLS.Fallback {
			client, closer, err := util.NewOptionalTLSGRPCClient(ctx, in, cfg.Address, config, opts...)
			if err != nil {
				return util.Default[T](), nil, err
			}

			return client, closer, nil
		} else {
			opts = append(opts, util.ClientTLS(config)...)
		}
	}

	client, closer, err := util.NewGRPCClient(ctx, in, cfg.Address, opts...)
	if err != nil {
		return util.Default[T](), nil, err
	}

	return client, closer, nil
}
