#!/usr/bin/env bash
set -euo pipefail

REPO_URL="${CLEAN_CODE_REPO:-https://github.com/shashank-stitch/clean-code.git}"
BRANCH="${CLEAN_CODE_BRANCH:-codex/initial-release}"
INSTALL_DIR="${CLEAN_CODE_INSTALL_DIR:-${HOME}/.clean-code-cli}"

echo "Installing Clean Code CLI from ${REPO_URL} (${BRANCH})..."

if ! command -v go >/dev/null 2>&1; then
  echo "Go 1.22+ is required. Install from https://go.dev/dl/ or: brew install go"
  exit 1
fi

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

git clone --depth 1 --branch "$BRANCH" "$REPO_URL" "$TMP/clean-code"
cd "$TMP/clean-code"

mkdir -p "$INSTALL_DIR"
go build -o "$INSTALL_DIR/clean-code" ./cmd/clean-code

BIN_DIR="${HOME}/.local/bin"
mkdir -p "$BIN_DIR"
ln -sf "$INSTALL_DIR/clean-code" "$BIN_DIR/clean-code"

echo ""
echo "Installed: $INSTALL_DIR/clean-code"
echo "Symlink:   $BIN_DIR/clean-code"
if ! echo ":$PATH:" | grep -q ":${BIN_DIR}:"; then
  echo ""
  echo "Add to your shell profile:"
  echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
fi
echo ""
"$INSTALL_DIR/clean-code" version
