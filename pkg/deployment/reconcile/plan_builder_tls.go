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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	memberTls "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tls"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tools"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

const CertificateRenewalMargin = 7 * 24 * time.Hour

func (r *Reconciler) createTLSStatusPropagatedFieldUpdate(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext, w WithPlanBuilder, builders ...planBuilder) api.Plan {
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
			actions.NewClusterAction(api.ActionTypeTLSPropagated, "Change propagated flag to false").AddParam(propagated, conditionFalse),
		}, plan...)
	}

	return plan
}

// createTLSStatusUpdate creates plan to update ca info
func (r *Reconciler) createTLSStatusUpdate(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	if r.createTLSStatusUpdateRequired(apiObject, spec, status, context) {
		return api.Plan{actions.NewClusterAction(api.ActionTypeTLSKeyStatusUpdate, "Update status")}
	}

	return nil
}

// createTLSStatusUpdate creates plan to update ca info
func (r *Reconciler) createTLSStatusPropagated(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	if !status.Hashes.TLS.Propagated {
		return api.Plan{
			actions.NewClusterAction(api.ActionTypeTLSPropagated, "Change propagated flag to true").AddParam(propagated, conditionTrue),
		}
	}

	return nil
}

func (r *Reconciler) createTLSStatusUpdateRequired(apiObject k8sutil.APIObject, spec api.DeploymentSpec,
	status api.DeploymentStatus, context PlanBuilderContext) bool {
	if !spec.TLS.IsSecure() {
		return false
	}

	trusted, exists := context.ACS().CurrentClusterCache().Secret().V1().GetSimple(resources.GetCASecretName(apiObject))
	if !exists {
		r.planLogger.Str("secret", resources.GetCASecretName(apiObject)).Warn("Folder with secrets does not exist")
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

	if !strings.CompareStringArray(status.Hashes.TLS.Truststore, keyHashes) {
		return true
	}

	return false
}

// createCAAppendPlan creates plan to append CA
func (r *Reconciler) createCAAppendPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	caSecret, exists := context.ACS().CurrentClusterCache().Secret().V1().GetSimple(spec.TLS.GetCASecretName())
	if !exists {
		r.planLogger.Str("secret", spec.TLS.GetCASecretName()).Warn("CA Secret does not exists")
		return nil
	}

	ca, _, err := resources.GetKeyCertFromSecret(caSecret, resources.CACertName, resources.CAKeyName)
	if err != nil {
		r.planLogger.Err(err).Str("secret", spec.TLS.GetCASecretName()).Warn("CA Secret does not contains Cert")
		return nil
	}

	if len(ca) == 0 {
		r.planLogger.Str("secret", spec.TLS.GetCASecretName()).Warn("CA does not contain any certs")
		return nil
	}

	trusted, exists := context.ACS().CurrentClusterCache().Secret().V1().GetSimple(resources.GetCASecretName(apiObject))
	if !exists {
		r.planLogger.Str("secret", resources.GetCASecretName(apiObject)).Warn("Folder with secrets does not exist")
		return nil
	}

	caData, err := ca.ToPem()
	if err != nil {
		r.planLogger.Err(err).Str("secret", spec.TLS.GetCASecretName()).Warn("Unable to parse cert")
		return nil
	}

	certSha := util.SHA256(caData)

	if _, exists := trusted.Data[certSha]; !exists {
		return api.Plan{actions.NewClusterAction(api.ActionTypeAppendTLSCACertificate, "Append CA to truststore").
			AddParam(checksum, certSha)}
	}

	return nil
}

