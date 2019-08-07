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

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/spf13/pflag"
)

var (
	options struct {
		OutputSuffix string
		TemplatesDir string

		Namespace                         string
		Image                             string
		ImagePullPolicy                   string
		ImageSHA256                       bool
		DeploymentOperatorName            string
		DeploymentReplicationOperatorName string
		StorageOperatorName               string
		BackupOperatorName                string
		RBAC                              bool
		AllowChaos                        bool
	}
	crdTemplateNames = []Template{
		Template{Name: "deployment.yaml"},
		Template{Name: "deployment-replication.yaml"},
		Template{Name: "backup.yaml"},
	}
	deploymentTemplateNames = []Template{
		Template{Name: "rbac.yaml", Predicate: hasRBAC},
		Template{Name: "deployment.yaml"},
		Template{Name: "service.yaml"},
	}
	deploymentReplicationTemplateNames = []Template{
		Template{Name: "rbac.yaml", Predicate: hasRBAC},
		Template{Name: "deployment-replication.yaml"},
		Template{Name: "service.yaml"},
	}
	storageTemplateNames = []Template{
		Template{Name: "crd.yaml"},
		Template{Name: "rbac.yaml", Predicate: hasRBAC},
		Template{Name: "deployment.yaml"},
		Template{Name: "service.yaml"},
	}
	backupTemplateNames = []Template{
		Template{Name: "rbac.yaml", Predicate: hasRBAC},
		Template{Name: "deployment.yaml"},
		Template{Name: "service.yaml"},
	}
	testTemplateNames = []Template{
		Template{Name: "rbac.yaml", Predicate: func(o TemplateOptions, isHelm bool) bool { return o.RBAC && !isHelm }},
	}
)

type Template struct {
	Name      string
	Predicate func(o TemplateOptions, isHelm bool) bool
}

type TemplateGroup struct {
	ChartName string
	Templates []Template
}

func hasRBAC(o TemplateOptions, isHelm bool) bool {
	return o.RBAC || isHelm
}

var (
	tmplFuncs = template.FuncMap{
		"quote": func(x string) string { return strconv.Quote(x) },
	}
)

type (
	chartTemplates map[string]string
)

