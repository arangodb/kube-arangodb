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

package openid

import (
	"context"
	"fmt"
	goHttp "net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"google.golang.org/genproto/googleapis/rpc/status"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/arangodb/shared"
	"github.com/arangodb/go-driver/v2/connection"

	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
	platformAuthenticationApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1/authentication"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
)

func New(configuration pbImplEnvoyAuthV3Shared.Configuration) (pbImplEnvoyAuthV3Shared.AuthHandler, bool) {
	c := cache.NewConfigFile[platformAuthenticationApi.OpenID](configuration.Auth.Path, time.Minute)

	i := &impl{
		fileConfig: c,
		oauth2Config: cache.NewHashedConfiguration(c, func(ctx context.Context, in platformAuthenticationApi.OpenID) (oauth2.Config, error) {
			return in.GetOAuth2Config(ctx)
		}),
		verifier: cache.NewHashedConfiguration(c, func(ctx context.Context, in platformAuthenticationApi.OpenID) (*oidc.IDTokenVerifier, error) {
			return in.GetIDTokenVerifier(ctx)
		}),
		client: cache.NewHashedConfiguration(c, func(ctx context.Context, in platformAuthenticationApi.OpenID) (*goHttp.Client, error) {
			return in.HTTP.Client()
		}),
	}

	i.authClient = cache.NewObject[pbAuthenticationV1.AuthenticationV1Client](configuration.GetAuthClientFetcher)
	i.session = cache.NewRemoteCache[*Session](cache.NewObject(func(ctx context.Context) (arangodb.Collection, time.Duration, error) {
		ac, err := i.authClient.Get(ctx)
		if err != nil {
			return nil, 0, err
		}

		client := arangodb.NewClient(connection.NewHttpConnection(connection.HttpConfiguration{
			Authentication: pbAuthenticationV1.NewRootRequestModifier(ac),
			Endpoint: connection.NewRoundRobinEndpoints([]string{
				fmt.Sprintf("%s://%s:%d", configuration.Database.Proto, configuration.Database.Endpoint, configuration.Database.Port),
			}),
			ContentType:    connection.ApplicationJSON,
			ArangoDBConfig: connection.ArangoDBConfiguration{},
			Transport:      operatorHTTP.RoundTripperWithShortTransport(operatorHTTP.WithTransportTLS(operatorHTTP.Insecure)),
		}))

		db, err := client.GetDatabase(ctx, "_system", nil)
		if err != nil {
			return nil, 0, err
		}

		if _, err := db.GetCollection(ctx, "_gateway_session", nil); err != nil {
			if !shared.IsNotFound(err) {
				return nil, 0, err
			}

			if _, err := db.CreateCollectionWithOptions(ctx, "_gateway_session", &arangodb.CreateCollectionProperties{
				IsSystem: true,
			}, nil); err != nil {
				if !shared.IsConflict(err) {
					return nil, 0, err
				}
			}
		}

		col, err := db.GetCollection(ctx, "_gateway_session", nil)
		if err != nil {
			return nil, 0, err
		}

		return col, 24 * time.Hour, nil
	}))

	i.id = cache.NewCache(func(ctx context.Context, in string) (*oidc.IDToken, error) {
		verifier, err := i.verifier.Get(ctx)
		if err != nil {
			logger.Err(err).Error("Unable to get verifier")
			return nil, err
		}

		return verifier.Verify(ctx, in)
	}, 5*time.Minute)

	return i, true
}

type impl struct {
	fileConfig cache.ConfigFile[platformAuthenticationApi.OpenID]

	oauth2Config cache.HashedConfiguration[oauth2.Config]

	verifier cache.HashedConfiguration[*oidc.IDTokenVerifier]

	client cache.HashedConfiguration[*goHttp.Client]

	id cache.Cache[string, *oidc.IDToken]

	session cache.RemoteCache[*Session]

	authClient cache.Object[pbAuthenticationV1.AuthenticationV1Client]
}

