#!/usr/bin/env bash
# setup.sh - Provision a Kind cluster and deploy Bulkhead for E2E benchmarks.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLUSTER_NAME="bulkhead-bench"
IMAGE_TAG="bench"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

SERVICES=(control-plane policy observability)

# ── Pre-flight checks ────────────────────────────────────────────────────────
for cmd in kind docker helm kubectl; do
  if ! command -v "$cmd" &>/dev/null; then
    echo "ERROR: $cmd is required but not found in PATH" >&2
    exit 1
  fi
done

# ── 1. Create Kind cluster ──────────────────────────────────────────────────
echo "==> Creating Kind cluster '${CLUSTER_NAME}' ..."
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
  echo "    Cluster already exists, deleting first ..."
  kind delete cluster --name "${CLUSTER_NAME}"
fi
kind create cluster --name "${CLUSTER_NAME}" --config "${SCRIPT_DIR}/kind-config.yaml"
echo "    Cluster created."

# ── 2. Build Docker images ──────────────────────────────────────────────────
echo "==> Building Docker images for all services ..."
for svc in "${SERVICES[@]}"; do
  echo "    Building ${svc} ..."
  docker build \
    -t "bulkhead-${svc}:${IMAGE_TAG}" \
    -f "${REPO_ROOT}/deploy/docker/Dockerfile.control-plane" \
    --build-arg SERVICE="${svc}" \
    "${REPO_ROOT}"
done
echo "    All images built."

# ── 3. Load images into Kind ────────────────────────────────────────────────
echo "==> Loading images into Kind cluster ..."
for svc in "${SERVICES[@]}"; do
  echo "    Loading bulkhead-${svc}:${IMAGE_TAG} ..."
  kind load docker-image --name "${CLUSTER_NAME}" "bulkhead-${svc}:${IMAGE_TAG}"
done
echo "    All images loaded."

# ── 4. Install Helm chart ───────────────────────────────────────────────────
echo "==> Installing Helm chart ..."
helm install bulkhead "${REPO_ROOT}/deploy/helm/bulkhead" \
  --set image.tag="${IMAGE_TAG}" \
  --set service.type=NodePort
echo "    Helm release installed."

# ── 5. Wait for pods ────────────────────────────────────────────────────────
echo "==> Waiting for all pods to become ready (timeout 300s) ..."
kubectl wait --for=condition=ready pod --all -n default --timeout=300s
echo "    All pods ready."

# ── Summary ─────────────────────────────────────────────────────────────────
echo ""
echo "============================================"
echo "  Bulkhead E2E benchmark cluster is ready"
echo "============================================"
echo ""
echo "  Service port mappings (localhost):"
echo "    control-plane  : localhost:50060  (Identity, Task, Workspace, Compute)"
echo "    policy         : localhost:50062  (Guardrails, Governance)"
echo "    observability  : localhost:50065  (Activity, Economics, Human)"
echo ""
echo "  Run benchmarks with:"
echo "    cd benchmarks && go test -tags e2e -bench=. -benchtime=5s -count=3 ./..."
echo ""
echo "  Tear down with:"
echo "    ./teardown.sh"
echo ""
