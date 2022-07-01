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
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

var (
	NotFoundError     = errors.New("not found")
	UnauthorizedError = errors.New("unauthorized")
)

func isNotFound(err error) bool {
	return err == NotFoundError || errors.Cause(err) == NotFoundError
}

func isUnauthorized(err error) bool {
	return err == UnauthorizedError || errors.Cause(err) == UnauthorizedError
}

// sendError sends an error on the given context
func sendError(c *gin.Context, err error) {
	// TODO proper status handling
	code := http.StatusInternalServerError
	if isNotFound(err) {
		code = http.StatusNotFound
	} else if isUnauthorized(err) {
		code = http.StatusUnauthorized
	}
	c.JSON(code, gin.H{
		"error": err.Error(),
	})
}
