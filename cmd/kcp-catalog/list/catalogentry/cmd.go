/*
Copyright 2022 The KCP Authors.

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

package catalogentry

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	bindExampleUses = `
	# list the catalog entries and the respective APIs exported by the catalog entry for the specified workspace.
 	%[1]s list catalogentry root:catalog:cert-manager
	`
)

func New(streams genericclioptions.IOStreams) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:              "list",
		Short:            "Operations related to listing APIs",
		SilenceUsage:     true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	listOpts := NewListOptions(streams)
	listCmd := &cobra.Command{
		Use:          "catalogentry <catalog_workspace_path> [name_of_catalog]",
		Short:        "Bind to a Catalog Entry",
		Example:      fmt.Sprintf(bindExampleUses, "kubectl kcp"),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := listOpts.Complete(args); err != nil {
				return err
			}
			if err := listOpts.Validate(); err != nil {
				return err
			}
			return listOpts.Run(cmd.Context())
		},
	}
	listOpts.BindFlags(listCmd)
	cmd.AddCommand(listCmd)
	return cmd, nil
}
