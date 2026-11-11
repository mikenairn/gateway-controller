# Control Plane for Praxis Gateways

This controller will manage the Kubernetes based control plane API for Gateway implementations such as AI Gateway.

## Development

Run `make help` for the full target list. The common ones:

```console
make build             # build the manager binary
make test              # unit tests (hermetic)
make test-integration  # integration tests against envtest
make kind-up           # create a kind cluster and deploy the manager into it
make test-e2e          # e2e tests against that cluster (skipped without one)
make kind-down         # delete the cluster
make lint              # golangci-lint
make verify            # fail if generated or derived output is uncommitted
```

Development follows the praxis-proxy [development conventions].

[development conventions]: https://github.com/praxis-proxy/conventions/blob/main/docs/conventions.md
