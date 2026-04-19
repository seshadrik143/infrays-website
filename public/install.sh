#!/usr/bin/env bash
# NodePulse Agent — remote install script
# Hosted at https://infrays.org/install.sh
#
# Usage:
#   curl -fsSL https://infrays.org/install.sh | sudo bash -s -- \
#     --server http://YOUR_SERVER:8080 \
#     --key    npk_YOUR_API_KEY
#
# Optional flags:
#   --agent-id  my-web-01      custom agent name (default: hostname)
#   --version   v0.34.1        pin to a specific release (default: latest)
#   --uninstall                remove NodePulse agent from this host
#
# Supported: any systemd-based Linux (amd64 / arm64)

set -euo pipefail

# ── Constants ──────────────────────────────────────────────────────────────────
BINARY_NAME="nodepulse-agent"
INSTALL_BIN="/usr/local/bin/${BINARY_NAME}"
CONFIG_DIR="/etc/nodepulse"
CONFIG_FILE="${CONFIG_DIR}/config.yaml"
DATA_DIR="/var/lib/nodepulse"
SERVICE_FILE="/etc/systemd/system/nodepulse-agent.service"
SERVICE_USER="nodepulse"
GITHUB_REPO="NodepulseRepo/NodePulse"
RELEASES_API="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
RELEASES_DL="https://github.com/${GITHUB_REPO}/releases/download"

# ── Colour helpers ─────────────────────────────────────────────────────────────
GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; BOLD='\033[1m'; NC='\033[0m'
info()    { echo -e "${GREEN}[nodepulse]${NC} $*"; }
warn()    { echo -e "${YELLOW}[nodepulse]${NC} $*"; }
error()   { echo -e "${RED}[nodepulse]${NC} $*" >&2; }
heading() { echo -e "\n${BOLD}$*${NC}"; }

# ── Argument parsing ───────────────────────────────────────────────────────────
SERVER_URL=""
API_KEY=""
AGENT_ID="$(hostname -s)"
VERSION=""
UNINSTALL=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --server)    SERVER_URL="$2";  shift 2 ;;
    --key)       API_KEY="$2";     shift 2 ;;
    --agent-id)  AGENT_ID="$2";   shift 2 ;;
    --version)   VERSION="$2";    shift 2 ;;
    --uninstall) UNINSTALL=true;  shift   ;;
    *) error "Unknown flag: $1"; exit 1   ;;
  esac
done

# ── Root check ─────────────────────────────────────────────────────────────────
if [[ $EUID -ne 0 ]]; then
  error "This script must be run as root. Re-run with: sudo bash -s -- $*"
  exit 1
fi

# ── systemd check ─────────────────────────────────────────────────────────────
if ! command -v systemctl &>/dev/null; then
  error "systemd not found — this installer requires a systemd-based Linux distribution."
  exit 1
fi

# ── Uninstall ─────────────────────────────────────────────────────────────────
if [[ "${UNINSTALL}" == "true" ]]; then
  heading "Uninstalling NodePulse Agent..."
  systemctl stop nodepulse-agent.service  2>/dev/null || true
  systemctl disable nodepulse-agent.service 2>/dev/null || true
  rm -f "${SERVICE_FILE}"
  systemctl daemon-reload
  rm -f "${INSTALL_BIN}"
  warn "Config and data directories preserved:"
  warn "  Config : ${CONFIG_DIR}"
  warn "  Data   : ${DATA_DIR}"
  warn "Remove them manually if no longer needed:"
  warn "  sudo rm -rf ${CONFIG_DIR} ${DATA_DIR}"
  info "NodePulse Agent uninstalled."
  exit 0
fi

# ── Required args check ────────────────────────────────────────────────────────
if [[ -z "${SERVER_URL}" ]]; then
  error "Missing required flag: --server <url>"
  echo "  Example: --server http://192.168.1.10:8080"
  exit 1
fi
if [[ -z "${API_KEY}" ]]; then
  error "Missing required flag: --key <api-key>"
  echo "  Generate one: Dashboard → Settings → API Keys → Generate Key"
  exit 1
fi

# ── Architecture detection ────────────────────────────────────────────────────
ARCH="$(uname -m)"
case "${ARCH}" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  *)
    error "Unsupported architecture: ${ARCH} (supported: x86_64, aarch64)"
    exit 1
    ;;
esac
info "Detected architecture: ${ARCH}"

# ── Resolve version ───────────────────────────────────────────────────────────
if [[ -z "${VERSION}" ]]; then
  info "Fetching latest release version..."
  if command -v curl &>/dev/null; then
    VERSION="$(curl -fsSL "${RELEASES_API}" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
  elif command -v wget &>/dev/null; then
    VERSION="$(wget -qO- "${RELEASES_API}" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
  else
    error "Neither curl nor wget found — cannot fetch release info."
    exit 1
  fi
  if [[ -z "${VERSION}" ]]; then
    error "Could not determine latest version. Specify one with --version v0.34.0"
    exit 1
  fi
fi
info "Installing NodePulse Agent ${VERSION} (${ARCH})"

# ── Download binary ───────────────────────────────────────────────────────────
BINARY_URL="${RELEASES_DL}/${VERSION}/${BINARY_NAME}-linux-${ARCH}"
TMP_BINARY="$(mktemp)"
trap 'rm -f "${TMP_BINARY}"' EXIT

heading "Downloading binary..."
if command -v curl &>/dev/null; then
  curl -fsSL --progress-bar "${BINARY_URL}" -o "${TMP_BINARY}"
