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

package server

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	tokenExpirationTime = time.Hour
	bearerPrefix        = "bearer "
)

var authLogger = logging.Global().RegisterAndGetLogger("server-authentication", logging.Info)

type serverAuthentication struct {
	secrets typedCore.SecretInterface
	admin   struct {
		mutex    sync.Mutex
		username string
		password string
	}
	tokens struct {
		mutex  sync.Mutex
		tokens map[string]*tokenEntry
	}
	adminSecretName string
	allowAnonymous  bool
}

type tokenEntry struct {
	Token     string
	ExpiresAt time.Time
}

func (t *tokenEntry) IsExpired() bool {
	return t.ExpiresAt.Before(time.Now())
}

// loginRequest is the JSON structure POSTed to `/login`.
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// loginResponse is the JSON structure returned from `/login`.
type loginResponse struct {
	Token string `json:"token"`
}

// newServerAuthentication creates a new server authentication service
// for the given arguments.
func newServerAuthentication(secrets typedCore.SecretInterface, adminSecretName string, allowAnonymous bool) *serverAuthentication {
	auth := &serverAuthentication{
		secrets:         secrets,
		adminSecretName: adminSecretName,
		allowAnonymous:  allowAnonymous,
	}
	auth.tokens.tokens = make(map[string]*tokenEntry)
	return auth
}

// fetchAdminSecret tries to fetch the admin username & password from the configured Secret.
// Returns username, password, error
func (s *serverAuthentication) fetchAdminSecret() (string, string, error) {
	if s.adminSecretName == "" {
		return "", "", errors.WithStack(errors.Newf("No admin secret name specified"))
	}
	secret, err := s.secrets.Get(context.Background(), s.adminSecretName, meta.GetOptions{})
	if err != nil {
		return "", "", errors.WithStack(err)
	}
	var username, password string
	if raw, found := secret.Data[core.BasicAuthUsernameKey]; !found {
		return "", "", errors.WithStack(errors.Newf("Secret '%s' contains no '%s' field", s.adminSecretName, core.BasicAuthUsernameKey))
	} else {
		username = string(raw)
	}
	if raw, found := secret.Data[core.BasicAuthPasswordKey]; !found {
		return "", "", errors.WithStack(errors.Newf("Secret '%s' contains no '%s' field", s.adminSecretName, core.BasicAuthPasswordKey))
	} else {
		password = string(raw)
	}
	s.admin.mutex.Lock()
	defer s.admin.mutex.Unlock()
	s.admin.username = username
	s.admin.password = password
	return username, password, nil
}

// checkLogin compares the given username+password with the admin credentials.
// If needed admin credentials are loaded first.
func (s *serverAuthentication) checkLogin(username, password string) error {
	s.admin.mutex.Lock()
	expectedUsername := s.admin.username
	expectedPassword := s.admin.password
	s.admin.mutex.Unlock()

	if expectedUsername == "" {
		var err error
		if expectedUsername, expectedPassword, err = s.fetchAdminSecret(); err != nil {
			authLogger.Err(err).Error("Failed to fetch secret")
			return errors.WithStack(errors.Wrap(UnauthorizedError, "admin secret cannot be loaded"))
		}
	}

	if expectedUsername != username || expectedPassword != password {
		return errors.WithStack(errors.Wrap(UnauthorizedError, "invalid credentials"))
	}
	return nil
}

// Handle the authentication check
func (s *serverAuthentication) checkAuthentication(c *gin.Context) {
	if s.allowAnonymous {
		// All ok
		return
	}
	// Fetch authorization token
	authHdr := strings.ToLower(c.Request.Header.Get("Authorization"))
	if !strings.HasPrefix(authHdr, bearerPrefix) {
		sendError(c, errors.WithStack(errors.Wrap(UnauthorizedError, "missing bearer token")))
		c.Abort()
		return
	}
	token := strings.TrimSpace(authHdr[len(bearerPrefix):])
	// Lookup token
	s.tokens.mutex.Lock()
	defer s.tokens.mutex.Unlock()
	if entry, found := s.tokens.tokens[token]; !found {
		authLogger.Str("token", token).Debug("Invalid token")
		sendError(c, errors.WithStack(errors.Wrap(UnauthorizedError, "invalid credentials")))
		c.Abort()
		return
	} else if entry.IsExpired() {
		authLogger.Str("token", token).Debug("Token expired")
		sendError(c, errors.WithStack(errors.Wrap(UnauthorizedError, "credentials expired")))
		c.Abort()
		return
	} else {
		// All good, renew expiration
		entry.ExpiresAt = time.Now().Add(tokenExpirationTime)
	}
}

// Handle a POST /login request
func (s *serverAuthentication) handleLogin(c *gin.Context) {
	var req loginRequest
	if err := c.BindJSON(&req); err != nil {
		sendError(c, err)
		return
	}
	if err := s.checkLogin(req.Username, req.Password); err != nil {
		sendError(c, err)
		return
	}
	// Create new token
	token := strings.ToLower(uniuri.New())
	s.tokens.mutex.Lock()
	defer s.tokens.mutex.Unlock()
	s.tokens.tokens[token] = &tokenEntry{
		Token:     token,
		ExpiresAt: time.Now().Add(tokenExpirationTime),
	}
	// Send response
	c.JSON(http.StatusOK, loginResponse{
		Token: token,
	})
}
