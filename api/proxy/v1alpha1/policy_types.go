/*
Copyright 2026.

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
	"k8s.io/apimachinery/pkg/runtime"
)

// PolicySpec defines a Praxis Policy Engine document and native filter options.
type PolicySpec struct {
	// document is the Praxis Policy Engine document as a YAML string.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Document string `json:"document"`

	// filter provides typed options for the native policy filter.
	// +optional
	Filter *PolicyFilterConfig `json:"filter,omitempty"`
}

// PolicyFilterConfig provides typed options for the native policy filter.
type PolicyFilterConfig struct {
	// bodyAccess controls body access mode: read_only (default) or read_write.
	// +optional
	// +kubebuilder:default="read_only"
	BodyAccess *BodyAccessMode `json:"bodyAccess,omitempty"`

	// initTimeoutSeconds is the policy initialization timeout.
	// +optional
	// +kubebuilder:default=30
	// +kubebuilder:validation:Minimum=1
	InitTimeoutSeconds *int32 `json:"initTimeoutSeconds,omitempty"`

	// maxBufferBytes is the request/response buffer ceiling in write mode.
	// +optional
	// +kubebuilder:default=10485760
	// +kubebuilder:validation:Minimum=1
	MaxBufferBytes *int64 `json:"maxBufferBytes,omitempty"`

	// allowPrivateIdp requires an explicitly authorized runtime capability.
	// +optional
	// +kubebuilder:default=false
	AllowPrivateIdp *bool `json:"allowPrivateIdp,omitempty"`

	// trustedPrivateEndpoints lists hosts the policy engine may reach at private (RFC 1918) addresses.
	// Narrower than allowPrivateIdp: only listed hosts may resolve to non-public addresses.
	// +optional
	TrustedPrivateEndpoints []string `json:"trustedPrivateEndpoints,omitempty"`

	// requireProtocolMetadata maps to the native classifier check.
	// +optional
	// +kubebuilder:default=true
	RequireProtocolMetadata *bool `json:"requireProtocolMetadata,omitempty"`

	// llm provides inference policy options.
	// +optional
	LLM *PolicyLLMOptions `json:"llm,omitempty"`

	// contextHeaders maps extension field names to HTTP header names for populating
	// host-supplied context slots. Valid keys: agent.conversation_id, agent.turn,
	// agent.agent_id, agent.parent_agent_id, agent.input, framework.framework,
	// framework.framework_version, framework.node_id, framework.graph_id.
	// Values must use a reserved header prefix (x-praxis-*, x-ext-agent-*, x-ext-protocol-*).
	// +optional
	ContextHeaders map[string]string `json:"contextHeaders,omitempty"`

	// contextMeta provides static operator-supplied metadata merged into the policy engine's
	// meta extension slot on every request.
	// +optional
	ContextMeta *PolicyContextMeta `json:"contextMeta,omitempty"`
}

// PolicyLLMOptions provides inference policy options.
type PolicyLLMOptions struct {
	// maxRequestBytes is the positive request limit; defaults to maxBufferBytes.
	// +optional
	// +kubebuilder:validation:Minimum=1
	MaxRequestBytes *int64 `json:"maxRequestBytes,omitempty"`

	// promoteParams lists request fields promoted to policy metadata.
	// +optional
	PromoteParams []string `json:"promoteParams,omitempty"`

	// provider is an operator-supplied provider value exposed to policy rules.
	// +optional
	Provider *string `json:"provider,omitempty"`

	// requireModel denies inference requests without a usable model.
	// +optional
	// +kubebuilder:default=true
	RequireModel *bool `json:"requireModel,omitempty"`

	// requireRoute denies models not selected by a policy route.
	// +optional
	// +kubebuilder:default=true
	RequireRoute *bool `json:"requireRoute,omitempty"`
}

// PolicyContextMeta provides static operator-supplied metadata for the meta extension slot.
type PolicyContextMeta struct {
	// tags are operational tags available as meta.tags in policy rules.
	// +optional
	Tags []string `json:"tags,omitempty"`

	// scope is a host-defined grouping used by the PPE for route resolution.
	// +optional
	Scope *string `json:"scope,omitempty"`

	// properties are arbitrary key-value metadata available as meta.properties.<key> in APL predicates.
	// +optional
	Properties map[string]string `json:"properties,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Accepted",type=string,JSONPath=`.status.conditions[?(@.type=="Accepted")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Policy is a reusable Policy Engine document. A chain activates it;
// the controller manages the policy file and generates config_path.
type Policy struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// +required
	Spec PolicySpec `json:"spec"`

	// +optional
	Status PraxisDefinitionStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// PolicyList contains a list of Policy.
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Policy `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &Policy{}, &PolicyList{})
		return nil
	})
}
