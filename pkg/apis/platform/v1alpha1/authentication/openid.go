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

package authentication

import (
	"context"
	"crypto/tls"
	"fmt"
	goHttp "net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	OpenIDJWTRedirect  = "X-ArangoDB-OpenID-Redirect"
	OpenIDJWTSessionID = "X-ArangoDB-OpenID-Session-ID"
	OpenIDRedirectURL  = "/oauth2/idpresponse"
)

func OpenIDDefaultDisabledPaths() []string {
	return []string{
		"/_login",
		"/_logout",
		"/_identity",
		"/_open/auth",
	}
}

type OpenID struct {
	// HTTP defines the HTTP Client Configuration
	HTTP OpenIDHTTPClient `json:"http,omitempty"`

	// Provider defines the OpenID Provider configuration
	Provider OpenIDProvider `json:"provider,omitempty"`

	// Client defines the OpenID Client configuration
	Client OpenIDClient `json:"client,omitempty"`

	// Endpoint defines the OpenID callback Endpoint
	Endpoint string `json:"endpoint,omitempty"`

	// Scope defines OpenID Scopes (OpenID is added by default).
	Scope []string `json:"scope,omitempty"`

	// DisabledPaths keeps the list of SSO disabled paths. By default, "_logout" endpoint is passed through
	DisabledPaths []string `json:"disabledPaths,omitempty"`

	// Features keeps the information about OpenID Features
	Features *OpenIDFeatures `json:"features,omitempty"`

	// Claims keeps the information about OpenID Claims Spec
	Claims *OpenIDClaims `json:"claims,omitempty"`
}

func (c *OpenID) GetDisabledPaths() []string {
	var r []string

	r = append(r, OpenIDDefaultDisabledPaths()...)

	if c != nil {
		r = append(r, c.DisabledPaths...)
	}

	return r
}

func (c *OpenID) IsDisabledPath(path string) bool {
	for _, p := range c.GetDisabledPaths() {
		if p == path {
			return true
		}
	}

	return false
}

func (c *OpenID) GetOAuth2Config(ctx context.Context) (oauth2.Config, error) {
	if c == nil {
		return oauth2.Config{}, errors.Errorf("Config cannot be empty")
	}

	endpoints, err := c.GetEndpoint(ctx)
	if err != nil {
		return oauth2.Config{}, err
	}

	if c.Endpoint == "" {
		return oauth2.Config{}, errors.Errorf("Endpoint cannot be empty")
	}

	return oauth2.Config{
		ClientID:     c.Client.ID,
		ClientSecret: c.Client.Secret,
		Endpoint:     endpoints,
		RedirectURL:  fmt.Sprintf("%s%s", c.Endpoint, OpenIDRedirectURL),
		Scopes:       append([]string{oidc.ScopeOpenID}, c.Scope...),
	}, nil
}

func (c *OpenID) GetIDTokenVerifier(ctx context.Context) (*oidc.IDTokenVerifier, error) {
	if c == nil {
		return nil, errors.Errorf("Provider cannot be empty")
	}

	if client, err := c.HTTP.Client(); err != nil {
		return nil, err
	} else if client != nil {
		ctx = oidc.ClientContext(ctx, client)
	}

	if c.Provider.Issuer == nil {
		return nil, errors.Errorf("Provider Issuer cannot be empty")
	}

	p, err := oidc.NewProvider(ctx, *c.Provider.Issuer)
	if err != nil {
		return nil, err
	}

	return p.Verifier(&oidc.Config{ClientID: c.Client.ID}), nil
}

