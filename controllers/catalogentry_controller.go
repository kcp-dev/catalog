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
	"fmt"
	"reflect"
	"strings"

	"github.com/kcp-dev/catalog/api/v1alpha1"
	catalogv1alpha1 "github.com/kcp-dev/catalog/api/v1alpha1"
	apisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	conditionsapi "github.com/kcp-dev/kcp/pkg/apis/third_party/conditions/apis/conditions/v1alpha1"
	"github.com/kcp-dev/kcp/pkg/apis/third_party/conditions/util/conditions"
	"github.com/kcp-dev/kcp/pkg/logging"
	"github.com/kcp-dev/logicalcluster/v2"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	controllerName = "kcp-catalogentry"
)

// CatalogEntryReconciler reconciles a CatalogEntry object
type CatalogEntryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile validates exports in CatalogEntry spec and add a condition to status
// to reflect the outcome of the validation.
// It also aggregates all permissionClaims and api resources from referenced APIExport
// to CatalogEntry status

//+kubebuilder:rbac:groups=catalog.kcp.dev,resources=catalogentries,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=catalog.kcp.dev,resources=catalogentries/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=catalog.kcp.dev,resources=catalogentries/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=configmaps/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=namespaces/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=namespaces/finalizers,verbs=update

func (r *CatalogEntryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logging.WithReconciler(klog.Background(), controllerName)
	logger = logger.WithValues("clusterName", req.ClusterName)
	ctx = logicalcluster.WithCluster(ctx, logicalcluster.New(req.ClusterName))

	// Fetch the catalog entry from the request
	catalogEntry := &v1alpha1.CatalogEntry{}
	err := r.Get(ctx, req.NamespacedName, catalogEntry)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected.
			logger.Info("CatalogEntry not found")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "failed to get resource")
		return ctrl.Result{}, err
	}

	oldEntry := catalogEntry.DeepCopy()
	resources := []metav1.GroupResource{}
	exportPermissionClaims := []apisv1alpha1.PermissionClaim{}
	invalidExports := []string{}
	for _, exportRef := range catalogEntry.Spec.Exports {
		// TODO: verify if path contains the entire heirarchy or just the clusterName.
		// If it contains the heirarchy then extract the clusterName
		path := exportRef.Workspace.Path
		name := exportRef.Workspace.ExportName
		logger = logger.WithValues(
			"path", path,
			"exportName", name,
		)
		logger.V(2).Info("reconciling CatalogEntry")
		export := apisv1alpha1.APIExport{}
		err := r.Get(logicalcluster.WithCluster(ctx, logicalcluster.New(path)), types.NamespacedName{Name: name, Namespace: req.Namespace}, &export)
		if err != nil {
			invalidExports = append(invalidExports, fmt.Sprintf("%s/%s", path, name))
			if errors.IsNotFound(err) {
				logger.Error(err, "APIExport referenced in catalog entry does not exist")
				continue
			}
			// Error reading the object - requeue the request.
			logger.Error(err, "failed to get resource")
			continue
		}

		// Extract permission and API resource info
		for _, claim := range export.Spec.PermissionClaims {
			exportPermissionClaims = append(exportPermissionClaims, claim)
		}
		catalogEntry.Status.ExportPermissionClaims = exportPermissionClaims

		for _, schemaName := range export.Spec.LatestResourceSchemas {
			_, resource, group, ok := split3(schemaName, ".")
			if !ok {
				continue
			}
			gr := metav1.GroupResource{
				Group:    group,
				Resource: resource,
			}
			resources = append(resources, gr)
		}
		catalogEntry.Status.Resources = resources
	}

	if len(invalidExports) == 0 {
		cond := conditionsapi.Condition{
			Type:               catalogv1alpha1.APIExportValidType,
			Status:             corev1.ConditionTrue,
			Severity:           conditionsapi.ConditionSeverityNone,
			LastTransitionTime: metav1.Now(),
		}
		conditions.Set(catalogEntry, &cond)
	} else {
		message := fmt.Sprintf("invalid export(s): %s", strings.Join(invalidExports, " ,"))
		invalidCond := conditionsapi.Condition{
			Type:               catalogv1alpha1.APIExportValidType,
			Status:             corev1.ConditionFalse,
			Severity:           conditionsapi.ConditionSeverityError,
			LastTransitionTime: metav1.Now(),
			Message:            message,
		}
		conditions.Set(catalogEntry, &invalidCond)
	}

	// Update the catalog entry if status is changed
	if !reflect.DeepEqual(catalogEntry.Status, oldEntry.Status) {
		err = r.Client.Status().Update(context.TODO(), catalogEntry)
		if err != nil {
			logger.Error(err, "failed to update CatalogEntry")
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

func split3(s string, sep string) (string, string, string, bool) {
	comps := strings.SplitN(s, sep, 3)
	if len(comps) != 3 {
		return "", "", "", false
	}
	return comps[0], comps[1], comps[2], true
}
