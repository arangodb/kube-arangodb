//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package service

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"golang.org/x/sys/unix"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

var logger = logging.Global().RegisterAndGetLogger("deployment-storage-service", logging.Info)

// Config for the storage provisioner
type Config struct {
	Address  string // Server address to listen on
	NodeName string // Name of the run I'm running now
}

// Provisioner implements a Local storage provisioner
type Provisioner struct {
	Log logging.Logger
	Config
}

// New creates a new local storage provisioner
func New(config Config) (*Provisioner, error) {
	p := &Provisioner{
		Config: config,
	}

	p.Log = logger.WrapObj(p)

	return p, nil
}

func (p *Provisioner) WrapLogger(in *zerolog.Event) *zerolog.Event {
	return in
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
	log := p.Log.Str("local-path", localPath)

	log.Debug("gettting info for local path")
	statfs := &unix.Statfs_t{}
	if err := unix.Statfs(localPath, statfs); err != nil {
		log.Err(err).Error("Statfs failed")
		return provisioner.Info{}, errors.WithStack(err)
	}

	// Available is blocks available * fragment size
	available := int64(statfs.Bavail) * statfs.Bsize // nolint:typecheck

	// Capacity is total block count * fragment size
	capacity := int64(statfs.Blocks) * statfs.Bsize // nolint:typecheck

	log.
		Str("node-name", p.NodeName).
		Int64("capacity", capacity).
		Int64("available", available).
		Debug("Returning info for local path")
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
	log := p.Log.Str("local-path", localPath)
	log.Debug("preparing local path")

	// Make sure directory is empty
	if err := os.RemoveAll(localPath); err != nil && !os.IsNotExist(err) {
		log.Err(err).Error("Failed to clean existing directory")
		return errors.WithStack(err)
	}
	// Make sure directory exists
	if err := os.MkdirAll(localPath, 0755); err != nil {
		log.Err(err).Error("Failed to make directory")
		return errors.WithStack(err)
	}
	// Set access rights
	if err := os.Chmod(localPath, 0777); err != nil {
		log.Err(err).Error("Failed to set directory access")
		return errors.WithStack(err)
	}
	return nil
}

// Remove a volume with the given local path
func (p *Provisioner) Remove(ctx context.Context, localPath string) error {
	log := p.Log.Str("local-path", localPath)
	log.Debug("cleanup local path")

	// Make sure directory is empty
	if err := os.RemoveAll(localPath); err != nil && !os.IsNotExist(err) {
		log.Err(err).Error("Failed to clean directory")
		return errors.WithStack(err)
	}
	return nil
}
