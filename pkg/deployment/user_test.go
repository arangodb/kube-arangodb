package deployment

import (
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "k8s.io/api/core/v1"
)

func TestDeployment_ChangeUserPassword(t *testing.T) {
	// Arrange
	testCases := []struct {
		Name        string
		OldUsername string
		OldPassword string
		NewUsername string
		NewPassword string
		ExpectedErr error
	}{
		{
			Name: "Old secret without credentials",
		},
		{
			Name:        "New secret without credentials",
			OldUsername: "root",
			OldPassword: "test",
		},
		{
			Name:        "Username has been changed",
			OldUsername: "root",
			OldPassword: "test",
			NewUsername: "user",
			NewPassword: "test",
		},
		{
			Name:        "Old and new passwords are the same",
			OldUsername: "root",
			OldPassword: "test",
			NewUsername: "root",
			NewPassword: "test",
		},
	}

	createSecret := func(username, password string) *v1.Secret {
		secret := &v1.Secret{}
		secret.Data = make(map[string][]byte)
		if len(username) > 0 {
			secret.Data[constants.SecretUsername] = []byte(username)
		}

		if len(password) > 0 {
			secret.Data[constants.SecretPassword] = []byte(password)
		}
		return secret
	}

	for _, testCase := range testCases {
		//nolint:scopelint
		t.Run(testCase.Name, func(t *testing.T) {
			// Arrange
			d := Deployment{}
			oldSecret := createSecret(testCase.OldUsername, testCase.OldPassword)
			newSecret := createSecret(testCase.NewUsername, testCase.NewPassword)

			// Act
			err := d.ChangeUserPassword(oldSecret, newSecret)

			// Assert
			if testCase.ExpectedErr != nil {
				assert.Error(t, testCase.ExpectedErr, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
