package e2e

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"testing"
	"time"

	kcpclienthelper "github.com/kcp-dev/apimachinery/pkg/client"
	"github.com/stretchr/testify/assert"

	apisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	tenancyv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	"github.com/kcp-dev/kcp/pkg/apis/third_party/conditions/util/conditions"

	"github.com/kcp-dev/logicalcluster/v2"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	catalogv1alpha1 "github.com/kcp-dev/catalog/api/v1alpha1"
)

// The tests in this package expect to be called when:
// - kcp is running
// - a kind cluster is up and running
// - it is hosting a syncer, and the SyncTarget is ready to go
// - the controller-manager from this repo is deployed to kcp
// - that deployment is synced to the kind cluster
// - the deployment is rolled out & ready
//
// We can then check that the controllers defined here are working as expected.

var workspaceName string

func init() {
	rand.Seed(time.Now().Unix())
	flag.StringVar(&workspaceName, "workspace", "", "Workspace in which to run these tests.")
}

func parentWorkspace(t *testing.T) logicalcluster.Name {
	flag.Parse()
	if workspaceName == "" {
		t.Fatal("--workspace cannot be empty")
	}

	return logicalcluster.New(workspaceName)
}

func loadClusterConfig(t *testing.T, clusterName logicalcluster.Name) *rest.Config {
	t.Helper()
	restConfig, err := config.GetConfigWithContext("base")
	if err != nil {
		t.Fatalf("failed to load *rest.Config: %v", err)
	}
	return rest.AddUserAgent(kcpclienthelper.SetCluster(rest.CopyConfig(restConfig), clusterName), t.Name())
}

func loadClient(t *testing.T, clusterName logicalcluster.Name) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add client go to scheme: %v", err)
	}
	if err := tenancyv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add %q to scheme: %v", tenancyv1alpha1.SchemeGroupVersion, err)
	}
	if err := catalogv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add %q to scheme: %v", catalogv1alpha1.GroupVersion, err)
	}
	if err := apisv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add %q to scheme: %v", apisv1alpha1.SchemeGroupVersion, err)
	}
	tenancyClient, err := client.New(loadClusterConfig(t, clusterName), client.Options{Scheme: scheme})
	if err != nil {
		t.Fatalf("failed to create a client: %v", err)
	}
	return tenancyClient
}

func createWorkspace(t *testing.T, clusterName logicalcluster.Name) client.Client {
	t.Helper()
	parent, ok := clusterName.Parent()
	if !ok {
		t.Fatalf("cluster %q has no parent", clusterName)
	}
	c := loadClient(t, parent)
	t.Logf("creating workspace %q", clusterName)
	if err := c.Create(context.TODO(), &tenancyv1alpha1.ClusterWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName.Base(),
		},
		Spec: tenancyv1alpha1.ClusterWorkspaceSpec{
			Type: tenancyv1alpha1.ClusterWorkspaceTypeReference{
				Name: "universal",
				Path: "root",
			},
		},
	}); err != nil {
		t.Fatalf("failed to create workspace: %q: %v", clusterName, err)
	}

	t.Logf("waiting for workspace %q to be ready", clusterName)
	var workspace tenancyv1alpha1.ClusterWorkspace
	if err := wait.PollImmediate(100*time.Millisecond, wait.ForeverTestTimeout, func() (done bool, err error) {
		fetchErr := c.Get(context.TODO(), client.ObjectKey{Name: clusterName.Base()}, &workspace)
		if fetchErr != nil {
			t.Logf("failed to get workspace %q: %v", clusterName, err)
			return false, fetchErr
		}
		var reason string
		if actual, expected := workspace.Status.Phase, tenancyv1alpha1.ClusterWorkspacePhaseReady; actual != expected {
			reason = fmt.Sprintf("phase is %q, not %q", actual, expected)
			t.Logf("not done waiting for workspace %q to be ready: %q", clusterName, reason)
		}
		return reason == "", nil
	}); err != nil {
		t.Fatalf("workspace %q never ready: %v", clusterName, err)
	}

	return createAPIBinding(t, clusterName)
}

