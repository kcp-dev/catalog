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

package main

import (
	goflags "flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/component-base/version"
	"k8s.io/klog/v2"

	"github.com/kcp-dev/catalog/cmd/kcp-catalog/bind/catalogentry"
	"github.com/kcp-dev/kcp/pkg/cmd/help"
)

func main() {
	cmd := &cobra.Command{
		Use:   "kubectl-catalog",
		Short: "kubectl-catalog",
		Long: help.Doc(`
			kcp is a CLI tool to manage Catalog API objects.
		`),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Setup klog
	fs := goflags.NewFlagSet("klog", goflags.PanicOnError)
	klog.InitFlags(fs)
	cmd.PersistentFlags().AddGoFlagSet(fs)

	if v := version.Get().String(); len(v) == 0 {
		cmd.Version = "<unknown>"
	} else {
		cmd.Version = v
	}

	bindCmd, err := catalogentry.New(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	cmd.AddCommand(bindCmd)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
