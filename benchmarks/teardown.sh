#!/usr/bin/env bash
# teardown.sh - Remove the Bulkhead E2E benchmark Kind cluster.
set -euo pipefail

CLUSTER_NAME="bulkhead-bench"

echo "==> Uninstalling Helm release ..."
helm uninstall bulkhead 2>/dev/null || true

echo "==> Deleting Kind cluster '${CLUSTER_NAME}' ..."
kind delete cluster --name "${CLUSTER_NAME}"

echo "==> Cleanup complete."
