#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="${KIND_CLUSTER_NAME:-praxis-dev}"
KIND="${KIND:-kind}"
CONTAINER_ENGINE="${CONTAINER_ENGINE:-$(command -v podman || command -v docker || true)}"
# Honor explicit KIND_EXPERIMENTAL_PROVIDER, otherwise derive from CONTAINER_ENGINE
KIND_PROVIDER="${KIND_EXPERIMENTAL_PROVIDER:-$(basename "${CONTAINER_ENGINE}")}"

echo "==> Deleting KIND cluster '${CLUSTER_NAME}'..."
KIND_EXPERIMENTAL_PROVIDER="${KIND_PROVIDER}" ${KIND} delete cluster --name "${CLUSTER_NAME}"
echo "==> Done."
