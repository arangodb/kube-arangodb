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

package reconcile

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"

	"github.com/arangodb-helper/go-certificates"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/rs/zerolog"
)

const CertificateRenewalMargin = 7 * 24 * time.Hour

type Certificates []*x509.Certificate

func (c Certificates) Contains(cert *x509.Certificate) bool {
	for _, localCert := range c {
		if !localCert.Equal(cert) {
			return false
		}
	}

	return true
}

func (c Certificates) ContainsAll(certs Certificates) bool {
	if len(certs) == 0 {
		return true
	}

	for _, cert := range certs {
		if !c.Contains(cert) {
			return false
		}
	}

	return true
}

func (c Certificates) ToPem() ([]byte, error) {
	bytes := bytes.NewBuffer([]byte{})

	for _, cert := range c {
		if err := pem.Encode(bytes, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}); err != nil {
			return nil, err
		}
	}

	return bytes.Bytes(), nil
}

func (c Certificates) AsCertPool() *x509.CertPool {
	cp := x509.NewCertPool()

	for _, cert := range c {
		cp.AddCert(cert)
	}

	return cp
}

func getCertsFromData(log zerolog.Logger, caPem []byte) Certificates {
	certs := make([]*x509.Certificate, 0, 2)

	for {
		pem, rest := pem.Decode(caPem)
		if pem == nil {
			break
		}

		caPem = rest

		cert, err := x509.ParseCertificate(pem.Bytes)
		if err != nil {
			// This error should be ignored
			log.Error().Err(err).Msg("Unable to parse certificate")
			continue
		}

		certs = append(certs, cert)
	}

	return certs
}

func getCertsFromSecret(log zerolog.Logger, secret *core.Secret) Certificates {
	caPem, exists := secret.Data[core.ServiceAccountRootCAKey]
	if !exists {
		return nil
	}

	return getCertsFromData(log, caPem)
}

func getKeyCertFromCache(log zerolog.Logger, cachedStatus inspector.Inspector, spec api.DeploymentSpec, certName, keyName string) (Certificates, interface{}, error) {
	caSecret, exists := cachedStatus.Secret(spec.TLS.GetCASecretName())
	if !exists {
		return nil, nil, errors.Errorf("CA Secret does not exists")
	}

	return getKeyCertFromSecret(log, caSecret, keyName, certName)
}

func getKeyCertFromSecret(log zerolog.Logger, secret *core.Secret, certName, keyName string) (Certificates, interface{}, error) {
	ca, exists := secret.Data[certName]
	if !exists {
		return nil, nil, errors.Errorf("Key %s missing in secret", certName)
	}

	key, exists := secret.Data[keyName]
	if !exists {
		return nil, nil, errors.Errorf("Key %s missing in secret", keyName)
	}

	cert, keys, err := certificates.LoadFromPEM(string(ca), string(key))
	if err != nil {
		return nil, nil, err
	}

	return cert, keys, nil
}

// createTLSStatusUpdate creates plan to update ca info
func createTLSStatusUpdate(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspector.Inspector, context PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	if createTLSStatusUpdateRequired(ctx, log, apiObject, spec, status, cachedStatus, context) {
		return api.Plan{api.NewAction(api.ActionTypeTLSKeyStatusUpdate, api.ServerGroupUnknown, "", "Update status")}
	}

	return nil
}

func createTLSStatusUpdateRequired(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspector.Inspector, context PlanBuilderContext) bool {
	if !spec.TLS.IsSecure() {
		return false
	}

	trusted, exists := cachedStatus.Secret(resources.GetCASecretName(apiObject))
	if !exists {
		log.Warn().Str("secret", resources.GetCASecretName(apiObject)).Msg("Folder with secrets does not exist")
		return false
	}

	keyHashes := secretKeysToListWithPrefix("sha256:", trusted)

	if len(keyHashes) == 0 {
		return false
	}

	if len(keyHashes) == 1 {
		if status.Hashes.TLS.CA == nil {
			return true
		}

		if *status.Hashes.TLS.CA != keyHashes[0] {
			return true
		}
	}

	if !util.CompareStringArray(status.Hashes.TLS.Truststore, keyHashes) {
		return true
	}

	return false
}

