package e2e

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Long enough to span several leader election lease renewals, which
	// default to every few seconds.
	stabilityWindow = 30 * time.Second
	stabilityPoll   = 5 * time.Second
)

var _ = Describe("Manager deployment", func() {
	It("is available", func(ctx context.Context) {
		deployment := &appsv1.Deployment{}
		key := client.ObjectKey{
			Namespace: managerNamespace,
			Name:      "gateway-controller-controller-manager",
		}

		Expect(k8sClient.Get(ctx, key, deployment)).To(Succeed())

		// Check that the deployment is fully rolled out and available
		var available bool
		for _, cond := range deployment.Status.Conditions {
			if cond.Type == appsv1.DeploymentAvailable && cond.Status == corev1.ConditionTrue {
				available = true
			}
		}
		Expect(available).To(BeTrue(), "the manager deployment should report Available")

		// Verify the rollout is complete: observed generation matches current generation
		Expect(deployment.Status.ObservedGeneration).To(Equal(deployment.Generation),
			"deployment should have observed the current generation")

		// Verify all desired replicas are updated and available
		desired := *deployment.Spec.Replicas
		Expect(deployment.Status.UpdatedReplicas).To(Equal(desired),
			"all desired replicas should be updated")
		Expect(deployment.Status.AvailableReplicas).To(Equal(desired),
			"all desired replicas should be available")
	})

	It("runs without restarting", func(ctx context.Context) {
		// Observed over a window rather than once: a single reading cannot
		// tell a healthy manager from one that has not crashed yet. Leader
		// election renews its lease every few seconds, so a missing lease
		// permission shows up shortly after the deployment reports Available,
		// not at the moment it does.
		Consistently(func(gom Gomega) {
			// Get the deployment to find its ReplicaSets
			deployment := &appsv1.Deployment{}
			key := client.ObjectKey{
				Namespace: managerNamespace,
				Name:      "gateway-controller-controller-manager",
			}
			gom.Expect(k8sClient.Get(ctx, key, deployment)).To(Succeed())

			// Get ReplicaSets owned by the deployment
			replicaSets := &appsv1.ReplicaSetList{}
			gom.Expect(k8sClient.List(ctx, replicaSets,
				client.InNamespace(managerNamespace),
				client.MatchingLabels{"app.kubernetes.io/name": "gateway-controller"},
			)).To(Succeed())

			// Find ReplicaSets owned by this deployment
			var deploymentReplicaSets []appsv1.ReplicaSet
			for _, rs := range replicaSets.Items {
				for _, owner := range rs.OwnerReferences {
					if owner.Kind == "Deployment" && owner.Name == deployment.Name {
						deploymentReplicaSets = append(deploymentReplicaSets, rs)
						break
					}
				}
			}

			// Get pods owned by the deployment's ReplicaSets
			allPods := &corev1.PodList{}
			gom.Expect(k8sClient.List(ctx, allPods,
				client.InNamespace(managerNamespace),
			)).To(Succeed())

			live := make([]corev1.Pod, 0)
			for _, pod := range allPods.Items {
				if pod.DeletionTimestamp != nil {
					continue
				}
				// Check if this pod is owned by one of the deployment's ReplicaSets
				for _, owner := range pod.OwnerReferences {
					if owner.Kind == "ReplicaSet" {
						for _, rs := range deploymentReplicaSets {
							if owner.Name == rs.Name {
								live = append(live, pod)
								break
							}
						}
						break
					}
				}
			}

			gom.Expect(live).NotTo(BeEmpty(), "expected at least one manager pod")

			for _, pod := range live {
				gom.Expect(pod.Status.Phase).To(Equal(corev1.PodRunning), "manager pod should be Running")
				for _, cs := range pod.Status.ContainerStatuses {
					gom.Expect(cs.RestartCount).To(BeZero(),
						"manager container restarted, which usually means RBAC or leader election is misconfigured")
				}
			}
		}).WithContext(ctx).
			WithTimeout(stabilityWindow).
			WithPolling(stabilityPoll).
			Should(Succeed())
	})
})
