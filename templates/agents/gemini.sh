#!/bin/bash
set -e

if command -v gemini &>/dev/null; then
  echo "  ✅ Gemini CLI already installed"
else
  echo "  📦 Installing Gemini CLI..."
  npm install -g @google/gemini-cli
fi
