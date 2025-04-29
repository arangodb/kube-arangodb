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

package crd

import (
	"context"
	"io"
	"sort"
	"strconv"

	authorization "k8s.io/api/authorization/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/arangodb/kube-arangodb/pkg/crd/crds"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/access"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
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
	// CRDOptions defines skip options
	Skip []string
}

func EnsureCRDWithOptions(ctx context.Context, client kclient.Client, opts EnsureCRDOptions) error {
	crdsLock.Lock()
	defer crdsLock.Unlock()

	for crdName, crdReg := range registeredCRDs {

		getAccess := verifyCRDAccess(ctx, client, crdName, access.Get)
		if !getAccess.Allowed {
			logger.
				Str("crd", crdName).
				Str("reason", getAccess.Reason).
				Str("evaluationError", getAccess.EvaluationError).
				Info("Get Operations is not allowed. Continue")
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

func GenerateCRDYAMLWithOptions(opts EnsureCRDOptions, out io.Writer) error {
	crds := GenerateCRDWithOptions(opts)

	for id, crd := range crds {
		obj := map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"labels": crd.Labels,
				"name":   crd.Name,
			},
			"spec": crd.Spec,
		}

		data, err := yaml.Marshal(obj)
		if err != nil {
			return err
		}

		if id > 0 {
			_, err = util.WriteAll(out, []byte("\n\n---\n\n"))
			if err != nil {
				return err
			}
		}

		_, err = util.WriteAll(out, data)
		if err != nil {
			return err
		}
	}

	return nil
}

func GenerateCRDWithOptions(opts EnsureCRDOptions) []apiextensions.CustomResourceDefinition {
	crdsLock.Lock()
	defer crdsLock.Unlock()

	ret := make([]apiextensions.CustomResourceDefinition, 0, len(registeredCRDs))

	for crdName, crdReg := range registeredCRDs {
		if util.ContainsList(opts.Skip, crdName) {
			continue
		}

		var opt = &crdReg.defaultOpts
		if o, ok := opts.CRDOptions[crdName]; ok {
			opt = &o
		}
		def := crdReg.getter(opt)

		ret = append(ret, *renderCRD(def, opt))
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].GetName() < ret[j].GetName()
	})

	return ret
}

func renderCRD(def crds.Definition, opts *crds.CRDOptions) *apiextensions.CustomResourceDefinition {
	crdName := def.CRD.Name

	definitionVersion, definitionSchemaVersion := def.DefinitionData.Checksum()

	schema := opts.GetWithSchema()
	preserve := !schema || opts.GetWithPreserve()

	c := &apiextensions.CustomResourceDefinition{
		ObjectMeta: meta.ObjectMeta{
			Name: crdName,
			Labels: map[string]string{
				Version:               definitionVersion,
				PreserveUnknownFields: strconv.FormatBool(preserve),
			},
		},
		Spec: def.CRD.Spec,
	}

	if schema {
		c.Labels[Schema] = definitionSchemaVersion
	}

	return c
}

func tryApplyCRD(ctx context.Context, client kclient.Client, def crds.Definition, opts *crds.CRDOptions, forceUpdate bool) error {
	crdDefinitions := client.KubernetesExtensions().ApiextensionsV1().CustomResourceDefinitions()

	crdName := def.CRD.Name

	definitionVersion, definitionSchemaVersion := def.DefinitionData.Checksum()

	logger := logger.Str("version", definitionVersion)

	schema := opts.GetWithSchema()
	preserve := !schema || opts.GetWithPreserve()

	if schema {
		logger = logger.Str("schema", definitionSchemaVersion)
	}

	logger = logger.Bool("preserve", preserve)

	c, err := crdDefinitions.Get(ctx, crdName, meta.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Err(err).
				Str("crd", crdName).
				Warn("Get Operations is not allowed due to error")
			return err
		}

		createAccess := verifyCRDAccess(ctx, client, crdName, access.Create)

		if !createAccess.Allowed {
			logger.
				Str("crd", crdName).
				Str("reason", createAccess.Reason).
				Str("evaluationError", createAccess.EvaluationError).
				Info("Create Operations is not allowed but CRD is missing. Continue")
			return nil
		}

		c = renderCRD(def, opts)

		if _, err := crdDefinitions.Create(ctx, c, meta.CreateOptions{}); err != nil {
			logger.Err(err).Str("crd", crdName).Warn("Create Operations is not allowed due to error")
			return err
		}

		logger.Str("crd", crdName).Info("CRD Created")
		return nil
	}

	updateAccess := verifyCRDAccess(ctx, client, crdName, access.Update)
	if !updateAccess.Allowed {
		logger.Str("crd", crdName).
			Str("reason", updateAccess.Reason).
			Str("evaluationError", updateAccess.EvaluationError).
			Info("Update Operations is not allowed. Continue")
		return nil
	}

	if c.Labels == nil {
		c.Labels = map[string]string{}
	}

	if !forceUpdate {
		if v, ok := c.Labels[Version]; ok && v == definitionVersion {
			if v, ok := c.Labels[Schema]; (schema && (ok && v == definitionSchemaVersion)) || (!schema && !ok) {
				if v, ok := c.Labels[PreserveUnknownFields]; (preserve && ok && v == "true") || (!preserve && (!ok || v != "true")) {
					logger.Str("crd", crdName).Info("CRD Update not required")
					return nil
				}
			}
		}
	}

	c.Labels[Version] = definitionVersion
	delete(c.Labels, Schema)
	if schema {
		c.Labels[Schema] = definitionSchemaVersion
	}
	c.Labels[PreserveUnknownFields] = strconv.FormatBool(preserve)
	c.Spec = def.CRD.Spec

	if _, err := crdDefinitions.Update(ctx, c, meta.UpdateOptions{}); err != nil {
		logger.Err(err).Str("crd", crdName).Warn("Failed to update CRD definition")
		return err
	}
	logger.Str("crd", crdName).Info("CRD Updated")
	return nil
}

func verifyCRDAccess(ctx context.Context, client kclient.Client, crd string, verb access.Verb) authorization.SubjectAccessReviewStatus {
	if c := verifyCRDAccessForTests; c != nil {
		return *c
	}

	return access.VerifyAccessRequest(ctx, client, access.GVR(constants.CustomResourceDefinitionGRv1(), crd, verb))
}

var verifyCRDAccessForTests *authorization.SubjectAccessReviewStatus
