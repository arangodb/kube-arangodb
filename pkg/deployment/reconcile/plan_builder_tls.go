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
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/pkg/deployment/features"

	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/rs/zerolog"
)

const CertificateRenewalMargin = 7 * 24 * time.Hour

func createTLSStatusPropagatedFieldUpdate(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext, w WithPlanBuilder, builders ...planBuilder) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	var plan api.Plan

	for _, builder := range builders {
		if !plan.IsEmpty() {
			continue
		}

		if p := w.Apply(builder); !p.IsEmpty() {
			plan = append(plan, p...)
		}
	}

	if plan.IsEmpty() {
		return nil
	}

	if status.Hashes.TLS.Propagated {
		plan = append(api.Plan{
			api.NewAction(api.ActionTypeTLSPropagated, api.ServerGroupUnknown, "", "Change propagated flag to false").AddParam(propagated, conditionFalse),
		}, plan...)
	}

	return plan
}

// createTLSStatusUpdate creates plan to update ca info
func createTLSStatusUpdate(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	if createTLSStatusUpdateRequired(log, apiObject, spec, status, cachedStatus) {
		return api.Plan{api.NewAction(api.ActionTypeTLSKeyStatusUpdate, api.ServerGroupUnknown, "", "Update status")}
	}

	return nil
}

// createTLSStatusUpdate creates plan to update ca info
func createTLSStatusPropagated(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	if !status.Hashes.TLS.Propagated {
		return api.Plan{
			api.NewAction(api.ActionTypeTLSPropagated, api.ServerGroupUnknown, "", "Change propagated flag to true").AddParam(propagated, conditionTrue),
		}
	}

	return nil
}

func createTLSStatusUpdateRequired(log zerolog.Logger, apiObject k8sutil.APIObject, spec api.DeploymentSpec,
	status api.DeploymentStatus, cachedStatus inspectorInterface.Inspector) bool {
	if !spec.TLS.IsSecure() {
		return false
	}

	trusted, exists := cachedStatus.Secret(resources.GetCASecretName(apiObject))
	if !exists {
		log.Warn().Str("secret", resources.GetCASecretName(apiObject)).Msg("Folder with secrets does not exist")
		return false
	}

	keyHashes := secretKeysToListWithPrefix(trusted)

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
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	caSecret, exists := cachedStatus.Secret(spec.TLS.GetCASecretName())
	if !exists {
		log.Warn().Str("secret", spec.TLS.GetCASecretName()).Msg("CA Secret does not exists")
		return nil
	}

	ca, _, err := resources.GetKeyCertFromSecret(log, caSecret, resources.CACertName, resources.CAKeyName)
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
			AddParam(checksum, certSha)}
	}

	return nil
}

// createCARenewalPlan creates plan to renew CA
func createCARenewalPlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	caSecret, exists := cachedStatus.Secret(spec.TLS.GetCASecretName())
	if !exists {
		log.Warn().Str("secret", spec.TLS.GetCASecretName()).Msg("CA Secret does not exists")
		return nil
	}

	if !k8sutil.IsOwner(apiObject.AsOwner(), caSecret) {
		log.Debug().Str("secret", spec.TLS.GetCASecretName()).Msg("CA Secret is not owned by Operator, we wont do anything")
		return nil
	}

	cas, _, err := resources.GetKeyCertFromSecret(log, caSecret, resources.CACertName, resources.CAKeyName)
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
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	caSecret, exists := cachedStatus.Secret(spec.TLS.GetCASecretName())
	if !exists {
		log.Warn().Str("secret", spec.TLS.GetCASecretName()).Msg("CA Secret does not exists")
		return nil
	}

	ca, _, err := resources.GetKeyCertFromSecret(log, caSecret, resources.CACertName, resources.CAKeyName)
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
				AddParam(checksum, sha)}
		}
	}

	return nil
}

func createKeyfileRenewalPlanDefault(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, planCtx PlanBuilderContext) api.Plan {
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

			lCtx, c := context.WithTimeout(ctx, 500*time.Millisecond)
			defer c()

			if renew, recreate := keyfileRenewalRequired(lCtx, log, apiObject, spec, cachedStatus, planCtx, group, member, api.TLSRotateModeRecreate); renew {
				log.Info().Msg("Renewal of keyfile required - Recreate")
				if recreate {
					plan = append(plan, api.NewAction(api.ActionTypeCleanTLSKeyfileCertificate, group, member.ID, "Remove server keyfile and enforce renewal"))
				}
				plan = append(plan, createRotateMemberPlan(log, member, group, "Restart server after keyfile removal")...)
			}
		}

		return nil
	})

	return plan
}

