#!/usr/bin/env bash
# bootstrap.sh — One-shot setup for a fresh EC2 instance (ARM64 / Amazon Linux 2 or Ubuntu)
# Usage: curl -fsSL https://raw.githubusercontent.com/gunni1/leipzig-library-media-search/main/scripts/bootstrap.sh | sudo bash

set -euo pipefail

REPO="gunni1/leipzig-library-media-search"
BINARY_NAME="lib-api-arm64-linux"
CHECKSUMS_NAME="SHA256SUMS"
INSTALL_PATH="/usr/local/bin/lib-api"
VERSION_FILE="/usr/local/share/lib-api-version"
UPDATE_SCRIPT="/usr/local/bin/lib-update.sh"

# Only allow downloads from GitHub's own domains
validate_url() {
  local url="$1"
  if [[ "$url" != https://github.com/* ]] && [[ "$url" != https://objects.githubusercontent.com/* ]]; then
    echo "ERROR: Unexpected download URL '${url}' — aborting." >&2
    exit 1
  fi
}

echo "==> Fetching latest release info from GitHub..."
RELEASE_JSON=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest")
LATEST_TAG=$(echo "$RELEASE_JSON" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\(.*\)".*/\1/')
DOWNLOAD_URL=$(echo "$RELEASE_JSON" | grep "browser_download_url" | grep "/${BINARY_NAME}\"" | sed 's/.*"browser_download_url": *"\(.*\)".*/\1/')
CHECKSUMS_URL=$(echo "$RELEASE_JSON" | grep "browser_download_url" | grep "/${CHECKSUMS_NAME}\"" | sed 's/.*"browser_download_url": *"\(.*\)".*/\1/')

if [ -z "$LATEST_TAG" ] || [ -z "$DOWNLOAD_URL" ] || [ -z "$CHECKSUMS_URL" ]; then
  echo "ERROR: Could not determine latest release tag or download URLs." >&2
  exit 1
fi

validate_url "$DOWNLOAD_URL"
validate_url "$CHECKSUMS_URL"

echo "==> Downloading ${LATEST_TAG}..."
TMP_BIN=$(mktemp)
TMP_SUMS=$(mktemp)
curl -fsSL "$DOWNLOAD_URL" -o "$TMP_BIN"
curl -fsSL "$CHECKSUMS_URL" -o "$TMP_SUMS"

echo "==> Verifying SHA256 checksum..."
# SHA256SUMS contains "hash  filename"; rewrite to point at our temp file for verification
EXPECTED_HASH=$(awk '{print $1}' "$TMP_SUMS")
ACTUAL_HASH=$(sha256sum "$TMP_BIN" | awk '{print $1}')
if [ "$EXPECTED_HASH" != "$ACTUAL_HASH" ]; then
  echo "ERROR: SHA256 mismatch! expected=${EXPECTED_HASH} got=${ACTUAL_HASH}" >&2
  rm -f "$TMP_BIN" "$TMP_SUMS"
  exit 1
fi
echo "Checksum OK (${ACTUAL_HASH})"
rm -f "$TMP_SUMS"

chmod +x "$TMP_BIN"

echo "==> Granting permission to bind port 80 without running as root..."
setcap 'cap_net_bind_service=+ep' "$TMP_BIN"
mv "$TMP_BIN" "$INSTALL_PATH"

mkdir -p "$(dirname "$VERSION_FILE")"
echo "$LATEST_TAG" > "$VERSION_FILE"

# ── systemd service unit ────────────────────────────────────────────────────
echo "==> Writing systemd service unit..."
cat > /etc/systemd/system/lib-api.service <<'EOF'
[Unit]
Description=Leipzig Library Media Search
After=network.target

[Service]
ExecStart=/usr/local/bin/lib-api -port=80
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# ── update script ────────────────────────────────────────────────────────────
echo "==> Writing update script to ${UPDATE_SCRIPT}..."
cat > "$UPDATE_SCRIPT" <<UPDATESCRIPT
#!/usr/bin/env bash
# lib-update.sh — Checks GitHub Releases for a newer version and hot-swaps the binary if found.
set -euo pipefail

REPO="${REPO}"
BINARY_NAME="${BINARY_NAME}"
CHECKSUMS_NAME="${CHECKSUMS_NAME}"
INSTALL_PATH="${INSTALL_PATH}"
VERSION_FILE="${VERSION_FILE}"

# Only allow downloads from GitHub's own domains
validate_url() {
  local url="\$1"
  if [[ "\$url" != https://github.com/* ]] && [[ "\$url" != https://objects.githubusercontent.com/* ]]; then
    echo "lib-update: unexpected download URL '\${url}' — aborting." >&2
    exit 1
  fi
}

CURRENT_TAG=\$(cat "\$VERSION_FILE" 2>/dev/null || echo "none")

RELEASE_JSON=\$(curl -fsSL "https://api.github.com/repos/\${REPO}/releases/latest")
LATEST_TAG=\$(echo "\$RELEASE_JSON" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\\(.*\\)".*/\\1/')

if [ -z "\$LATEST_TAG" ]; then
  echo "lib-update: could not fetch latest release tag, skipping." >&2
  exit 0
fi

if [ "\$LATEST_TAG" = "\$CURRENT_TAG" ]; then
  exit 0
fi

echo "lib-update: new version \${LATEST_TAG} (current: \${CURRENT_TAG}), updating..."

DOWNLOAD_URL=\$(echo "\$RELEASE_JSON" | grep "browser_download_url" | grep "/\${BINARY_NAME}\"" | sed 's/.*"browser_download_url": *"\\(.*\\)".*/\\1/')
CHECKSUMS_URL=\$(echo "\$RELEASE_JSON" | grep "browser_download_url" | grep "/\${CHECKSUMS_NAME}\"" | sed 's/.*"browser_download_url": *"\\(.*\\)".*/\\1/')

if [ -z "\$DOWNLOAD_URL" ] || [ -z "\$CHECKSUMS_URL" ]; then
  echo "lib-update: could not find download URLs in release \${LATEST_TAG}." >&2
  exit 1
fi

validate_url "\$DOWNLOAD_URL"
validate_url "\$CHECKSUMS_URL"

TMP_BIN=\$(mktemp)
TMP_SUMS=\$(mktemp)
curl -fsSL "\$DOWNLOAD_URL" -o "\$TMP_BIN"
curl -fsSL "\$CHECKSUMS_URL" -o "\$TMP_SUMS"

EXPECTED_HASH=\$(awk '{print \$1}' "\$TMP_SUMS")
ACTUAL_HASH=\$(sha256sum "\$TMP_BIN" | awk '{print \$1}')
if [ "\$EXPECTED_HASH" != "\$ACTUAL_HASH" ]; then
  echo "lib-update: SHA256 mismatch! expected=\${EXPECTED_HASH} got=\${ACTUAL_HASH} — aborting." >&2
  rm -f "\$TMP_BIN" "\$TMP_SUMS"
  exit 1
fi
echo "lib-update: checksum OK (\${ACTUAL_HASH})"
rm -f "\$TMP_SUMS"

chmod +x "\$TMP_BIN"
setcap 'cap_net_bind_service=+ep' "\$TMP_BIN"
mv "\$TMP_BIN" "\$INSTALL_PATH"
echo "\$LATEST_TAG" > "\$VERSION_FILE"

systemctl restart lib-api
echo "lib-update: restarted lib-api at version \${LATEST_TAG}."
UPDATESCRIPT

chmod +x "$UPDATE_SCRIPT"

# ── systemd timer + service for the update script ───────────────────────────
echo "==> Writing lib-update systemd units..."
cat > /etc/systemd/system/lib-update.service <<'EOF'
[Unit]
Description=Leipzig Library Media Search — update check

[Service]
Type=oneshot
ExecStart=/usr/local/bin/lib-update.sh
StandardOutput=journal
StandardError=journal
EOF

cat > /etc/systemd/system/lib-update.timer <<'EOF'
[Unit]
Description=Leipzig Library Media Search — poll for updates every 5 minutes

[Timer]
OnBootSec=2min
OnUnitActiveSec=5min
Unit=lib-update.service

[Install]
WantedBy=timers.target
EOF

# ── enable and start everything ──────────────────────────────────────────────
echo "==> Enabling and starting services..."
systemctl daemon-reload
systemctl enable --now lib-api
systemctl enable --now lib-update.timer

echo ""
echo "Done. lib-api is running at version ${LATEST_TAG}."
echo "  Status:  systemctl status lib-api"
echo "  Logs:    journalctl -u lib-api -f"
echo "  Updates: systemctl status lib-update.timer"
