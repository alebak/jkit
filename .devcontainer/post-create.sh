#!/bin/bash
set -e

# Persist PATH for future shells
mkdir -p "$HOME/.local/bin"
if ! grep -qF '.local/bin' "$HOME/.bashrc" 2>/dev/null; then
  echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.bashrc"
fi
export PATH="$HOME/.local/bin:$PATH"

# Install Claude Code
if command -v claude &>/dev/null; then
  echo "✅ Claude Code already installed"
else
  echo "📦 Installing Claude Code..."
  curl -fsSL https://claude.ai/install.sh | bash
fi

# Install OpenCode
if command -v opencode &>/dev/null; then
  echo "✅ OpenCode already installed"
else
  echo "📦 Installing OpenCode..."
  curl -fsSL https://opencode.ai/install | bash
fi

# Install gentle-ai
if command -v gentle-ai &>/dev/null; then
  echo "✅ gentle-ai already installed"
else
  echo "📦 Installing gentle-ai..."
  curl -fsSL https://raw.githubusercontent.com/Gentleman-Programming/gentle-ai/main/scripts/install.sh | bash
fi
