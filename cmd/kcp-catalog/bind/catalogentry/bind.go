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
	"context"
	"time"

	"errors"
	"fmt"
	"strings"

	kcpclienthelper "github.com/kcp-dev/apimachinery/pkg/client"
	catalogv1alpha1 "github.com/kcp-dev/catalog/api/v1alpha1"
	apisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	pluginhelpers "github.com/kcp-dev/kcp/pkg/cliplugins/helpers"
	"github.com/kcp-dev/logicalcluster/v2"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlcfg "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/kcp-dev/kcp/pkg/cliplugins/base"
	"k8s.io/client-go/rest"
)

// BindOptions contains the options for creating APIBindings for CE
type BindOptions struct {
	*base.Options
	// CatalogEntryRef is the argument accepted by the command. It contains the
	// reference to where APIExport exists. For ex: <absolute_ref_to_workspace>:<apiexport>.
	CatalogEntryRef string
}

// NewBindOptions returns new BindOptions.
func NewBindOptions(streams genericclioptions.IOStreams) *BindOptions {
	return &BindOptions{
		Options: base.NewOptions(streams),
	}
}

// BindFlags binds fields to cmd's flagset.
func (b *BindOptions) BindFlags(cmd *cobra.Command) {
	b.Options.BindFlags(cmd)
}

// Complete ensures all fields are initialized.
func (b *BindOptions) Complete(args []string) error {
	if err := b.Options.Complete(); err != nil {
		return err
	}

	if len(args) > 0 {
		b.CatalogEntryRef = args[0]
	}
	return nil
}

// Validate validates the BindOptions are complete and usable.
func (b *BindOptions) Validate() error {
	if b.CatalogEntryRef == "" {
		return errors.New("`root:ws:catalogentry_object` reference to bind is required as an argument")
	}

	if !strings.HasPrefix(b.CatalogEntryRef, "root") || !logicalcluster.New(b.CatalogEntryRef).IsValid() {
		return fmt.Errorf("fully qualified reference to workspace where catalog entry exists is required. The format is `root:<ws>:<catalogentry>`")
	}

	return b.Options.Validate()
}

// Run creates an apibinding for the user.
func (b *BindOptions) Run(ctx context.Context) error {
	config, err := b.ClientConfig.ClientConfig()
	if err != nil {
		return err
	}

	_, currentClusterName, err := pluginhelpers.ParseClusterURL(config.Host)
	if err != nil {
		return err
	}

	// get the base config, which is needed for creation of clients.
	baseConfig, err := ctrlcfg.GetConfigWithContext("base")
	if err != nil {
		return fmt.Errorf("unable to get base config %v", err)
	}

	// get the entry referenced in the command to which the user wants to bind.
	path, entryName := logicalcluster.New(b.CatalogEntryRef).Split()
	client, err := newCatalogClient(baseConfig, path)
	if err != nil {
		return err
	}

	entry := catalogv1alpha1.CatalogEntry{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: entryName}, &entry)
	if err != nil {
		return fmt.Errorf("cannot find the catalog entry %q referenced in the command in the workspace %q", entryName, path)
	}

	kcpclient, err := newKCPClusterClient(baseConfig, currentClusterName)
	if err != nil {
		return err
	}

	apibindings := []apisv1alpha1.APIBinding{}
	for _, ref := range entry.Spec.Exports {
		// check if ref is valid. Skip if invalid by logging error.
		if ref.Workspace.Path == "" || ref.Workspace.ExportName == "" {
			klog.Infof("invalid reference %q/%q", ref.Workspace.Path, ref.Workspace.ExportName)
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

		apibindings = append(apibindings, *apibinding)
	}

	// Create bindings to the target workspace
	for _, binding := range apibindings {
		err := kcpclient.Create(ctx, &binding)
		if err != nil {
			// If an APIBinding already exists, intentionally not updating it since we would not like reset AcceptablePermissionClaims.
			klog.Infof("Failed to create API binding %s: %w", binding.Name, err)
			continue
		}

		if err := wait.PollImmediate(time.Millisecond*500, time.Second*5, func() (done bool, err error) {
			createdBinding := apisv1alpha1.APIBinding{}
			err = kcpclient.Get(ctx, types.NamespacedName{Name: binding.Name}, &createdBinding)
			if err != nil {
				return false, err
			}
			if createdBinding.Status.Phase == apisv1alpha1.APIBindingPhaseBound {
				return true, nil
			}
			return false, nil
		}); err != nil {
			return fmt.Errorf("could not bind %s: %w", binding.Name, err)
		}
	}
	if _, err := fmt.Fprintf(b.Out, "%s created and bound to catalog entry.\n", entryName); err != nil {
		return err
	}
	return nil
}

func newKCPClusterClient(cfg *rest.Config, clusterName logicalcluster.Name) (client.Client, error) {
	scheme := runtime.NewScheme()
	err := apisv1alpha1.AddToScheme(scheme)
	if err != nil {
		return nil, err
	}
	return client.New(kcpclienthelper.SetCluster(rest.CopyConfig(cfg), clusterName), client.Options{
		Scheme: scheme,
	})
}

func newCatalogClient(cfg *rest.Config, clusterName logicalcluster.Name) (client.Client, error) {
	scheme := runtime.NewScheme()
	err := catalogv1alpha1.AddToScheme(scheme)
	if err != nil {
		return nil, err
	}

	return client.New(kcpclienthelper.SetCluster(rest.CopyConfig(cfg), clusterName), client.Options{
		Scheme: scheme,
	})
}