// createCARenewalPlan creates plan to renew CA
func (r *Reconciler) createCARenewalPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	caSecret, exists := context.ACS().CurrentClusterCache().Secret().V1().GetSimple(spec.TLS.GetCASecretName())
	if !exists {
		r.planLogger.Str("secret", spec.TLS.GetCASecretName()).Warn("CA Secret does not exists")
		return nil
	}

	if !tools.IsOwner(apiObject.AsOwner(), caSecret) {
		r.planLogger.Str("secret", spec.TLS.GetCASecretName()).Debug("CA Secret is not owned by Operator, we wont do anything")
		return nil
	}

	cas, _, err := resources.GetKeyCertFromSecret(caSecret, resources.CACertName, resources.CAKeyName)
	if err != nil {
		r.planLogger.Err(err).Str("secret", spec.TLS.GetCASecretName()).Warn("CA Secret does not contains Cert")
		return nil
	}

	for _, ca := range cas {
		if time.Now().Add(CertificateRenewalMargin).After(ca.NotAfter) {
			// CA will expire soon, renewal needed
			return api.Plan{actions.NewClusterAction(api.ActionTypeRenewTLSCACertificate, "Renew CA Certificate")}
		}
	}

	return nil
}

// createCACleanPlan creates plan to remove old CA's
func (r *Reconciler) createCACleanPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	caSecret, exists := context.ACS().CurrentClusterCache().Secret().V1().GetSimple(spec.TLS.GetCASecretName())
	if !exists {
		r.planLogger.Str("secret", spec.TLS.GetCASecretName()).Warn("CA Secret does not exists")
		return nil
	}

	ca, _, err := resources.GetKeyCertFromSecret(caSecret, resources.CACertName, resources.CAKeyName)
	if err != nil {
		r.planLogger.Err(err).Str("secret", spec.TLS.GetCASecretName()).Warn("CA Secret does not contains Cert")
		return nil
	}

	if len(ca) == 0 {
		r.planLogger.Str("secret", spec.TLS.GetCASecretName()).Warn("CA does not contain any certs")
		return nil
	}

	trusted, exists := context.ACS().CurrentClusterCache().Secret().V1().GetSimple(resources.GetCASecretName(apiObject))
	if !exists {
		r.planLogger.Str("secret", resources.GetCASecretName(apiObject)).Warn("Folder with secrets does not exist")
		return nil
	}

	caData, err := ca.ToPem()
	if err != nil {
		r.planLogger.Err(err).Str("secret", spec.TLS.GetCASecretName()).Warn("Unable to parse cert")
		return nil
	}

	certSha := util.SHA256(caData)

	for sha := range trusted.Data {
		if certSha != sha {
			return api.Plan{actions.NewClusterAction(api.ActionTypeCleanTLSCACertificate, "Clean CA from truststore").
				AddParam(checksum, sha)}
		}
	}

	return nil
}

func (r *Reconciler) createKeyfileRenewalPlanSynced(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	planCtx PlanBuilderContext) api.Plan {

	if !planCtx.IsSyncEnabled() || !spec.Sync.TLS.IsSecure() {
		return nil
	}

	var plan api.Plan
	group := api.ServerGroupSyncMasters

	for _, member := range status.Members.MembersOfGroup(group) {
		if !plan.IsEmpty() {
			continue
		}

		cache, ok := planCtx.ACS().ClusterCache(member.ClusterID)
		if !ok {
			continue
		}

		lCtx, c := context.WithTimeout(ctx, 500*time.Millisecond)
		defer c()

		if renew, _ := r.keyfileRenewalRequired(lCtx, apiObject, spec.Sync.TLS, spec, cache, planCtx, group, member, api.TLSRotateModeRecreate); renew {
			r.planLogger.Info("Renewal of keyfile required - Recreate (sync master)")
			plan = append(plan, tlsRotateConditionAction(group, member.ID, "Restart sync master after keyfile removal"))
		}
	}

	return plan
}

func (r *Reconciler) createKeyfileRenewalPlanDefault(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	planCtx PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	var plan api.Plan

	for _, e := range status.Members.AsListInGroups(api.AllArangoDServerGroups...) {
		cache, ok := planCtx.ACS().ClusterCache(e.Member.ClusterID)
		if !ok {
			continue
		}

		lCtx, c := context.WithTimeout(ctx, 500*time.Millisecond)
		defer c()

		if renew, _ := r.keyfileRenewalRequired(lCtx, apiObject, spec.TLS, spec, cache, planCtx, e.Group, e.Member, api.TLSRotateModeRecreate); renew {
			r.planLogger.Info("Renewal of keyfile required - Recreate (server)")
			plan = append(plan, tlsRotateConditionAction(e.Group, e.Member.ID, "Restart server after keyfile removal"))
			break
		}
	}

	return plan
}

