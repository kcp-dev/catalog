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

package v1alpha1

import (
	kcpv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	conditionsv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/third_party/conditions/apis/conditions/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// These are valid conditions of CatalogEntry.
const (
	// CatalogEntryValid is a condition for CatalogEntry that reflects the validity
	// of the referenced APIExport.
	APIExportValidType conditionsv1alpha1.ConditionType = "APIExportValid"
	// CatalogEntryInvalidReferenceReason is a reason for the CatalogEntryValid
	// condition of APIBinding that the referenced CatalogEntry reference is invalid.
	APIExportNotFoundReason = "APIExportNotFound"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// CatalogEntry is the Schema for the catalogentries API
type CatalogEntry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CatalogEntrySpec   `json:"spec,omitempty"`
	Status CatalogEntryStatus `json:"status,omitempty"`
}

// CatalogEntrySpec defines the desired state of CatalogEntry
type CatalogEntrySpec struct {
	// exports is a list of references to APIExports.
	// +kubebuilder:validation:MinItems:=1
	Exports []kcpv1alpha1.ExportReference `json:"exports"`
	// description is a human-readable message to describe the information regarding
	// the capabilities and features that the API provides
	// +optional
	Description string `json:"description,omitempty"`
}

// CatalogEntryStatus defines the observed state of CatalogEntry
type CatalogEntryStatus struct {
	// exportPermissionClaims is a list of permissions requested by the API provider(s)
	// for this catalog entry.
	// +optional
	ExportPermissionClaims []kcpv1alpha1.PermissionClaim `json:"exportPermissionClaims,omitempty"`
	// resources is the list of APIs that are provided by this catalog entry.
	// +optional
	Resources []metav1.GroupResource `json:"resources,omitempty"`
	// conditions is a list of conditions that apply to the CatalogEntry.
	//
	// +optional
	Conditions conditionsv1alpha1.Conditions `json:"conditions,omitempty"`
}

func (in *CatalogEntry) GetConditions() conditionsv1alpha1.Conditions {
	return in.Status.Conditions
}

func (in *CatalogEntry) SetConditions(conditions conditionsv1alpha1.Conditions) {
	in.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// CatalogEntryList contains a list of CatalogEntry
type CatalogEntryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CatalogEntry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CatalogEntry{}, &CatalogEntryList{})
}
