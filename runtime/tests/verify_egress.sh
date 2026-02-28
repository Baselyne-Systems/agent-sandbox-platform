#!/bin/bash
# Live verification of egress allowlist enforcement.
# Runs inside a privileged Docker container with iptables + Docker socket access.
#
# Usage: docker run --rm --privileged --network=host \
#          -v /var/run/docker.sock:/var/run/docker.sock \
#          alpine:3.20 sh /test/verify_egress.sh

set -e

echo "=== Egress Allowlist Live Verification ==="

# Install dependencies
apk add --no-cache docker-cli iptables curl > /dev/null 2>&1
echo "[OK] Dependencies installed"

SANDBOX_ID="verify-egress-$(date +%s)"
CHAIN="BH-${SANDBOX_ID:0:12}"
IMAGE="alpine:3.20"

echo ""
echo "--- Step 1: Start container ---"
CONTAINER_ID=$(docker run -d --name "bulkhead-${SANDBOX_ID}" "${IMAGE}" sleep 300)
echo "[OK] Container: ${CONTAINER_ID:0:12}"

# Get container IP
CONTAINER_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "${CONTAINER_ID}")
echo "[OK] Container IP: ${CONTAINER_IP}"

if [ -z "${CONTAINER_IP}" ]; then
    echo "[FAIL] No container IP — skipping iptables test"
    docker rm -f "${CONTAINER_ID}" > /dev/null 2>&1
    exit 1
fi

echo ""
echo "--- Step 2: Apply iptables egress rules ---"
echo "    Allowlist: 1.1.1.1 (Cloudflare DNS)"
echo "    Chain: ${CHAIN}"

# Create chain
iptables -N "${CHAIN}" 2>/dev/null || true
echo "[OK] Chain created"

# Jump from FORWARD
iptables -I FORWARD -s "${CONTAINER_IP}" -j "${CHAIN}"
echo "[OK] FORWARD jump added"

# Allow established connections
iptables -A "${CHAIN}" -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
echo "[OK] ESTABLISHED allowed"

# Allow DNS
iptables -A "${CHAIN}" -p udp --dport 53 -j ACCEPT
iptables -A "${CHAIN}" -p tcp --dport 53 -j ACCEPT
echo "[OK] DNS allowed"

# Allow 1.1.1.1
iptables -A "${CHAIN}" -d 1.1.1.1 -j ACCEPT
echo "[OK] 1.1.1.1 allowed"

# Default DROP
iptables -A "${CHAIN}" -j DROP
echo "[OK] Default DROP added"

echo ""
echo "--- Step 3: Verify rules ---"
iptables -L "${CHAIN}" -n --line-numbers

echo ""
echo "--- Step 4: Test connectivity ---"

# Test allowed destination
echo -n "    1.1.1.1 (allowed):  "
docker exec "${CONTAINER_ID}" wget -q -O /dev/null --timeout=5 http://1.1.1.1/ 2>/dev/null && echo "REACHABLE [OK]" || echo "UNREACHABLE [UNEXPECTED]"

# Test blocked destination
echo -n "    8.8.8.8 (blocked):  "
docker exec "${CONTAINER_ID}" wget -q -O /dev/null --timeout=5 http://8.8.8.8/ 2>/dev/null && echo "REACHABLE [FAIL — should be blocked!]" || echo "BLOCKED [OK]"

# Test another blocked destination
echo -n "    93.184.216.34 (example.com, blocked): "
docker exec "${CONTAINER_ID}" wget -q -O /dev/null --timeout=5 http://93.184.216.34/ 2>/dev/null && echo "REACHABLE [FAIL — should be blocked!]" || echo "BLOCKED [OK]"

echo ""
echo "--- Step 5: Cleanup ---"

# Remove FORWARD jump
iptables -D FORWARD -s "${CONTAINER_IP}" -j "${CHAIN}" 2>/dev/null || true
echo "[OK] FORWARD jump removed"

# Flush and delete chain
iptables -F "${CHAIN}" 2>/dev/null || true
iptables -X "${CHAIN}" 2>/dev/null || true
echo "[OK] Chain ${CHAIN} deleted"

# Stop container
docker rm -f "${CONTAINER_ID}" > /dev/null 2>&1
echo "[OK] Container removed"

echo ""
echo "=== Verification Complete ==="
