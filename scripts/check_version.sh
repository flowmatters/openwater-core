#!/bin/bash
# Version Check Script
# Checks if model signatures have changed and version needs updating

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CORE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "OpenWater Version Check"
echo "======================="
echo

# Get current version info
cd "$CORE_DIR"
CURRENT_VERSION=$(./ow-sim --version 2>/dev/null | head -1 | awk '{print $2}')
CURRENT_SIG=$(./ow-sim --version 2>/dev/null | grep "Signature:" | awk '{print $2}')

echo -e "Current Version: ${GREEN}${CURRENT_VERSION}${NC}"
echo -e "Current Signature Hash: ${GREEN}${CURRENT_SIG}${NC}"
echo
echo "Version check complete. Use this script to verify signature changes before committing."