func createAPIBinding(t *testing.T, workspaceCluster logicalcluster.Name) client.Client {
	c := loadClient(t, workspaceCluster)
	apiName := "catalog.kcp.dev"
	t.Logf("creating APIBinding %q|%q", workspaceCluster, apiName)
	if err := c.Create(context.TODO(), &apisv1alpha1.APIBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: apiName,
		},
		Spec: apisv1alpha1.APIBindingSpec{
			Reference: apisv1alpha1.ExportReference{
				Workspace: &apisv1alpha1.WorkspaceExportReference{
					Path:       parentWorkspace(t).String(),
					ExportName: apiName,
				},
			},
		},
	}); err != nil {
		t.Fatalf("could not create APIBinding %q|%q: %v", workspaceCluster, apiName, err)
	}

	t.Logf("waiting for APIBinding %q|%q to be bound", workspaceCluster, apiName)
	var apiBinding apisv1alpha1.APIBinding
	if err := wait.PollImmediate(100*time.Millisecond, wait.ForeverTestTimeout, func() (done bool, err error) {
		fetchErr := c.Get(context.TODO(), client.ObjectKey{Name: apiName}, &apiBinding)
		if fetchErr != nil {
			t.Logf("failed to get APIBinding %q|%q: %v", workspaceCluster, apiName, err)
			return false, fetchErr
		}
		var reason string
		if !conditions.IsTrue(&apiBinding, apisv1alpha1.InitialBindingCompleted) {
			condition := conditions.Get(&apiBinding, apisv1alpha1.InitialBindingCompleted)
			if condition != nil {
				reason = fmt.Sprintf("%q: %q", condition.Reason, condition.Message)
			} else {
				reason = "no condition present"
			}
			t.Logf("not done waiting for APIBinding %q|%q to be bound: %q", workspaceCluster, apiName, reason)
		}
		return conditions.IsTrue(&apiBinding, apisv1alpha1.InitialBindingCompleted), nil
	}); err != nil {
		t.Fatalf("APIBinding %q|%q never bound: %v", workspaceCluster, apiName, err)
	}

	return c
}

func createAPIResourceSchema(t *testing.T, c client.Client, workspaceCluster logicalcluster.Name, schema *apisv1alpha1.APIResourceSchema) error {
	apiName := schema.GetName()
	t.Logf("creating APIResourceSchema %q|%q", workspaceCluster, apiName)
	if err := c.Create(context.TODO(), schema); err != nil {
		return fmt.Errorf("could not create APIResourceSchema %q|%q: %v", workspaceCluster, apiName, err)
	}

	t.Logf("waiting for APIResourceSchema %q|%q to be found", workspaceCluster, apiName)
	var apiResourceSchema apisv1alpha1.APIResourceSchema
	if err := wait.PollImmediate(100*time.Millisecond, wait.ForeverTestTimeout, func() (done bool, err error) {
		fetchErr := c.Get(context.TODO(), client.ObjectKey{Name: apiName}, &apiResourceSchema)
		if fetchErr != nil {
			t.Logf("failed to get APIResourceSchema %q|%q: %v", workspaceCluster, apiName, err)
			return false, fetchErr
		}
		return true, nil
	}); err != nil {
		return fmt.Errorf("APIResourceSchema %q|%q not found: %v", workspaceCluster, apiName, err)
	}

	return nil
}

func createAPIExport(t *testing.T, c client.Client, workspaceCluster logicalcluster.Name, export *apisv1alpha1.APIExport) error {
	apiName := export.GetName()
	t.Logf("creating APIExport %q|%q", workspaceCluster, apiName)
	if err := c.Create(context.TODO(), export); err != nil {
		return fmt.Errorf("could not create APIExport %q|%q: %v", workspaceCluster, apiName, err)
	}

	t.Logf("waiting for APIExport %q|%q to be found", workspaceCluster, apiName)
	var apiExport apisv1alpha1.APIExport
	if err := wait.PollImmediate(100*time.Millisecond, wait.ForeverTestTimeout, func() (done bool, err error) {
		fetchErr := c.Get(context.TODO(), client.ObjectKey{Name: apiName}, &apiExport)
		if fetchErr != nil {
			t.Logf("failed to get APIExport %q|%q: %v", workspaceCluster, apiName, err)
			return false, fetchErr
		}
		return true, nil
	}); err != nil {
		return fmt.Errorf("APIExport %q|%q not found: %v", workspaceCluster, apiName, err)
	}

	return nil
}

