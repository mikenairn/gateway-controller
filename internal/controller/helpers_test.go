package controller

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proxyv1alpha1 "github.com/praxis-proxy/gateway-controller/api/proxy/v1alpha1"
)

func TestComputeRevision(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		inputA    []any
		inputB    []any
		wantEqual bool
	}{
		{
			name:      "deterministic: same input produces same output",
			inputA:    []any{"hello"},
			inputB:    []any{"hello"},
			wantEqual: true,
		},
		{
			name:      "unique: different inputs produce different outputs",
			inputA:    []any{"hello"},
			inputB:    []any{"world"},
			wantEqual: false,
		},
		{
			name:      "order-sensitive: swapped inputs produce different outputs",
			inputA:    []any{"a", "b"},
			inputB:    []any{"b", "a"},
			wantEqual: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			revA, err := computeRevision(tt.inputA...)
			if err != nil {
				t.Fatalf("unexpected error for inputA: %v", err)
			}

			revB, err := computeRevision(tt.inputB...)
			if err != nil {
				t.Fatalf("unexpected error for inputB: %v", err)
			}

			if tt.wantEqual && revA != revB {
				t.Errorf("expected equal revisions, got %q and %q", revA, revB)
			}
			if !tt.wantEqual && revA == revB {
				t.Errorf("expected different revisions, both are %q", revA)
			}
		})
	}
}

func TestIsPolicyRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ref  *proxyv1alpha1.PraxisFilterReference
		want bool
	}{
		{
			name: "nil ref returns false",
			ref:  nil,
			want: false,
		},
		{
			name: "correct group and kind returns true",
			ref: &proxyv1alpha1.PraxisFilterReference{
				Group: proxyv1alpha1.PraxisGroup,
				Kind:  proxyv1alpha1.PolicyKind,
				Name:  "test",
			},
			want: true,
		},
		{
			name: "wrong group returns false",
			ref: &proxyv1alpha1.PraxisFilterReference{
				Group: "wrong.io",
				Kind:  proxyv1alpha1.PolicyKind,
				Name:  "test",
			},
			want: false,
		},
		{
			name: "wrong kind returns false",
			ref: &proxyv1alpha1.PraxisFilterReference{
				Group: proxyv1alpha1.PraxisGroup,
				Kind:  "Filter",
				Name:  "test",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isPolicyRef(tt.ref); got != tt.want {
				t.Errorf("isPolicyRef() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRefNamespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ref       *proxyv1alpha1.PraxisFilterReference
		defaultNS string
		want      string
	}{
		{
			name: "explicit namespace is returned",
			ref: &proxyv1alpha1.PraxisFilterReference{
				Namespace: "custom-ns",
			},
			defaultNS: "default",
			want:      "custom-ns",
		},
		{
			name:      "empty namespace falls back to default",
			ref:       &proxyv1alpha1.PraxisFilterReference{},
			defaultNS: "default",
			want:      "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := refNamespace(tt.ref, tt.defaultNS); got != tt.want {
				t.Errorf("refNamespace() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSetCondition(t *testing.T) {
	t.Parallel()

	var conditions []metav1.Condition
	setCondition(&conditions, "Ready", metav1.ConditionTrue, "AllGood", "everything is fine", 1)

	if len(conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(conditions))
	}
	if conditions[0].Type != "Ready" {
		t.Errorf("expected type Ready, got %s", conditions[0].Type)
	}
	if conditions[0].Status != metav1.ConditionTrue {
		t.Errorf("expected status True, got %s", conditions[0].Status)
	}
	if conditions[0].Reason != "AllGood" {
		t.Errorf("expected reason AllGood, got %s", conditions[0].Reason)
	}
	if conditions[0].ObservedGeneration != 1 {
		t.Errorf("expected observedGeneration 1, got %d", conditions[0].ObservedGeneration)
	}
}
