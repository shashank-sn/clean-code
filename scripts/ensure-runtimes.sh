#!/usr/bin/env bash
# Bootstrap Node.js and Go without an existing Node runtime (curl + tar only).
set -euo pipefail

GO_VERSION="${CLEAN_CODE_GO_VERSION:-1.22.10}"
NODE_VERSION="${CLEAN_CODE_NODE_VERSION:-20.18.0}"
CLEAN_CODE_HOME="${CLEAN_CODE_HOME:-${HOME}/.clean-code-cli}"
RUNTIME_HOME="${CLEAN_CODE_HOME}/runtime"
CACHE_DIR="${RUNTIME_HOME}/cache"

log() {
  echo "clean-code: $*"
}

ensure_dir() {
  mkdir -p "$1"
}

detect_platform() {
  local os arch go_os go_arch node_os node_arch
  case "$(uname -s)" in
    Linux) os="linux"; go_os="linux"; node_os="linux" ;;
    Darwin) os="darwin"; go_os="darwin"; node_os="darwin" ;;
    MINGW*|MSYS*|CYGWIN*) os="windows"; go_os="windows"; node_os="win" ;;
    *) echo "unsupported OS: $(uname -s)" >&2; exit 1 ;;
  esac
  case "$(uname -m)" in
    x86_64|amd64) arch="amd64"; go_arch="amd64"; node_arch="x64" ;;
    arm64|aarch64) arch="arm64"; go_arch="arm64"; node_arch="arm64" ;;
    *) echo "unsupported architecture: $(uname -m)" >&2; exit 1 ;;
  esac
  PLATFORM_OS="$os"
  GO_OS="$go_os"
  GO_ARCH="$go_arch"
  NODE_OS="$node_os"
  NODE_ARCH="$node_arch"
}

download() {
  local url="$1"
  local dest="$2"
  ensure_dir "$(dirname "$dest")"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$dest"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO "$dest" "$url"
  else
    echo "curl or wget is required to download runtimes" >&2
    exit 1
  fi
}

# verify_sha256 fails closed: the archive is removed and the script exits
# unless the local file's SHA-256 exactly matches the expected hex digest.
verify_sha256() {
  local archive_path="$1"
  local expected_hex="$2"
  local label="$3"
  expected_hex="$(printf '%s' "$expected_hex" | tr '[:upper:]' '[:lower:]')"
  if ! [[ "$expected_hex" =~ ^[0-9a-f]{64}$ ]]; then
    echo "clean-code: invalid or missing SHA-256 checksum for ${label}" >&2
    rm -f "$archive_path"
    exit 1
  fi
  local actual_hex
  if command -v shasum >/dev/null 2>&1; then
    actual_hex="$(shasum -a 256 "$archive_path" | awk '{print $1}')"
  elif command -v sha256sum >/dev/null 2>&1; then
    actual_hex="$(sha256sum "$archive_path" | awk '{print $1}')"
  else
    echo "clean-code: shasum or sha256sum is required to verify runtimes" >&2
    rm -f "$archive_path"
    exit 1
  fi
  if [[ "$actual_hex" != "$expected_hex" ]]; then
    echo "clean-code: checksum mismatch for ${label}: expected ${expected_hex}, got ${actual_hex}" >&2
    rm -f "$archive_path"
    exit 1
  fi
}

# download_checksum fetches a small checksum text file to stdout. On failure it
# emits nothing and returns non-zero so callers fail closed via verify_sha256.
download_checksum() {
  local url="$1"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO- "$url"
  else
    return 1
  fi
}

install_go() {
  if [[ -x "${RUNTIME_HOME}/go/bin/go" ]]; then
    return 0
  fi
  if command -v go >/dev/null 2>&1; then
    local version major minor rest
    version="$(go version | awk '{print $3}' | sed 's/^go//')"
    major="${version%%.*}"
    rest="${version#*.}"
    minor="${rest%%.*}"
    if [[ "$major" -gt 1 ]] || { [[ "$major" -eq 1 ]] && [[ "$minor" -ge 22 ]]; }; then
      return 0
    fi
  fi

  ensure_dir "$CACHE_DIR"
  local archive="go${GO_VERSION}.${GO_OS}-${GO_ARCH}.tar.gz"
  local url="https://go.dev/dl/${archive}"
  local archive_path="${CACHE_DIR}/${archive}"
  local extract_root="${CACHE_DIR}/extract-go"

  log "Downloading Go ${GO_VERSION}..."
  download "$url" "$archive_path"
  local go_expected
  go_expected="$(download_checksum "https://go.dev/dl/${archive}.sha256" 2>/dev/null | awk '{print $1}')"
  verify_sha256 "$archive_path" "$go_expected" "Go ${archive}"
  rm -rf "$extract_root"
  ensure_dir "$extract_root"
  tar -xzf "$archive_path" -C "$extract_root"
  rm -rf "${RUNTIME_HOME}/go"
  mv "${extract_root}/go" "${RUNTIME_HOME}/go"
  log "Go ${GO_VERSION} installed under ${RUNTIME_HOME}/go"
}

install_node() {
  if command -v node >/dev/null 2>&1; then
    local major
    major="$(node -p 'process.versions.node.split(".")[0]')"
    if [[ "$major" -ge 18 ]]; then
      return 0
    fi
  fi
  if [[ -x "${RUNTIME_HOME}/node/bin/node" ]]; then
    return 0
  fi

  ensure_dir "$CACHE_DIR"
  local folder="node-v${NODE_VERSION}-${NODE_OS}-${NODE_ARCH}"
  local archive="${folder}.tar.xz"
  local url="https://nodejs.org/dist/v${NODE_VERSION}/${archive}"
  local archive_path="${CACHE_DIR}/${archive}"
  local extract_root="${CACHE_DIR}/extract-node"

  log "Downloading Node.js ${NODE_VERSION}..."
  download "$url" "$archive_path"
  local node_expected
  node_expected="$(download_checksum "https://nodejs.org/dist/v${NODE_VERSION}/SHASUMS256.txt" 2>/dev/null | awk -v want="$archive" '$2 == want { print $1 }')"
  verify_sha256 "$archive_path" "$node_expected" "Node.js ${archive}"
  rm -rf "$extract_root"
  ensure_dir "$extract_root"
  tar -xJf "$archive_path" -C "$extract_root"
  rm -rf "${RUNTIME_HOME}/node"
  mv "${extract_root}/${folder}" "${RUNTIME_HOME}/node"
  log "Node.js ${NODE_VERSION} installed under ${RUNTIME_HOME}/node"
}

detect_platform
install_node
install_go
