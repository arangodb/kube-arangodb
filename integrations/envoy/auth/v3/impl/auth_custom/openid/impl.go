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

	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	"github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/impl/session"
	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
	platformAuthenticationApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1/authentication"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func New(ctx context.Context, configuration pbImplEnvoyAuthV3Shared.Configuration) (pbImplEnvoyAuthV3Shared.AuthHandler, bool) {
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

	i.authClient = session.NewAuthClientFetcherObject(configuration)
	i.session = session.NewManager[*Session](ctx, "Auth_Custom_OpenID", session.NewConnectionObject(configuration, i.authClient))

	i.id = cache.NewCache(func(ctx context.Context, in string) (*oidc.IDToken, time.Time, error) {
		verifier, err := i.verifier.Get(ctx)
		if err != nil {
			logger.Err(err).Error("Unable to get verifier")
			return nil, util.Default[time.Time](), err
		}

		v, err := verifier.Verify(ctx, in)
		return v, time.Now().Add(5 * time.Minute), err
	})

	return i, true
}

type impl struct {
	fileConfig cache.ConfigFile[platformAuthenticationApi.OpenID]

	oauth2Config cache.HashedConfiguration[oauth2.Config]

	verifier cache.HashedConfiguration[*oidc.IDTokenVerifier]

	client cache.HashedConfiguration[*goHttp.Client]

	id cache.Cache[string, *oidc.IDToken]

	session session.Manager[*Session]

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
		if err := i.handleOpenIDResponse(ctx, request, requestUrl, cfg, file); err != nil {
			return err
		}
	} else {
		if err := i.handleOpenIDAuthentication(ctx, request, current, cfg, file); err != nil {
			return err
		}

		// Authenticated
		if current.Authenticated() {
			return nil
		}
	}

	if file.IsDisabledPath(requestUrl.Path) {
		// Skip Authentication
		return nil
	}

	return i.initOpenIDFlow(request, cfg)
}

