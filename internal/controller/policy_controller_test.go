package controller

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	proxyv1alpha1 "github.com/praxis-proxy/gateway-controller/api/proxy/v1alpha1"
)

func TestMapProxyToPolicies(t *testing.T) {
	t.Parallel()

	reconciler := &PolicyReconciler{}

	tests := []struct {
		name     string
		proxy    *proxyv1alpha1.Proxy
		wantKeys []types.NamespacedName
	}{
		{
			name: "two different policy refs produce two requests",
			proxy: &proxyv1alpha1.Proxy{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default"},
				Spec: proxyv1alpha1.ProxySpec{
					FilterChains: []proxyv1alpha1.ProxyFilterChain{
						{
							Name: "chain-a",
							Filters: []proxyv1alpha1.PraxisFilterEntry{{
								Name: "f1",
								Ref: &proxyv1alpha1.PraxisFilterReference{
									Group: proxyv1alpha1.PraxisGroup,
									Kind:  proxyv1alpha1.PolicyKind,
									Name:  "policy-a",
								},
							}},
						},
						{
							Name: "chain-b",
							Filters: []proxyv1alpha1.PraxisFilterEntry{{
								Name: "f2",
								Ref: &proxyv1alpha1.PraxisFilterReference{
									Group: proxyv1alpha1.PraxisGroup,
									Kind:  proxyv1alpha1.PolicyKind,
									Name:  "policy-b",
								},
							}},
						},
					},
				},
			},
			wantKeys: []types.NamespacedName{
				{Name: "policy-a", Namespace: "default"},
				{Name: "policy-b", Namespace: "default"},
			},
		},
		{
			name: "no policy refs returns empty slice",
			proxy: &proxyv1alpha1.Proxy{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default"},
				Spec: proxyv1alpha1.ProxySpec{
					FilterChains: []proxyv1alpha1.ProxyFilterChain{
						{
							Name: "chain",
							Filters: []proxyv1alpha1.PraxisFilterEntry{{
								Name: "f1",
								InlineFilter: &proxyv1alpha1.PraxisInlineFilter{
									Type: "rate_limit",
								},
							}},
						},
					},
				},
			},
			wantKeys: nil,
		},
		{
			name: "duplicate policy ref across chains produces one request",
			proxy: &proxyv1alpha1.Proxy{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default"},
				Spec: proxyv1alpha1.ProxySpec{
					FilterChains: []proxyv1alpha1.ProxyFilterChain{
						{
							Name: "chain-a",
							Filters: []proxyv1alpha1.PraxisFilterEntry{{
								Name: "f1",
								Ref: &proxyv1alpha1.PraxisFilterReference{
									Group: proxyv1alpha1.PraxisGroup,
									Kind:  proxyv1alpha1.PolicyKind,
									Name:  "same-policy",
								},
							}},
						},
						{
							Name: "chain-b",
							Filters: []proxyv1alpha1.PraxisFilterEntry{{
								Name: "f2",
								Ref: &proxyv1alpha1.PraxisFilterReference{
									Group: proxyv1alpha1.PraxisGroup,
									Kind:  proxyv1alpha1.PolicyKind,
									Name:  "same-policy",
								},
							}},
						},
					},
				},
			},
			wantKeys: []types.NamespacedName{
				{Name: "same-policy", Namespace: "default"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requests := reconciler.mapProxyToPolicies(context.Background(), tt.proxy)

			if len(requests) != len(tt.wantKeys) {
				t.Fatalf("expected %d requests, got %d", len(tt.wantKeys), len(requests))
			}
			for i, want := range tt.wantKeys {
				if requests[i].NamespacedName != want {
					t.Errorf("request[%d] = %v, want %v", i, requests[i].NamespacedName, want)
				}
			}
		})
	}
}
