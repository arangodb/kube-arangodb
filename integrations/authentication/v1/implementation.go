//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package v1

import (
	"context"
	"fmt"
	goHttp "net/http"
	"slices"
	goStrings "strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/connection"

	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	"github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/impl/auth_cookie"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	platformAuthenticationApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1/authentication"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/token"
)

func New(ctx context.Context, cfg Configuration) (svc.Handler, error) {
	return newInternal(ctx, cfg)
}

func newInternal(ctx context.Context, cfg Configuration) (*implementation, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	obj := &implementation{
		cfg: cfg,
		ctx: ctx,

		cache: cache.NewObject(newCache(cfg)),

		userClient: cache.NewObject(func(ctx context.Context) (arangodb.Requests, time.Duration, error) {
			client := arangodb.NewClient(connection.NewHttpConnection(connection.HttpConfiguration{
				Endpoint: connection.NewRoundRobinEndpoints([]string{
					fmt.Sprintf("%s://%s:%d", cfg.Database.Proto, cfg.Database.Endpoint, cfg.Database.Port),
				}),
				ContentType:    connection.ApplicationJSON,
				ArangoDBConfig: connection.ArangoDBConfiguration{},
				Transport:      operatorHTTP.RoundTripperWithShortTransport(operatorHTTP.WithTransportTLS(operatorHTTP.Insecure)),
			}))

			return client, 24 * time.Hour, nil
		}),
	}

	return obj, nil
}

var _ pbAuthenticationV1.AuthenticationV1Server = &implementation{}
var _ svc.Handler = &implementation{}

type implementation struct {
	pbAuthenticationV1.UnimplementedAuthenticationV1Server

	ctx context.Context
	cfg Configuration

	userClient cache.Object[arangodb.Requests]
	cache      cache.Object[token.Secret]
}

func (i *implementation) Name() string {
	return pbAuthenticationV1.Name
}

func (i *implementation) Health() svc.HealthState {
	return svc.Healthy
}

func (i *implementation) Register(registrar *grpc.Server) {
	pbAuthenticationV1.RegisterAuthenticationV1Server(registrar, i)
}

func (i *implementation) Gateway(ctx context.Context, mux *runtime.ServeMux) error {
	return pbAuthenticationV1.RegisterAuthenticationV1HandlerServer(ctx, mux, i)
}

func (i *implementation) Validate(ctx context.Context, request *pbAuthenticationV1.ValidateRequest) (*pbAuthenticationV1.ValidateResponse, error) {
	if request == nil {
		return nil, errors.Errorf("Request is nil")
	}

	if !i.cfg.Enabled {
		return &pbAuthenticationV1.ValidateResponse{
			IsValid: true,
			Details: &pbAuthenticationV1.ValidateResponseDetails{
				Lifetime: durationpb.New(DefaultTokenMaxTTL),
				User:     DefaultAdminUser,
			},
		}, nil
	}

	cache, err := i.cache.Get(ctx)
	if err != nil {
		return nil, err
	}

	user, roles, exp, err := i.extractTokenDetails(cache, request.GetToken())
	if err != nil {
		return &pbAuthenticationV1.ValidateResponse{
			IsValid: false,
			Message: err.Error(),
		}, nil
	}

	return &pbAuthenticationV1.ValidateResponse{
		IsValid: true,
		Details: &pbAuthenticationV1.ValidateResponseDetails{
			Lifetime: durationpb.New(exp),
			User:     user,
			Roles:    roles,
		},
	}, nil
}

func (i *implementation) CreateToken(ctx context.Context, request *pbAuthenticationV1.CreateTokenRequest) (*pbAuthenticationV1.CreateTokenResponse, error) {
	if request == nil {
		return nil, errors.Errorf("Request is nil")
	}

	if !i.cfg.Enabled {
		// Authentication is not enabled, pass with empty token

		return &pbAuthenticationV1.CreateTokenResponse{
			Lifetime: durationpb.New(DefaultTokenMaxTTL),
			User:     DefaultAdminUser,
			Token:    "",
		}, nil
	}

	cache, err := i.cache.Get(ctx)
	if err != nil {
		return nil, err
	}

	user := util.TypeOrDefault(request.User, i.cfg.Create.DefaultUser)
	duration := i.cfg.Create.DefaultTTL
	if v := request.Lifetime; v != nil {
		duration = v.AsDuration()
	}

	// Check configuration
	if v := i.cfg.Create.AllowedUsers; len(v) > 0 {
		if !slices.Contains(v, user) {
			return nil, errors.Errorf("User %s is not allowed", user)
		}
	}

	if v := i.cfg.Create.MaxTTL; duration > v {
		duration = v
	}

	if v := i.cfg.Create.MinTTL; duration < v {
		duration = v
	}

	signedToken, err := token.NewClaims().With(
		token.WithDefaultClaims(),
		token.WithCurrentIAT(),
		token.WithDuration(duration),
		token.WithUsername(user),
		token.WithRoles(request.GetRoles()...)).Sign(cache)
	if err != nil {
		return nil, err
	}

	user, roles, _, err := i.extractTokenDetails(cache, signedToken)
	if err != nil {
		return nil, err
	}

	return &pbAuthenticationV1.CreateTokenResponse{
		Lifetime: durationpb.New(duration),
		User:     user,
		Token:    signedToken,
		Roles:    roles,
	}, nil
}

