package e2e

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlconfig "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/praxis-proxy/gateway-controller/test/utils"
)

const managerNamespace = "gateway-controller-system"

var k8sClient client.Client

func TestE2E(t *testing.T) {
	cfg, err := ctrlconfig.GetConfig()
	if err != nil {
		t.Skipf("no reachable cluster, skipping e2e: %v", err)
	}

	k8sClient, err = client.New(cfg, client.Options{Scheme: utils.NewScheme(t)})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite")
}
