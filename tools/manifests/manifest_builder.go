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
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/pflag"
)

var (
	options struct {
		OutputSuffix string
		TemplatesDir string

		Namespace              string
		Image                  string
		ImagePullPolicy        string
		ImageSHA256            bool
		DeploymentOperatorName string
		StorageOperatorName    string
		RBAC                   bool
	}
	deploymentTemplateNames = []string{
		"rbac.yaml",
		"deployment.yaml",
	}
	storageTemplateNames = []string{
		"rbac.yaml",
		"deployment.yaml",
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
	pflag.StringVar(&options.StorageOperatorName, "storage-operator-name", "arango-storage-operator", "Name of the ArangoLocalStorage operator deployment")
	pflag.BoolVar(&options.RBAC, "rbac", true, "Use role based access control")

	pflag.Parse()
}

type TemplateOptions struct {
	Image           string
	ImagePullPolicy string
	RBAC            bool
	Deployment      ResourceOptions
	Storage         ResourceOptions
}

type CommonOptions struct {
	Namespace          string
	RoleName           string
	RoleBindingName    string
	ServiceAccountName string
}

type ResourceOptions struct {
	User                   CommonOptions
	Operator               CommonOptions
	OperatorDeploymentName string
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
	templateNameSet := map[string][]string{
		"deployment": deploymentTemplateNames,
		"storage":    storageTemplateNames,
	}

	// Process templates
	templateOptions := TemplateOptions{
		Image:           options.Image,
		ImagePullPolicy: options.ImagePullPolicy,
		RBAC:            options.RBAC,
		Deployment: ResourceOptions{
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
			},
			OperatorDeploymentName: "arango-deployment-operator",
		},
		Storage: ResourceOptions{
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
			},
			OperatorDeploymentName: "arango-storage-operator",
		},
	}
	for group, templateNames := range templateNameSet {
		output := &bytes.Buffer{}
		for i, name := range templateNames {
			t, err := template.New(name).ParseFiles(filepath.Join(options.TemplatesDir, group, name))
			if err != nil {
				log.Fatalf("Failed to parse template %s: %v", name, err)
			}
			if i > 0 {
				output.WriteString("\n---\n\n")
			}
			output.WriteString(fmt.Sprintf("## %s/%s\n", group, name))
			t.Execute(output, templateOptions)
			output.WriteString("\n")
		}

		// Save output
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
