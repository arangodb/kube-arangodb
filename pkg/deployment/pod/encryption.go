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
	"path/filepath"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GroupEncryptionSupported(mode api.DeploymentMode, group api.ServerGroup) bool {
	switch mode {
	case api.DeploymentModeCluster:
		switch group {
		case api.ServerGroupDBServers, api.ServerGroupAgents:
			return true
		default:
			return false
		}

	case api.DeploymentModeSingle:
		return group == api.ServerGroupSingle

	case api.DeploymentModeActiveFailover:
		switch group {
		case api.ServerGroupSingle, api.ServerGroupAgents:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func GetEncryptionKey(secrets k8sutil.SecretInterface, name string) (string, []byte, bool, error) {
	keyfile, err := secrets.Get(name, meta.GetOptions{})
	if err != nil {
		if k8sutil.IsNotFound(err) {
			return "", nil, false, nil
		}
		return "", nil, false, errors.Wrapf(err, "Unable to fetch secret")
	}

	sha, data, err := GetEncryptionKeyFromSecret(keyfile)

	return sha, data, true, err
}

func GetEncryptionKeyFromSecret(keyfile *core.Secret) (string, []byte, error) {
	if len(keyfile.Data) == 0 {
		return "", nil, nil
	}

	d, ok := keyfile.Data[constants.SecretEncryptionKey]
	if !ok {
		return "", nil, nil
	}

	if len(d) != 32 {
		return "", nil, errors.Errorf("Current encryption key is not valid")
	}

	sha := fmt.Sprintf("%0x", sha256.Sum256(d))

	return sha, d, nil
}

func GetKeyfolderSecretName(name string) string {
	n := fmt.Sprintf("%s-encryption-folder", name)

	return n
}

func IsEncryptionEnabled(i Input) bool {
	return i.Deployment.RocksDB.IsEncrypted()
}

func MultiFileMode(i Input) bool {
	return i.Enterprise && i.Version.CompareTo("3.7.0") >= 0
}

func Encryption() Builder {
	return encryption{}
}

type encryption struct{}

func (e encryption) Args(i Input) k8sutil.OptionPairs {
	if !IsEncryptionEnabled(i) {
		return nil
	}
	if !MultiFileMode(i) {
		keyPath := filepath.Join(k8sutil.RocksDBEncryptionVolumeMountDir, constants.SecretEncryptionKey)
		return k8sutil.NewOptionPair(k8sutil.OptionPair{
			Key:   "--rocksdb.encryption-keyfile",
			Value: keyPath,
		})
	} else {
		return k8sutil.NewOptionPair(k8sutil.OptionPair{
			Key:   "--rocksdb.encryption-keyfolder",
			Value: k8sutil.RocksDBEncryptionVolumeMountDir,
		})
	}
}

func (e encryption) Volumes(i Input) ([]core.Volume, []core.VolumeMount) {
	if !IsEncryptionEnabled(i) {
		return nil, nil
	}
	if !MultiFileMode(i) {
		vol := k8sutil.CreateVolumeWithSecret(k8sutil.RocksdbEncryptionVolumeName, i.Deployment.RocksDB.Encryption.GetKeySecretName())
		return []core.Volume{vol}, []core.VolumeMount{k8sutil.RocksdbEncryptionVolumeMount()}
	} else {
		vol := k8sutil.CreateVolumeWithSecret(k8sutil.RocksdbEncryptionVolumeName, GetKeyfolderSecretName(i.ApiObject.GetName()))
		return []core.Volume{vol}, []core.VolumeMount{k8sutil.RocksdbEncryptionReadOnlyVolumeMount()}
	}
}

func (e encryption) Verify(i Input, cachedStatus inspector.Inspector) error {
	if !IsEncryptionEnabled(i) {
		return nil
	}

	secret, exists := cachedStatus.Secret(i.Deployment.RocksDB.Encryption.GetKeySecretName())
	if !exists {
		return errors.Errorf("Encryption key secret does not exist %s", i.Deployment.RocksDB.Encryption.GetKeySecretName())
	}

	if !MultiFileMode(i) {
		if err := k8sutil.ValidateEncryptionKeyFromSecret(secret); err != nil {
			return errors.Wrapf(err, "RocksDB encryption key secret validation failed")
		}
		return nil
	} else {
		return nil
	}
}
