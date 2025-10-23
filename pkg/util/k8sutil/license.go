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

package k8sutil

import (
	"encoding/base64"
	"encoding/json"

	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret"
)

type License string

func (l License) IsV2Set() bool {
	return l != ""
}

func (l License) V2Hash() string {
	return util.SHA256FromString(string(l))
}

type LicenseSecret struct {
	V1     string
	V2     License
	Master *LicenseSecretMaster
}

type LicenseSecretMaster struct {
	ClientID     string
	ClientSecret string
}

func GetLicenseFromSecret(secret secret.Inspector, name string) (LicenseSecret, error) {
	s, ok := secret.Secret().V1().GetSimple(name)
	if !ok {
		return LicenseSecret{}, errors.Errorf("Secret %s not found", name)
	}

	var l LicenseSecret

	if v, ok := s.Data[utilConstants.SecretKeyToken]; ok {
		l.V1 = string(v)
	}

	if cid, ok := s.Data[utilConstants.SecretKeyMasterClientID]; ok {
		if cs, ok := s.Data[utilConstants.SecretKeyMasterClientSecret]; ok {
			l.Master = &LicenseSecretMaster{
				ClientID:     string(cid),
				ClientSecret: string(cs),
			}
		}
	}

	if v1, ok1 := s.Data[utilConstants.SecretKeyV2License]; ok1 {
		// some customers put the raw JSON-encoded value, but operator and DB servers expect the base64-encoded value
		if IsJSON(v1) {
			l.V2 = License(base64.StdEncoding.EncodeToString(v1))
		} else {
			l.V2 = License(v1)
		}
	} else if v2, ok2 := s.Data[utilConstants.SecretKeyV2Token]; ok2 {
		// some customers put the raw JSON-encoded value, but operator and DB servers expect the base64-encoded value
		if IsJSON(v2) {
			l.V2 = License(base64.StdEncoding.EncodeToString(v2))
		} else {
			l.V2 = License(v2)
		}
	} else if l.Master == nil {
		return LicenseSecret{}, errors.Errorf("Key (%s, %s, %s, or %s+%s) is missing in the license secret (%s)",
			utilConstants.SecretKeyToken, utilConstants.SecretKeyV2License, utilConstants.SecretKeyV2Token, utilConstants.SecretKeyMasterClientID, utilConstants.SecretKeyMasterClientSecret, name)
	}

	return l, nil
}

func IsJSON(s []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(s, &js) == nil
}
