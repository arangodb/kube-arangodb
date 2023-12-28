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

	"github.com/fsnotify/fsnotify"

	"github.com/arangodb/kube-arangodb/pkg/logging"
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
	w         *fsnotify.Watcher
	log       logging.Logger

	changed bool
	lock    sync.RWMutex
}

// NewFileContentWatcher returns FileContentWatcher, which tracks changes in file
// Returns error if filePath is a directory.
// Caller must Close() the watcher once work finished.
func NewFileContentWatcher(filePath string, log logging.Logger) (FileContentWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("unable to setup fsnotify: %s", err)
	}
	err = watcher.Add(filePath)
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
					fw.log.Err(err).Info("error while closing fsnotify watcher")
				} else {
					fw.log.Info("fsnotify watcher closed")
				}
				return
			case err, ok := <-fw.w.Errors:
				if !ok {
					return
				}
				fw.log.Err(err).Debug("error while watching for file content")
			case event, ok := <-fw.w.Events:
				if !ok {
					return
				}

				// File attributes were changed - skip it
				if event.Op == fsnotify.Chmod {
					continue
				}

				fw.log.Info("modified file: %s", event.Name)
				fw.markAsChanged()

				if event.Op == fsnotify.Remove {
					// restart watch on removed file
					if err := fw.w.Remove(fw.p); err != nil {
						fw.log.Err(err).Error("unable to remove watch")
					}

					if err := fw.w.Add(fw.p); err != nil {
						fw.log.Err(err).Error("could not start watch again")
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
