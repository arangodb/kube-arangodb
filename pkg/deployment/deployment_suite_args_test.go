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

package deployment

import (
	"fmt"
	"sort"
	"testing"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
)

type TestArgs func(t *testing.T) map[string]string

func buildTestArgs(t *testing.T, args ...TestArgs) map[string]string {
	m := map[string]string{}

	for _, arg := range args {
		n := arg(t)
		if len(n) == 0 {
			continue
		}

		for key, value := range n {
			m[key] = value
		}
	}

	return m
}

func testArgsToList(m map[string]string) []string {
	if len(m) == 0 {
		return []string{}
	}

	r := make([]string, 0, len(m))
	for key, value := range m {
		r = append(r, fmt.Sprintf("--%s=%s", key, value))
	}

	sort.Strings(r)

	return r
}

func buildTestArgsList(t *testing.T, args ...TestArgs) []string {
	return testArgsToList(buildTestArgs(t, args...))
}

func BuildTestArgs(t *testing.T, args ...TestArgs) []string {
	z := buildTestArgsList(t, args...)

	return append([]string{resources.ArangoDExecutor}, z...)
}

func agentTestArgs(name string) TestArgs {
	return func(t *testing.T) map[string]string {
		return map[string]string{
			"agency.activate":             "true",
			"agency.disaster-recovery-id": name,
			"agency.size":                 "3",
			"agency.supervision":          "true",
			"database.directory":          "/data",
			"foxx.queues":                 "false",
			"log.level":                   "INFO",
			"log.output":                  "+",
			"server.statistics":           "false",
			"server.storage-engine":       "rocksdb",
		}
	}
}

func BuildTestAgentArgs(t *testing.T, name string, args ...TestArgs) []string {
	return BuildTestArgs(t, append([]TestArgs{agentTestArgs(name)}, args...)...)
}

func ArgsWithAuth(auth bool) TestArgs {
	return func(t *testing.T) map[string]string {
		v := "true"
		if !auth {
			v = "false"
		}

		m := map[string]string{
			"server.authentication": v,
		}

		if auth {
			m["server.jwt-secret-keyfile"] = "/secrets/cluster/jwt/token"
		}

		return m
	}
}

func AgentArgsWithTLS(name string, tls bool) TestArgs {
	return func(t *testing.T) map[string]string {
		p := "ssl"
		if !tls {
			p = "tcp"
		}

		m := map[string]string{
			"agency.my-address": fmt.Sprintf("%s://%s-%s-%s.%s-int.%s.svc:8529", p, testDeploymentName, api.ServerGroupAgentsString, name, testDeploymentName, testNamespace),
			"server.endpoint":   fmt.Sprintf("%s://[::]:8529", p),
		}

		if tls {
			m["ssl.ecdh-curve"] = ""
			m["ssl.keyfile"] = "/secrets/tls/tls.keyfile"
		}

		return m
	}
}

func ArgsWithEncryptionKey() TestArgs {
	return func(t *testing.T) map[string]string {
		return map[string]string{
			"rocksdb.encryption-keyfile": "/secrets/rocksdb/encryption/key",
		}
	}
}

func ArgsWithEncryptionFolder() TestArgs {
	return func(t *testing.T) map[string]string {
		return map[string]string{
			"rocksdb.encryption-keyfolder": "/secrets/rocksdb/encryption",
		}
	}
}
