#!/bin/bash
set -e

if command -v opencode &>/dev/null; then
  echo "  ✅ OpenCode already installed"
else
  echo "  📦 Installing OpenCode..."
  curl -fsSL https://opencode.ai/install | bash
fi
