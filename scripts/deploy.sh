#!/usr/bin/env bash
# Build diary-api + diary-web and rsync to the VPS, then restart the API.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="${DEPLOY_ENV_FILE:-$ROOT/deploy/env}"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Missing $ENV_FILE — copy deploy/env.example to deploy/env and edit it." >&2
  exit 1
fi

# shellcheck disable=SC1090
source "$ENV_FILE"

: "${DEPLOY_HOST:?}"
: "${DEPLOY_USER:?}"
: "${REMOTE_API_DIR:?}"
: "${REMOTE_WEB_DIR:?}"

GOOS="${GOOS:-linux}"
GOARCH="${GOARCH:-amd64}"
SSH_TARGET="${DEPLOY_USER}@${DEPLOY_HOST}"
BUILD_DIR="$ROOT/.deploy-build"
API_BIN="$BUILD_DIR/diary-api"

cleanup() { rm -rf "$BUILD_DIR"; }
trap cleanup EXIT

mkdir -p "$BUILD_DIR"

echo "==> Building API (${GOOS}/${GOARCH})"
(
  cd "$ROOT/diary-api"
  CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -o "$API_BIN" ./cmd/server
)

echo "==> Building web"
(
  cd "$ROOT/diary-web"
  npm ci
  npm run build
)

echo "==> Ensuring remote dirs"
ssh "$SSH_TARGET" "mkdir -p '$REMOTE_API_DIR/migrations' '$REMOTE_WEB_DIR' /var/lib/3xn"

echo "==> Uploading API binary + migrations"
rsync -avz --delete \
  "$API_BIN" \
  "$SSH_TARGET:$REMOTE_API_DIR/diary-api"
rsync -avz --delete \
  "$ROOT/diary-api/migrations/" \
  "$SSH_TARGET:$REMOTE_API_DIR/migrations/"

echo "==> Uploading web dist"
rsync -avz --delete \
  "$ROOT/diary-web/dist/" \
  "$SSH_TARGET:$REMOTE_WEB_DIR/"

echo "==> Restarting API"
ssh "$SSH_TARGET" bash -s <<EOF
set -euo pipefail
chmod +x '$REMOTE_API_DIR/diary-api'
systemctl restart diary-api
systemctl --no-pager --full status diary-api | head -n 20
EOF

echo "==> Done. Open https://${DEPLOY_DOMAIN:-$DEPLOY_HOST}"
