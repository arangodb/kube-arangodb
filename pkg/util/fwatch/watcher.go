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

package fwatch

import (
	"context"
	"fmt"
	"os"
	"sync"

	"k8s.io/utils/inotify"

	"github.com/arangodb/kube-arangodb/pkg/logging"
)

const (
	inotifyEventsModification = inotify.InCreate | inotify.InDelete | inotify.InDeleteSelf |
		inotify.InCloseWrite | inotify.InMove | inotify.InMove | inotify.InMoveSelf | inotify.InUnmount
)

type FileContentWatcher interface {
	// Start a routine to watch for file content changes. It will be stopped when context finishes
	Start(ctx context.Context)
	// IsChanged returns true if file content has been changed since last read
	IsChanged() bool
	// ReadAll tries to read all content from file
	ReadAll() ([]byte, error)
}

type fileContentWatcher struct {
	isRunning bool
	p         string
	w         *inotify.Watcher
	log       logging.Logger

	changed bool
	lock    sync.RWMutex
}

// NewFileContentWatcher returns FileContentWatcher, which tracks changes in file
// Returns error if filePath is a directory.
// Caller must Close() the watcher once work finished.
func NewFileContentWatcher(filePath string, log logging.Logger) (FileContentWatcher, error) {
	watcher, err := inotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("unable to setup inotify: %s", err)
	}
	err = watcher.AddWatch(filePath, inotifyEventsModification)
	if err != nil {
		return nil, fmt.Errorf("unable to AddWatch: %s", err)
	}

	// This returns an *os.FileInfo type
	fileInfo, err := os.Stat(filePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("unexpected err while Stat: %s", err)
	}
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("can't operator on directories")
	}

	fw := &fileContentWatcher{p: filePath, w: watcher, log: log, changed: true}

	return fw, nil
}

func (fw *fileContentWatcher) Start(ctx context.Context) {
	fw.isRunning = true
	go func() {
		defer func() {
			fw.isRunning = false
		}()

		defer fw.w.Close()

		fw.log.Info("Starting to watch for file content")

		for {
			select {
			case <-ctx.Done():
				err := fw.w.Close()
				if err != nil {
					fw.log.Err(err).Info("error while closing inotify watcher")
				}
				return
			case err := <-fw.w.Error:
				fw.log.Err(err).Debug("error while watching for file content")
			case e := <-fw.w.Event:
				fw.log.Info("changed: %s", e.String())
				fw.markAsChanged()

				if e.Mask&inotify.InIgnored == 0 {
					// IN_IGNORED can happen if file is deleted
					// restart watch:
					err := fw.w.RemoveWatch(fw.p)
					if err != nil {
						fw.log.Err(err).Warn("RemoveWatch failed")
					}
					err = fw.w.AddWatch(fw.p, inotifyEventsModification)
					if err != nil {
						fw.log.Err(err).Error("Could not start watch again after getting IN_IGNORED")
					}
				}
			}
		}
	}()
}

func (fw *fileContentWatcher) markAsChanged() {
	fw.lock.Lock()
	defer fw.lock.Unlock()
	fw.changed = true
}

func (fw *fileContentWatcher) IsChanged() bool {
	fw.lock.RLock()
	defer fw.lock.RUnlock()
	return fw.changed
}

func (fw *fileContentWatcher) ReadAll() ([]byte, error) {
	fw.lock.Lock()
	defer fw.lock.Unlock()

	result, err := os.ReadFile(fw.p)
	if err != nil {
		return nil, err
	}

	fw.changed = false

	return result, nil
}
