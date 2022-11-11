/*
Copyright 2019 The Knative Authors

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// SeccompProfile represents a seccomp profile to distribute to nodes.
//
// +genclient
// +genclient:nonNamespaced
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SeccompProfile struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the SeccompProfile (from the client).
	// +optional
	Spec SeccompProfileSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the SeccompProfile.
	// +optional
	Status SeccompProfileStatus `json:"status,omitempty"`
}

var (
	// Check that SeccompProfile can be validated and defaulted.
	_ apis.Validatable   = (*SeccompProfile)(nil)
	_ apis.Defaultable   = (*SeccompProfile)(nil)
	_ kmeta.OwnerRefable = (*SeccompProfile)(nil)
	// Check that the type conforms to the duck Knative Resource shape.
	_ duckv1.KRShaped = (*SeccompProfile)(nil)
)

// SeccompProfileSpec holds the desired state of the SeccompProfileSpec (from the client).
type SeccompProfileSpec struct {
	// Contents contains the contents of the policy as JSON.
	Contents *SeccompProfileJSON `json:"contents,omitempty"`
}

type SeccompProfileJSON struct {
	DefaultAction Action                  `json:"defaultAction"`
	Architectures []string                `json:"architectures,omitempty"`
	Syscalls      []SeccompProfileSyscall `json:"syscalls,omitempty"`
}

type SeccompProfileSyscall struct {
	Name   string   `json:"name"`
	Names  []string `json:"names,omitempty"`
	Action Action   `json:"action"`
	Args   []string `json:"args,omitempty"`
}

// SeccompProfileStatus communicates the observed state of the SeccompProfile (from the controller).
type SeccompProfileStatus struct {
	duckv1.Status `json:",inline"`
}

// GetStatus retrieves the status of the resource. Implements the KRShaped interface.
func (sp *SeccompProfile) GetStatus() *duckv1.Status {
	return &sp.Status.Status
}

// SeccompProfileList is a list of SeccompProfile resources
//
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SeccompProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []SeccompProfile `json:"items"`
}