func (i *implementation) Identity(ctx context.Context, _ *pbSharedV1.Empty) (*pbAuthenticationV1.IdentityResponse, error) {
	if !i.cfg.Enabled {
		// Auth is disabled, return static response
		return &pbAuthenticationV1.IdentityResponse{User: "root"}, nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	auth, ok := md["authorization"]
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	if len(auth) != 1 {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	for _, a := range auth {
		if !goStrings.HasPrefix(a, "bearer ") {
			continue
		}

		a = goStrings.TrimSpace(goStrings.TrimPrefix(a, "bearer "))

		resp, err := i.Validate(ctx, &pbAuthenticationV1.ValidateRequest{
			Token: a,
		})
		if err != nil {
			logger.Err(err).Warn("Error during identity fetch")
			continue
		}

		if !resp.GetIsValid() {
			continue
		}

		return &pbAuthenticationV1.IdentityResponse{User: resp.GetDetails().GetUser()}, nil
	}

	return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
}

func (i *implementation) Login(ctx context.Context, login *pbAuthenticationV1.LoginRequest) (*pbAuthenticationV1.LoginResponse, error) {
	client, err := i.userClient.Get(ctx)
	if err != nil {
		return nil, err
	}

	if login.Credentials == nil {
		return nil, status.Errorf(codes.InvalidArgument, "missing credentials")
	}

	var req authRequest
	req.Username = login.Credentials.Username
	req.Password = login.Credentials.Password
	var resp authResponse

	response, err := client.Post(ctx, &resp, &req, "_open", "auth")
	if err != nil {
		logger.Err(err).Warn("Unable to Login via /_open/auth")
		return nil, status.Errorf(codes.Internal, "Internal error while authenticating")
	}

	switch response.Code() {
	case goHttp.StatusUnauthorized:
		return nil, status.Errorf(codes.Unauthenticated, "Invalid credentials")
	case goHttp.StatusOK:
		if login.Options != nil {
			if login.Options.GetCookies() {
				// Set the cookie for response
				cookie := goHttp.Cookie{
					Name:   auth_cookie.JWTAuthorizationCookieName,
					Value:  resp.Token,
					Path:   "/",
					MaxAge: 3600,
				}
				if err := grpc.SetHeader(ctx, metadata.Pairs("Set-Cookie", cookie.String())); err != nil {
					logger.Err(err).Warn("Unable to set the cookie")
				}
			}
		}

		return &pbAuthenticationV1.LoginResponse{
			Token: resp.Token,
		}, nil
	}

	return nil, status.Errorf(codes.Internal, "Authentication failed")
}

func (i *implementation) Logout(ctx context.Context, req *pbAuthenticationV1.LogoutRequest) (*pbSharedV1.Empty, error) {
	in, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "incoming context not passed")
	}

	for _, value := range in.Get("grpcgateway-cookie") {
		lines, err := goHttp.ParseCookie(value)
		if err != nil {
			continue
		}

		for _, l := range lines {
			switch l.Name {
			case auth_cookie.JWTAuthorizationCookieName, platformAuthenticationApi.OpenIDJWTSessionID, platformAuthenticationApi.OpenIDJWTRedirect:
				l.MaxAge = -1
				if err := grpc.SetHeader(ctx, metadata.Pairs("Set-Cookie", l.String())); err != nil {
					logger.Err(err).Warn("Unable to set the cookie")
				}
			default:
				continue
			}
		}
	}

	if req.Location != nil {
		if err := grpc.SetHeader(ctx, metadata.Pairs("Location", req.GetLocation())); err != nil {
			logger.Err(err).Warn("Unable to set the cookie")
		}
	}

	return &pbSharedV1.Empty{}, nil
}

func (i *implementation) extractTokenDetails(cache token.Secret, t string) (string, []string, time.Duration, error) {
	// Let's check if token is signed properly
	p, err := cache.Validate(t)
	if err != nil {
		return "", nil, 0, err
	}

	user := DefaultAdminUser
	if v, ok := p.Claims()[token.ClaimPreferredUsername]; ok {
		if s, ok := v.(string); ok {
			user = s
		}
	}

	duration := DefaultTokenMaxTTL

	claims := p.Claims()

	if v, ok := claims[token.ClaimEXP]; ok {
		switch o := v.(type) {
		case int64:
			duration = time.Until(time.Unix(o, 0))
		case float64:
			duration = time.Until(time.Unix(int64(o), 0))
		}
	}

	var roles []string

	if v, ok := claims[token.ClaimRoles]; ok {
		switch o := v.(type) {
		case []string:
			roles = o
		}
	}

	return user, roles, duration, nil
}
