//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package probes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func TestCreate(t *testing.T) {
	path := "/api/version"
	secret := "the secret"

	// http
	config := HTTPProbeConfig{
		LocalPath:     path,
		Secure:        false,
		Authorization: secret,
	}
	probe := config.Create()

	assert.Equal(t, probe.InitialDelaySeconds, int32(15*60))
	assert.Equal(t, probe.TimeoutSeconds, int32(2))
	assert.Equal(t, probe.PeriodSeconds, int32(60))
	assert.Equal(t, probe.SuccessThreshold, int32(1))
	assert.Equal(t, probe.FailureThreshold, int32(10))

	assert.Equal(t, probe.ProbeHandler.HTTPGet.Path, path)
	assert.Equal(t, probe.ProbeHandler.HTTPGet.HTTPHeaders[0].Name, "Authorization")
	assert.Equal(t, probe.ProbeHandler.HTTPGet.HTTPHeaders[0].Value, secret)
	assert.Equal(t, probe.ProbeHandler.HTTPGet.Port.String(), shared.ServerPortName)
	assert.Equal(t, probe.ProbeHandler.HTTPGet.Scheme, core.URISchemeHTTP)

	// https
	config = HTTPProbeConfig{
		LocalPath:     path,
		Secure:        true,
		Authorization: secret,
	}
	probe = config.Create()

	assert.Equal(t, probe.ProbeHandler.HTTPGet.Scheme, core.URISchemeHTTPS)

	// http, custom timing
	config = HTTPProbeConfig{
		LocalPath:     path,
		Secure:        true,
		Authorization: secret,
		Common:        Common{util.NewType[int32](1), util.NewType[int32](2), util.NewType[int32](3), util.NewType[int32](4), util.NewType[int32](5)},
	}
	probe = config.Create()

	assert.Equal(t, probe.InitialDelaySeconds, int32(1))
	assert.Equal(t, probe.TimeoutSeconds, int32(2))
	assert.Equal(t, probe.PeriodSeconds, int32(3))
	assert.Equal(t, probe.SuccessThreshold, int32(4))
	assert.Equal(t, probe.FailureThreshold, int32(5))
}