func (i *impl) Handle(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest, current *pbImplEnvoyAuthV3Shared.Response) error {
	if current.Authenticated() {
		return nil
	}

	file, _, err := i.fileConfig.Get(ctx)
	if err != nil {
		logger.Err(err).Error("Unable to get config")
		return err
	}

	cfg, err := i.oauth2Config.Get(ctx)
	if err != nil {
		logger.Err(err).Error("Unable to get config")
		return err
	}

	if client, err := i.client.Get(ctx); err != nil {
		logger.Err(err).Error("Unable to get client")
		return err
	} else if client != nil {
		ctx = oidc.ClientContext(ctx, client)
	}

	requestUrl, err := url.ParseRequestURI(request.GetAttributes().GetRequest().GetHttp().GetPath())
	if err != nil {
		logger.Err(err).Error("Unable to parse request path")
		return err
	}

	if requestUrl.Path == platformAuthenticationApi.OpenIDRedirectURL {
		// We got a response, auth flow initiated

		oauth2Token, err := cfg.Exchange(ctx, requestUrl.Query().Get("code"))
		if err != nil {
			logger.Str("token", requestUrl.Query().Get("code")).Err(err).Error("Request failure")
			return pbImplEnvoyAuthV3Shared.DeniedResponse{
				Code: goHttp.StatusForbidden,
			}
		}

		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			logger.Error("Request `id_token` is not present")
			return pbImplEnvoyAuthV3Shared.DeniedResponse{
				Code: goHttp.StatusForbidden,
			}
		}

		// Parse and verify ID Token payload.
		token, err := i.id.Get(ctx, rawIDToken)
		if err != nil {
			logger.Err(err).Error("Request `id_token` is not able to be verified")
			return pbImplEnvoyAuthV3Shared.DeniedResponse{
				Code: goHttp.StatusForbidden,
			}
		}

		if err := token.VerifyAccessToken(oauth2Token.AccessToken); err != nil {
			logger.Err(err).Error("Unable to verify access token")
			return pbImplEnvoyAuthV3Shared.DeniedResponse{
				Code: goHttp.StatusForbidden,
			}
		}

		resultToken, _, err := jwt.NewParser(jwt.WithIssuedAt(), jwt.WithIssuer(token.Issuer)).ParseUnverified(oauth2Token.AccessToken, jwt.MapClaims{})
		if err != nil {
			logger.Err(err).Error("Unable to parse token")
			return pbImplEnvoyAuthV3Shared.DeniedResponse{
				Code: goHttp.StatusForbidden,
			}
		}

		claims, ok := resultToken.Claims.(jwt.MapClaims)
		if !ok {
			logger.Err(err).Error("Unable to parse token")
			return pbImplEnvoyAuthV3Shared.DeniedResponse{
				Code: goHttp.StatusForbidden,
			}
		}

		u, ok := claims["username"]
		if !ok {
			logger.Err(err).Error("Unable to get username token")
			return pbImplEnvoyAuthV3Shared.DeniedResponse{
				Code: goHttp.StatusForbidden,
			}
		}

		user, ok := u.(string)
		if !ok {
			logger.Err(err).Error("Unable to get username token")
			return pbImplEnvoyAuthV3Shared.DeniedResponse{
				Code: goHttp.StatusForbidden,
			}
		}

		sid := string(uuid.NewUUID())

		logger.JSON("token", token).Str("access", oauth2Token.AccessToken).Str("sid", sid).Info("Token Received, saving session")

		if err := i.session.Put(ctx, util.SHA256FromString(sid), &Session{
			Key:              util.SHA256FromString(sid),
			IDToken:          rawIDToken,
			Username:         user,
			ExpiresAt:        meta.NewTime(token.Expiry),
			ExpiresAtSeconds: token.Expiry.Unix(),
		}); err != nil {
			return err
		}

		cookie := goHttp.Cookie{
			Name:     platformAuthenticationApi.OpenIDJWTSessionID,
			Value:    sid,
			Secure:   true,
			SameSite: goHttp.SameSiteNoneMode,
			Expires:  token.Expiry,
			Path:     "/",
		}

		redirect := "/"

		for _, cookie := range pbImplEnvoyAuthV3Shared.ExtractRequestCookies(request).Filter(func(in *goHttp.Cookie) bool {
			return in.Name == platformAuthenticationApi.OpenIDJWTRedirect
		}).Get() {
			redirect = cookie.Value
		}

		// We are able to get the session, continue
		return pbImplEnvoyAuthV3Shared.NewCustomStaticResponse(&pbEnvoyAuthV3.CheckResponse{
			Status: &status.Status{
				Code: goHttp.StatusUnauthorized,
			},

			HttpResponse: &pbEnvoyAuthV3.CheckResponse_DeniedResponse{
				DeniedResponse: &pbEnvoyAuthV3.DeniedHttpResponse{
					Status: &typev3.HttpStatus{
						Code: typev3.StatusCode_TemporaryRedirect,
					},
					Headers: []*pbEnvoyCoreV3.HeaderValueOption{
						{
							Header: &pbEnvoyCoreV3.HeaderValue{
								Key:   "Location",
								Value: redirect,
							},
						},
						{
							Header: &pbEnvoyCoreV3.HeaderValue{
								Key:   "Set-Cookie",
								Value: cookie.String(),
							},
						},
					},
				},
			},
		})
	} else {
		for _, cookie := range pbImplEnvoyAuthV3Shared.ExtractRequestCookies(request).Filter(func(in *goHttp.Cookie) bool {
			return in.Name == platformAuthenticationApi.OpenIDJWTSessionID
		}).Get() {
			session, ok, err := i.session.Get(ctx, util.SHA256FromString(cookie.Value))
			if err != nil {
				return err
			}

			if !ok || session.ExpiresAt.Time.Before(time.Now()) {
				continue
			}

			current.User = &pbImplEnvoyAuthV3Shared.ResponseAuth{
				User: session.Username,
			}

			return nil
		}
	}

	if file.IsDisabledPath(requestUrl.Path) {
		// Skip Authentication
		return nil
	}

	var headers = []*pbEnvoyCoreV3.HeaderValueOption{
		{
			Header: &pbEnvoyCoreV3.HeaderValue{
				Key:   "Location",
				Value: cfg.AuthCodeURL(""),
			},
		},
		{
			Header: &pbEnvoyCoreV3.HeaderValue{
				Key:   "Access-Control-Allow-Origin",
				Value: request.GetAttributes().GetRequest().GetHttp().GetHeaders()["Origin"],
			},
		},
		{
			Header: &pbEnvoyCoreV3.HeaderValue{
				Key:   "Access-Control-Allow-Credentials",
				Value: "true",
			},
		},
	}

	headers = append(headers, &pbEnvoyCoreV3.HeaderValueOption{
		Header: &pbEnvoyCoreV3.HeaderValue{
			Key: "Set-Cookie",
			Value: (&goHttp.Cookie{
				Name:     platformAuthenticationApi.OpenIDJWTRedirect,
				Value:    request.GetAttributes().GetRequest().GetHttp().GetPath(),
				Secure:   true,
				SameSite: goHttp.SameSiteNoneMode,
				MaxAge:   15,
				Path:     "/",
			}).String(),
		},
	},
	)

	// Cleanup old cookies
	for _, cookie := range pbImplEnvoyAuthV3Shared.ExtractRequestCookies(request).Filter(func(in *goHttp.Cookie) bool {
		return in.Name == platformAuthenticationApi.OpenIDJWTSessionID
	}).Get() {
		cookie.MaxAge = 0
		headers = append(headers, &pbEnvoyCoreV3.HeaderValueOption{
			Header: &pbEnvoyCoreV3.HeaderValue{
				Key:   "Set-Cookie",
				Value: cookie.String(),
			},
		},
		)
	}

	// Redirect
	return pbImplEnvoyAuthV3Shared.NewCustomStaticResponse(&pbEnvoyAuthV3.CheckResponse{
		Status: &status.Status{
			Code: goHttp.StatusUnauthorized,
		},

		HttpResponse: &pbEnvoyAuthV3.CheckResponse_DeniedResponse{
			DeniedResponse: &pbEnvoyAuthV3.DeniedHttpResponse{
				Status: &typev3.HttpStatus{
					Code: typev3.StatusCode_TemporaryRedirect,
				},
				Headers: headers,
			},
		},
	})
}
