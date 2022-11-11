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

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	catalogv1alpha1 "github.com/kcp-dev/catalog/api/v1alpha1"
)

// CatalogEntryReconciler reconciles a CatalogEntry object
type CatalogEntryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=catalog.kcp.dev,resources=catalogentries,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=catalog.kcp.dev,resources=catalogentries/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=catalog.kcp.dev,resources=catalogentries/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CatalogEntry object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *CatalogEntryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	ctx = logicalcluster.WithCluster(ctx, logicalcluster.New(req.ClusterName))
	// Fetch the catalog entry from the request
	catalogEntry := &v1alpha1.CatalogEntry{}
	err := r.Get(ctx, req.NamespacedName, catalogEntry)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected.
			log.Info("Catalog Entry not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get resource")
		return ctrl.Result{}, err
	}

	apiExportNameReferences := catalogEntry.Spec.References

	for _, exportRef := range apiExportNameReferences {
		// TODO: verify if path contains the entire heirarchy or just the clusterName.
		// If it contains the heirarchy then extract the clusterName
		path := exportRef.Workspace.Path
		name := exportRef.Workspace.ExportName
		clusterApiExport := apisv1alpha1.APIExport{}
		err := r.Get(logicalcluster.WithCluster(ctx, logicalcluster.New(path)), types.NamespacedName{Name: name, Namespace: req.Namespace}, &clusterApiExport)
		if err != nil {
			if errors.IsNotFound(err) {
				log.Error(err, "APIExport referenced in catalog entry does not exist")
				return ctrl.Result{}, err
			}
			// Error reading the object - requeue the request.
			log.Error(err, "Failed to get resource")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CatalogEntryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&catalogv1alpha1.CatalogEntry{}).
		Complete(r)
}
