#!/bin/bash
set -e

# JKit installer — curl | bash
# Usage: curl -fsSL https://raw.githubusercontent.com/alebak/jkit/main/scripts/install.sh | bash

REPO="alebak/jkit"
BIN_NAME="jkit"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${VERSION:-latest}"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}📦 JKit installer${NC}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64)   ARCH="arm64" ;;
    *)       echo -e "${RED}❌ Unsupported architecture: $ARCH${NC}"; exit 1 ;;
esac

case "$OS" in
    linux)   OS="linux" ;;
    darwin)  OS="darwin" ;;
    *)       echo -e "${RED}❌ Unsupported OS: $OS${NC}"; exit 1 ;;
esac

# Determine download URL
if [ "$VERSION" = "latest" ]; then
    DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${BIN_NAME}-${OS}-${ARCH}"
else
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BIN_NAME}-${OS}-${ARCH}"
fi

# Create install directory
mkdir -p "$INSTALL_DIR"

# Download binary
TMP_DIR=$(mktemp -d)
TMP_BIN="$TMP_DIR/$BIN_NAME"

echo -e "${YELLOW}⬇️  Downloading $BIN_NAME ($OS/$ARCH)...${NC}"
if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$DOWNLOAD_URL" -o "$TMP_BIN"
elif command -v wget >/dev/null 2>&1; then
    wget -q "$DOWNLOAD_URL" -O "$TMP_BIN"
else
    echo -e "${RED}❌ Neither curl nor wget found. Install one and try again.${NC}"
    rm -rf "$TMP_DIR"
    exit 1
fi

# Make executable
chmod +x "$TMP_BIN"

# Install
mv "$TMP_BIN" "$INSTALL_DIR/$BIN_NAME"
rm -rf "$TMP_DIR"

# Verify installation
if [ -x "$INSTALL_DIR/$BIN_NAME" ]; then
    echo -e "${GREEN}✅ JKit installed to $INSTALL_DIR/$BIN_NAME${NC}"
    INSTALLED_VERSION=$("$INSTALL_DIR/$BIN_NAME" --version 2>/dev/null || echo "unknown")
    echo -e "${GREEN}   Version: $INSTALLED_VERSION${NC}"
else
    echo -e "${RED}❌ Installation failed${NC}"
    exit 1
fi

# PATH check
if ! echo "$PATH" | tr ':' '\n' | grep -qx "$INSTALL_DIR"; then
    echo -e "${YELLOW}⚠️  $INSTALL_DIR is not in your PATH${NC}"
    echo -e "${YELLOW}   Add this to your shell config:${NC}"
    echo -e "${YELLOW}   export PATH=\"\$HOME/.local/bin:\$PATH\"${NC}"
fi

echo ""
echo -e "${GREEN}Run 'jkit init' to start your Joomla project!${NC}"
