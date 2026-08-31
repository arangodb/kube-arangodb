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

package authentication

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io"
	goHttp "net/http"
	"regexp"
	goStrings "strings"

	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	// ALBDataHeader is the header AWS Application Load Balancer injects after it terminates OIDC. It
	// carries a JWT signed by the ALB with the authenticated user claims.
	ALBDataHeader = "x-amzn-oidc-data"

	// ALBSigningMethod is the only JWT signing method AWS ALB uses for the x-amzn-oidc-data token.
	ALBSigningMethod = "ES256"
)

// albRegionRegexp and albKeyIDRegexp constrain the values that are interpolated into the public key
// URL. Region comes from the trusted operator config; the key id is taken from the attacker-supplied
// token header, so it is validated to a strict charset to keep it a single, harmless path segment.
var (
	albRegionRegexp = regexp.MustCompile(`^[a-z0-9-]+$`)
	albKeyIDRegexp  = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
)

// ALB configures how the gateway trusts identity forwarded by an AWS Application Load Balancer that
// terminates OIDC in front of it. The ALB signs the x-amzn-oidc-data JWT with a per-region key, so
// the gateway verifies that signature rather than performing the OIDC flow itself.
type ALB struct {
	// HTTP defines the HTTP Client Configuration used to fetch the ALB signing keys.
	HTTP OpenIDHTTPClient `json:"http,omitempty"`

	// Region defines the AWS Region of the Application Load Balancer. It is used to resolve the
	// public key endpoint `public-keys.auth.elb.<region>.amazonaws.com`, or, for AWS GovCloud (US)
	// regions (`us-gov-*`), the S3-hosted endpoint `s3-<region>.amazonaws.com/aws-elb-public-keys-prod-<region>`.
	Region string `json:"region,omitempty"`

	// Signer optionally pins the expected ALB ARN carried in the token `signer` header. When set,
	// tokens signed by any other load balancer are rejected.
	Signer *string `json:"signer,omitempty"`

	// Claims keeps the information about ALB Claims Spec.
	Claims *ALBClaims `json:"claims,omitempty"`
}

// GetPublicKeyURL returns the AWS endpoint that serves the PEM-encoded public key for the given key
// id. The key id originates from the untrusted token header, so it is validated before use.
func (c *ALB) GetPublicKeyURL(kid string) (string, error) {
	if c == nil || c.Region == "" {
		return "", errors.Errorf("Region cannot be empty")
	}

	if !albRegionRegexp.MatchString(c.Region) {
		return "", errors.Errorf("Invalid Region `%s`", c.Region)
	}

	if kid == "" {
		return "", errors.Errorf("Key ID cannot be empty")
	}

	if !albKeyIDRegexp.MatchString(kid) {
		return "", errors.Errorf("Invalid Key ID")
	}

	// AWS GovCloud (US) does not serve the ALB signing keys from the public-keys.auth.elb host; they
	// are hosted in a per-region S3 bucket instead. GovCloud regions are prefixed `us-gov-`.
	if goStrings.HasPrefix(c.Region, "us-gov-") {
		return fmt.Sprintf("https://s3-%s.amazonaws.com/aws-elb-public-keys-prod-%s/%s", c.Region, c.Region, kid), nil
	}

	return fmt.Sprintf("https://public-keys.auth.elb.%s.amazonaws.com/%s", c.Region, kid), nil
}

// ALBKeyResolver resolves the ALB signing public key for the given key id (`kid` token header).
type ALBKeyResolver func(ctx context.Context, kid string) (*ecdsa.PublicKey, error)

// KeyResolver returns a resolver that fetches and parses the ALB signing key over HTTP.
func (c *ALB) KeyResolver() (ALBKeyResolver, error) {
	client, err := c.HTTP.Client()
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, kid string) (*ecdsa.PublicKey, error) {
		url, err := c.GetPublicKeyURL(kid)
		if err != nil {
			return nil, err
		}

		req, err := goHttp.NewRequestWithContext(ctx, goHttp.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != goHttp.StatusOK {
			return nil, errors.Errorf("Unable to fetch ALB public key: status %d", resp.StatusCode)
		}

		data, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
		if err != nil {
			return nil, err
		}

		return jwt.ParseECPublicKeyFromPEM(data)
	}, nil
}

// VerifyToken verifies an AWS ALB x-amzn-oidc-data token using the provided key resolver and returns
// its claims. The token signature, signing method and expiry are all enforced; when Signer is set the
// token `signer` header must match. The resolver is only reached for a token whose signer is accepted.
func (c *ALB) VerifyToken(ctx context.Context, token string, resolver ALBKeyResolver) (jwt.MapClaims, error) {
	if c == nil {
		return nil, errors.Errorf("Config cannot be empty")
	}

	claims := jwt.MapClaims{}

	if _, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		signer, _ := t.Header["signer"].(string)
		if c.Signer != nil && *c.Signer != signer {
			return nil, errors.Errorf("Unexpected signer `%s`", signer)
		}

		kid, ok := t.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, errors.Errorf("Missing key id")
		}

		return resolver(ctx, kid)
		// AWS ALB base64url-encodes the x-amzn-oidc-data segments WITH padding, which the default
		// RawURLEncoding rejects ("illegal base64 data"); WithPaddingAllowed accepts the padding.
	}, jwt.WithValidMethods([]string{ALBSigningMethod}), jwt.WithPaddingAllowed()); err != nil {
		return nil, err
	}

	return claims, nil
}

// ALBClaims maps the ALB token claims onto the ArangoDB identity.
type ALBClaims struct {
	// Username defines the claim key to extract the username from.
	// +doc/default: sub
	Username *string `json:"username,omitempty"`

	// Groups defines the claim key to extract the groups from. When empty, no groups are extracted.
	Groups *string `json:"groups,omitempty"`
}

func (o *ALBClaims) GetUsernameClaim() string {
	if o == nil || o.Username == nil {
		return "sub"
	}

	return *o.Username
}

func (o *ALBClaims) GetGroupsClaim() string {
	if o == nil || o.Groups == nil {
		return ""
	}

	return *o.Groups
}