// createCAAppendPlan creates plan to append CA
func createCAAppendPlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspector.Inspector, context PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	caSecret, exists := cachedStatus.Secret(spec.TLS.GetCASecretName())
	if !exists {
		log.Warn().Str("secret", spec.TLS.GetCASecretName()).Msg("CA Secret does not exists")
		return nil
	}

	ca, _, err := getKeyCertFromSecret(log, caSecret, resources.CACertName, resources.CAKeyName)
	if err != nil {
		log.Warn().Err(err).Str("secret", spec.TLS.GetCASecretName()).Msg("CA Secret does not contains Cert")
		return nil
	}

	if len(ca) == 0 {
		log.Warn().Str("secret", spec.TLS.GetCASecretName()).Msg("CA does not contain any certs")
		return nil
	}

	trusted, exists := cachedStatus.Secret(resources.GetCASecretName(apiObject))
	if !exists {
		log.Warn().Str("secret", resources.GetCASecretName(apiObject)).Msg("Folder with secrets does not exist")
		return nil
	}

	caData, err := ca.ToPem()
	if err != nil {
		log.Warn().Err(err).Str("secret", spec.TLS.GetCASecretName()).Msg("Unable to parse cert")
		return nil
	}

	certSha := util.SHA256(caData)

	if _, exists := trusted.Data[certSha]; !exists {
		return api.Plan{api.NewAction(api.ActionTypeAppendTLSCACertificate, api.ServerGroupUnknown, "", "Append CA to truststore").
			AddParam(actionTypeAppendTLSCACertificateChecksum, certSha)}
	}

	return nil
}

// createCARenewalPlan creates plan to renew CA
func createCARenewalPlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspector.Inspector, context PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	caSecret, exists := cachedStatus.Secret(spec.TLS.GetCASecretName())
	if !exists {
		log.Warn().Str("secret", spec.TLS.GetCASecretName()).Msg("CA Secret does not exists")
		return nil
	}

	if !k8sutil.IsOwner(apiObject.AsOwner(), caSecret) {
		log.Warn().Str("secret", spec.TLS.GetCASecretName()).Msg("CA Secret is not owned by Operator, we wont do anything")
		return nil
	}

	cas, _, err := getKeyCertFromSecret(log, caSecret, resources.CACertName, resources.CAKeyName)
	if err != nil {
		log.Warn().Err(err).Str("secret", spec.TLS.GetCASecretName()).Msg("CA Secret does not contains Cert")
		return nil
	}

	for _, ca := range cas {
		if time.Now().Add(CertificateRenewalMargin).After(ca.NotAfter) {
			// CA will expire soon, renewal needed
			return api.Plan{api.NewAction(api.ActionTypeRenewTLSCACertificate, api.ServerGroupUnknown, "", "Renew CA Certificate")}
		}
	}

	return nil
}

// createCACleanPlan creates plan to remove old CA's
func createCACleanPlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspector.Inspector, context PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	caSecret, exists := cachedStatus.Secret(spec.TLS.GetCASecretName())
	if !exists {
		log.Warn().Str("secret", spec.TLS.GetCASecretName()).Msg("CA Secret does not exists")
		return nil
	}

	ca, _, err := getKeyCertFromSecret(log, caSecret, resources.CACertName, resources.CAKeyName)
	if err != nil {
		log.Warn().Err(err).Str("secret", spec.TLS.GetCASecretName()).Msg("CA Secret does not contains Cert")
		return nil
	}

	if len(ca) == 0 {
		log.Warn().Str("secret", spec.TLS.GetCASecretName()).Msg("CA does not contain any certs")
		return nil
	}

	trusted, exists := cachedStatus.Secret(resources.GetCASecretName(apiObject))
	if !exists {
		log.Warn().Str("secret", resources.GetCASecretName(apiObject)).Msg("Folder with secrets does not exist")
		return nil
	}

	caData, err := ca.ToPem()
	if err != nil {
		log.Warn().Err(err).Str("secret", spec.TLS.GetCASecretName()).Msg("Unable to parse cert")
		return nil
	}

	certSha := util.SHA256(caData)

	for sha := range trusted.Data {
		if certSha != sha {
			return api.Plan{api.NewAction(api.ActionTypeCleanTLSCACertificate, api.ServerGroupUnknown, "", "Clean CA from truststore").
				AddParam(actionTypeAppendTLSCACertificateChecksum, sha)}
		}
	}

	return nil
}

// createKeyfileRenewalPlan creates plan to renew server keyfile
func createKeyfileRenewalPlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspector.Inspector, context PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	var plan api.Plan

	status.Members.ForeachServerGroup(func(group api.ServerGroup, members api.MemberStatusList) error {
		if !group.IsArangod() {
			return nil
		}

		for _, member := range members {
			if !plan.IsEmpty() {
				return nil
			}

			if renew, recreate := keyfileRenewalRequired(ctx, log, apiObject, spec, status, cachedStatus, context, group, member); renew {
				log.Info().Msg("Renewal of keyfile required")
				plan = append(plan, createKeyfileRotationPlan(log, spec, status, group, member, recreate)...)
			}
		}

		return nil
	})

	return plan
}