func createCatalogEntry(t *testing.T, c client.Client, workspaceCluster logicalcluster.Name, entry *catalogv1alpha1.CatalogEntry) (*catalogv1alpha1.CatalogEntry, error) {
	entryName := entry.GetName()
	t.Logf("creating CatalogEntry %q|%q", workspaceCluster, entryName)
	if err := c.Create(context.TODO(), entry); err != nil {
		return nil, fmt.Errorf("could not create CatalogEntry %q|%q: %v", workspaceCluster, entryName, err)
	}

	t.Logf("waiting for CatalogEntry %q|%q to be found", workspaceCluster, entryName)
	var catalogEntry catalogv1alpha1.CatalogEntry
	if err := wait.PollImmediate(100*time.Millisecond, wait.ForeverTestTimeout, func() (done bool, err error) {
		fetchErr := c.Get(context.TODO(), client.ObjectKey{Name: entryName}, &catalogEntry)
		if fetchErr != nil {
			t.Logf("failed to get CatalogEntry %q|%q: %v", workspaceCluster, entryName, err)
			return false, fetchErr
		}
		// Waiting for `APIExportValid` condition which signals the entry has been
		// reconciled by the controller
		if !conditions.Has(&catalogEntry, catalogv1alpha1.APIExportValidType) {
			t.Logf("not done waiting for CatalogEntry %q|%q to be reconciled", workspaceCluster, entryName)
			return false, fmt.Errorf("CatalogEntry %q|%q hasn't been reconciled: %v", workspaceCluster, entryName, err)
		}
		return true, nil
	}); err != nil {
		return nil, fmt.Errorf("CatalogEntry %q|%q not found: %v", workspaceCluster, entryName, err)
	}

	return &catalogEntry, nil
}

const characters = "abcdefghijklmnopqrstuvwxyz"

func randomName() string {
	b := make([]byte, 10)
	for i := range b {
		b[i] = characters[rand.Intn(len(characters))]
	}
	return string(b)
}

