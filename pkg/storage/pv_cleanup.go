//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package storage

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/timer"
	"github.com/arangodb/kube-arangodb/pkg/util/trigger"
)

var pcLogger = logging.Global().RegisterAndGetLogger("deployment-storage-pc", logging.Info)

type pvCleaner struct {
	mutex        sync.Mutex
	log          logging.Logger
	cli          kubernetes.Interface
	items        []*core.PersistentVolume
	trigger      trigger.Trigger
	clientGetter func(ctx context.Context, nodeName string) (provisioner.API, error)
}

// newPVCleaner creates a new cleaner of persistent volumes.
func newPVCleaner(cli kubernetes.Interface, clientGetter func(ctx context.Context, nodeName string) (provisioner.API, error)) *pvCleaner {
	c := &pvCleaner{
		cli:          cli,
		clientGetter: clientGetter,
	}

	c.log = pcLogger.WrapObj(c)

	return c
}

// Run continues cleaning PV's until the given channel is closed.
func (c *pvCleaner) Run(stopCh <-chan struct{}) {
	for {
		delay := time.Hour
		hasMore, err := c.cleanFirst()
		if err != nil {
			c.log.Err(err).Error("Failed to clean PersistentVolume")
		}
		if hasMore {
			delay = time.Millisecond * 5
		}

		select {
		case <-stopCh:
			// We're done
			return
		case <-c.trigger.Done():
			// Continue
		case <-timer.After(delay):
			// Continue
		}
	}
}

// Add the given volume to the list of items to clean.
func (c *pvCleaner) Add(pv *core.PersistentVolume) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check the existing list first, ignore if already found
	for _, x := range c.items {
		if x.GetUID() == pv.GetUID() {
			return
		}
	}

	// Is new, add it
	c.items = append(c.items, pv.DeepCopy())
	c.trigger.Trigger()
}

// cleanFirst tries to clean the first PV in the list.
// Returns (hasMore, error)
func (c *pvCleaner) cleanFirst() (bool, error) {
	var first *core.PersistentVolume
	c.mutex.Lock()
	if len(c.items) > 0 {
		first = c.items[0]
	}
	c.mutex.Unlock()

	if first == nil {
		// Nothing todo
		return false, nil
	}

	// Do actual cleaning
	if err := c.clean(*first); err != nil {
		return true, errors.WithStack(err)
	}

	// Remove first from list
	c.mutex.Lock()
	c.items = c.items[1:]
	remaining := len(c.items)
	c.mutex.Unlock()

	return remaining > 0, nil
}

// clean tries to clean the given PV.
func (c *pvCleaner) clean(pv core.PersistentVolume) error {
	log := c.log.Str("name", pv.GetName())
	log.Debug("Cleaning PersistentVolume")

	// Find local path
	localSource := pv.Spec.PersistentVolumeSource.Local
	if localSource == nil {
		return errors.WithStack(errors.Newf("PersistentVolume has no local source"))
	}
	localPath := localSource.Path

	// Find client that serves the node
	nodeName := pv.GetAnnotations()[nodeNameAnnotation]
	if nodeName == "" {
		return errors.WithStack(errors.Newf("PersistentVolume has no node-name annotation"))
	}
	client, err := c.clientGetter(context.Background(), nodeName)
	if err != nil {
		log.Err(err).Str("node", nodeName).Debug("Failed to get client for node")
		return errors.WithStack(err)
	}

	// Clean volume through client
	ctx := context.Background()
	if err := client.Remove(ctx, localPath); err != nil {
		log.Err(err).
			Str("node", nodeName).
			Str("local-path", localPath).
			Debug("Failed to remove local path")
		return errors.WithStack(err)
	}

	// Remove persistent volume
	if err := c.cli.CoreV1().PersistentVolumes().Delete(context.Background(), pv.GetName(), meta.DeleteOptions{}); err != nil && !kerrors.IsNotFound(err) {
		log.Err(err).
			Str("name", pv.GetName()).
			Debug("Failed to remove PersistentVolume")
		return errors.WithStack(err)
	}

	return nil
}

func (c *pvCleaner) WrapLogger(in *zerolog.Event) *zerolog.Event {
	return in
}
