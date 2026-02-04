//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package tests

import (
	"os"
	"path"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
	utilTokenLoader "github.com/arangodb/kube-arangodb/pkg/util/token/loader"
)

type Token []byte

func GenerateJWTToken() Token {
	var tokenData = make([]byte, 32)
	util.Rand().Read(tokenData)
	return tokenData
}

func NewTokenManager(t *testing.T) TokenManager {
	return &tokenManager{
		path: t.TempDir(),
	}
}

type TokenManager interface {
	Path() string

	Activate(t *testing.T, token Token) TokenManager
	Clean(t *testing.T) TokenManager
	Save(t *testing.T, tokens ...Token) TokenManager

	Set(t *testing.T, active Token, tokens ...Token) TokenManager

	Sign(t *testing.T, claims utilToken.Claims) string
}

type tokenManager struct {
	lock sync.Mutex

	path string
}

func (m *tokenManager) Sign(t *testing.T, claims utilToken.Claims) string {
	m.lock.Lock()
	defer m.lock.Unlock()

	secrets, err := utilTokenLoader.LoadSecretSetFromDirectory(m.Path())
	require.NoError(t, err)

	z, err := claims.Sign(secrets)
	require.NoError(t, err)

	return z
}

func (m *tokenManager) Set(t *testing.T, active Token, tokens ...Token) TokenManager {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.clean(t)

	m.save(t, utilConstants.ActiveJWTKey, active)

	m.save(t, util.SHA256(active), active)

	for _, token := range tokens {
		m.save(t, util.SHA256(token), token)
	}

	return m
}

func (m *tokenManager) Path() string {
	return m.path
}

func (m *tokenManager) Clean(t *testing.T) TokenManager {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.clean(t)
	return m
}

func (m *tokenManager) clean(t *testing.T) {
	files, err := os.ReadDir(m.path)
	require.NoError(t, err)

	for _, f := range files {
		require.NoError(t, os.Remove(path.Join(m.path, f.Name())))
	}

	files, err = os.ReadDir(m.path)
	require.NoError(t, err)
	require.Len(t, files, 0)
}

func (m *tokenManager) Activate(t *testing.T, token Token) TokenManager {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.save(t, utilConstants.ActiveJWTKey, token)

	return m
}

func (m *tokenManager) Save(t *testing.T, tokens ...Token) TokenManager {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, token := range tokens {
		m.save(t, util.SHA256(token), token)
	}

	return m
}

func (m *tokenManager) save(t *testing.T, name string, token Token) {
	fn := path.Join(m.path, name)
	require.NoError(t, os.WriteFile(fn, token, 0644))
}
