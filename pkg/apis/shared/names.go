//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package shared

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"unicode"
)

var (
	arangodPrefixes = []string{"CRDN-", "PRMR-", "AGNT-", "SNGL-"}
)

const (
	qualifiedNameMaxLength int = 63
)

// StripArangodPrefix removes well know arangod ID prefixes from the given id.
func StripArangodPrefix(id string) string {
	for _, prefix := range arangodPrefixes {
		if strings.HasPrefix(id, prefix) {
			return id[len(prefix):]
		}
	}
	return id
}

// FixupResourceName ensures that the given name
// complies with kubernetes name requirements.
// If the name is too long or contains invalid characters,
// it will be adjusted and a hash will be added.
func FixupResourceName(name string) string {
	sb := strings.Builder{}
	needHash := len(name) > qualifiedNameMaxLength
	for _, ch := range name {
		if unicode.IsDigit(ch) || unicode.IsLower(ch) || ch == '-' {
			sb.WriteRune(ch)
		} else if unicode.IsUpper(ch) {
			sb.WriteRune(unicode.ToLower(ch))
			needHash = true
		} else {
			needHash = true
		}
	}
	result := sb.String()
	if needHash {
		hash := sha1.Sum([]byte(name))
		h := fmt.Sprintf("-%0x", hash[:3])
		if len(result)+len(h) > qualifiedNameMaxLength {
			result = result[:qualifiedNameMaxLength-(len(h))]
		}
		result = result + h
	}
	return result
}

// CreatePodHostName returns the hostname of the pod for a member with
// a given id in a deployment with a given name.
func CreatePodHostName(deploymentName, role, id string) string {
	suffix := "-" + role + "-" + StripArangodPrefix(id)
	maxDeplNameLen := qualifiedNameMaxLength - len(suffix)
	// shorten deployment name part if resulting name is too big:
	if maxDeplNameLen > 1 && len(deploymentName) > maxDeplNameLen {
		deploymentName = deploymentName[:maxDeplNameLen-1]
	}
	return deploymentName + suffix
}

// CreatePersistentVolumeClaimName returns the name of the persistent volume claim for a member with
// a given id in a deployment with a given name.
func CreatePersistentVolumeClaimName(deploymentName, role, id string) string {
	return deploymentName + "-" + role + "-" + StripArangodPrefix(id)
}

func RenderResourceName(in string, keys map[string]string) string {
	for k, v := range keys {
		in = strings.ReplaceAll(in, fmt.Sprintf("${%s}", k), v)
	}

	return in
}
