//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

package pod

import (
	"fmt"

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
)

func GroupSNISupported(spec api.DeploymentSpec, group api.ServerGroup) bool {
	switch spec.Mode.Get() {
	case api.DeploymentModeCluster:
		if features.IsGatewayEnabled(spec) {
			return group == api.ServerGroupGateways
		}
		return group == api.ServerGroupCoordinators

	case api.DeploymentModeSingle:
		if features.IsGatewayEnabled(spec) {
			return group == api.ServerGroupGateways
		}
		fallthrough
	case api.DeploymentModeActiveFailover:
		return group == api.ServerGroupSingle
	default:
		return false
	}
}

func SNI() Builder {
	return sni{}
}

type sni struct{}

func (s sni) Envs(i Input) []core.EnvVar {
	return nil
}

func (s sni) isSupported(i Input) bool {
	if !i.Deployment.TLS.IsSecure() {
		return false
	}

	if !features.TLSSNI().ImageSupported(&i.Image) {
		// We need 3.7.0+ and Enterprise to support this
		return false
	}

	return GroupSNISupported(i.Deployment, i.Group)
}

func (s sni) Verify(i Input, cachedStatus interfaces.Inspector) error {
	if !s.isSupported(i) {
		return nil
	}

	for _, secret := range util.SortKeys(i.Deployment.TLS.GetSNI().Mapping) {
		kubeSecret, exists := cachedStatus.Secret().V1().GetSimple(secret)
		if !exists {
			return errors.Errorf("SNI Secret not found %s", secret)
		}

		_, ok := kubeSecret.Data[constants.SecretTLSKeyfile]
		if !ok {
			return errors.Errorf("Unable to find secret key %s/%s for SNI", secret, constants.SecretTLSKeyfile)
		}
	}
	return nil
}

func (s sni) Volumes(i Input) ([]core.Volume, []core.VolumeMount) {
	if !s.isSupported(i) {
		return nil, nil
	}

	sni := i.Deployment.TLS.GetSNI()
	volumes := make([]core.Volume, 0, len(sni.Mapping))
	volumeMounts := make([]core.VolumeMount, 0, len(sni.Mapping))

	for _, secret := range util.SortKeys(sni.Mapping) {
		secretNameSha := util.SHA256FromString(secret)

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
			MountPath: fmt.Sprintf("%s/%s", shared.TLSSNIKeyfileVolumeMountDir, secret),
			ReadOnly:  true,
		}

		volumes = append(volumes, vol)
		volumeMounts = append(volumeMounts, volMount)
	}

	return volumes, volumeMounts
}

func (s sni) Args(i Input) k8sutil.OptionPairs {
	if !s.isSupported(i) {
		return nil
	}

	opts := k8sutil.CreateOptionPairs()

	for _, volume := range util.SortKeys(i.Deployment.TLS.GetSNI().Mapping) {
		servers, ok := i.Deployment.TLS.SNI.Mapping[volume]
		if !ok {
			continue
		}

		for _, server := range servers {
			opts.Addf("--ssl.server-name-indication", "%s=%s/%s/%s", server, shared.TLSSNIKeyfileVolumeMountDir, volume, constants.SecretTLSKeyfile)
		}
	}

	return opts
}
