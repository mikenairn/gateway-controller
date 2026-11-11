package integration

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/praxis-proxy/gateway-controller/test/utils"
)

var k8sClient client.Client

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)

	k8sClient, _ = utils.StartEnvtest(t)

	RunSpecs(t, "Integration Suite")
}
