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

# Install jkit (Joomla scaffolding toolkit)
if command -v jkit &>/dev/null; then
  echo "✅ jkit already installed"
else
  echo "📦 Installing jkit..."
  curl -fsSL https://raw.githubusercontent.com/alebak/jkit/main/scripts/install.sh | bash
fi

# Install gentle-ai (always required)
if command -v gentle-ai &>/dev/null; then
  echo "✅ gentle-ai already installed"
else
  echo "📦 Installing gentle-ai..."
  curl -fsSL https://raw.githubusercontent.com/Gentleman-Programming/gentle-ai/main/scripts/install.sh | bash
fi

# Link JKit skills into agent directories
echo "🔗 Linking JKit skills..."
if [ -d ".jkit/agents/skills" ]; then
  for skill_dir in .jkit/agents/skills/*/; do
    skill_name=$(basename "$skill_dir")
    for agent_dir in "$HOME/.claude/skills" "$HOME/.config/opencode/skills" "$HOME/.gemini/skills"; do
      if [ -d "$agent_dir" ] || mkdir -p "$agent_dir" 2>/dev/null; then
        ln -sf "$(pwd)/$skill_dir" "$agent_dir/$skill_name" 2>/dev/null || true
      fi
    done
  done
  echo "✅ JKit skills linked"
fi