func (r *Reconciler) createKeyfileRenewalPlanInPlace(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	planCtx PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	var plan api.Plan

	for _, e := range status.Members.AsListInGroups(api.AllArangoDServerGroups...) {
		cache, ok := planCtx.ACS().ClusterCache(e.Member.ClusterID)
		if !ok {
			continue
		}

		lCtx, c := context.WithTimeout(ctx, 500*time.Millisecond)
		defer c()

		if renew, recreate := r.keyfileRenewalRequired(lCtx, apiObject, spec.TLS, spec, cache, planCtx, e.Group, e.Member, api.TLSRotateModeInPlace); renew {
			r.planLogger.Info("Renewal of keyfile required - InPlace (server)")
			if recreate {
				plan = append(plan, actions.NewAction(api.ActionTypeCleanTLSKeyfileCertificate, e.Group, e.Member, "Remove server keyfile and enforce renewal"))
			}
			plan = append(plan, actions.NewAction(api.ActionTypeRefreshTLSKeyfileCertificate, e.Group, e.Member, "Renew Member Keyfile"))
		}
	}

	return plan
}

func (r *Reconciler) createKeyfileRenewalPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	planCtx PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	gCtx, c := context.WithTimeout(ctx, 2*time.Second)
	defer c()

	plan := r.createKeyfileRenewalPlanSynced(gCtx, apiObject, spec, status, planCtx)

	switch createKeyfileRenewalPlanMode(spec, status) {
	case api.TLSRotateModeInPlace:
		plan = append(plan, r.createKeyfileRenewalPlanInPlace(gCtx, apiObject, spec, status, planCtx)...)
	default:
		plan = append(plan, r.createKeyfileRenewalPlanDefault(gCtx, apiObject, spec, status, planCtx)...)
	}

	return plan
}

func createKeyfileRenewalPlanMode(
	spec api.DeploymentSpec, status api.DeploymentStatus) api.TLSRotateMode {
	if !spec.TLS.IsSecure() {
		return api.TLSRotateModeRecreate
	}

	mode := spec.TLS.Mode.Get()

	for _, e := range status.Members.AsList() {
		if mode != api.TLSRotateModeInPlace {
			break
		}

		if i, ok := status.Images.GetByImageID(e.Member.ImageID); !ok {
			mode = api.TLSRotateModeRecreate
		} else {
			if !features.TLSRotation().Supported(i.ArangoDBVersion, i.Enterprise) {
				mode = api.TLSRotateModeRecreate
			}
		}
	}

	return mode
}

