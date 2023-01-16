//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package debug_package

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/debug_package/cli"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/generators/kubernetes"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

var rootFactories = []shared.Factory{
	kubernetes.Events(),
	kubernetes.Pods(),
	kubernetes.Secrets(),
	kubernetes.Services(),
	kubernetes.Deployments(),
	kubernetes.AgencyDump(),
}

func InitCommand(cmd *cobra.Command) {
	cli.Register(cmd)
	f := cmd.Flags()

	for _, factory := range rootFactories {
		n := fmt.Sprintf("generator.%s", factory.Name())

		f.Bool(n, factory.Enabled(), fmt.Sprintf("Define if generator %s is enabled", factory.Name()))
	}
}

func GenerateD(cmd *cobra.Command, out io.Writer) error {
	return Generate(cmd, out, rootFactories...)
}

func Generate(cmd *cobra.Command, out io.Writer, factories ...shared.Factory) error {
	tw := tar.NewWriter(out)

	data := bytes.NewBuffer(nil)

	log := NewLogger(data)

	files := make(chan shared.File)

	fileErrors := map[string]error{}
	factoryErrors := map[string]error{}

	done := make(chan struct{})

	n := time.Now()

	go func() {
		defer close(done)
		for file := range files {
			log.Info().Msgf("Fetching file %s", file.Path())
			data, err := file.Write()
			if err != nil {
				fileErrors[file.Path()] = err
				continue
			}

			if err := tw.WriteHeader(&tar.Header{
				Name:       file.Path(),
				ModTime:    n,
				AccessTime: n,
				ChangeTime: n,
				Mode:       0644,
				Uid:        1000,
				Gid:        1000,
				Size:       int64(len(data)),
			}); err != nil {
				fileErrors[file.Path()] = err
				continue
			}

			if _, err := tw.Write(data); err != nil {
				fileErrors[file.Path()] = err
				continue
			}
		}
	}()

	for _, f := range factories {
		ok, _ := cmd.Flags().GetBool(fmt.Sprintf("generator.%s", f.Name()))

		if !ok {
			log.Info().Msgf("Factory %s disabled", f.Name())
			continue
		}

		log.Info().Msgf("Fetching factory %s", f.Name())

		if err := f.Generate(log, files); err != nil {
			factoryErrors[f.Name()] = err
		}
	}

	close(files)

	<-done

	if len(fileErrors) > 0 {
		log.Error().Msgf("%d errors while fetching files:", len(fileErrors))

		parsedErrors := map[string]string{}

		for f, n := range fileErrors {
			parsedErrors[f] = n.Error()
			log.Error().Msgf("\t%s: %s", f, n.Error())
		}

	}

	if len(factoryErrors) > 0 {
		log.Error().Msgf("%d errors while fetching factories:", len(factoryErrors))
		for f, n := range factoryErrors {
			log.Error().Msgf("\t%s: %s", f, n.Error())
		}
	}

	if err := tw.WriteHeader(&tar.Header{
		Name:       "logs",
		ModTime:    n,
		AccessTime: n,
		ChangeTime: n,
		Mode:       0644,
		Uid:        1000,
		Gid:        1000,
		Size:       int64(len(data.Bytes())),
	}); err != nil {
		return err
	}

	if _, err := io.Copy(tw, data); err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}

	if len(fileErrors) > 0 || len(factoryErrors) > 0 {
		return errors.Newf("Error while receiving data")
	}

	return nil
}
