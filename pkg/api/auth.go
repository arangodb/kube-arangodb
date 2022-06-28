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

package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jg "github.com/golang-jwt/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type authorization struct {
	jwtSigningKey string
}

func (a *authorization) isValid(token string) bool {
	t, err := jg.Parse(token, func(_ *jg.Token) (interface{}, error) {
		return []byte(a.jwtSigningKey), nil
	})
	if err != nil {
		apiLogger.Err(err).Info("invalid JWT: %s", token)
		return false
	}
	return t.Valid
}

// ensureHTTPAuth ensure a valid token exists within HTTP request header
func (a *authorization) ensureHTTPAuth(c *gin.Context) {
	h := c.Request.Header.Values("Authorization")
	bearerToken := extractBearerToken(h)
	if !a.isValid(bearerToken) {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

// ensureGRPCAuth ensures a valid token exists within a GRPC request's metadata
func (a *authorization) ensureGRPCAuth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
	}

	// The keys within metadata.MD are normalized to lowercase.
	// See: https://godoc.org/google.golang.org/grpc/metadata#New
	bearerToken := extractBearerToken(md["authorization"])
	if !a.isValid(bearerToken) {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}
	// Continue execution of handler after ensuring a valid token.
	return handler(ctx, req)
}

func extractBearerToken(authorization []string) string {
	if len(authorization) < 1 {
		return ""
	}
	return strings.TrimPrefix(authorization[0], "Bearer ")
}
