//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package pod

import (
	"crypto/sha256"
	"fmt"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SNI() Builder {
	return sni{}
}

type sni struct{}

func (s sni) Verify(i Input, secrets k8sutil.SecretInterface) error {
	if !i.Deployment.TLS.IsSecure() {
		return nil
	}

	for _, secret := range util.SortKeys(i.Deployment.TLS.GetTLSSNISpec().Mapping) {
		kubeSecret, err := secrets.Get(secret, meta.GetOptions{})
		if err != nil {
			return err
		}

		_, ok := kubeSecret.Data[constants.SecretTLSKeyfile]
		if !ok {
			return errors.Errorf("Unable to find secret key %s/%s for SNI", secret, constants.SecretTLSKeyfile)
		}
	}
	return nil
}

func (s sni) Volumes(i Input) ([]core.Volume, []core.VolumeMount) {
	sni := i.Deployment.TLS.GetTLSSNISpec()
	volumes := make([]core.Volume, 0, len(sni.Mapping))
	volumeMounts := make([]core.VolumeMount, 0, len(sni.Mapping))

	if i.Deployment.TLS.IsSecure() {
		for _, secret := range util.SortKeys(sni.Mapping) {
			secretNameSha := fmt.Sprintf("%0x", sha256.Sum256([]byte(secret)))

			secretNameSha = fmt.Sprintf("sni-%s", secretNameSha[:48])

			vol := core.Volume{
				Name: secretNameSha,
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: secret,
					},
				},
			}

			volMount := core.VolumeMount{
				Name:      secretNameSha,
				MountPath: fmt.Sprintf("%s/%s", k8sutil.TLSSNIKeyfileVolumeMountDir, secret),
				ReadOnly:  true,
			}

			volumes = append(volumes, vol)
			volumeMounts = append(volumeMounts, volMount)
		}
	}

	return volumes, volumeMounts
}

func (s sni) Args(i Input) k8sutil.OptionPairs {
	if !i.Deployment.TLS.IsSecure() {
		return nil
	}

	if i.Version.CompareTo("3.7.0") < 0 || !i.Enterprise {
		return nil
	}

	opts := k8sutil.CreateOptionPairs()

	for _, volume := range util.SortKeys(i.Deployment.TLS.GetTLSSNISpec().Mapping) {
		servers, ok := i.Deployment.TLS.SNI.Mapping[volume]
		if !ok {
			continue
		}

		for _, server := range servers {
			opts.Addf("--ssl.server-name-indication", "%s=%s/%s/%s", server, k8sutil.TLSSNIKeyfileVolumeMountDir, volume, constants.SecretTLSKeyfile)
		}
	}

	return opts
}