func checkServerValidCertRequest(ctx context.Context, context PlanBuilderContext, apiObject k8sutil.APIObject, group api.ServerGroup, member api.MemberStatus, ca resources.Certificates) (*tls.ConnectionState, error) {
	endpoint := fmt.Sprintf("https://%s:%d", k8sutil.CreatePodDNSNameWithDomain(apiObject, context.GetSpec().ClusterDomain, group.AsRole(), member.ID), shared.ArangoPort)
	if group == api.ServerGroupSyncMasters {
		endpoint = fmt.Sprintf("https://%s:%d%s", k8sutil.CreatePodDNSNameWithDomain(apiObject, context.GetSpec().ClusterDomain, group.AsRole(), member.ID), shared.ArangoSyncMasterPort, shared.ArangoSyncStatusEndpoint)
	}

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

// keyfileRenewalRequired checks if a keyfile renewal is required and if recreation should be made
func (r *Reconciler) keyfileRenewalRequired(ctx context.Context, apiObject k8sutil.APIObject, tls api.TLSSpec,
	spec api.DeploymentSpec, cachedStatus inspectorInterface.Inspector,
	context PlanBuilderContext,
	group api.ServerGroup, member api.MemberStatus, mode api.TLSRotateMode) (bool, bool) {
	if !tls.IsSecure() {
		return false, false
	}

	memberName := member.ArangoMemberName(apiObject.GetName(), group)

	service, ok := cachedStatus.Service().V1().GetSimple(memberName)
	if !ok {
		r.planLogger.Str("service", memberName).Warn("Service does not exists")
		return false, false
	}

	caSecret, exists := cachedStatus.Secret().V1().GetSimple(tls.GetCASecretName())
	if !exists {
		r.planLogger.Str("secret", tls.GetCASecretName()).Warn("CA Secret does not exists")
		return false, false
	}

	ca, _, err := resources.GetKeyCertFromSecret(caSecret, resources.CACertName, resources.CAKeyName)
	if err != nil {
		r.planLogger.Err(err).Str("secret", tls.GetCASecretName()).Warn("CA Secret does not contains Cert")
		return false, false
	}

	res, err := checkServerValidCertRequest(ctx, context, apiObject, group, member, ca)
	if err != nil {
		switch v := err.(type) {
		case *url.Error:
			switch v.Err.(type) {
			case x509.UnknownAuthorityError, x509.CertificateInvalidError:
				r.planLogger.Err(v.Err).Str("type", reflect.TypeOf(v.Err).String()).Debug("Validation of cert for %s failed, renewal is required", memberName)
				return true, true
			default:
				r.planLogger.Err(v.Err).Str("type", reflect.TypeOf(v.Err).String()).Debug("Validation of cert for %s failed, but cert looks fine - continuing", memberName)
			}
		default:
			r.planLogger.Err(err).Str("type", reflect.TypeOf(err).String()).Debug("Validation of cert for %s failed, will try again next time", memberName)
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
			r.planLogger.Info("Renewal margin exceeded")
			return true, true
		}

		// Verify AltNames
		var altNames memberTls.KeyfileInput
		if group.IsArangosync() {
			altNames, err = memberTls.GetSyncAltNames(apiObject, spec, tls, group, member)
		} else {
			altNames, err = memberTls.GetServerAltNames(apiObject, spec, tls, service, group, member)
		}

		if err != nil {
			r.planLogger.Warn("Unable to render alt names")
			return false, false
		}

		var dnsNames = cert.DNSNames

		for _, ip := range cert.IPAddresses {
			dnsNames = append(dnsNames, ip.String())
		}

		if a := strings.DiffStrings(altNames.AltNames, dnsNames); len(a) > 0 {
			r.planLogger.Strs("AltNames Current", cert.DNSNames...).
				Strs("AltNames Expected", altNames.AltNames...).
				Info("Alt names are different")
			return true, true
		}
	}

	// Ensure secret is propagated only on 3.7.0+ enterprise and inplace mode
	if mode == api.TLSRotateModeInPlace && group.IsArangod() {
		conn, err := context.GetMembersState().GetMemberClient(member.ID)
		if err != nil {
			r.planLogger.Err(err).Warn("Unable to get client")
			return false, false
		}

		s, exists := cachedStatus.Secret().V1().GetSimple(k8sutil.CreateTLSKeyfileSecretName(apiObject.GetName(), group.AsRole(), member.ID))
		if !exists {
			r.planLogger.Warn("Keyfile secret is missing")
			return false, false
		}

		c := client.NewClient(conn.Connection(), r.log)
		tls, err := c.GetTLS(ctx)
		if err != nil {
			r.planLogger.Err(err).Warn("Unable to get tls details")
			return false, false
		}

		keyfile, ok := s.Data[constants.SecretTLSKeyfile]
		if !ok {
			r.planLogger.Warn("Keyfile secret is invalid")
			return false, false
		}

		keyfileSha := util.SHA256(keyfile)

		if tls.Result.KeyFile.GetSHA().Checksum() != keyfileSha {
			r.planLogger.Str("current", tls.Result.KeyFile.GetSHA().Checksum()).Str("desired", keyfileSha).Debug("Unable to get tls details")
			return true, false
		}
	}

	return false, false
}
