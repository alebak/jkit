#!/bin/bash
set -e
 
# npm global prefix → ~/.local (persisted via home volume)
mkdir -p "$HOME/.local"
npm config set prefix "$HOME/.local"
export PATH="$HOME/.local/bin:$PATH"
 
# Persist PATH for future shells
if ! grep -qF '.local/bin' "$HOME/.bashrc" 2>/dev/null; then
  echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.bashrc"
fi
 
# Install tools only if not present
if command -v claude &>/dev/null; then
  echo "✅ Claude Code already installed"
else
  echo "📦 Installing Claude Code..."
  npm install -g @anthropic-ai/claude-code
fi

if command -v opencode &>/dev/null; then
  echo "✅ OpenCode already installed"
else
  echo "📦 Installing OpenCode..."
  curl -fsSL https://opencode.ai/install | bash
fi
 
# if command -v openspec &>/dev/null; then
#   echo "✅ OpenSpec already installed"
# else
#   echo "📦 Installing OpenSpec..."
#   npm install -g @fission-ai/openspec
# fi
 
if command -v gentle-ai &>/dev/null; then
  echo "✅ gentle-ai already installed"
else
  echo "📦 Installing gentle-ai..."
  curl -fsSL https://raw.githubusercontent.com/Gentleman-Programming/gentle-ai/main/scripts/install.sh | bash
fi
 