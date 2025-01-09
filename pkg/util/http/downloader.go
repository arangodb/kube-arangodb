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

package http

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func Download(ctx context.Context, client *http.Client, format string, args ...interface{}) ([]byte, error) {
	if client == nil {
		client = http.DefaultClient
	}

	url := fmt.Sprintf(format, args...)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		if err := resp.Body.Close(); err != nil {
			return nil, err
		}

		return nil, err
	}

	if err := resp.Body.Close(); err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("Unexpected code: %d for `%s`", resp.StatusCode, url)
	}

	return data, nil
}
