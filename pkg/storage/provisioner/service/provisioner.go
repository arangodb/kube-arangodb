//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package service

import (
	"context"
	"os"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/rs/zerolog"
	"golang.org/x/sys/unix"

	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner"
)

// Config for the storage provisioner
type Config struct {
	Address  string // Server address to listen on
	NodeName string // Name of the run I'm running now
}

// Dependencies for the storage provisioner
type Dependencies struct {
	Log zerolog.Logger
}

// Provisioner implements a Local storage provisioner
type Provisioner struct {
	Config
	Dependencies
}

// New creates a new local storage provisioner
func New(config Config, deps Dependencies) (*Provisioner, error) {
	return &Provisioner{
		Config:       config,
		Dependencies: deps,
	}, nil
}

// Run the provisioner until the given context is canceled.
func (p *Provisioner) Run(ctx context.Context) {
	runServer(ctx, p.Log, p.Address, p)
}

// GetNodeInfo fetches information from the current node.
func (p *Provisioner) GetNodeInfo(ctx context.Context) (provisioner.NodeInfo, error) {
	return provisioner.NodeInfo{
		NodeName: p.NodeName,
	}, nil
}

// GetInfo fetches information from the filesystem containing
// the given local path.
func (p *Provisioner) GetInfo(ctx context.Context, localPath string) (provisioner.Info, error) {
	log := p.Log.With().Str("local-path", localPath).Logger()

	log.Debug().Msg("gettting info for local path")
	statfs := &unix.Statfs_t{}
	if err := unix.Statfs(localPath, statfs); err != nil {
		log.Error().Err(err).Msg("Statfs failed")
		return provisioner.Info{}, errors.WithStack(err)
	}

	// Available is blocks available * fragment size
	available := int64(statfs.Bavail) * statfs.Bsize

	// Capacity is total block count * fragment size
	capacity := int64(statfs.Blocks) * statfs.Bsize

	log.Debug().
		Str("node-name", p.NodeName).
		Int64("capacity", capacity).
		Int64("available", available).
		Msg("Returning info for local path")
	return provisioner.Info{
		NodeInfo: provisioner.NodeInfo{
			NodeName: p.NodeName,
		},
		Available: available,
		Capacity:  capacity,
	}, nil
}

// Prepare a volume at the given local path
func (p *Provisioner) Prepare(ctx context.Context, localPath string) error {
	log := p.Log.With().Str("local-path", localPath).Logger()
	log.Debug().Msg("preparing local path")

	// Make sure directory is empty
	if err := os.RemoveAll(localPath); err != nil && !os.IsNotExist(err) {
		log.Error().Err(err).Msg("Failed to clean existing directory")
		return errors.WithStack(err)
	}
	// Make sure directory exists
	if err := os.MkdirAll(localPath, 0755); err != nil {
		log.Error().Err(err).Msg("Failed to make directory")
		return errors.WithStack(err)
	}
	// Set access rights
	if err := os.Chmod(localPath, 0777); err != nil {
		log.Error().Err(err).Msg("Failed to set directory access")
		return errors.WithStack(err)
	}
	return nil
}

// Remove a volume with the given local path
func (p *Provisioner) Remove(ctx context.Context, localPath string) error {
	log := p.Log.With().Str("local-path", localPath).Logger()
	log.Debug().Msg("cleanup local path")

	// Make sure directory is empty
	if err := os.RemoveAll(localPath); err != nil && !os.IsNotExist(err) {
		log.Error().Err(err).Msg("Failed to clean directory")
		return errors.WithStack(err)
	}
	return nil
}
