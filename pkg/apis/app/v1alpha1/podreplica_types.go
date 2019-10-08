package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PodReplicaSpec defines the desired state of PodReplica
// +k8s:openapi-gen=true
type PodReplicaSpec struct {
	Size int32 `json:"size"`
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// PodReplicaStatus defines the observed state of PodReplica
// +k8s:openapi-gen=true
type PodReplicaStatus struct {
	Replicas []string `json:"replicas"`
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodReplica is the Schema for the podreplicas API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type PodReplica struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PodReplicaSpec   `json:"spec,omitempty"`
	Status PodReplicaStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodReplicaList contains a list of PodReplica
type PodReplicaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodReplica `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodReplica{}, &PodReplicaList{})
}
