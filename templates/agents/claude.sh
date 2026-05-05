#!/bin/bash
set -e

if command -v claude &>/dev/null; then
  echo "  ✅ Claude Code already installed"
else
  echo "  📦 Installing Claude Code..."
  npm install -g @anthropic-ai/claude-code
fi
