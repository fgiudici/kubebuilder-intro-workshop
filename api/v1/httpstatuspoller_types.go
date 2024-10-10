/*
Copyright 2024.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HttpStatusPollerSpec defines the desired state of HttpStatusPoller
type HttpStatusPollerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// URLs is a slice containing the URLs to check
	// +kubebuilder:validation:MinItems=1
	URLs []string `json:"urls"`

	// IntervalSeconds specifies how many seconds to wait between subsequent checks
	// +optional
	IntervalSeconds int `json:"intervalSeconds,omitempty"`
}

// HttpStatusPollerStatus defines the observed state of HttpStatusPoller
type HttpStatusPollerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// HttpStatusPoller is the Schema for the httpstatuspollers API
type HttpStatusPoller struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HttpStatusPollerSpec   `json:"spec,omitempty"`
	Status HttpStatusPollerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HttpStatusPollerList contains a list of HttpStatusPoller
type HttpStatusPollerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HttpStatusPoller `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HttpStatusPoller{}, &HttpStatusPollerList{})
}