// TestValidCatalogEntry verifies that our catalog controller can validate valid
// catalogEntry by adding correct conditions, resources and exportPermissionClaims
func TestValidCatalogEntry(t *testing.T) {
	t.Run("testing valid catalog entry", func(t *testing.T) {
		workspaceCluster := parentWorkspace(t).Join(randomName())
		c := createWorkspace(t, workspaceCluster)

		// Create test apiResourceSchema
		schema := &apisv1alpha1.APIResourceSchema{
			ObjectMeta: metav1.ObjectMeta{
				Name: "today.tests.catalog.kcp.dev",
			},
			Spec: apisv1alpha1.APIResourceSchemaSpec{
				Group: "catalog.kcp.dev",
				Names: apiextensionsv1.CustomResourceDefinitionNames{
					Plural:   "tests",
					Singular: "test",
					Kind:     "Test",
					ListKind: "testlist",
				},
				Scope: apiextensionsv1.ClusterScoped,
				Versions: []apisv1alpha1.APIResourceVersion{{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: runtime.RawExtension{
						Raw: []byte(`{"description":"foo","type":"object"}`),
					},
				}},
			},
		}
		err := createAPIResourceSchema(t, c, workspaceCluster, schema)
		if err != nil {
			t.Fatalf("failed to create APIResourceSchema in cluster %q: %v", workspaceCluster, err)
		}

		// Create test APIExport
		export := &apisv1alpha1.APIExport{
			ObjectMeta: metav1.ObjectMeta{
				Name: "export-test",
			},
			Spec: apisv1alpha1.APIExportSpec{
				LatestResourceSchemas: []string{"tests.catalog.kcp.dev"},
				PermissionClaims: []apisv1alpha1.PermissionClaim{
					{
						GroupResource: apisv1alpha1.GroupResource{Resource: "configmaps"},
					},
				},
			},
		}
		err = createAPIExport(t, c, workspaceCluster, export)
		if err != nil {
			t.Fatalf("failed to create APIExport in cluster %q: %v", workspaceCluster, err)
		}

		// Create catalogentry
		path := "root:" + workspaceCluster.Base()
		newEntry := &catalogv1alpha1.CatalogEntry{
			ObjectMeta: metav1.ObjectMeta{
				Name: "entry-test",
			},
			Spec: catalogv1alpha1.CatalogEntrySpec{
				Exports: []apisv1alpha1.ExportReference{
					{
						Workspace: &apisv1alpha1.WorkspaceExportReference{
							Path:       path,
							ExportName: "export-test",
						},
					},
				},
			},
		}
		entry, err := createCatalogEntry(t, c, workspaceCluster, newEntry)
		if err != nil {
			t.Fatalf("unable to create CatalogEntry: %v", err)
		}

		// Check APIExportValid condition status to be True
		if !conditions.IsTrue(entry, catalogv1alpha1.APIExportValidType) {
			t.Fatalf("expect APIExportValid condition in entry %qin cluster %q to be True", entry.GetName(), workspaceCluster)
		}
		// Check ExportPermissionClaims status
		gr := metav1.GroupResource{
			Group:    schema.Spec.Group,
			Resource: schema.Spec.Names.Plural,
		}
		if len(entry.Status.Resources) > 0 {
			assert.Equal(t, entry.Status.Resources[0], gr, "two resources should be the same")
		} else {
			t.Fatalf("expect entry %q in cluster %q to has one resource", entry.GetName(), workspaceCluster)
		}
		// Check Resource status
		if len(entry.Status.ExportPermissionClaims) > 0 {
			claim := apisv1alpha1.PermissionClaim{
				GroupResource: apisv1alpha1.GroupResource{
					Resource: "configmaps",
				},
			}
			assert.Equal(t, entry.Status.ExportPermissionClaims[0], claim, "two ExportPermissionClaims should be the same")
		} else {
			t.Fatalf("expect entry %q in cluster %q to has one ExportPermissionClaim", entry.GetName(), workspaceCluster)
		}
	})
}

// TestInvalidCatalogEntry verifies that our catalog controller can validate
// invalid catalogEntry (bad export reference) but add a False condition in status
func TestInvalidCatalogEntry(t *testing.T) {
	t.Run("testing invalid catalog entry", func(t *testing.T) {
		workspaceCluster := parentWorkspace(t).Join(randomName())
		c := createWorkspace(t, workspaceCluster)

		// Create catalogentry
		path := "root:" + workspaceCluster.Base()
		newEntry := &catalogv1alpha1.CatalogEntry{
			ObjectMeta: metav1.ObjectMeta{
				Name: "entry-test",
			},
			Spec: catalogv1alpha1.CatalogEntrySpec{
				Exports: []apisv1alpha1.ExportReference{
					{
						Workspace: &apisv1alpha1.WorkspaceExportReference{
							Path:       path,
							ExportName: "export-test",
						},
					},
				},
			},
		}
		entry, err := createCatalogEntry(t, c, workspaceCluster, newEntry)
		if err != nil {
			t.Fatalf("catalogEntry %q failed to be reconciled in cluster %q: %v", entry.GetName(), workspaceCluster, err)
		}

		// Check APIExportValid condition status to be False due to bad export ref
		if conditions.IsTrue(entry, catalogv1alpha1.APIExportValidType) {
			t.Fatalf("expect APIExportValid condition in entry %q in cluster %q to be False", entry.GetName(), workspaceCluster)
		}
	})
}
