package controller

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proxyv1alpha1 "github.com/praxis-proxy/gateway-controller/api/proxy/v1alpha1"
)

func computeRevision(data ...any) (string, error) {
	hash := sha256.New()
	for _, d := range data {
		encoded, err := json.Marshal(d)
		if err != nil {
			return "", fmt.Errorf("marshal revision input: %w", err)
		}
		hash.Write(encoded)
	}
	return hex.EncodeToString(hash.Sum(nil))[:12], nil
}

func setCondition(conditions *[]metav1.Condition, condType string, status metav1.ConditionStatus, reason, message string, generation int64) {
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:               condType,
		Status:             status,
		ObservedGeneration: generation,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Reason:             reason,
		Message:            message,
	})
}

func isPolicyRef(ref *proxyv1alpha1.PraxisFilterReference) bool {
	return ref != nil && ref.Group == proxyv1alpha1.PraxisGroup && ref.Kind == proxyv1alpha1.PolicyKind
}

func refNamespace(ref *proxyv1alpha1.PraxisFilterReference, defaultNS string) string {
	if ref.Namespace != "" {
		return ref.Namespace
	}
	return defaultNS
}
