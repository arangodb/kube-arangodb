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

package crd

import (
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func init() {
	registerCRDWithPanic("arangobackups.backup.arangodb.com", crd{
		version: "1.0.1",
		spec: apiextensions.CustomResourceDefinitionSpec{
			Group: "backup.arangodb.com",
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural:   "arangobackups",
				Singular: "arangobackup",
				Kind:     "ArangoBackup",
				ListKind: "ArangoBackupList",
				ShortNames: []string{
					"arangobackup",
				},
			},
			Scope: apiextensions.NamespaceScoped,
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name: "v1",
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type:                   "object",
							XPreserveUnknownFields: util.NewBool(true),
						},
					},
					Served:  true,
					Storage: true,
					AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
						{
							JSONPath:    ".spec.policyName",
							Description: "Policy name",
							Name:        "Policy",
							Type:        "string",
						},
						{
							JSONPath:    ".spec.deployment.name",
							Description: "Deployment name",
							Name:        "Deployment",
							Type:        "string",
						},
						{
							JSONPath:    ".status.backup.version",
							Description: "Backup Version",
							Name:        "Version",
							Type:        "string",
						},
						{
							JSONPath:    ".status.backup.createdAt",
							Description: "Backup Creation Timestamp",
							Name:        "Created",
							Type:        "string",
						},
						{
							JSONPath:    ".status.backup.sizeInBytes",
							Description: "Backup Size in Bytes",
							Name:        "Size",
							Type:        "integer",
							Format:      "byte",
						},
						{
							JSONPath:    ".status.backup.numberOfDBServers",
							Description: "Backup Number of the DB Servers",
							Name:        "DBServers",
							Type:        "integer",
						},
						{
							JSONPath:    ".status.state",
							Description: "The actual state of the ArangoBackup",
							Name:        "State",
							Type:        "string",
						},
						{
							JSONPath:    ".status.message",
							Priority:    1,
							Description: "Message of the ArangoBackup object",
							Name:        "Message",
							Type:        "string",
						},
					},
					Subresources: &apiextensions.CustomResourceSubresources{
						Status: &apiextensions.CustomResourceSubresourceStatus{},
					},
				},
				{
					Name: "v1alpha",
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type:                   "object",
							XPreserveUnknownFields: util.NewBool(true),
						},
					},
					Served:  true,
					Storage: false,
					AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
						{
							JSONPath:    ".spec.policyName",
							Description: "Policy name",
							Name:        "Policy",
							Type:        "string",
						},
						{
							JSONPath:    ".spec.deployment.name",
							Description: "Deployment name",
							Name:        "Deployment",
							Type:        "string",
						},
						{
							JSONPath:    ".status.backup.version",
							Description: "Backup Version",
							Name:        "Version",
							Type:        "string",
						},
						{
							JSONPath:    ".status.backup.createdAt",
							Description: "Backup Creation Timestamp",
							Name:        "Created",
							Type:        "string",
						},
						{
							JSONPath:    ".status.backup.sizeInBytes",
							Description: "Backup Size in Bytes",
							Name:        "Size",
							Type:        "integer",
							Format:      "byte",
						},
						{
							JSONPath:    ".status.backup.numberOfDBServers",
							Description: "Backup Number of the DB Servers",
							Name:        "DBServers",
							Type:        "integer",
						},
						{
							JSONPath:    ".status.state",
							Description: "The actual state of the ArangoBackup",
							Name:        "State",
							Type:        "string",
						},
						{
							JSONPath:    ".status.message",
							Priority:    1,
							Description: "Message of the ArangoBackup object",
							Name:        "Message",
							Type:        "string",
						},
					},
					Subresources: &apiextensions.CustomResourceSubresources{
						Status: &apiextensions.CustomResourceSubresourceStatus{},
					},
				},
			},
		},
	})
}
