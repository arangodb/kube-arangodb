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
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	goHttp "net/http"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func albTestKey(t *testing.T) *ecdsa.PrivateKey {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	return key
}

func albSign(t *testing.T, key *ecdsa.PrivateKey, header map[string]interface{}, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	for k, v := range header {
		token.Header[k] = v
	}
	signed, err := token.SignedString(key)
	require.NoError(t, err)
	return signed
}

func albResolver(key *ecdsa.PrivateKey) ALBKeyResolver {
	return func(ctx context.Context, kid string) (*ecdsa.PublicKey, error) {
		return &key.PublicKey, nil
	}
}

func Test_ALB_VerifyToken_Valid(t *testing.T) {
	key := albTestKey(t)
	c := &ALB{Region: "eu-central-1"}

	token := albSign(t, key, map[string]interface{}{"kid": "abc-123", "signer": "arn:aws:elb"}, jwt.MapClaims{
		"sub": "user-1",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	claims, err := c.VerifyToken(context.Background(), token, albResolver(key))
	require.NoError(t, err)
	require.Equal(t, "user-1", claims["sub"])
}

func Test_ALB_VerifyToken_Expired(t *testing.T) {
	key := albTestKey(t)
	c := &ALB{Region: "eu-central-1"}

	token := albSign(t, key, map[string]interface{}{"kid": "abc-123"}, jwt.MapClaims{
		"sub": "user-1",
		"exp": time.Now().Add(-time.Hour).Unix(),
	})

	_, err := c.VerifyToken(context.Background(), token, albResolver(key))
	require.Error(t, err)
}

func Test_ALB_VerifyToken_WrongKey(t *testing.T) {
	key := albTestKey(t)
	other := albTestKey(t)
	c := &ALB{Region: "eu-central-1"}

	token := albSign(t, key, map[string]interface{}{"kid": "abc-123"}, jwt.MapClaims{
		"sub": "user-1",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	// Resolver returns a public key that does not match the signing key -> signature check fails.
	_, err := c.VerifyToken(context.Background(), token, albResolver(other))
	require.Error(t, err)
}

func Test_ALB_VerifyToken_RejectsNonES256(t *testing.T) {
	key := albTestKey(t)
	c := &ALB{Region: "eu-central-1"}

	// A token signed with HS256 (alg confusion) must be rejected by the ValidMethods restriction; the
	// resolver would never even be consulted for a matching key type.
	hs := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user-1",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	hs.Header["kid"] = "abc-123"
	signed, err := hs.SignedString([]byte("secret"))
	require.NoError(t, err)

	_, err = c.VerifyToken(context.Background(), signed, albResolver(key))
	require.Error(t, err)
}

func Test_ALB_VerifyToken_SignerPin(t *testing.T) {
	key := albTestKey(t)
	c := &ALB{Region: "eu-central-1", Signer: util.NewType("arn:aws:elb:expected")}

	token := albSign(t, key, map[string]interface{}{"kid": "abc-123", "signer": "arn:aws:elb:attacker"}, jwt.MapClaims{
		"sub": "user-1",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	_, err := c.VerifyToken(context.Background(), token, albResolver(key))
	require.Error(t, err)

	// The same token with the expected signer verifies.
	ok := albSign(t, key, map[string]interface{}{"kid": "abc-123", "signer": "arn:aws:elb:expected"}, jwt.MapClaims{
		"sub": "user-1",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	_, err = c.VerifyToken(context.Background(), ok, albResolver(key))
	require.NoError(t, err)
}

func Test_ALB_GetPublicKeyURL(t *testing.T) {
	c := &ALB{Region: "eu-central-1"}

	url, err := c.GetPublicKeyURL("abc-123")
	require.NoError(t, err)
	require.Equal(t, "https://public-keys.auth.elb.eu-central-1.amazonaws.com/abc-123", url)

	// Empty region / key id and injection attempts are rejected.
	_, err = (&ALB{}).GetPublicKeyURL("abc-123")
	require.Error(t, err)
	_, err = c.GetPublicKeyURL("")
	require.Error(t, err)
	_, err = c.GetPublicKeyURL("../../evil")
	require.Error(t, err)
	_, err = (&ALB{Region: "bad/region"}).GetPublicKeyURL("abc-123")
	require.Error(t, err)
}

func Test_ALB_Claims_Defaults(t *testing.T) {
	require.Equal(t, "sub", (*ALBClaims)(nil).GetUsernameClaim())
	require.Equal(t, "", (*ALBClaims)(nil).GetGroupsClaim())
	require.Equal(t, "email", (&ALBClaims{Username: util.NewType("email")}).GetUsernameClaim())
	require.Equal(t, "groups", (&ALBClaims{Groups: util.NewType("groups")}).GetGroupsClaim())
}

func Test_OpenIDHTTPClient_Proxy(t *testing.T) {
	req, _ := goHttp.NewRequest(goHttp.MethodGet, "https://public-keys.auth.elb.eu-central-1.amazonaws.com/kid", nil)

	t.Run("no proxy by default", func(t *testing.T) {
		c, err := (&OpenIDHTTPClient{}).Client()
		require.NoError(t, err)
		require.Nil(t, c.Transport.(*goHttp.Transport).Proxy)
	})

	t.Run("explicit proxy URL", func(t *testing.T) {
		c, err := (&OpenIDHTTPClient{Proxy: util.NewType("http://proxy.local:3128")}).Client()
		require.NoError(t, err)
		u, err := c.Transport.(*goHttp.Transport).Proxy(req)
		require.NoError(t, err)
		require.NotNil(t, u)
		require.Equal(t, "http://proxy.local:3128", u.String())
	})

	t.Run("invalid proxy URL errors", func(t *testing.T) {
		_, err := (&OpenIDHTTPClient{Proxy: util.NewType("://not a url")}).Client()
		require.Error(t, err)
	})
}
