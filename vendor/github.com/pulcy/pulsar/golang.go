// Copyright (c) 2016 Pulcy.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"
	"sync"

	"github.com/spf13/cobra"

	"github.com/pulcy/pulsar/golang"
)

var (
	goCmd = &cobra.Command{
		Use:   "go",
		Short: "Execute `go get` with cache support",
		Run:   UsageFunc,
	}
	goGetCmd = &cobra.Command{
		Use:   "get",
		Short: "Execute `go get` with cache support",
		Run:   runGoGet,
	}
	goPathCmd = &cobra.Command{
		Use:   "path",
		Short: "Create a local GOPATH for the repository in the current directory",
		Run:   runGoPath,
	}
	goVendorCmd = &cobra.Command{
		Use:   "vendor",
		Short: "Update a package in the vendor directory",
		Run:   runGoVendor,
	}
	goFlattenCmd = &cobra.Command{
		Use:   "flatten",
		Short: "Copy directories found in the given vendor directory to the GOPATH and flatten all vendor directories found in the GOPATH",
		Long:  "Copy directories found in the given vendor directory to the GOPATH and flatten all vendor directories found in the GOPATH, or flatten vendor directories in a given target directory.",
		Run:   runGoFlatten,
	}

	vendorDir     string
	vendorFlatten bool
	goPathPkg     string
)

func init() {
	goCmd.PersistentFlags().StringVarP(&vendorDir, "vendor-dir", "V", golang.DefaultVendorDir, "Specify vendor directory")
	goPathCmd.Flags().StringVarP(&goPathPkg, "package", "p", "", "If set, use this package instead of the origin URL from the local repo")
	goVendorCmd.Flags().BoolVarP(&vendorFlatten, "flatten", "F", false, "If set, flattens all directories in the vendor directory")

	mainCmd.AddCommand(goCmd)
	goCmd.AddCommand(goGetCmd)
	goCmd.AddCommand(goPathCmd)
	goCmd.AddCommand(goVendorCmd)
	goCmd.AddCommand(goFlattenCmd)
}

func runGoGet(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		CommandError(cmd, "Expected <package> argument\n")
	} else {
		wg := sync.WaitGroup{}
		errors := make(chan error, len(args))
		for _, pkg := range args {
			wg.Add(1)
			go func(pkg string) {
				defer wg.Done()
				gogetFlags := &golang.GetFlags{Package: pkg}
				if err := golang.Get(log, gogetFlags); err != nil {
					errors <- err
				}
			}(pkg)
		}
		wg.Wait()
		close(errors)
		failed := false
		for err := range errors {
			Printf("Go get failed: %v\n", err)
			failed = true
		}
		if failed {
			os.Exit(1)
		}
	}
}

func runGoPath(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		CommandError(cmd, "Expected no arguments\n")
	} else {
		flags := &golang.GoPathFlags{
			Package: goPathPkg,
		}
		if err := golang.CreateLocalGoPath(log, flags); err != nil {
			Printf("creating GOPATH failed: %#v\n", err)
			os.Exit(1)
		}
	}
}

func runGoVendor(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		CommandError(cmd, "Expected <package> argument\n")
	} else {
		wg := sync.WaitGroup{}
		errors := make(chan error, len(args))
		for _, pkg := range args {
			wg.Add(1)
			go func(pkg string) {
				defer wg.Done()
				goVendorFlags := &golang.VendorFlags{Package: pkg, VendorDir: vendorDir}
				if err := golang.Vendor(log, goVendorFlags); err != nil {
					errors <- err
				}
			}(pkg)
		}
		wg.Wait()
		close(errors)
		failed := false
		for err := range errors {
			Printf("Go vendor failed: %v\n", err)
			failed = true
		}
		if failed {
			os.Exit(1)
		}
		if vendorFlatten {
			flags := &golang.FlattenFlags{
				VendorDir: vendorDir,
				TargetDir: vendorDir,
			}
			if err := golang.Flatten(log, flags); err != nil {
				Printf("flatten failed: %#v\n", err)
				os.Exit(1)
			}
		}
	}
}

func runGoFlatten(cmd *cobra.Command, args []string) {
	targetDir := ""
	if len(args) == 1 {
		targetDir = args[0]
	}
	flags := &golang.FlattenFlags{
		VendorDir: vendorDir,
		TargetDir: targetDir,
	}
	if err := golang.Flatten(log, flags); err != nil {
		Printf("flatten failed: %#v\n", err)
		os.Exit(1)
	}
}