const (
	kubeArangoDBChartTemplate = `
apiVersion: v1
name: kube-arangodb
version: "{{ .Version }}"
description: |
  Kube-ArangoDB is a set of operators to easily deploy ArangoDB deployments on Kubernetes
home: https://arangodb.com
`
	kubeArangoDBStorageChartTemplate = `
apiVersion: v1
name: kube-arangodb-storage
version: "{{ .Version }}"
description: |
  Kube-ArangoDB-Storage is a cluster-wide operator used to provision PersistentVolumes on disks attached locally to Nodes
home: https://arangodb.com
`
	kubeArangoDBCRDChartTemplate = `
apiVersion: v1
name: kube-arangodb-crd
version: "{{ .Version }}"
description: |
  Kube-ArangoDB-crd contains the custom resource definitions for ArangoDeployment and ArangoDeploymentReplication resources.
home: https://arangodb.com
`

	kubeArangoDBValuesTemplate = `
# Image containing the kube-arangodb operators
Image: {{ .Image | quote }}
# Image pull policy for Image
ImagePullPolicy: {{ .ImagePullPolicy | quote }}
RBAC:
  Create: {{ .RBAC }}
Deployment:
  Create: {{ .Deployment.Create }}
  User:
    ServiceAccountName: {{ .Deployment.User.ServiceAccountName | quote }}
  Operator:
    ServiceAccountName: {{ .Deployment.Operator.ServiceAccountName | quote }}
    ServiceType: {{ .Deployment.Operator.ServiceType | quote }}
  AllowChaos: {{ .Deployment.AllowChaos }}
DeploymentReplication:
  Create: {{ .DeploymentReplication.Create }}
  User:
    ServiceAccountName: {{ .DeploymentReplication.User.ServiceAccountName | quote }}
  Operator:
    ServiceAccountName: {{ .DeploymentReplication.Operator.ServiceAccountName | quote }}
    ServiceType: {{ .DeploymentReplication.Operator.ServiceType | quote }}
`

	kubeArangoDBBackupValuesTemplate = `
# Image containing the kube-arangodb operators
Image: {{ .Image | quote }}
# Image pull policy for Image
ImagePullPolicy: {{ .ImagePullPolicy | quote }}
RBAC:
  Create: {{ .RBAC }}
Backup:
  Create: {{ .Backup.Create }}
  User:
    ServiceAccountName: {{ .Backup.User.ServiceAccountName | quote }}
  Operator:
    ServiceAccountName: {{ .Backup.Operator.ServiceAccountName | quote }}
    ServiceType: {{ .Backup.Operator.ServiceType | quote }}
`

	kubeArangoDBStorageValuesTemplate = `
Image: {{ .Image | quote }}
ImagePullPolicy: {{ .ImagePullPolicy | quote }}
RBAC:
  Create: {{ .RBAC }}
Storage:
  Create: {{ .Storage.Create }}
  User:
    ServiceAccountName: {{ .Storage.User.ServiceAccountName | quote }}
  Operator:
    ServiceAccountName: {{ .Storage.Operator.ServiceAccountName | quote }}
    ServiceType: {{ .Storage.Operator.ServiceType | quote }}
`
	kubeArangoDBCRDValuesTemplate = ``

	kubeArangoDBNotesText = `
kube-arangodb has been deployed successfully!

Your release is named '{{ .Release.Name }}'.

{{ if and .Values.Deployment.Create .Values.DeploymentReplication.Create -}}
You can now deploy ArangoDeployment & ArangoDeploymentReplication resources.
{{- else if and .Values.Deployment.Create (not .Values.DeploymentReplication.Create) -}}
You can now deploy ArangoDeployment resources.
{{- else if and (not .Values.Deployment.Create) .Values.DeploymentReplication.Create -}}
You can now deploy ArangoDeploymentReplication resources.
{{- end }}

See https://docs.arangodb.com/devel/Manual/Tutorials/Kubernetes/
for how to get started.
`
	kubeArangoDBStorageNotesText = `
kube-arangodb-storage has been deployed successfully!

Your release is named '{{ .Release.Name }}'.

You can now deploy an ArangoLocalStorage resource.

See https://docs.arangodb.com/devel/Manual/Deployment/Kubernetes/StorageResource.html
for further instructions.
`
	kubeArangoDBCRDNotesText = `
kube-arangodb-crd has been deployed successfully!

Your release is named '{{ .Release.Name }}'.

You can now continue install kube-arangodb chart.
`
)

var (
	chartTemplateGroups = map[string]chartTemplates{
		"kube-arangodb-crd": chartTemplates{
			"Chart.yaml":          kubeArangoDBCRDChartTemplate,
			"values.yaml":         kubeArangoDBCRDValuesTemplate,
			"templates/NOTES.txt": kubeArangoDBCRDNotesText,
		},
		"kube-arangodb": chartTemplates{
			"Chart.yaml":          kubeArangoDBChartTemplate,
			"values.yaml":         kubeArangoDBValuesTemplate,
			"templates/NOTES.txt": kubeArangoDBNotesText,
		},
		"kube-arangodb-storage": chartTemplates{
			"Chart.yaml":          kubeArangoDBStorageChartTemplate,
			"values.yaml":         kubeArangoDBStorageValuesTemplate,
			"templates/NOTES.txt": kubeArangoDBStorageNotesText,
		},
		"kube-arangodb-backup": chartTemplates{
			"Chart.yaml":          kubeArangoDBChartTemplate,
			"values.yaml":         kubeArangoDBBackupValuesTemplate,
			"templates/NOTES.txt": kubeArangoDBNotesText,
		},
	}
)

