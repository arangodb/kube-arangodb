//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package pack

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"
	"path"
	"sync"

	"github.com/pkg/errors"
)

func NewCache(path string) Cache {
	return &cache{
		lock:    sync.Mutex{},
		path:    path,
		files:   map[string]string{},
		writers: map[string]*cacheWriter{},
	}
}

type Cache interface {
	CacheObject(checksum string, path string, args ...any) (io.WriteCloser, error)

	Get(checksum string, path string, args ...any) (io.ReadCloser, error)

	Saved() int
}

type cache struct {
	lock sync.Mutex

	path string

	saved int

	files map[string]string

	writers map[string]*cacheWriter
}

func (c *cache) Saved() int {
	return c.saved
}

func (c *cache) Get(checksum string, p string, args ...any) (io.ReadCloser, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	z := fmt.Sprintf(p, args...)

	if v, ok := c.files[z]; !ok || v != checksum {
		return nil, os.ErrNotExist
	}

	return os.Open(path.Join(c.path, z))
}

func (c *cache) CacheObject(checksum string, p string, args ...any) (io.WriteCloser, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	z := fmt.Sprintf(p, args...)

	if v, ok := c.files[z]; ok {
		if v == checksum {
			return nil, os.ErrExist
		} else {
			return nil, errors.Errorf("cache object %s already exists with checksum %s, expected %s", z, v, checksum)
		}
	}

	if _, ok := c.writers[z]; ok {
		return nil, nil
	}

	if current, err := c.readChecksum(p, args...); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else if current == checksum {
		c.files[z] = checksum
		return nil, os.ErrExist
	}

	fd, err := os.CreateTemp("", "tmp-")
	if err != nil {
		return nil, err
	}

	q := &cacheWriter{
		cache:       c,
		remote:      fd,
		dest:        z,
		hash:        checksum,
		currentHash: sha256.New(),
	}

	c.writers[z] = q

	return q, nil
}

func (c *cache) readChecksum(p string, args ...any) (string, error) {
	f, err := os.Open(path.Join(c.path, fmt.Sprintf(p, args...)))
	if err != nil {
		return "", err
	}

	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (c *cache) complete(z *cacheWriter) error {
	if err := os.MkdirAll(path.Dir(path.Join(c.path, z.dest)), 0755); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	if _, err := z.remote.Seek(0, 0); err != nil {
		return err
	}

	hash := fmt.Sprintf("%x", z.currentHash.Sum(nil))

	if hash != z.hash {
		return errors.Errorf("checksum mismatch for %s != %s", hash, z.hash)
	}

	out, err := os.OpenFile(path.Join(c.path, z.dest), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	defer out.Close()

	if _, err := io.Copy(out, z.remote); err != nil {
		return err
	}

	if err := z.remote.Close(); err != nil {
		return err
	}

	if err := os.Remove(z.remote.Name()); err != nil {
		return err
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	c.saved += 1

	c.files[z.dest] = hash

	delete(c.writers, z.dest)

	return nil
}

type cacheWriter struct {
	lock sync.Mutex

	cache *cache

	remote      *os.File
	currentHash hash.Hash

	dest string
	hash string
}

func (c *cacheWriter) Write(p []byte) (n int, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	n, err = c.remote.Write(p)
	if err != nil {
		return n, err
	}

	c.currentHash.Write(p[:n])

	return n, nil
}

func (c *cacheWriter) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.cache.complete(c)
}
