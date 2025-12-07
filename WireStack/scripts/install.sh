#!/usr/bin/env bash
# Wirestack installer/builder
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
INSTALL_PREFIX="/usr/local/bin"
TARGET_BIN="$REPO_ROOT/wirestack"
IS_ROOT="${EUID:-$(id -u)}"

log()  { printf "[INFO] %s\n" "$*"; }
warn() { printf "[WARN] %s\n" "$*"; }
err()  { printf "[ERR ] %s\n" "$*" >&2; }

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    err "Missing required command: $1"
    return 1
  fi
}

prompt_continue() {
  read -r -p "$1 [Y/n]: " ans
  ans=${ans:-y}
  [[ "${ans,,}" == y* ]]
}

have_working_sudo() {
  if ! command -v sudo >/dev/null 2>&1; then
    return 1
  fi
  sudo -n true >/dev/null 2>&1 || return 1
  return 0
}

run_as_root() {
  if [ "$IS_ROOT" -eq 0 ]; then
    "$@"
    return
  fi
  if have_working_sudo; then
    sudo "$@"
    return
  fi
  err "No sudo available and not running as root; cannot execute: $*"
  return 1
}

install_linux_ubuntu() {
  if [ "$IS_ROOT" -ne 0 ] && ! have_working_sudo; then
    warn "Cannot gain privileges (no sudo/root). Skipping package install. Install deps manually: golang-go wireguard wireguard-tools tor torsocks python3"
    return
  fi
  log "Updating apt repositories..."
  run_as_root apt-get update -y
  log "Installing dependencies (Go, WireGuard tools, Python)..."
  run_as_root apt-get install -y --no-install-recommends \
    golang-go wireguard wireguard-tools python3 python3-pip
}

build_wirestack() {
  if ! command -v go >/dev/null 2>&1; then
    err "Go is not installed or not in PATH. Install Go then rerun."
    exit 1
  fi
  log "Building Wirestack binary..."
  (cd "$REPO_ROOT" && go build -o "$TARGET_BIN" ./cmd/wirestack)
  log "Built $TARGET_BIN"
}

install_binary() {
  local dest="$INSTALL_PREFIX/wirestack"
  if prompt_continue "Copy binary to $dest (may require sudo)?"; then
    if run_as_root cp "$TARGET_BIN" "$dest"; then
      run_as_root chmod +x "$dest"
      log "Installed to $dest"
      return
    fi
    warn "Could not copy to $dest. Trying user-local ~/.local/bin"
  fi
  local user_bin="$HOME/.local/bin"
  mkdir -p "$user_bin"
  cp "$TARGET_BIN" "$user_bin/wirestack"
  chmod +x "$user_bin/wirestack"
  log "Installed to $user_bin/wirestack. Ensure $user_bin is in your PATH."
}

ensure_path_contains_local_bin() {
  case ":$PATH:" in
    *":$HOME/.local/bin:"*) return ;;
  esac
  warn "$HOME/.local/bin is not in your PATH."
  if prompt_continue "Add export PATH to ~/.bashrc?"; then
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.bashrc"
    log "Appended to ~/.bashrc. Restart your shell or run: source ~/.bashrc"
  else
    warn "Skipped PATH update. Add it manually if needed."
  fi
}

validate_tools() {
  local missing=0
  for tool in go wg wg-quick tor torsocks python3; do
    if ! command -v "$tool" >/dev/null 2>&1; then
      warn "Missing tool: $tool"
      missing=1
    fi
  done
  if [ "$missing" -eq 1 ]; then
    warn "Some tools are missing. Install them via your package manager."
  else
    log "All required tools detected."
  fi
}

main() {
  log "Wirestack installer (Linux Ubuntu/Debian)"
  install_linux_ubuntu
  build_wirestack
  install_binary
  validate_tools
  ensure_path_contains_local_bin
}

main
