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
	# binds to the mentioned catalog entry in the command, e.g the below command will create
 	# APIBindings referenced in catalog entry "certificates" present in "root:catalog:cert-manager" workspace.
 	%[1]s bind catalogentry root:catalog:cert-manager:certificates
	`
)

func New(streams genericclioptions.IOStreams) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:              "bind",
		Short:            "Operations related to binding with API",
		SilenceUsage:     true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	bindOpts := NewBindOptions(streams)
	bindCmd := &cobra.Command{
		Use:          "catalogentry <workspace_path:catalogentry-name>",
		Short:        "Bind to a Catalog Entry",
		Example:      fmt.Sprintf(bindExampleUses, "kubectl catalog"),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := bindOpts.Complete(args); err != nil {
				return err
			}
			if err := bindOpts.Validate(); err != nil {
				return err
			}
			return bindOpts.Run(cmd.Context())
		},
	}
	bindOpts.BindFlags(bindCmd)
	cmd.AddCommand(bindCmd)
	return cmd, nil
}
