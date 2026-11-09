// Package utils provides shared test utilities for the integration and e2e
// suites.
package utils

import (
	"path/filepath"
	goruntime "runtime"
	"testing"

	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	aigatewayv1alpha1 "github.com/praxis-proxy/gateway-controller/api/aigateway/v1alpha1"
	proxyv1alpha1 "github.com/praxis-proxy/gateway-controller/api/proxy/v1alpha1"
)

// CRDDirectoryPaths locates the generated CRDs relative to this file, so the
// suites do not depend on the directory the tests are run from.
func CRDDirectoryPaths() []string {
	_, thisFile, _, _ := goruntime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	return []string{filepath.Join(repoRoot, "config", "crd", "bases")}
}

// NewScheme returns a scheme with the Kubernetes built-ins and both project
// API groups registered.
func NewScheme(tb testing.TB) *k8sruntime.Scheme {
	tb.Helper()

	scheme := k8sruntime.NewScheme()
	for _, add := range []func(*k8sruntime.Scheme) error{
		clientgoscheme.AddToScheme,
		proxyv1alpha1.AddToScheme,
		aigatewayv1alpha1.AddToScheme,
	} {
		if err := add(scheme); err != nil {
			tb.Fatalf("build scheme: %v", err)
		}
	}

	return scheme
}

// StartEnvtest brings up a control plane with the project CRDs installed and
// returns a client backed by the project scheme. The environment is stopped
// when the test finishes.
func StartEnvtest(tb testing.TB) (client.Client, *rest.Config) {
	tb.Helper()

	env := &envtest.Environment{
		CRDDirectoryPaths:     CRDDirectoryPaths(),
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := env.Start()
	if err != nil {
		tb.Fatalf("start envtest: %v", err)
	}
	tb.Cleanup(func() {
		if err := env.Stop(); err != nil {
			tb.Errorf("stop envtest: %v", err)
		}
	})

	c, err := client.New(cfg, client.Options{Scheme: NewScheme(tb)})
	if err != nil {
		tb.Fatalf("new client: %v", err)
	}

	return c, cfg
}