func createKeyfileRenewalPlanInPlace(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, planCtx PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	var plan api.Plan

	status.Members.ForeachServerGroup(func(group api.ServerGroup, members api.MemberStatusList) error {
		if !group.IsArangod() {
			return nil
		}

		for _, member := range members {
			lCtx, c := context.WithTimeout(ctx, 500*time.Millisecond)
			defer c()

			if renew, recreate := keyfileRenewalRequired(lCtx, log, apiObject, spec, cachedStatus, planCtx, group, member, api.TLSRotateModeInPlace); renew {
				log.Info().Msg("Renewal of keyfile required - InPlace")
				if recreate {
					plan = append(plan, api.NewAction(api.ActionTypeCleanTLSKeyfileCertificate, group, member.ID, "Remove server keyfile and enforce renewal"))
				}
				plan = append(plan, api.NewAction(api.ActionTypeRefreshTLSKeyfileCertificate, group, member.ID, "Renew Member Keyfile"))
			}
		}

		return nil
	})

	return plan
}

func createKeyfileRenewalPlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, planCtx PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	gCtx, c := context.WithTimeout(ctx, 2*time.Second)
	defer c()

	switch createKeyfileRenewalPlanMode(spec, status) {
	case api.TLSRotateModeInPlace:
		return createKeyfileRenewalPlanInPlace(gCtx, log, apiObject, spec, status, cachedStatus, planCtx)
	default:
		return createKeyfileRenewalPlanDefault(gCtx, log, apiObject, spec, status, cachedStatus, planCtx)
	}
}

func createKeyfileRenewalPlanMode(
	spec api.DeploymentSpec, status api.DeploymentStatus) api.TLSRotateMode {
	if !spec.TLS.IsSecure() {
		return api.TLSRotateModeRecreate
	}

	mode := spec.TLS.Mode.Get()

	status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		if mode != api.TLSRotateModeInPlace {
			return nil
		}

		for _, member := range list {
			if mode != api.TLSRotateModeInPlace {
				return nil
			}

			if i, ok := status.Images.GetByImageID(member.ImageID); !ok {
				mode = api.TLSRotateModeRecreate
			} else {
				if !features.TLSRotation().Supported(i.ArangoDBVersion, i.Enterprise) {
					mode = api.TLSRotateModeRecreate
				}
			}
		}

		return nil
	})

	return mode
}

func checkServerValidCertRequest(ctx context.Context, context PlanBuilderContext, apiObject k8sutil.APIObject, group api.ServerGroup, member api.MemberStatus, ca resources.Certificates) (*tls.ConnectionState, error) {
	endpoint := fmt.Sprintf("https://%s:%d", k8sutil.CreatePodDNSNameWithDomain(apiObject, context.GetSpec().ClusterDomain, group.AsRole(), member.ID), k8sutil.ArangoPort)

	tlsConfig := &tls.Config{
		RootCAs: ca.AsCertPool(),
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport, Timeout: time.Second}

	auth, err := context.GetAuthentication()()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	if auth != nil && auth.Type() == driver.AuthenticationTypeRaw {
		if h := auth.Get("value"); h != "" {
			req.Header.Add("Authorization", h)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp.TLS, nil
}

func keyfileRenewalRequired(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext,
	group api.ServerGroup, member api.MemberStatus, mode api.TLSRotateMode) (bool, bool) {
	if !spec.TLS.IsSecure() {
		return false, false
	}

	caSecret, exists := cachedStatus.Secret(spec.TLS.GetCASecretName())
	if !exists {
		log.Warn().Str("secret", spec.TLS.GetCASecretName()).Msg("CA Secret does not exists")
		return false, false
	}

	ca, _, err := resources.GetKeyCertFromSecret(log, caSecret, resources.CACertName, resources.CAKeyName)
	if err != nil {
		log.Warn().Err(err).Str("secret", spec.TLS.GetCASecretName()).Msg("CA Secret does not contains Cert")
		return false, false
	}

	res, err := checkServerValidCertRequest(ctx, context, apiObject, group, member, ca)
	if err != nil {
		switch v := err.(type) {
		case *url.Error:
			switch v.Err.(type) {
			case x509.UnknownAuthorityError, x509.CertificateInvalidError:
				log.Debug().Err(v.Err).Str("type", reflect.TypeOf(v.Err).String()).Msg("Validation of server cert failed")
				return true, true
			default:
				log.Debug().Err(v.Err).Str("type", reflect.TypeOf(v.Err).String()).Msg("Validation of server cert failed")
			}
		default:
			log.Debug().Err(err).Str("type", reflect.TypeOf(err).String()).Msg("Validation of server cert failed")
		}
		return false, false
	}

	// Check if cert is not expired
	for _, cert := range res.PeerCertificates {
		if cert == nil {
			continue
		}

		if ca.Contains(cert) {
			continue
		}

		if time.Now().Add(CertificateRenewalMargin).After(cert.NotAfter) {
			log.Info().Msg("Renewal margin exceeded")
			return true, true
		}
	}

	// Ensure secret is propagated only on 3.7.0+ enterprise and inplace mode
	if mode == api.TLSRotateModeInPlace {
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

		if tls.Result.KeyFile.GetSHA().Checksum() != keyfileSha {
			log.Debug().Str("current", tls.Result.KeyFile.GetSHA().Checksum()).Str("desired", keyfileSha).Msg("Unable to get tls details")
			return true, false
		}
	}

	return false, false
}
