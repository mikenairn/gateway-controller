package controller

import (
	"context"
	"fmt"
	"slices"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"

	proxyv1alpha1 "github.com/praxis-proxy/gateway-controller/api/proxy/v1alpha1"
)

const IndexPolicyRef = ".spec.filterChains.policyRef"

type PolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=praxis.io,resources=policies,verbs=get;list;watch
// +kubebuilder:rbac:groups=praxis.io,resources=policies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=praxis.io,resources=proxies,verbs=get;list;watch

func (r *PolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var policy proxyv1alpha1.Policy
	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	patch := client.MergeFrom(policy.DeepCopy())

	log.V(1).Info("reconciling policy")

	policy.Status.ObservedGeneration = policy.Generation

	revision, err := computeRevision(policy.Spec)
	if err != nil {
		return ctrl.Result{}, err
	}
	policy.Status.EffectiveRevision = revision

	var doc any
	if err := yaml.Unmarshal([]byte(policy.Spec.Document), &doc); err != nil {
		setCondition(&policy.Status.Conditions, proxyv1alpha1.ConditionTypeAccepted,
			metav1.ConditionFalse, "InvalidDocument", err.Error(), policy.Generation)
	} else {
		setCondition(&policy.Status.Conditions, proxyv1alpha1.ConditionTypeAccepted,
			metav1.ConditionTrue, "Valid", "Policy definition is valid", policy.Generation)
	}

	policyKey := fmt.Sprintf("%s/%s", policy.Namespace, policy.Name)
	var proxies proxyv1alpha1.ProxyList
	if err := r.List(ctx, &proxies, client.MatchingFields{IndexPolicyRef: policyKey}); err != nil {
		return ctrl.Result{}, err
	}
	var usedBy []string
	for _, proxy := range proxies.Items {
		usedBy = append(usedBy, fmt.Sprintf("%s/%s", proxy.Namespace, proxy.Name))
	}
	slices.Sort(usedBy)
	policy.Status.UsedBy = usedBy

	if err := r.Status().Patch(ctx, &policy, patch); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *PolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := SetupIndexes(context.Background(), mgr.GetFieldIndexer()); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&proxyv1alpha1.Policy{}).
		Watches(&proxyv1alpha1.Proxy{}, handler.EnqueueRequestsFromMapFunc(r.mapProxyToPolicies)).
		Named("policy").
		Complete(r)
}

func SetupIndexes(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &proxyv1alpha1.Proxy{}, IndexPolicyRef, extractPolicyRefs)
}

func extractPolicyRefs(obj client.Object) []string {
	proxy, ok := obj.(*proxyv1alpha1.Proxy)
	if !ok {
		return nil
	}
	var refs []string
	seen := make(map[string]bool)
	for _, chain := range proxy.Spec.FilterChains {
		for _, entry := range chain.Filters {
			if isPolicyRef(entry.Ref) {
				key := refNamespace(entry.Ref, proxy.Namespace) + "/" + entry.Ref.Name
				if !seen[key] {
					seen[key] = true
					refs = append(refs, key)
				}
			}
		}
	}
	return refs
}

func (r *PolicyReconciler) mapProxyToPolicies(ctx context.Context, obj client.Object) []reconcile.Request {
	proxy, ok := obj.(*proxyv1alpha1.Proxy)
	if !ok {
		return nil
	}
	seen := make(map[types.NamespacedName]bool)
	var requests []reconcile.Request
	for _, chain := range proxy.Spec.FilterChains {
		for _, entry := range chain.Filters {
			if isPolicyRef(entry.Ref) {
				nn := types.NamespacedName{
					Name:      entry.Ref.Name,
					Namespace: refNamespace(entry.Ref, proxy.Namespace),
				}
				if !seen[nn] {
					seen[nn] = true
					requests = append(requests, reconcile.Request{NamespacedName: nn})
				}
			}
		}
	}
	return requests
}
