package deployment

import (
	"context"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	v1 "k8s.io/api/core/v1"
)

// ChangeUserPassword changes password in the database for a user according to the new secret
func (d *Deployment) ChangeUserPassword(old *v1.Secret, new *v1.Secret) error {

	oldUsername, oldPassword, err := k8sutil.GetSecretAuthCredentials(old)
	if err != nil {
		return nil
	}

	username, password, err := k8sutil.GetSecretAuthCredentials(new)
	if err != nil {
		return nil
	}

	if oldUsername != username {
		// TODO Is it not possible to change username?. What we should do here?
		return nil
	}

	if oldPassword == password {
		// Password has not been changed
		return nil
	}

	// TODO Below when error occurs then passwords in secret and database are different
	//  so maybe we should restore old password in the secret?
	ctx := context.Background()
	client, err := d.clientCache.GetDatabase(ctx)
	if err != nil {
		return maskAny(err)
	}

	user, err := client.User(ctx, username)
	if err != nil {
		if driver.IsNotFound(err) {
			// TODO should we delete secret if there is no user in the database?
			return nil
		}
		return err
	}

	err = user.Update(ctx, driver.UserOptions{
		Password: password,
	})

	return maskAny(err)
}