elif command -v wget &>/dev/null; then
  wget -q --show-progress "${BINARY_URL}" -O "${TMP_BINARY}"
fi

# Verify it's an ELF binary (basic sanity check)
if ! file "${TMP_BINARY}" 2>/dev/null | grep -q "ELF"; then
  # file command may not be available; fall back to magic bytes check
  if ! head -c 4 "${TMP_BINARY}" | grep -q $'\x7fELF'; then
    error "Downloaded file does not appear to be a valid ELF binary."
    error "Check that version ${VERSION} exists at:"
    error "  ${BINARY_URL}"
    exit 1
  fi
fi
info "Binary downloaded and verified."

# ── System user ───────────────────────────────────────────────────────────────
heading "Creating system user..."
if ! id -u "${SERVICE_USER}" &>/dev/null; then
  useradd \
    --system \
    --no-create-home \
    --shell /usr/sbin/nologin \
    --comment "NodePulse monitoring agent" \
    "${SERVICE_USER}"
  info "Created user '${SERVICE_USER}'"
else
  info "User '${SERVICE_USER}' already exists"
fi

# ── Directories ───────────────────────────────────────────────────────────────
heading "Creating directories..."
install -d -m 755 "${CONFIG_DIR}"
install -d -m 750 -o "${SERVICE_USER}" -g "${SERVICE_USER}" "${DATA_DIR}"

# ── Install binary ────────────────────────────────────────────────────────────
heading "Installing binary..."
install -m 755 "${TMP_BINARY}" "${INSTALL_BIN}"
info "Installed: ${INSTALL_BIN}"

# ── Write config ──────────────────────────────────────────────────────────────
heading "Writing config..."
if [[ -f "${CONFIG_FILE}" ]]; then
  warn "Config already exists at ${CONFIG_FILE} — updating server_url and api_key only."
  sed -i "s|^server_url:.*|server_url: \"${SERVER_URL}\"|" "${CONFIG_FILE}"
  sed -i "s|api_key:.*|api_key: \"${API_KEY}\"|"           "${CONFIG_FILE}"
  sed -i "s|^agent_id:.*|agent_id: ${AGENT_ID}|"          "${CONFIG_FILE}"
else
  cat > "${CONFIG_FILE}" <<EOF
# NodePulse Agent Configuration
# Generated by install.sh on $(date -u +"%Y-%m-%dT%H:%M:%SZ")
# Full reference: https://infrays.org/docs/setup-guide.html#agent

agent_id: ${AGENT_ID}
server_url: "${SERVER_URL}"

auth:
  api_key: "${API_KEY}"

modules:
  system_metrics: true
  process_monitoring: true
  docker: false          # set true if Docker is running on this host
  kubernetes: false
  network: true
  logs: true
  systemd: true
  kernel: true

intervals:
  system_metrics: 5      # seconds
  process: 10
  kernel: 5

transport:
  use_gzip: true
  flush_interval_sec: 10
  tls_skip_verify: false

queue:
  max_memory_items: 10000
  max_disk_mb: 500
  data_dir: ${DATA_DIR}

auto_update:
  enabled: true
  check_interval: 1h
EOF
  chmod 640 "${CONFIG_FILE}"
  chown root:"${SERVICE_USER}" "${CONFIG_FILE}"
  info "Config written to ${CONFIG_FILE}"
fi

# ── Systemd unit ─────────────────────────────────────────────────────────────
heading "Installing systemd service..."
cat > "${SERVICE_FILE}" <<'UNIT'
[Unit]
Description=NodePulse Monitoring Agent
Documentation=https://infrays.org/docs/setup-guide.html
After=network-online.target
Wants=network-online.target
StartLimitIntervalSec=60
StartLimitBurst=5

[Service]
Type=simple
User=nodepulse
Group=nodepulse
ExecStart=/usr/local/bin/nodepulse-agent -config /etc/nodepulse/config.yaml
Restart=on-failure
RestartSec=5s

ReadWritePaths=/var/lib/nodepulse
ReadOnlyPaths=/etc/nodepulse
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
NoNewPrivileges=true
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true
ProcSubset=all
ProtectProc=default
AmbientCapabilities=
CapabilityBoundingSet=

StandardOutput=journal
StandardError=journal
SyslogIdentifier=nodepulse-agent

LimitNOFILE=65536
MemoryMax=256M
CPUQuota=20%

[Install]
WantedBy=multi-user.target
UNIT

# ── Enable & start ────────────────────────────────────────────────────────────
heading "Starting service..."
systemctl daemon-reload
systemctl enable nodepulse-agent.service
systemctl restart nodepulse-agent.service

# Brief wait then check status
sleep 2
if systemctl is-active --quiet nodepulse-agent.service; then
  STATUS="${GREEN}running${NC}"
else
  STATUS="${RED}failed to start${NC}"
fi

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e " NodePulse Agent ${VERSION} installed successfully"
echo -e " Status : $(systemctl is-active nodepulse-agent.service 2>/dev/null || echo 'unknown')"
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "  Agent ID  : ${AGENT_ID}"
echo "  Server    : ${SERVER_URL}"
echo "  Binary    : ${INSTALL_BIN}"
echo "  Config    : ${CONFIG_FILE}"
echo "  Data      : ${DATA_DIR}"
echo ""
echo "  Check status : systemctl status nodepulse-agent"
echo "  View logs    : journalctl -u nodepulse-agent -f"
echo "  Edit config  : nano ${CONFIG_FILE}"
echo "  Uninstall    : curl -fsSL https://infrays.org/install.sh | sudo bash -s -- --uninstall"
echo ""
