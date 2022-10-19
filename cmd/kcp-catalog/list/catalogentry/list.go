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
	"errors"
	"fmt"
	"io"
	"strings"

	kcpclienthelper "github.com/kcp-dev/apimachinery/pkg/client"
	catalogv1alpha1 "github.com/kcp-dev/catalog/api/v1alpha1"
	apisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	"github.com/kcp-dev/kcp/pkg/cliplugins/base"
	"github.com/kcp-dev/logicalcluster/v2"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlcfg "sigs.k8s.io/controller-runtime/pkg/client/config"
)

// ListOptions contains the options for listing APIs for CE
type ListOptions struct {
	*base.Options
	// CatalogWorkspace refers to the workspace path of the catalog
	// whose catalog entries we are supposed to list.
	CatalogWorkspace string
	// CatalogEntry is the optional parameter specified by the user, whose
	// referenced APIs we are to list.
	CatalogEntry string
}

// NewListOptions returns new ListOptions.
func NewListOptions(streams genericclioptions.IOStreams) *ListOptions {
	return &ListOptions{
		Options: base.NewOptions(streams),
	}
}

// BindFlags binds fields to cmd's flagset.
func (l *ListOptions) BindFlags(cmd *cobra.Command) {
	l.Options.BindFlags(cmd)
}

// Complete ensures all fields are initialized.
func (l *ListOptions) Complete(args []string) error {
	if err := l.Options.Complete(); err != nil {
		return err
	}

	switch len := len(args); len {
	case 1:
		l.CatalogWorkspace = args[0]
	case 2:
		l.CatalogWorkspace = args[0]
		l.CatalogEntry = args[1]
	}
	return nil
}

// Validate validates the BindOptions are complete and usable.
func (l *ListOptions) Validate() error {
	if l.CatalogWorkspace == "" {
		return errors.New("workspace path of the catalog is a required argument.")
	}

	if !strings.HasPrefix(l.CatalogWorkspace, "root") || !logicalcluster.New(l.CatalogWorkspace).IsValid() {
		return fmt.Errorf("fully qualified reference to workspace where catalog exists is required. The format is `root:<catalog_ws>`")
	}
	return l.Options.Validate()
}

// Run lists the referenced catalog entries
func (l *ListOptions) Run(ctx context.Context) error {
	// get the base config, which is needed for creation of clients.
	baseConfig, err := ctrlcfg.GetConfigWithContext("base")
	if err != nil {
		return fmt.Errorf("unable to get base config %v", err)
	}

	client, err := newCatalogClient(baseConfig, logicalcluster.New(l.CatalogWorkspace))
	if err != nil {
		return err
	}

	out := printers.GetNewTabWriter(l.Out)
	defer out.Flush()

	err = printHeaders(out)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	catalogEntries := []catalogv1alpha1.CatalogEntry{}
	allErrors := []error{}

	if l.CatalogEntry != "" {
		catalogEntryObj := catalogv1alpha1.CatalogEntry{}
		err := client.Get(ctx, types.NamespacedName{Name: l.CatalogEntry}, &catalogEntryObj)
		if err != nil {
			return fmt.Errorf("error finding the specified catalogentry %q", l.CatalogEntry)
		}
		catalogEntries = append(catalogEntries, catalogEntryObj)
	} else {
		list := catalogv1alpha1.CatalogEntryList{}
		err := client.List(ctx, &list)
		if err != nil {
			return fmt.Errorf("error listing catalog entries in workspace %q", l.CatalogEntry)
		}
		catalogEntries = append(catalogEntries, list.Items...)
	}

	for _, ce := range catalogEntries {
		for _, apis := range ce.Spec.Exports {

			cl, err := newAPIExportClient(baseConfig, logicalcluster.New(apis.Workspace.Path))
			allErrors = append(allErrors, err)

			exposedSchemas, err := getExposedGV(ctx, cl, apis.Workspace.ExportName)
			if err != nil {
				allErrors = append(allErrors, err)
			}
			if err := printDetails(l.Out, ce.Name, getAPISchema(exposedSchemas)); err != nil {
				allErrors = append(allErrors, err)
			}
		}
	}

	return utilerrors.NewAggregate(allErrors)
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

func printHeaders(out io.Writer) error {
	columnNames := []string{"NAME", "AVAILABLE API"}
	_, err := fmt.Fprintf(out, "%s\n", strings.Join(columnNames, "\t"))
	return err
}

func printDetails(w io.Writer, name, apis string) error {
	_, err := fmt.Fprintf(w, "%s\t%s\n", name, apis)
	return err
}

func newAPIExportClient(cfg *rest.Config, clusterName logicalcluster.Name) (client.Client, error) {
	scheme := runtime.NewScheme()
	if err := apisv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	return client.New(kcpclienthelper.SetCluster(rest.CopyConfig(cfg), clusterName), client.Options{
		Scheme: scheme,
	})
}

func getExposedGV(ctx context.Context, cl client.Client, apiexportName string) ([]string, error) {
	apiExport := apisv1alpha1.APIExport{}
	if err := cl.Get(ctx, types.NamespacedName{Name: apiexportName}, &apiExport); err != nil {
		return nil, err
	}
	return apiExport.Spec.LatestResourceSchemas, nil
}

func getAPISchema(gv []string) string {
	return strings.Join(gv, "\t")
}
