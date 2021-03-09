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
	"path/filepath"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"

	"github.com/arangodb/kube-arangodb/pkg/deployment/features"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	core "k8s.io/api/core/v1"
)

func IsRuntimeTLSKeyfileUpdateSupported(i Input) bool {
	return IsTLSEnabled(i) && features.TLSRotation().Supported(i.Version, i.Enterprise) &&
		i.Deployment.TLS.Mode.Get() == api.TLSRotateModeInPlace
}

func IsTLSEnabled(i Input) bool {
	return i.Deployment.TLS.IsSecure()
}

func GetTLSKeyfileSecretName(i Input) string {
	return k8sutil.AppendTLSKeyfileSecretPostfix(i.ArangoMember.GetName())
}

func TLS() Builder {
	return tls{}
}

type tls struct{}

func (s tls) Envs(i Input) []core.EnvVar {
	return nil
}

func (s tls) Verify(i Input, cachedStatus interfaces.Inspector) error {
	if !IsTLSEnabled(i) {
		return nil
	}

	return nil
}

func (s tls) Volumes(i Input) ([]core.Volume, []core.VolumeMount) {
	if !IsTLSEnabled(i) {
		return nil, nil
	}

	return []core.Volume{k8sutil.CreateVolumeWithSecret(k8sutil.TlsKeyfileVolumeName, GetTLSKeyfileSecretName(i))},
		[]core.VolumeMount{k8sutil.TlsKeyfileVolumeMount()}
}

func (s tls) Args(i Input) k8sutil.OptionPairs {
	if !IsTLSEnabled(i) {
		return nil
	}

	opts := k8sutil.CreateOptionPairs()

	keyPath := filepath.Join(k8sutil.TLSKeyfileVolumeMountDir, constants.SecretTLSKeyfile)
	opts.Add("--ssl.keyfile", keyPath)
	opts.Add("--ssl.ecdh-curve", "") // This way arangod accepts curves other than P256 as well.

	return opts
}
