package v1alpha

import (
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestLicenseSpecValidation(t *testing.T) {
	assert.Nil(t, LicenseSpec{SecretName: nil}.Validate())
	assert.Nil(t, LicenseSpec{SecretName: util.NewString("some-name")}.Validate())

	assert.Error(t, LicenseSpec{SecretName: util.NewString("@@")}.Validate())
}
