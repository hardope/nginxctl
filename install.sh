#!/usr/bin/env bash
set -euo pipefail

REPO="hardope/nginxctl"
BIN_DIR="/usr/local/bin"
BIN_NAME="nginxctl"

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH_SUFFIX="linux-amd64" ;;
  aarch64) ARCH_SUFFIX="linux-arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [[ -z "$LATEST" ]]; then
  echo "Could not determine latest release" >&2
  exit 1
fi

URL="https://github.com/${REPO}/releases/download/${LATEST}/nginxctl-${ARCH_SUFFIX}"

echo "Downloading nginxctl ${LATEST} (${ARCH_SUFFIX})..."
curl -fsSL "$URL" -o "${BIN_DIR}/${BIN_NAME}"
chmod +x "${BIN_DIR}/${BIN_NAME}"

echo "Installed → ${BIN_DIR}/${BIN_NAME}"
echo "Run: sudo nginxctl setup"
