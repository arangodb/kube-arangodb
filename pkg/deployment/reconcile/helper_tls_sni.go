package reconcile

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/tls"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func mapTLSSNIConfig(log zerolog.Logger, sni api.TLSSNISpec, secrets k8sutil.SecretInterface) (map[string]string, error) {
	fetchedSecrets := map[string]string{}

	mapping := sni.Mapping
	if len(mapping) == 0 {
		return fetchedSecrets, nil
	}

	for name, servers := range mapping {
		secret, err := secrets.Get(name, meta.GetOptions{})
		if err != nil {
			return nil, errors.WithMessage(err, "Unable to get SNI secret")
		}

		tlsKey, ok := secret.Data[constants.SecretTLSKeyfile]
		if !ok {
			return nil, errors.Errorf("Not found tls keyfile key in SNI secret")
		}

		tlsKeyChecksum := fmt.Sprintf("%0x", sha256.Sum256(tlsKey))

		for _, server := range servers {
			if _, ok := fetchedSecrets[server]; ok {
				return nil, errors.Errorf("Not found tls key in SNI secret")
			}
			fetchedSecrets[server] = tlsKeyChecksum
		}
	}

	return fetchedSecrets, nil
}

func compareTLSSNIConfig(ctx context.Context, c driver.Connection, m map[string]string, refresh bool) (bool, error) {
	tlsClient := tls.NewClient(c)

	f := tlsClient.GetTLS
	if refresh {
		f = tlsClient.RefreshTLS
	}

	tlsDetails, err := f(ctx)
	if err != nil {
		return false, errors.WithMessage(err, "Unable to fetch TLS SNI state")
	}

	if len(m) != len(tlsDetails.Result.SNI) {
		return false, errors.Errorf("Count of SNI mounted secrets does not match")
	}

	for key, value := range tlsDetails.Result.SNI {
		currentValue, ok := m[key]
		if !ok {
			return false, errors.Errorf("Unable to fetch TLS SNI state")
		}

		if value.Checksum != currentValue {
			return false, nil
		}
	}

	return true, nil
}
