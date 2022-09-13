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

package bind

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	apisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	kcpclient "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
	"github.com/kcp-dev/kcp/pkg/cmd/help"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func NewCmd() {
	bindCmd := &cobra.Command{
		Use:   "bind",
		Short: "Bind a catalog entry to a workspace",
		Long: help.Doc(`
			Bind a catalog entry to a workspace

			The bind process creates an APIBinding using the ExportReference information
			in the CatalogEntry. The RBAC that is required to use APIBinding is not created
			as a part of this process and must be handled separately.
		`),
		RunE: runBindCmdFunc,
	}

	// TODO(dinhxuanvu): Use flags to accept inputs for now. Need to reevaluate this
	// approach later to see if this is a good UX
	bindCmd.Flags().StringP("entryname", "-e", "", "the name of the catalog entry")
	if err := bindCmd.MarkFlagRequired("entryname"); err != nil {
		klog.Fatalf("Failed to set required `entryname` flag for `bind` command: %s", err.Error())
	}
	bindCmd.Flags().StringP("workspace", "-w", "", "the workspace path where catalog entry locates")
	if err := bindCmd.MarkFlagRequired("workspace"); err != nil {
		klog.Fatalf("Failed to set required `workspace` flag for `bind` command: %s", err.Error())
	}
	bindCmd.Flags().StringP("target", "-t", "", "the targeted workspace where the entry should be binded to")
	if err := bindCmd.MarkFlagRequired("target"); err != nil {
		klog.Fatalf("Failed to set required `target` flag for `bind` command: %s", err.Error())
	}

	return bindCmd
}

// runBindCmdFunc creates APIBindings from the list of References in CatalogEntry
// User is responsible to create all of permissons/RBAC necessary to enable
// APIBindings to work in the particular workspace.
// TODO(dinhxuanvu): construct a permission/RBAC request model
func runBindCmdFunc(cmd *cobra.Command, _ []string) error {
	// Get kubeconfig to access the cluster
	// Assume this cluster has kcp all setup and running
	kubeconfigPath := cmd.Flag("kubeconfig").Value.String()
	configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}, nil)
	config, err := configLoader.ClientConfig()
	if err != nil {
		return err
	}

	kcpClient, err := kcpclient.NewClusterForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kcp client: %w", err)
	}

	entryName := cmd.Flag("entryname").Value.String()
	workspace := cmd.Flag("workspace").Value.String()
	targetWS := cmd.Flag("target").Value.String()
	// Get the catalog entry from the worksplace using RESTClient for now
	// TODO(dinhxuanvu): potentially want to add catalog entry into typed clientset
	// or create its own clienset in the repo
	// TODO(dinhxuanvu): figure out the AbsPath for workspace and api info
	entry, err := kcpClient.RESTClient().
		Get().
		AbsPath("/apis/<api>/<version>").
		Resource("catalogentries").
		Name(entryName).
		DoRaw(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to retrieve catalog entry: %w", err)
	}

	// Ensure that the target workspace exists
	wp, err := kcpClient.TenancyV1beta1().Workspaces().Get(context.TODO(), targetWS, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to retrieve target workspace: %w", err)
	}

	// Generate the APIBinding for each reference in catalog entry
	apiBindings := []apisv1alpha1.APIBinding{}
	for _, ref := range entry.References {
		// Check if ref is valid. Skip if invalid
		if ref.Workspace.Path == "" || ref.Workspace.ExportName == "" {
			klog.Infof("Invalid reference: %q/%q", ref.Workspace.Path, ref.Workspace.ExportName)
			continue
		}
		apibinding := &apisv1alpha1.APIBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: ref.Workspace.ExportName,
			},
			Spec: apisv1alpha1.APIBindingSpec{
				Reference: ref,
			},
		}
		apiBindings = append(apiBindings, apibinding)
	}

	// Apply the APIBinding(s) to the target workspace
	for _, binding := range apiBindings {
		created, err := kcpClient.ApisV1alpha1().APIBindings().Create(context.TODO(), binding, metav1.CreateOptions{})
		if err != nil {
			klog.Infof("Failed to create API binding %s: %w", binding.Name, err)
			continue
		}

		if apierrors.IsAlreadyExists(err) {
			klog.Infof("APIBinding %s already exists in workspace %s", binding.Name, targetWS)
		}

		existing, err := kcpClient.ApisV1alpha1().APIBindings().Get(context.TODO(), binding.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		klog.Infof("Updating API binding %s", binding.Name)
		existing.Spec = binding.Spec
		if _, err := kcpClient.ApisV1alpha1().APIBindings().Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("could not update API binding %s in workspace %s: %w", existing.Name, targetWS, err)
		}
	}
}
