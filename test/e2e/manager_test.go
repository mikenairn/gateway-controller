package e2e

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Manager deployment", func() {
	It("is available", func(ctx context.Context) {
		deployment := &appsv1.Deployment{}
		key := client.ObjectKey{
			Namespace: managerNamespace,
			Name:      "gateway-controller-controller-manager",
		}

		Expect(k8sClient.Get(ctx, key, deployment)).To(Succeed())

		var available bool
		for _, cond := range deployment.Status.Conditions {
			if cond.Type == appsv1.DeploymentAvailable && cond.Status == corev1.ConditionTrue {
				available = true
			}
		}
		Expect(available).To(BeTrue(), "the manager deployment should report Available")
	})

	It("runs without restarting", func(ctx context.Context) {
		pods := &corev1.PodList{}
		Expect(k8sClient.List(ctx, pods,
			client.InNamespace(managerNamespace),
			client.MatchingLabels{"control-plane": "controller-manager"},
		)).To(Succeed())

		Expect(pods.Items).NotTo(BeEmpty(), "expected at least one manager pod")

		for _, pod := range pods.Items {
			Expect(pod.Status.Phase).To(Equal(corev1.PodRunning), "manager pod should be Running")
			for _, cs := range pod.Status.ContainerStatuses {
				Expect(cs.RestartCount).To(BeZero(),
					"manager container restarted, which usually means RBAC or leader election is misconfigured")
			}
		}
	})
})