func createKeyfileRenewalPlanMode(
	spec api.DeploymentSpec, status api.DeploymentStatus,
	member api.MemberStatus) api.TLSRotateMode {
	if !spec.TLS.IsSecure() {
		return api.TLSRotateModeRecreate
	}

	if spec.TLS.Mode.Get() != api.TLSRotateModeInPlace {
		return api.TLSRotateModeRecreate
	}

	if i := status.CurrentImage; i == nil {
		return api.TLSRotateModeRecreate
	} else {
		if !i.Enterprise || i.ArangoDBVersion.CompareTo("3.7.0") < 0 || i.ImageID != member.ImageID {
			return api.TLSRotateModeRecreate
		}
	}

	return api.TLSRotateModeInPlace
}

func createKeyfileRotationPlan(log zerolog.Logger, spec api.DeploymentSpec, status api.DeploymentStatus, group api.ServerGroup, member api.MemberStatus, recreate bool) api.Plan {
	p := api.Plan{}

	if recreate {
		p = append(p,
			api.NewAction(api.ActionTypeCleanTLSKeyfileCertificate, group, member.ID, "Remove server keyfile and enforce renewal"))
	}

	switch createKeyfileRenewalPlanMode(spec, status, member) {
	case api.TLSRotateModeInPlace:
		p = append(p, api.NewAction(api.ActionTypeRefreshTLSKeyfileCertificate, group, member.ID, "Renew Member Keyfile"))
	default:
		p = append(p, createRotateMemberPlan(log, member, group, "Restart server after keyfile removal")...)
	}
	return p
}

func checkServerValidCertRequest(ctx context.Context, apiObject k8sutil.APIObject, group api.ServerGroup, member api.MemberStatus, ca Certificates) (*tls.ConnectionState, error) {
	endpoint := fmt.Sprintf("https://%s:%d", k8sutil.CreatePodDNSName(apiObject, group.AsRole(), member.ID), k8sutil.ArangoPort)

	tlsConfig := &tls.Config{
		RootCAs: ca.AsCertPool(),
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport, Timeout: time.Second}

	resp, err := client.Get(endpoint)
	if err != nil {
		return nil, err
	}

	return resp.TLS, nil
}

func keyfileRenewalRequired(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspector.Inspector, context PlanBuilderContext,
	group api.ServerGroup, member api.MemberStatus) (bool, bool) {
	if !spec.TLS.IsSecure() {
		return false, false
	}

	caSecret, exists := cachedStatus.Secret(spec.TLS.GetCASecretName())
	if !exists {
		log.Warn().Str("secret", spec.TLS.GetCASecretName()).Msg("CA Secret does not exists")
		return false, false
	}

	ca, _, err := getKeyCertFromSecret(log, caSecret, resources.CACertName, resources.CAKeyName)
	if err != nil {
		log.Warn().Err(err).Str("secret", spec.TLS.GetCASecretName()).Msg("CA Secret does not contains Cert")
		return false, false
	}

	res, err := checkServerValidCertRequest(ctx, apiObject, group, member, ca)
	if err != nil {
		switch v := err.(type) {
		case *url.Error:
			switch v.Err.(type) {
			case x509.UnknownAuthorityError, x509.CertificateInvalidError:
				return true, true
			default:
				log.Warn().Err(v.Err).Str("type", reflect.TypeOf(v.Err).String()).Msg("Validation of server cert failed")
			}
		default:
			log.Warn().Err(err).Str("type", reflect.TypeOf(err).String()).Msg("Validation of server cert failed")
		}
		return false, false
	}

	// Check if cert is not expired
	for _, cert := range res.PeerCertificates {
		if cert == nil {
			continue
		}

		if time.Now().Add(CertificateRenewalMargin).After(cert.NotAfter) {
			return true, true
		}
	}

	// Ensure secret is propagated only on 3.7.0+ enterprise and inplace mode
	if createKeyfileRenewalPlanMode(spec, status, member) == api.TLSRotateModeInPlace {
		conn, err := context.GetServerClient(ctx, group, member.ID)
		if err != nil {
			log.Warn().Err(err).Msg("Unable to get client")
			return false, false
		}

		s, exists := cachedStatus.Secret(k8sutil.CreateTLSKeyfileSecretName(apiObject.GetName(), group.AsRole(), member.ID))
		if !exists {
			log.Warn().Msg("Keyfile secret is missing")
			return false, false
		}

		c := client.NewClient(conn.Connection())
		tls, err := c.GetTLS(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("Unable to get tls details")
			return false, false
		}

		keyfile, ok := s.Data[constants.SecretTLSKeyfile]
		if !ok {
			log.Warn().Msg("Keyfile secret is invalid")
			return false, false
		}

		keyfileSha := util.SHA256(keyfile)

		if tls.Result.KeyFile.Checksum != keyfileSha {
			return true, false
		}
	}

	return false, false
}
