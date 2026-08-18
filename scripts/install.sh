#!/usr/bin/env bash
set -euo pipefail

REPO_URL="${CLEAN_CODE_REPO:-https://github.com/shashank-sn/clean-code.git}"
BRANCH="${CLEAN_CODE_BRANCH:-codex/initial-release}"
INSTALL_DIR="${CLEAN_CODE_INSTALL_DIR:-${HOME}/.clean-code-cli}"
CLEAN_CODE_HOME="${CLEAN_CODE_HOME:-${INSTALL_DIR}}"

echo "Installing Clean Code CLI from ${REPO_URL} (${BRANCH})..."

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

git clone --depth 1 --branch "$BRANCH" "$REPO_URL" "$TMP/clean-code"
cd "$TMP/clean-code"

if command -v node >/dev/null 2>&1; then
  node bin/runtime.js ensure-all
else
  bash scripts/ensure-runtimes.sh
fi

RUNTIME_PATH=""
if [[ -d "${CLEAN_CODE_HOME}/runtime/go/bin" ]]; then
  RUNTIME_PATH="${CLEAN_CODE_HOME}/runtime/go/bin"
fi
if [[ -d "${CLEAN_CODE_HOME}/runtime/node/bin" ]]; then
  if [[ -n "$RUNTIME_PATH" ]]; then
    RUNTIME_PATH="${RUNTIME_PATH}:${CLEAN_CODE_HOME}/runtime/node/bin"
  else
    RUNTIME_PATH="${CLEAN_CODE_HOME}/runtime/node/bin"
  fi
fi
if [[ -n "$RUNTIME_PATH" ]]; then
  export PATH="${RUNTIME_PATH}:${PATH}"
fi

mkdir -p "$INSTALL_DIR"
PKG_VERSION="$(node -p "require('./package.json').version" 2>/dev/null || sed -n 's/.*"version": "\([^"]*\)".*/\1/p' package.json | head -1)"
go build -ldflags "-X main.version=${PKG_VERSION}" -o "$INSTALL_DIR/clean-code" ./cmd/clean-code

BIN_DIR="${HOME}/.local/bin"
mkdir -p "$BIN_DIR"
if [[ -L "${BIN_DIR}/clean-code" || -e "${BIN_DIR}/clean-code" ]]; then
  echo ""
  echo "Note: ${BIN_DIR}/clean-code already exists and will be replaced."
  echo "If you use npm, prefer: npm install -g @shashanksn/clean-code"
  echo "and remove ${BIN_DIR}/clean-code to avoid PATH shadowing."
fi
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
