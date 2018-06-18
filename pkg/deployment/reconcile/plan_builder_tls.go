//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package reconcile

import (
	"crypto/x509"
	"encoding/pem"
	"net"
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/rs/zerolog"
)

// createRotateTLSServerCertificatePlan creates plan to rotate a server because of an (soon to be) expired TLS certificate.
func createRotateTLSServerCertificatePlan(log zerolog.Logger, spec api.DeploymentSpec, status api.DeploymentStatus,
	getTLSKeyfile func(group api.ServerGroup, member api.MemberStatus) (string, error)) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}
	var plan api.Plan
	status.Members.ForeachServerGroup(func(group api.ServerGroup, members api.MemberStatusList) error {
		for _, m := range members {
			if len(plan) > 0 {
				// Only 1 change at a time
				continue
			}
			if m.Phase != api.MemberPhaseCreated {
				// Only make changes when phase is created
				continue
			}
			if group == api.ServerGroupSyncWorkers {
				// SyncWorkers have no externally created TLS keyfile
				continue
			}
			// Load keyfile
			keyfile, err := getTLSKeyfile(group, m)
			if err != nil {
				log.Warn().Err(err).
					Str("role", group.AsRole()).
					Str("id", m.ID).
					Msg("Failed to get TLS secret")
				continue
			}
			tlsSpec := spec.TLS
			if group.IsArangosync() {
				tlsSpec = spec.Sync.TLS
			}
			renewalNeeded, reason := tlsKeyfileNeedsRenewal(log, keyfile, tlsSpec)
			if renewalNeeded {
				plan = append(append(plan,
					api.NewAction(api.ActionTypeRenewTLSCertificate, group, m.ID, reason)),
					createRotateMemberPlan(log, m, group, "TLS certificate renewal")...,
				)
			}
		}
		return nil
	})
	return plan
}

// createRotateTLSCAPlan creates plan to replace a TLS CA and rotate all server.
func createRotateTLSCAPlan(log zerolog.Logger, spec api.DeploymentSpec, status api.DeploymentStatus,
	getTLSCA func(string) (string, string, bool, error)) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}
	secretName := spec.TLS.GetCASecretName()
	cert, _, isOwned, err := getTLSCA(secretName)
	if err != nil {
		log.Warn().Err(err).Str("secret-name", secretName).Msg("Failed to fetch TLS CA secret")
		return nil
	}
	if !isOwned {
		// TLS CA is not owned by the deployment, we cannot change it
		return nil
	}
	var plan api.Plan
	if renewalNeeded, reason := tlsCANeedsRenewal(log, cert, spec.TLS); renewalNeeded {
		var planSuffix api.Plan
		plan = append(plan,
			api.NewAction(api.ActionTypeRenewTLSCACertificate, 0, "", reason),
		)
		status.Members.ForeachServerGroup(func(group api.ServerGroup, members api.MemberStatusList) error {
			for _, m := range members {
				if m.Phase != api.MemberPhaseCreated {
					// Only make changes when phase is created
					continue
				}
				if !group.IsArangod() {
					// Sync master/worker is not applicable here
					continue
				}
				plan = append(plan,
					api.NewAction(api.ActionTypeRenewTLSCertificate, group, m.ID),
					api.NewAction(api.ActionTypeRotateMember, group, m.ID, "TLS CA certificate changed"),
				)
				planSuffix = append(planSuffix,
					api.NewAction(api.ActionTypeWaitForMemberUp, group, m.ID, "TLS CA certificate changed"),
				)
			}
			return nil
		})
		plan = append(plan, planSuffix...)
	}
	return plan
}

// tlsKeyfileNeedsRenewal decides if the certificate in the given keyfile
// should be renewed.
func tlsKeyfileNeedsRenewal(log zerolog.Logger, keyfile string, spec api.TLSSpec) (bool, string) {
	raw := []byte(keyfile)
	// containsAll returns true when all elements in the expected list
	// are in the actual list.
	containsAll := func(actual []string, expected []string) bool {
		for _, x := range expected {
			found := false
			for _, y := range actual {
				if x == y {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}
	ipsToStringSlice := func(list []net.IP) []string {
		result := make([]string, len(list))
		for i, x := range list {
			result[i] = x.String()
		}
		return result
	}
	for {
		var derBlock *pem.Block
		derBlock, raw = pem.Decode(raw)
		if derBlock == nil {
			break
		}
		if derBlock.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(derBlock.Bytes)
			if err != nil {
				// We do not understand the certificate, let's renew it
				log.Warn().Err(err).Msg("Failed to parse x509 certificate. Renewing it")
				return true, "Cannot parse x509 certificate: " + err.Error()
			}
			if cert.IsCA {
				// Only look at the server certificate, not CA or intermediate
				continue
			}
			// Check expiration date. Renewal at 2/3 of lifetime.
			ttl := cert.NotAfter.Sub(cert.NotBefore)
			expirationDate := cert.NotBefore.Add((ttl / 3) * 2)
			if expirationDate.Before(time.Now()) {
				// We should renew now
				log.Debug().
					Str("not-before", cert.NotBefore.String()).
					Str("not-after", cert.NotAfter.String()).
					Str("expiration-date", expirationDate.String()).
					Msg("TLS certificate renewal needed")
				return true, "Server certificate about to expire"
			}
			// Check alternate names against spec
			dnsNames, ipAddresses, emailAddress, err := spec.GetParsedAltNames()
			if err == nil {
				if !containsAll(cert.DNSNames, dnsNames) {
					return true, "Some alternate DNS names are missing"
				}
				if !containsAll(ipsToStringSlice(cert.IPAddresses), ipAddresses) {
					return true, "Some alternate IP addresses are missing"
				}
				if !containsAll(cert.EmailAddresses, emailAddress) {
					return true, "Some alternate email addresses are missing"
				}
			}
		}
	}
	return false, ""
}

// tlsCANeedsRenewal decides if the given CA certificate
// should be renewed.
// Returns: shouldRenew, reason
func tlsCANeedsRenewal(log zerolog.Logger, cert string, spec api.TLSSpec) (bool, string) {
	raw := []byte(cert)
	for {
		var derBlock *pem.Block
		derBlock, raw = pem.Decode(raw)
		if derBlock == nil {
			break
		}
		if derBlock.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(derBlock.Bytes)
			if err != nil {
				// We do not understand the certificate, let's renew it
				log.Warn().Err(err).Msg("Failed to parse x509 certificate. Renewing it")
				return true, "Cannot parse x509 certificate: " + err.Error()
			}
			if !cert.IsCA {
				// Only look at the CA certificate
				continue
			}
			// Check expiration date. Renewal at 90% of lifetime.
			ttl := cert.NotAfter.Sub(cert.NotBefore)
			expirationDate := cert.NotBefore.Add((ttl / 10) * 9)
			if expirationDate.Before(time.Now()) {
				// We should renew now
				log.Debug().
					Str("not-before", cert.NotBefore.String()).
					Str("not-after", cert.NotAfter.String()).
					Str("expiration-date", expirationDate.String()).
					Msg("TLS CA certificate renewal needed")
				return true, "CA Certificate about to expire"
			}
		}
	}
	return false, ""
}