func init() {
	pflag.StringVar(&options.OutputSuffix, "output-suffix", "", "Suffix of the generated manifest files")
	pflag.StringVar(&options.TemplatesDir, "templates-dir", "manifests/templates", "Directory containing manifest templates")
	pflag.StringVar(&options.Namespace, "namespace", "default", "Namespace in which the operator will be deployed")
	pflag.StringVar(&options.Image, "image", "arangodb/arangodb-operator:latest", "Fully qualified image name of the ArangoDB operator")
	pflag.StringVar(&options.ImagePullPolicy, "image-pull-policy", "IfNotPresent", "Pull policy of the ArangoDB operator image")
	pflag.BoolVar(&options.ImageSHA256, "image-sha256", true, "Use SHA256 syntax for image")
	pflag.StringVar(&options.DeploymentOperatorName, "deployment-operator-name", "arango-deployment-operator", "Name of the ArangoDeployment operator deployment")
	pflag.StringVar(&options.DeploymentReplicationOperatorName, "deployment-replication-operator-name", "arango-deployment-replication-operator", "Name of the ArangoDeploymentReplication operator deployment")
	pflag.StringVar(&options.StorageOperatorName, "storage-operator-name", "arango-storage-operator", "Name of the ArangoLocalStorage operator deployment")
	pflag.StringVar(&options.BackupOperatorName, "backup-operator-name", "arango-backup-operator", "Name of the ArangoBackup operator deployment")
	pflag.BoolVar(&options.RBAC, "rbac", true, "Use role based access control")
	pflag.BoolVar(&options.AllowChaos, "allow-chaos", false, "If set, allows chaos in deployments")

	pflag.Parse()
}

type TemplateOptions struct {
	Version               string
	Image                 string
	ImagePullPolicy       string
	RBAC                  bool
	RBACFilterStart       string
	RBACFilterEnd         string
	Deployment            ResourceOptions
	DeploymentReplication ResourceOptions
	Storage               ResourceOptions
	Backup                ResourceOptions
	Test                  CommonOptions
}

type CommonOptions struct {
	Namespace          string
	RoleName           string
	RoleBindingName    string
	ServiceAccountName string
	ServiceType        string
}

type ResourceOptions struct {
	Create                 string
	FilterStart            string
	FilterEnd              string
	User                   CommonOptions
	Operator               CommonOptions
	OperatorDeploymentName string
	AllowChaos             string
}

