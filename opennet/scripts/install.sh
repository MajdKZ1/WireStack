#!/usr/bin/env bash
# OpenNET installer/builder
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
INSTALL_PREFIX="/usr/local/bin"
TARGET_BIN="$REPO_ROOT/opennet"
IS_ROOT="${EUID:-$(id -u)}"

banner() {
  cat <<'EOF'
   ___              _   _   _ ______ _______
  / _ \  ___ _ __  | | | | | |  _ \ \_   _/ \
 | | | |/ _ \ '__| | | | | | | |_) | | | |/  /
 | |_| |  __/ |    | |_| | |_| |  _ <  | | /\ \
  \___/ \___|_|     \___/ \___/|_| \_\ |_| \/ /
           OpenNET installer / by Majd Alzadjali
EOF
}

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
  log "Installing dependencies (Go, WireGuard tools, Tor, torsocks, Python)..."
  run_as_root apt-get install -y --no-install-recommends \
    golang-go wireguard wireguard-tools tor torsocks python3 python3-pip
}

build_opennet() {
  if ! command -v go >/dev/null 2>&1; then
    err "Go is not installed or not in PATH. Install Go then rerun."
    exit 1
  fi
  log "Building OpenNET binary..."
  (cd "$REPO_ROOT" && go build -o "$TARGET_BIN" ./cmd/opennet)
  log "Built $TARGET_BIN"
}

install_binary() {
  local dest="$INSTALL_PREFIX/opennet"
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
  cp "$TARGET_BIN" "$user_bin/opennet"
  chmod +x "$user_bin/opennet"
  log "Installed to $user_bin/opennet. Ensure $user_bin is in your PATH."
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

linux_menu() {
  while true; do
    cat <<'EOF'
Linux distro:
 1) Ubuntu / Debian (supported)
 2) Arch-based (coming soon)
 b) Back
EOF
    read -r -p "Select option: " choice
    case "$choice" in
      1)
        install_linux_ubuntu
        build_opennet
        install_binary
        validate_tools
        ensure_path_contains_local_bin
        return
        ;;
      2)
        warn "Arch support coming soon. Nothing to do."
        ;;
      b|B)
        return
        ;;
      *)
        warn "Invalid choice."
        ;;
    esac
  done
}

main_menu() {
  banner
  while true; do
    cat <<'EOF'
Choose platform:
 1) Linux (Ubuntu/Debian supported now)
 2) macOS (coming soon)
 3) Windows (coming soon)
 q) Quit
EOF
    read -r -p "Select option: " choice
    case "$choice" in
      1) linux_menu ;;
      2) warn "macOS support coming soon. Nothing to do yet." ;;
      3) warn "Windows support coming soon. Nothing to do yet." ;;
      q|Q) exit 0 ;;
      *) warn "Invalid choice." ;;
    esac
  done
}

main_menu
