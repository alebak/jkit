#!/bin/bash
set -e

echo "🚀 Setting up {{.ProjectName}} development environment..."

# npm global prefix → ~/.local (persisted via home volume)
mkdir -p "$HOME/.local"
npm config set prefix "$HOME/.local"
export PATH="$HOME/.local/bin:$PATH"

# Persist PATH for future shells
if ! grep -qF '.local/bin' "$HOME/.bashrc" 2>/dev/null; then
  echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.bashrc"
fi

# Install gentle-ai (always required)
if command -v gentle-ai &>/dev/null; then
  echo "✅ gentle-ai already installed"
else
  echo "📦 Installing gentle-ai..."
  curl -fsSL https://raw.githubusercontent.com/Gentleman-Programming/gentle-ai/main/scripts/install.sh | bash
fi
