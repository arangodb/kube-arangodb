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

package shared

import (
	goHttp "net/http"

	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func FilterCookiesHeader(cookies []*goHttp.Cookie, filter ...func(cookie *goHttp.Cookie) bool) []*pbEnvoyCoreV3.HeaderValueOption {
	if len(cookies) == 0 {
		return nil
	}

	filteredCookies := util.FilterList(cookies, util.MultiFilterList(filter...))

	if len(filteredCookies) == 0 {
		return nil
	}

	cookieStrings := util.FormatList(filteredCookies, func(a *goHttp.Cookie) string {
		return a.String()
	})

	var r []*pbEnvoyCoreV3.HeaderValueOption

	r = append(r, &pbEnvoyCoreV3.HeaderValueOption{
		Header: &pbEnvoyCoreV3.HeaderValue{
			Key: CookieHeader,
		},
		AppendAction:   pbEnvoyCoreV3.HeaderValueOption_OVERWRITE_IF_EXISTS,
		KeepEmptyValue: false,
	})

	for _, v := range cookieStrings {
		r = append(r, &pbEnvoyCoreV3.HeaderValueOption{
			Header: &pbEnvoyCoreV3.HeaderValue{
				Key:   CookieHeader,
				Value: v,
			},
			AppendAction: pbEnvoyCoreV3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD,
		})
	}

	return r
}

func ExtractRequestCookies(request *pbEnvoyAuthV3.CheckRequest) util.Filter[*goHttp.Cookie] {
	rawCookies := request.GetAttributes().GetRequest().GetHttp().GetHeaders()["cookie"]
	// Convert raw cookie string into map of http cookies
	header := goHttp.Header{}
	header.Add("Cookie", rawCookies)
	req := goHttp.Request{Header: header}
	cookies := req.Cookies()

	return util.NewFilter(cookies).Filter(filterOutInvalidCookies)
}

func filterOutInvalidCookies(c *goHttp.Cookie) bool {
	if c == nil {
		return false
	}

	if c.Valid() != nil {
		return false
	}

	return true
}
