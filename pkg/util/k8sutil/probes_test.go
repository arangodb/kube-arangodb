//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Jan Christoph Uhde
//

package k8sutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
)

func TestCreate(t *testing.T) {
	path := "/api/version"
	secret := "the secret"

	// http
	config := HTTPProbeConfig{path, false, secret, 0}
	probe := config.Create()

	assert.Equal(t, probe.InitialDelaySeconds, int32(30))
	assert.Equal(t, probe.TimeoutSeconds, int32(2))
	assert.Equal(t, probe.PeriodSeconds, int32(10))
	assert.Equal(t, probe.SuccessThreshold, int32(1))
	assert.Equal(t, probe.FailureThreshold, int32(3))

	assert.Equal(t, probe.Handler.HTTPGet.Path, path)
	assert.Equal(t, probe.Handler.HTTPGet.HTTPHeaders[0].Name, "Authorization")
	assert.Equal(t, probe.Handler.HTTPGet.HTTPHeaders[0].Value, secret)
	assert.Equal(t, probe.Handler.HTTPGet.Port.IntValue(), 8529)
	assert.Equal(t, probe.Handler.HTTPGet.Scheme, v1.URISchemeHTTP)

	// https
	config = HTTPProbeConfig{path, true, secret, 0}
	probe = config.Create()

	assert.Equal(t, probe.Handler.HTTPGet.Scheme, v1.URISchemeHTTPS)
}
