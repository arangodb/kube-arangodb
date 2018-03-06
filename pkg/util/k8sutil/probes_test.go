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
	config := HTTPProbeConfig{path, false, secret}
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
	config = HTTPProbeConfig{path, true, secret}
	probe = config.Create()

	assert.Equal(t, probe.Handler.HTTPGet.Scheme, v1.URISchemeHTTPS)
}
