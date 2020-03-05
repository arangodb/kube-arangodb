//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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

package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/trigger"
)

type pvCleaner struct {
	mutex        sync.Mutex
	log          zerolog.Logger
	cli          kubernetes.Interface
	items        []v1.PersistentVolume
	trigger      trigger.Trigger
	clientGetter func(nodeName string) (provisioner.API, error)
}

// newPVCleaner creates a new cleaner of persistent volumes.
func newPVCleaner(log zerolog.Logger, cli kubernetes.Interface, clientGetter func(nodeName string) (provisioner.API, error)) *pvCleaner {
	return &pvCleaner{
		log:          log,
		cli:          cli,
		clientGetter: clientGetter,
	}
}

// Run continues cleaning PV's until the given channel is closed.
func (c *pvCleaner) Run(stopCh <-chan struct{}) {
	for {
		delay := time.Hour
		hasMore, err := c.cleanFirst()
		if err != nil {
			c.log.Error().Err(err).Msg("Failed to clean PersistentVolume")
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
		case <-time.After(delay):
			// Continue
		}
	}
}

// Add the given volume to the list of items to clean.
func (c *pvCleaner) Add(pv v1.PersistentVolume) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check the existing list first, ignore if already found
	for _, x := range c.items {
		if x.GetUID() == pv.GetUID() {
			return
		}
	}

	// Is new, add it
	c.items = append(c.items, pv)
	c.trigger.Trigger()
}

// cleanFirst tries to clean the first PV in the list.
// Returns (hasMore, error)
func (c *pvCleaner) cleanFirst() (bool, error) {
	var first *v1.PersistentVolume
	c.mutex.Lock()
	if len(c.items) > 0 {
		first = &c.items[0]
	}
	c.mutex.Unlock()

	if first == nil {
		// Nothing todo
		return false, nil
	}

	// Do actual cleaning
	if err := c.clean(*first); err != nil {
		return true, maskAny(err)
	}

	// Remove first from list
	c.mutex.Lock()
	c.items = c.items[1:]
	remaining := len(c.items)
	c.mutex.Unlock()

	return remaining > 0, nil
}

// clean tries to clean the given PV.
func (c *pvCleaner) clean(pv v1.PersistentVolume) error {
	log := c.log.With().Str("name", pv.GetName()).Logger()
	log.Debug().Msg("Cleaning PersistentVolume")

	// Find local path
	localSource := pv.Spec.PersistentVolumeSource.Local
	if localSource == nil {
		return maskAny(fmt.Errorf("PersistentVolume has no local source"))
	}
	localPath := localSource.Path

	// Find client that serves the node
	nodeName := pv.GetAnnotations()[nodeNameAnnotation]
	if nodeName == "" {
		return maskAny(fmt.Errorf("PersistentVolume has no node-name annotation"))
	}
	client, err := c.clientGetter(nodeName)
	if err != nil {
		log.Debug().Err(err).Str("node", nodeName).Msg("Failed to get client for node")
		return maskAny(err)
	}

	// Clean volume through client
	ctx := context.Background()
	if err := client.Remove(ctx, localPath); err != nil {
		log.Debug().Err(err).
			Str("node", nodeName).
			Str("local-path", localPath).
			Msg("Failed to remove local path")
		return maskAny(err)
	}

	// Remove persistent volume
	if err := c.cli.CoreV1().PersistentVolumes().Delete(pv.GetName(), &metav1.DeleteOptions{}); err != nil && !k8sutil.IsNotFound(err) {
		log.Debug().Err(err).
			Str("name", pv.GetName()).
			Msg("Failed to remove PersistentVolume")
		return maskAny(err)
	}

	return nil
}
