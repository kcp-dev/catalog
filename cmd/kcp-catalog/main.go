/*
Copyright 2021 The KCP Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"os"

	"github.com/spf13/cobra"

	"k8s.io/component-base/cli"
	"k8s.io/component-base/version"

	"github.com/kcp-dev/kcp/pkg/cmd/help"
)

func main() {
	cmd := &cobra.Command{
		Use:   "kcp-catalog",
		Short: "KCP Catalog",
		Long: help.Doc(`
			kcp-catalog is a CLI tool to manage Catalog API objects.
		`),
	}

	// TODO(dinhxuanvu): Use kubeconfig flag to get access to the kcp cluster.
	// Later, potentially expand to other options such as KUBECONFIG env or .kcp
	// directory
	cmd.PersistentFlags().String("kubeconfig", ".kubeconfig", "kubeconfig file used to contact the cluster.")
	cmd.AddCommand(bind.NewCmd())

	help.FitTerminal(cmd.OutOrStdout())

	if v := version.Get().String(); len(v) == 0 {
		cmd.Version = "<unknown>"
	} else {
		cmd.Version = v
	}

	os.Exit(cli.Run(cmd))
}