func (i *impl) initOpenIDFlow(request *pbEnvoyAuthV3.CheckRequest, cfg oauth2.Config) error {
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
				MaxAge:   60,
				Path:     "/",
			}).String(),
		},
	},
	)

	// Cleanup old cookies
	for _, cookie := range pbImplEnvoyAuthV3Shared.ExtractRequestCookies(request).Filter(func(in *goHttp.Cookie) bool {
		return in.Name == platformAuthenticationApi.OpenIDJWTSessionID
	}).Get() {
		cookie.MaxAge = -1
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

func (i *impl) handleOpenIDResponse(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest, requestUrl *url.URL, cfg oauth2.Config, ocfg platformAuthenticationApi.OpenID) error {
	oauth2Token, err := cfg.Exchange(ctx, requestUrl.Query().Get("code"))
	if err != nil {
		logger.Str("token", requestUrl.Query().Get("code")).Err(err).Error("Request failure")
		return pbImplEnvoyAuthV3Shared.DeniedResponse{
			Code: goHttp.StatusForbidden,
		}
	}

	session, err := i.extractSessionFromToken(ctx, oauth2Token, ocfg)
	if err != nil {
		logger.Err(err).Error("Unable to extract token")
		return pbImplEnvoyAuthV3Shared.DeniedResponse{
			Code: goHttp.StatusForbidden,
		}
	}

	logger.Info("Token Received, saving session")

	cookie := goHttp.Cookie{
		Name:     platformAuthenticationApi.OpenIDJWTSessionID,
		Secure:   true,
		SameSite: goHttp.SameSiteNoneMode,
		Expires:  session.Expires(),
		Path:     "/",
	}

	if key, err := i.session.Put(ctx, session.Expires(), session); err != nil {
		return err
	} else {
		logger.Info("Token Received, saved session")

		cookie.Value = key
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
}

func (i *impl) handleOpenIDAuthentication(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest, current *pbImplEnvoyAuthV3Shared.Response, cfg oauth2.Config, ocfg platformAuthenticationApi.OpenID) error {
	for _, cookie := range pbImplEnvoyAuthV3Shared.ExtractRequestCookies(request).Filter(func(in *goHttp.Cookie) bool {
		return in.Name == platformAuthenticationApi.OpenIDJWTSessionID
	}).Get() {
		session, ok, eol, err := i.session.Get(ctx, cookie.Value)
		if err != nil {
			return err
		}

		if !ok || eol <= 0 {
			continue
		}

		// Do not override authentication
		if current.Authenticated() {
			continue
		}

		if session.Token.Valid() {
			current.User = session.AsResponse()
		}
	}

	for _, cookie := range pbImplEnvoyAuthV3Shared.ExtractRequestCookies(request).Filter(func(in *goHttp.Cookie) bool {
		return in.Name == platformAuthenticationApi.OpenIDJWTSessionID
	}).Get() {
		session, ok, eol, err := i.session.Get(ctx, cookie.Value)
		if err != nil {
			return err
		}

		if !ok || eol <= 0 {
			// Override the cookie
			cookie.MaxAge = -1
			cookie.Path = "/"
			cookie.SameSite = goHttp.SameSiteNoneMode
			cookie.Secure = true

			current.ResponseHeaders = append(current.ResponseHeaders, &pbEnvoyCoreV3.HeaderValueOption{
				Header: &pbEnvoyCoreV3.HeaderValue{
					Key:   "Set-Cookie",
					Value: cookie.String(),
				},
			})

			continue
		}

		if !session.Token.Valid() {
			// Override the cookie
			cookie.MaxAge = -1
			cookie.Path = "/"
			cookie.SameSite = goHttp.SameSiteNoneMode
			cookie.Secure = true

			current.ResponseHeaders = append(current.ResponseHeaders, &pbEnvoyCoreV3.HeaderValueOption{
				Header: &pbEnvoyCoreV3.HeaderValue{
					Key:   "Set-Cookie",
					Value: cookie.String(),
				},
			})

			if !ocfg.Features.GetRefreshEnabled() {
				continue
			}

			// Do not refresh if authenticated
			if current.Authenticated() {
				continue
			}

			session, cookie, err := i.handleOpenIDRefresh(ctx, cfg, ocfg, session)
			if err != nil {
				return err
			}

			current.User = session.AsResponse()

			if cookie != nil {
				current.ResponseHeaders = append(current.ResponseHeaders, &pbEnvoyCoreV3.HeaderValueOption{
					Header: &pbEnvoyCoreV3.HeaderValue{
						Key:   "Set-Cookie",
						Value: cookie.String(),
					},
				})
			}

			continue
		}
	}

	return nil
}

func (i *impl) handleOpenIDRefresh(ctx context.Context, cfg oauth2.Config, ocfg platformAuthenticationApi.OpenID, session *Session) (*Session, *goHttp.Cookie, error) {
	// Check refresh
	token, err := cfg.TokenSource(ctx, &session.Token).Token()
	if err != nil {
		logger.Err(err).Error("Unable to Refresh the token")
		return nil, nil, nil
	}

	cookie := goHttp.Cookie{
		Name:     platformAuthenticationApi.OpenIDJWTSessionID,
		Secure:   true,
		SameSite: goHttp.SameSiteNoneMode,
		Expires:  token.Expiry.Add(24 * time.Hour),
		Path:     "/",
	}

	session, err = i.extractSessionFromToken(ctx, token, ocfg)
	if err != nil {
		logger.Err(err).Error("Unable to extract token")
		return nil, nil, nil
	}

	if key, err := i.session.Put(ctx, token.Expiry.Add(24*time.Hour), session); err != nil {
		return nil, nil, err
	} else {
		logger.Info("Token Received, saved session")

		cookie.Value = key
	}

	return session, &cookie, nil
}

func (i *impl) extractSessionFromToken(ctx context.Context, oauth2Token *oauth2.Token, ocfg platformAuthenticationApi.OpenID) (*Session, error) {
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		logger.Error("Request `id_token` is not present")
		return nil, errors.Errorf("Invalid response format")
	}

	// Parse and verify ID Token payload.
	token, err := i.id.Get(ctx, rawIDToken)
	if err != nil {
		logger.Err(err).Error("Request `id_token` is not able to be verified")
		return nil, err
	}

	if err := token.VerifyAccessToken(oauth2Token.AccessToken); err != nil {
		logger.Err(err).Error("Unable to verify access token")
		return nil, err
	}

	resultToken, _, err := jwt.NewParser(jwt.WithIssuedAt(), jwt.WithIssuer(token.Issuer)).ParseUnverified(oauth2Token.AccessToken, jwt.MapClaims{})
	if err != nil {
		logger.Err(err).Error("Unable to parse token")
		return nil, err
	}

	claims, ok := resultToken.Claims.(jwt.MapClaims)
	if !ok {
		logger.Error("Unable to parse token")
		return nil, errors.Errorf("Unable to parse token")
	}

	u, ok := claims[ocfg.Claims.GetUsernameClaim()]
	if !ok {
		logger.Str("key", ocfg.Claims.GetUsernameClaim()).Error("Unable to find claim in the token")
		return nil, errors.Errorf("Unable to find claim in the token")
	}

	user, ok := u.(string)
	if !ok {
		logger.Error("Unable to parse token username")
		return nil, errors.Errorf("Unable to parse token username")
	}

	var session = Session{
		Token:     *oauth2Token,
		ExpiresAt: meta.NewTime(token.Expiry),
		Username:  user,
	}

	if oauth2Token.RefreshToken != "" && ocfg.Features.GetRefreshEnabled() {
		session.ExpiresAt = meta.NewTime(token.Expiry.Add(24 * time.Hour))
	} else {
		session.Token.RefreshToken = ""
	}

	return &session, nil
}
