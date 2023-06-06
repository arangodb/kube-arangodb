//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package exporter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
	"sync/atomic"
	"time"

	"github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/timer"
)

const (
	monitorMetricTemplate  = "arangodb_member_health{role=\"%s\",id=\"%s\"} %d \n"
	successRefreshInterval = time.Second * 120
	failRefreshInterval    = time.Second * 15
)

var logger = logging.Global().RegisterAndGetLogger("monitor", logging.Info)

var currentMembersStatus atomic.Value

func NewMonitor(arangodbEndpoint string, auth Authentication, sslVerify bool, timeout time.Duration) *monitor {
	uri, err := setPath(arangodbEndpoint, shared.ArangoExporterClusterHealthEndpoint)
	if err != nil {
		logger.Err(err).Error("Fatal")
		os.Exit(1)
	}

	return &monitor{
		factory:   newHttpClientFactory(arangodbEndpoint, auth, sslVerify, timeout),
		healthURI: uri,
	}
}

type monitor struct {
	factory   httpClientFactory
	healthURI *url.URL
}

// UpdateMonitorStatus load monitor metrics for current cluster into currentMembersStatus
func (m monitor) UpdateMonitorStatus(ctx context.Context) {
	for {
		sleep := successRefreshInterval

		health, err := m.GetClusterHealth()
		if err != nil {
			logger.Err(err).Error("GetClusterHealth error")
			sleep = failRefreshInterval
		} else {
			var output strings.Builder
			for key, value := range health.Health {
				entry, err := m.GetMemberStatus(key, value)
				if err != nil {
					logger.Err(err).Error("GetMemberStatus error")
					sleep = failRefreshInterval
				}
				output.WriteString(entry)
			}
			currentMembersStatus.Store(output.String())
		}

		select {
		case <-ctx.Done():
			return
		case <-timer.After(sleep):
			continue
		}
	}
}

// GetClusterHealth returns current ArangoDeployment cluster health status
func (m monitor) GetClusterHealth() (*driver.ClusterHealth, error) {
	c, req, err := m.factory()
	if err != nil {
		return nil, err
	}
	req.URL = m.healthURI
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result driver.ClusterHealth
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, err
}

// GetMemberStatus returns Prometheus monitor metric for specific member
func (m monitor) GetMemberStatus(id driver.ServerID, member driver.ServerHealth) (string, error) {
	result := fmt.Sprintf(monitorMetricTemplate, member.Role, id, 0)

	c, req, err := m.factory()
	if err != nil {
		return result, err
	}

	req.URL, err = setPath(member.Endpoint, shared.ArangoExporterStatusEndpoint)
	if err != nil {
		return result, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return result, err
	}

	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return result, err
		}
		return result, errors.New(string(body))
	}
	return fmt.Sprintf(monitorMetricTemplate, member.Role, id, 1), nil
}

func setPath(uri, uriPath string) (*url.URL, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return u, err
	}
	u.Path = path.Join(uriPath)
	return u, nil
}