func (c *OpenID) GetEndpoint(ctx context.Context) (oauth2.Endpoint, error) {
	if c == nil {
		return oauth2.Endpoint{}, errors.Errorf("Provider cannot be empty")
	}

	if client, err := c.HTTP.Client(); err != nil {
		return oauth2.Endpoint{}, err
	} else if client != nil {
		ctx = oidc.ClientContext(ctx, client)
	}

	if c.Provider.Issuer == nil {
		return oauth2.Endpoint{}, errors.Errorf("Provider Issuer cannot be empty")
	}

	if e := c.Provider.ConfigurationProviderEndpoints; e != nil {
		// Lets discover endpoints ourself
		if e.AuthorizationEndpoint == nil {
			return oauth2.Endpoint{}, errors.Errorf("Provider AuthorizationEndpoint cannot be empty if any is provided")
		}
		if e.TokenEndpoint == nil {
			return oauth2.Endpoint{}, errors.Errorf("Provider TokenEndpoint cannot be empty if any is provided")
		}
		if e.UserInfoEndpoint == nil {
			return oauth2.Endpoint{}, errors.Errorf("Provider UserInfoEndpoint cannot be empty if any is provided")
		}

		return oauth2.Endpoint{AuthURL: *e.AuthorizationEndpoint, DeviceAuthURL: *e.AuthorizationEndpoint, TokenURL: *e.TokenEndpoint}, nil
	}

	p, err := oidc.NewProvider(ctx, *c.Provider.Issuer)
	if err != nil {
		return oauth2.Endpoint{}, err
	}

	return p.Endpoint(), nil
}

type OpenIDHTTPClient struct {
	// Insecure defines if insecure HTTP Client is used
	// +doc/default: false
	Insecure *bool `json:"insecure,omitempty"`
}

func (c *OpenIDHTTPClient) Client() (*goHttp.Client, error) {
	var transport goHttp.Transport

	var tls tls.Config

	if c != nil {
		if q := c.Insecure; q != nil {
			tls.InsecureSkipVerify = *c.Insecure
		}
	}

	transport.TLSClientConfig = &tls

	return &goHttp.Client{
		Transport: &transport,
	}, nil
}

type OpenIDProvider struct {
	*ConfigurationProviderEndpoints `json:",omitempty,inline"`

	// Issuer defines OpenID Issuer
	Issuer *string `json:"issuer,omitempty"`
}

type ConfigurationProviderEndpoints struct {
	// AuthorizationEndpoint defines OpenID Authorization Endpoint
	// +doc/link: Documentation|https://www.ibm.com/docs/en/was-liberty/base?topic=connect-openid-endpoint-urls#rwlp_oidc_endpoint_urls__auth_endpoint__title__1
	AuthorizationEndpoint *string `json:"authorizationEndpoint,omitempty"`

	// TokenEndpoint defines OpenID Token Endpoint
	// +doc/link: Documentation|https://www.ibm.com/docs/en/was-liberty/base?topic=connect-openid-endpoint-urls#rwlp_oidc_endpoint_urls__token_endpoint__title__1
	TokenEndpoint *string `json:"tokenEndpoint,omitempty"`

	// UserInfoEndpoint defines OpenID UserInfo Endpoint
	// +doc/link: Documentation|https://www.ibm.com/docs/en/was-liberty/base?topic=connect-openid-endpoint-urls#rwlp_oidc_endpoint_urls__userinfo_endpoint__title__1
	UserInfoEndpoint *string `json:"userInfoEndpoint,omitempty"`
}

type OpenIDClient struct {
	// ID defines OpenID Client ID
	ID string `json:"id,omitempty"`

	// Secret defines OpenID Client Secret
	Secret string `json:"secret,omitempty"`
}

type OpenIDFeatures struct {
	// RefreshEnabled defines if the Refresh OpenID Functionality is enabled
	// +doc/default: false
	// +doc/grade: Alpha
	// +doc/grade: Experimental Feature, in development
	RefreshEnabled *bool `json:"refreshEnabled,omitempty"`
}

func (o *OpenIDFeatures) GetRefreshEnabled() bool {
	if o == nil || o.RefreshEnabled == nil {
		return false
	}

	return *o.RefreshEnabled
}

type OpenIDClaims struct {
	// Username defines the claim key to extract username
	// +doc/default: username
	Username *string `json:"username,omitempty"`
}

func (o *OpenIDClaims) GetUsernameClaim() string {
	if o == nil || o.Username == nil {
		return "username"
	}

	return *o.Username
}
