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
	"k8s.io/apimachinery/pkg/runtime"
)

// PraxisFilterEntry is a single use of a filter or policy definition in a chain.
// Exactly one of ref or inlineFilter must be set.
// +kubebuilder:validation:XValidation:rule="has(self.ref) != has(self.inlineFilter)",message="exactly one of ref or inlineFilter must be set"
type PraxisFilterEntry struct {
	// name is a unique entry name for diagnostics and repeated uses of one definition.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// ref identifies the filter or policy definition to instantiate.
	// +optional
	Ref *PraxisFilterReference `json:"ref,omitempty"`

	// inlineFilter defines a filter directly without a separate resource.
	// +optional
	InlineFilter *PraxisInlineFilter `json:"inlineFilter,omitempty"`

	// conditions are request gates for this entry. Empty means always eligible.
	// +optional
	Conditions []PraxisRequestCondition `json:"conditions,omitempty"`

	// responseConditions are additional response gates.
	// A filter skipped on request cannot run on response.
	// +optional
	ResponseConditions []PraxisResponseCondition `json:"responseConditions,omitempty"`

	// failureMode controls runtime behavior on filter errors: closed (default) or open.
	// +optional
	// +kubebuilder:default="closed"
	FailureMode *FailureMode `json:"failureMode,omitempty"`
}

// PraxisInlineFilter defines a filter type and configuration inline in a chain entry.
type PraxisInlineFilter struct {
	// type is the registered Praxis filter name.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Type string `json:"type"`

	// config is filter-specific configuration, flattened beside filter: in native config.
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Type=object
	Config *runtime.RawExtension `json:"config,omitempty"`
}

// PraxisRequestCondition gates a filter entry on request properties.
// Exactly one of when or unless must be set.
// +kubebuilder:validation:XValidation:rule="has(self.when) != has(self.unless)",message="exactly one of when or unless must be set"
type PraxisRequestCondition struct {
	// when runs the entry only when the request matches.
	// +optional
	When *PraxisRequestMatch `json:"when,omitempty"`

	// unless skips the entry when the request matches.
	// +optional
	Unless *PraxisRequestMatch `json:"unless,omitempty"`
}

// PraxisRequestMatch defines request matching criteria. All fields are ANDed.
type PraxisRequestMatch struct {
	// grpc matches or excludes gRPC requests.
	// +optional
	GRPC *bool `json:"grpc,omitempty"`

	// path is an exact request path match.
	// +optional
	Path *string `json:"path,omitempty"`

	// pathPrefix is a segment-boundary path prefix match.
	// +optional
	PathPrefix *string `json:"pathPrefix,omitempty"`

	// methods matches any listed HTTP method.
	// +optional
	Methods []string `json:"methods,omitempty"`

	// headers requires all named request headers to match.
	// +optional
	Headers map[string]string `json:"headers,omitempty"`

	// headersPresent requires all named request headers to be present, regardless of value.
	// +optional
	HeadersPresent []string `json:"headersPresent,omitempty"`

	// boundUpstream matches the request's bound logical upstream metadata.
	// +optional
	BoundUpstream *PraxisSelectedUpstreamMatch `json:"boundUpstream,omitempty"`

	// selectedUpstream matches metadata published by an earlier load balancer.
	// +optional
	SelectedUpstream *PraxisSelectedUpstreamMatch `json:"selectedUpstream,omitempty"`
}

// PraxisSelectedUpstreamMatch matches metadata from an earlier load balancer.
// At least one field is required.
type PraxisSelectedUpstreamMatch struct {
	// applicationProtocol is the selected cluster's application protocol.
	// +optional
	ApplicationProtocol *string `json:"applicationProtocol,omitempty"`

	// applicationProvider is the selected cluster's application provider.
	// +optional
	ApplicationProvider *string `json:"applicationProvider,omitempty"`
}

// PraxisResponseCondition gates a filter entry on response properties.
// Exactly one of when or unless must be set.
// +kubebuilder:validation:XValidation:rule="has(self.when) != has(self.unless)",message="exactly one of when or unless must be set"
type PraxisResponseCondition struct {
	// when runs response hooks only when the response matches.
	// +optional
	When *PraxisResponseMatch `json:"when,omitempty"`

	// unless skips response hooks when the response matches.
	// +optional
	Unless *PraxisResponseMatch `json:"unless,omitempty"`
}

// PraxisResponseMatch defines response matching criteria. All fields are ANDed.
type PraxisResponseMatch struct {
	// status matches one of the listed HTTP status codes (100-599).
	// +optional
	// +kubebuilder:validation:items:Minimum=100
	// +kubebuilder:validation:items:Maximum=599
	Status []int32 `json:"status,omitempty"`

	// headers requires all named response headers to match.
	// +optional
	Headers map[string]string `json:"headers,omitempty"`

	// headersPresent requires all named response headers to be present, regardless of value.
	// +optional
	HeadersPresent []string `json:"headersPresent,omitempty"`
}
