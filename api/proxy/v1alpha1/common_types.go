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
)

// FailureMode controls runtime behavior on filter errors.
// +kubebuilder:validation:Enum=closed;open
type FailureMode string

const (
	FailureModeClosed FailureMode = "closed"
	FailureModeOpen   FailureMode = "open"
)

// ServiceType for the proxy's Service.
// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
type ServiceType string

const (
	ServiceTypeClusterIP    ServiceType = "ClusterIP"
	ServiceTypeNodePort     ServiceType = "NodePort"
	ServiceTypeLoadBalancer ServiceType = "LoadBalancer"
)

// BodyAccessMode for policy filters.
// +kubebuilder:validation:Enum=read_only;read_write
type BodyAccessMode string

const (
	BodyAccessReadOnly  BodyAccessMode = "read_only"
	BodyAccessReadWrite BodyAccessMode = "read_write"
)

const (
	// PraxisGroup is the API group for all Praxis CRDs.
	PraxisGroup = "praxis.io"
	// PolicyKind is the Kind string for Policy resources.
	PolicyKind = "Policy"
)

// PraxisFilterReference identifies a resource by GVK and namespace/name.
type PraxisFilterReference struct {
	// group is the API group of the referenced resource (e.g. "praxis.io").
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Group string `json:"group"`

	// version is the API version of the referenced resource (e.g. "v1alpha1").
	// +optional
	Version string `json:"version,omitempty"`

	// kind is the kind of the referenced resource (e.g. "Policy").
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Kind string `json:"kind"`

	// namespace is the namespace of the referenced resource.
	// Defaults to the namespace of the containing Proxy.
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// name is the name of the referenced resource.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

// PraxisLocalReference references a same-namespace resource of a kind implied by the containing field.
type PraxisLocalReference struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

// PraxisDefinitionStatus is the shared status for Policy.
type PraxisDefinitionStatus struct {
	// observedGeneration is the latest generation inspected by this resource's controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// effectiveRevision is a digest of this definition and its resolved transitive inputs.
	// +optional
	EffectiveRevision string `json:"effectiveRevision,omitempty"`

	// usedBy lists the proxies that reference this definition.
	// Each entry is "namespace/name" of a Proxy.
	// +optional
	// +listType=set
	UsedBy []string `json:"usedBy,omitempty"`

	// programmedBy lists the proxies that have rendered this definition into their published config.
	// Each entry is "namespace/name" of a Proxy.
	// +optional
	// +listType=set
	ProgrammedBy []string `json:"programmedBy,omitempty"`

	// conditions report Accepted (valid specification) and ResolvedRefs (available and authorized dependencies).
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

const (
	ConditionTypeAccepted     = "Accepted"
	ConditionTypeResolvedRefs = "ResolvedRefs"
	ConditionTypeProgrammed   = "Programmed"
	ConditionTypeReady        = "Ready"
)
