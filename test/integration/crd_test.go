package integration

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	proxyv1alpha1 "github.com/praxis-proxy/gateway-controller/api/proxy/v1alpha1"
)

var _ = Describe("Proxy CRD", func() {
	It("accepts a minimal valid Proxy and reads it back", func(ctx context.Context) {
		proxy := &proxyv1alpha1.Proxy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "valid",
				Namespace: "default",
			},
			Spec: proxyv1alpha1.ProxySpec{
				Image: "example.com/praxis:latest",
				Listeners: []proxyv1alpha1.PraxisListener{
					{Name: "http", Port: 8080},
				},
			},
		}

		Expect(k8sClient.Create(ctx, proxy)).To(Succeed(), "a minimal valid Proxy should be admitted")

		got := &proxyv1alpha1.Proxy{}
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(proxy), got)).To(Succeed())
		Expect(got.Spec.Image).To(Equal("example.com/praxis:latest"))
	})
})