func main() {
	// Check options
	if options.Namespace == "" {
		log.Fatal("--namespace not specified.")
	}
	if options.Image == "" {
		log.Fatal("--image not specified.")
	}

	// Fetch image sha256
	if options.ImageSHA256 {
		cmd := exec.Command(
			"docker",
			"inspect",
			"--format={{index .RepoDigests 0}}",
			options.Image,
		)
		result, err := cmd.CombinedOutput()
		if err != nil {
			log.Println(string(result))
			log.Fatalf("Failed to fetch image SHA256: %v", err)
		}
		options.Image = strings.TrimSpace(string(result))
	}

	// Prepare templates to include
	templateInfoSet := map[string]TemplateGroup{
		"crd":                    TemplateGroup{ChartName: "kube-arangodb-crd", Templates: crdTemplateNames},
		"deployment":             TemplateGroup{ChartName: "kube-arangodb", Templates: deploymentTemplateNames},
		"deployment-replication": TemplateGroup{ChartName: "kube-arangodb", Templates: deploymentReplicationTemplateNames},
		"storage":                TemplateGroup{ChartName: "kube-arangodb-storage", Templates: storageTemplateNames},
		"backup":                 TemplateGroup{ChartName: "kube-arangodb-backup", Templates: backupTemplateNames},
		"test":                   TemplateGroup{ChartName: "", Templates: testTemplateNames},
	}

	// Read VERSION
	version, err := ioutil.ReadFile("VERSION")
	if err != nil {
		log.Fatalf("Failed to read VERSION file: %v", err)
	}

	// Prepare chart tars
	chartTarBufs := make(map[string]*bytes.Buffer)
	chartTars := make(map[string]*tar.Writer)
	for groupName := range chartTemplateGroups {
		buf := &bytes.Buffer{}
		chartTarBufs[groupName] = buf
		chartTars[groupName] = tar.NewWriter(buf)
	}

	// Process templates
	templateOptions := TemplateOptions{
		Version:         strings.TrimSpace(string(version)),
		Image:           options.Image,
		ImagePullPolicy: options.ImagePullPolicy,
		RBAC:            options.RBAC,
		Deployment: ResourceOptions{
			Create: "true",
			User: CommonOptions{
				Namespace:          options.Namespace,
				RoleName:           "arango-deployments",
				RoleBindingName:    "arango-deployments",
				ServiceAccountName: "default",
			},
			Operator: CommonOptions{
				Namespace:          options.Namespace,
				RoleName:           "arango-deployment-operator",
				RoleBindingName:    "arango-deployment-operator",
				ServiceAccountName: "default",
				ServiceType:        "ClusterIP",
			},
			OperatorDeploymentName: "arango-deployment-operator",
			AllowChaos:             strconv.FormatBool(options.AllowChaos),
		},
		DeploymentReplication: ResourceOptions{
			Create: "true",
			User: CommonOptions{
				Namespace:          options.Namespace,
				RoleName:           "arango-deployment-replications",
				RoleBindingName:    "arango-deployment-replications",
				ServiceAccountName: "default",
			},
			Operator: CommonOptions{
				Namespace:          options.Namespace,
				RoleName:           "arango-deployment-replication-operator",
				RoleBindingName:    "arango-deployment-replication-operator",
				ServiceAccountName: "default",
				ServiceType:        "ClusterIP",
			},
			OperatorDeploymentName: "arango-deployment-replication-operator",
		},
		Storage: ResourceOptions{
			Create: "true",
			User: CommonOptions{
				Namespace:          options.Namespace,
				RoleName:           "arango-storages",
				RoleBindingName:    "arango-storages",
				ServiceAccountName: "default",
			},
			Operator: CommonOptions{
				Namespace:          "kube-system",
				RoleName:           "arango-storage-operator",
				RoleBindingName:    "arango-storage-operator",
				ServiceAccountName: "arango-storage-operator",
				ServiceType:        "ClusterIP",
			},
			OperatorDeploymentName: "arango-storage-operator",
		},
		Backup: ResourceOptions{
			Create: "true",
			User: CommonOptions{
				Namespace:          options.Namespace,
				RoleName:           "arango-backups",
				RoleBindingName:    "arango-backups",
				ServiceAccountName: "default",
			},
			Operator: CommonOptions{
				Namespace:          options.Namespace,
				RoleName:           "arango-backup-operator",
				RoleBindingName:    "arango-backup-operator",
				ServiceAccountName: "default",
				ServiceType:        "ClusterIP",
			},
			OperatorDeploymentName: "arango-backup-operator",
		},
		Test: CommonOptions{
			Namespace:          options.Namespace,
			RoleName:           "arango-operator-test",
			RoleBindingName:    "arango-operator-test",
			ServiceAccountName: "default",
		},
	}
	chartTemplateOptions := TemplateOptions{
		Version:         strings.TrimSpace(string(version)),
		RBACFilterStart: "{{- if .Values.RBAC.Create }}",
		RBACFilterEnd:   "{{- end }}",
		Image:           "{{ .Values.Image }}",
		ImagePullPolicy: "{{ .Values.ImagePullPolicy }}",
		Deployment: ResourceOptions{
			Create:      "{{ .Values.Deployment.Create }}",
			FilterStart: "{{- if .Values.Deployment.Create }}",
			FilterEnd:   "{{- end }}",
			User: CommonOptions{
				Namespace:          "{{ .Release.Namespace }}",
				RoleName:           `{{ printf "%s-%s" .Release.Name "deployments" | trunc 63 | trimSuffix "-" }}`,
				RoleBindingName:    `{{ printf "%s-%s" .Release.Name "deployments" | trunc 63 | trimSuffix "-" }}`,
				ServiceAccountName: "{{ .Values.Deployment.User.ServiceAccountName }}",
			},
			Operator: CommonOptions{
				Namespace:          "{{ .Release.Namespace }}",
				RoleName:           `{{ printf "%s-%s" .Release.Name "deployment-operator" | trunc 63 | trimSuffix "-" }}`,
				RoleBindingName:    `{{ printf "%s-%s" .Release.Name "deployment-operator" | trunc 63 | trimSuffix "-" }}`,
				ServiceAccountName: "{{ .Values.Deployment.Operator.ServiceAccountName }}",
				ServiceType:        "{{ .Values.Deployment.Operator.ServiceType }}",
			},
			OperatorDeploymentName: "arango-deployment-operator", // Fixed name because only 1 is allowed per namespace
			AllowChaos:             "{{ .Values.Deployment.AllowChaos }}",
		},
		DeploymentReplication: ResourceOptions{
			Create:      "{{ .Values.DeploymentReplication.Create }}",
			FilterStart: "{{- if .Values.DeploymentReplication.Create }}",
			FilterEnd:   "{{- end }}",
			User: CommonOptions{
				Namespace:          "{{ .Release.Namespace }}",
				RoleName:           `{{ printf "%s-%s" .Release.Name "deployment-replications" | trunc 63 | trimSuffix "-" }}`,
				RoleBindingName:    `{{ printf "%s-%s" .Release.Name "deployment-replications" | trunc 63 | trimSuffix "-" }}`,
				ServiceAccountName: "{{ .Values.DeploymentReplication.User.ServiceAccountName }}",
			},
			Operator: CommonOptions{
				Namespace:          "{{ .Release.Namespace }}",
				RoleName:           `{{ printf "%s-%s" .Release.Name "deployment-replication-operator" | trunc 63 | trimSuffix "-" }}`,
				RoleBindingName:    `{{ printf "%s-%s" .Release.Name "deployment-replication-operator" | trunc 63 | trimSuffix "-" }}`,
				ServiceAccountName: "{{ .Values.DeploymentReplication.Operator.ServiceAccountName }}",
				ServiceType:        "{{ .Values.DeploymentReplication.Operator.ServiceType }}",
			},
			OperatorDeploymentName: "arango-deployment-replication-operator", // Fixed name because only 1 is allowed per namespace
		},
		Storage: ResourceOptions{
			Create:      "{{ .Values.Storage.Create }}",
			FilterStart: "{{- if .Values.Storage.Create }}",
			FilterEnd:   "{{- end }}",
			User: CommonOptions{
				Namespace:          "{{ .Release.Namespace }}",
				RoleName:           `{{ printf "%s-%s" .Release.Name "storages" | trunc 63 | trimSuffix "-" }}`,
				RoleBindingName:    `{{ printf "%s-%s" .Release.Name "storages" | trunc 63 | trimSuffix "-" }}`,
				ServiceAccountName: "{{ .Values.Storage.User.ServiceAccountName }}",
			},
			Operator: CommonOptions{
				Namespace:          "kube-system",
				RoleName:           `{{ printf "%s-%s" .Release.Name "storage-operator" | trunc 63 | trimSuffix "-" }}`,
				RoleBindingName:    `{{ printf "%s-%s" .Release.Name "storage-operator" | trunc 63 | trimSuffix "-" }}`,
				ServiceAccountName: "{{ .Values.Storage.Operator.ServiceAccountName }}",
				ServiceType:        "{{ .Values.Storage.Operator.ServiceType }}",
			},
			OperatorDeploymentName: "arango-storage-operator", // Fixed name because only 1 is allowed per namespace
		},
		Backup: ResourceOptions{
			Create:      "{{ .Values.Backup.Create }}",
			FilterStart: "{{- if .Values.Backup.Create }}",
			FilterEnd:   "{{- end }}",
			User: CommonOptions{
				Namespace:          "{{ .Release.Namespace }}",
				RoleName:           `{{ printf "%s-%s" .Release.Name "backup" | trunc 63 | trimSuffix "-" }}`,
				RoleBindingName:    `{{ printf "%s-%s" .Release.Name "backup" | trunc 63 | trimSuffix "-" }}`,
				ServiceAccountName: "{{ .Values.Backup.User.ServiceAccountName }}",
			},
			Operator: CommonOptions{
				Namespace:          "{{ .Release.Namespace }}",
				RoleName:           `{{ printf "%s-%s" .Release.Name "backup-operator" | trunc 63 | trimSuffix "-" }}`,
				RoleBindingName:    `{{ printf "%s-%s" .Release.Name "backup-operator" | trunc 63 | trimSuffix "-" }}`,
				ServiceAccountName: "{{ .Values.Backup.Operator.ServiceAccountName }}",
				ServiceType:        "{{ .Values.Backup.Operator.ServiceType }}",
			},
			OperatorDeploymentName: "arango-backup-operator", // Fixed name because only 1 is allowed per namespace
		},
	}

	for group, templateGroup := range templateInfoSet {
		// Build standalone yaml file for this group
		{
			output := &bytes.Buffer{}
			for _, tempInfo := range templateGroup.Templates {
				if tempInfo.Predicate == nil || tempInfo.Predicate(templateOptions, false) {
					name := tempInfo.Name
					t, err := template.New(name).ParseFiles(filepath.Join(options.TemplatesDir, group, name))
					if err != nil {
						log.Fatalf("Failed to parse template %s: %v", name, err)
					}
					// Execute to tmp buffer
					tmpBuf := &bytes.Buffer{}
					t.Execute(tmpBuf, templateOptions)
					// Add tmp buffer to output, unless empty
					if strings.TrimSpace(tmpBuf.String()) != "" {
						if output.Len() > 0 {
							output.WriteString("\n---\n\n")
						}
						output.WriteString(fmt.Sprintf("## %s/%s\n", group, name))
						tmpBuf.WriteTo(output)
						output.WriteString("\n")
					}
				}
			}

			// Save output
			if output.Len() > 0 {
				outputDir, err := filepath.Abs("manifests")
				if err != nil {
					log.Fatalf("Failed to get absolute output dir: %v\n", err)
				}
				outputPath := filepath.Join(outputDir, "arango-"+group+options.OutputSuffix+".yaml")
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					log.Fatalf("Failed to create output directory: %v\n", err)
				}
				if err := ioutil.WriteFile(outputPath, output.Bytes(), 0644); err != nil {
					log.Fatalf("Failed to write output file: %v\n", err)
				}
			}
		}

		// Build helm template file for this group
		{
			output := &bytes.Buffer{}
			for _, tempInfo := range templateGroup.Templates {
				if tempInfo.Predicate == nil || tempInfo.Predicate(chartTemplateOptions, true) {
					name := tempInfo.Name
					t, err := template.New(name).ParseFiles(filepath.Join(options.TemplatesDir, group, name))
					if err != nil {
						log.Fatalf("Failed to parse template %s: %v", name, err)
					}
					// Execute to tmp buffer
					tmpBuf := &bytes.Buffer{}
					t.Execute(tmpBuf, chartTemplateOptions)
					// Add tmp buffer to output, unless empty
					if strings.TrimSpace(tmpBuf.String()) != "" {
						if output.Len() > 0 {
							output.WriteString("\n---\n\n")
						}
						output.WriteString(fmt.Sprintf("## %s/%s\n", group, name))
						tmpBuf.WriteTo(output)
						output.WriteString("\n")
					}
				}
			}

			// Save output
			if output.Len() > 0 {
				tarPath := path.Join(templateGroup.ChartName, "templates", group+".yaml")
				hdr := &tar.Header{
					Name: tarPath,
					Mode: 0644,
					Size: int64(output.Len()),
				}
				tw := chartTars[templateGroup.ChartName]
				if err := tw.WriteHeader(hdr); err != nil {
					log.Fatal(err)
				}
				if _, err := tw.Write(output.Bytes()); err != nil {
					log.Fatal(err)
				}
			}
		}
	}

	// Build Chart files
	for groupName, chartTemplates := range chartTemplateGroups {
		for name, templateSource := range chartTemplates {
			output := &bytes.Buffer{}
			if strings.HasSuffix(name, ".txt") {
				// Plain text
				output.WriteString(templateSource)
			} else {
				// Template
				t, err := template.New(name).Funcs(tmplFuncs).Parse(templateSource)
				if err != nil {
					log.Fatalf("Failed to parse template %s: %v", name, err)
				}
				t.Execute(output, templateOptions)
			}

			// Save output
			tarPath := path.Join(groupName, name)
			hdr := &tar.Header{
				Name: tarPath,
				Mode: 0644,
				Size: int64(output.Len()),
			}
			tw := chartTars[groupName]
			if err := tw.WriteHeader(hdr); err != nil {
				log.Fatal(err)
			}
			if _, err := tw.Write(output.Bytes()); err != nil {
				log.Fatal(err)
			}
		}
	}

	// Save charts
	for groupName, tw := range chartTars {
		if err := tw.Close(); err != nil {
			log.Fatal(err)
		}
		// Gzip tarball
		tarBytes := chartTarBufs[groupName].Bytes()
		output := &bytes.Buffer{}
		gw := gzip.NewWriter(output)
		if _, err := gw.Write(tarBytes); err != nil {
			log.Fatal(err)
		}
		gw.Close()
		outputDir, err := filepath.Abs("bin/charts")
		if err != nil {
			log.Fatalf("Failed to get absolute output dir: %v\n", err)
		}
		outputPath := filepath.Join(outputDir, groupName+".tgz")
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			log.Fatalf("Failed to create output directory: %v\n", err)
		}
		if err := ioutil.WriteFile(outputPath, output.Bytes(), 0644); err != nil {
			log.Fatalf("Failed to write output file: %v\n", err)
		}
	}
}
