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

package shared

import (
	goHttp "net/http"
	"testing"

	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/stretchr/testify/require"
)

func requestWithCookieHeader(value string) *pbEnvoyAuthV3.CheckRequest {
	return &pbEnvoyAuthV3.CheckRequest{
		Attributes: &pbEnvoyAuthV3.AttributeContext{
			Request: &pbEnvoyAuthV3.AttributeContext_Request{
				Http: &pbEnvoyAuthV3.AttributeContext_HttpRequest{
					Headers: map[string]string{
						"cookie": value,
					},
				},
			},
		},
	}
}

func Test_ExtractRequestCookies(t *testing.T) {
	t.Run("Nil request", func(t *testing.T) {
		require.Empty(t, ExtractRequestCookies(nil).Get())
	})

	t.Run("No cookie header", func(t *testing.T) {
		require.Empty(t, ExtractRequestCookies(&pbEnvoyAuthV3.CheckRequest{}).Get())
	})

	t.Run("Single cookie", func(t *testing.T) {
		cookies := ExtractRequestCookies(requestWithCookieHeader("a=b")).Get()
		require.Len(t, cookies, 1)
		require.Equal(t, "a", cookies[0].Name)
		require.Equal(t, "b", cookies[0].Value)
	})

	t.Run("Multiple cookies", func(t *testing.T) {
		cookies := ExtractRequestCookies(requestWithCookieHeader("a=b; c=d")).Get()
		require.Len(t, cookies, 2)
		require.Equal(t, "a", cookies[0].Name)
		require.Equal(t, "c", cookies[1].Name)
	})

	t.Run("Filter by name", func(t *testing.T) {
		cookies := ExtractRequestCookies(requestWithCookieHeader("a=b; c=d")).Filter(func(in *goHttp.Cookie) bool {
			return in.Name == "c"
		}).Get()
		require.Len(t, cookies, 1)
		require.Equal(t, "c", cookies[0].Name)
		require.Equal(t, "d", cookies[0].Value)
	})
}

func Test_filterOutInvalidCookies(t *testing.T) {
	require.False(t, filterOutInvalidCookies(nil))
	require.False(t, filterOutInvalidCookies(&goHttp.Cookie{Name: "\n", Value: "x"}))
	require.True(t, filterOutInvalidCookies(&goHttp.Cookie{Name: "a", Value: "b"}))
}

func Test_FilterCookiesHeader(t *testing.T) {
	t.Run("No cookies", func(t *testing.T) {
		require.Nil(t, FilterCookiesHeader(nil))
	})

	t.Run("Everything filtered out", func(t *testing.T) {
		cookies := []*goHttp.Cookie{{Name: "a", Value: "b"}}
		require.Nil(t, FilterCookiesHeader(cookies, func(cookie *goHttp.Cookie) bool {
			return false
		}))
	})

	t.Run("Keeps matching cookies", func(t *testing.T) {
		cookies := []*goHttp.Cookie{{Name: "a", Value: "b"}, {Name: "c", Value: "d"}}
		r := FilterCookiesHeader(cookies, func(cookie *goHttp.Cookie) bool {
			return cookie.Name == "a"
		})

		// A reset (OVERWRITE_IF_EXISTS with empty value) followed by one append per kept cookie.
		require.Len(t, r, 2)

		require.Equal(t, CookieHeader, r[0].GetHeader().GetKey())
		require.Equal(t, "", r[0].GetHeader().GetValue())
		require.Equal(t, pbEnvoyCoreV3.HeaderValueOption_OVERWRITE_IF_EXISTS, r[0].GetAppendAction())

		require.Equal(t, CookieHeader, r[1].GetHeader().GetKey())
		require.Equal(t, "a=b", r[1].GetHeader().GetValue())
		require.Equal(t, pbEnvoyCoreV3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD, r[1].GetAppendAction())
	})
}
