//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

package crd

import (
	"context"

	authorization "k8s.io/api/authorization/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/crd/crds"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

var logger = logging.Global().RegisterAndGetLogger("crd", logging.Info)

// Deprecated: use EnsureCRDWithOptions instead
func EnsureCRD(ctx context.Context, client kclient.Client, ignoreErrors bool) error {
	return EnsureCRDWithOptions(ctx, client, EnsureCRDOptions{
		IgnoreErrors: ignoreErrors,
	})
}

type EnsureCRDOptions struct {
	// IgnoreErrors do not return errors if could not apply CRD
	IgnoreErrors bool
	// ForceUpdate if true, CRD will be updated even if definitions versions are the same
	ForceUpdate bool
	// CRDOptions defines options per each CRD
	CRDOptions map[string]crds.CRDOptions
}

func EnsureCRDWithOptions(ctx context.Context, client kclient.Client, opts EnsureCRDOptions) error {
	crdsLock.Lock()
	defer crdsLock.Unlock()

	for crdName, crdReg := range registeredCRDs {
		getAccess := verifyCRDAccess(ctx, client, crdName, "get")
		if !getAccess.Allowed {
			logger.Str("crd", crdName).Info("Get Operations is not allowed. Continue")
			continue
		}

		var opt = &crdReg.defaultOpts
		if o, ok := opts.CRDOptions[crdName]; ok {
			opt = &o
		}
		def := crdReg.getter(opt)

		err := tryApplyCRD(ctx, client, def, opt, opts.ForceUpdate)
		if !opts.IgnoreErrors && err != nil {
			return err
		}
	}
	return nil
}

func tryApplyCRD(ctx context.Context, client kclient.Client, def crds.Definition, opts *crds.CRDOptions, forceUpdate bool) error {
	crdDefinitions := client.KubernetesExtensions().ApiextensionsV1().CustomResourceDefinitions()

	crdName := def.CRD.Name

	definitionVersion, definitionSchemaVersion := def.DefinitionData.Checksum()

	logger := logger.Str("version", definitionVersion)

	if opts.GetWithSchema() {
		logger = logger.Str("schema", definitionSchemaVersion)
	}

	c, err := crdDefinitions.Get(ctx, crdName, meta.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Err(err).Str("crd", crdName).Warn("Get Operations is not allowed due to error")
			return err
		}

		createAccess := verifyCRDAccess(ctx, client, crdName, "create")

		if !createAccess.Allowed {
			logger.Str("crd", crdName).Info("Create Operations is not allowed but CRD is missing. Continue")
			return nil
		}

		c = &apiextensions.CustomResourceDefinition{
			ObjectMeta: meta.ObjectMeta{
				Name: crdName,
				Labels: map[string]string{
					Version: definitionVersion,
				},
			},
			Spec: def.CRD.Spec,
		}

		if opts.GetWithSchema() {
			c.Labels[Schema] = definitionSchemaVersion
		}

		if _, err := crdDefinitions.Create(ctx, c, meta.CreateOptions{}); err != nil {
			logger.Err(err).Str("crd", crdName).Warn("Create Operations is not allowed due to error")
			return err
		}

		logger.Str("crd", crdName).Info("CRD Created")
		return nil
	}

	updateAccess := verifyCRDAccess(ctx, client, crdName, "update")
	if !updateAccess.Allowed {
		logger.Str("crd", crdName).Info("Update Operations is not allowed. Continue")
		return nil
	}

	if c.ObjectMeta.Labels == nil {
		c.ObjectMeta.Labels = map[string]string{}
	}

	if !forceUpdate {
		if v, ok := c.ObjectMeta.Labels[Version]; ok && v == definitionVersion {
			if v, ok := c.ObjectMeta.Labels[Schema]; (opts.GetWithSchema() && (ok && v == definitionSchemaVersion)) || (!opts.GetWithSchema() && !ok) {
				logger.Str("crd", crdName).Info("CRD Update not required")
				return nil
			}
		}
	}

	c.ObjectMeta.Labels[Version] = definitionVersion
	delete(c.ObjectMeta.Labels, Schema)
	if opts.GetWithSchema() {
		c.ObjectMeta.Labels[Schema] = definitionSchemaVersion
	}
	c.Spec = def.CRD.Spec

	if _, err := crdDefinitions.Update(ctx, c, meta.UpdateOptions{}); err != nil {
		logger.Err(err).Str("crd", crdName).Warn("Failed to update CRD definition")
		return err
	}
	logger.Str("crd", crdName).Info("CRD Updated")
	return nil
}

func verifyCRDAccess(ctx context.Context, client kclient.Client, crd string, verb string) authorization.SubjectAccessReviewStatus {
	if c := verifyCRDAccessForTests; c != nil {
		return *c
	}

	return kresources.VerifyAccessRequestStatus(ctx, client.Kubernetes(), verb, "apiextensions.k8s.io", "v1", "customresourcedefinitions", "", crd, "")
}

var verifyCRDAccessForTests *authorization.SubjectAccessReviewStatus
