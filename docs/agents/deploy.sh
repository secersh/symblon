#!/usr/bin/env bash
# Deploy all agent packages in this directory to the registrar.
#
# Usage:
#   SYMBLON_TOKEN=<jwt> ./deploy.sh
#   SYMBLON_TOKEN=<jwt> REGISTRAR_URL=http://localhost:8082 ./deploy.sh
#
# Each subdirectory containing an agent.yaml is zipped and posted to the
# registrar. Existing versions are skipped (409 Conflict).

set -euo pipefail

REGISTRAR_URL="${REGISTRAR_URL:-https://api.symblon.cc}"
SYMBLON_TOKEN="${SYMBLON_TOKEN:-}"

if [ -z "$SYMBLON_TOKEN" ]; then
  echo "error: SYMBLON_TOKEN is required"
  echo "usage: SYMBLON_TOKEN=<jwt> ./deploy.sh"
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
deployed=0
failed=0

for agent_dir in "$SCRIPT_DIR"/*/; do
  agent_name="$(basename "$agent_dir")"

  if [ ! -f "$agent_dir/agent.yaml" ]; then
    continue
  fi

  echo "→ deploying $agent_name ..."

  zip_file="/tmp/agent-${agent_name}-$$.zip"
  (cd "$SCRIPT_DIR" && zip -qr "$zip_file" "$agent_name/" --exclude "*/.DS_Store" "*.DS_Store")

  response=$(curl -s -o /tmp/agent-response.json -w "%{http_code}" \
    -X POST "$REGISTRAR_URL/registrar/v1/agents" \
    -H "Authorization: Bearer $SYMBLON_TOKEN" \
    -F "package=@$zip_file")

  rm -f "$zip_file"

  if [ "$response" = "201" ]; then
    ref=$(python3 -c "import json,sys; d=json.load(open('/tmp/agent-response.json')); print(d.get('ref',''))" 2>/dev/null || echo "")
    echo "  ok  $agent_name → $ref"
    deployed=$((deployed + 1))
  elif [ "$response" = "409" ]; then
    echo "  --  $agent_name already exists, skipping"
  else
    echo "  err $agent_name (HTTP $response)"
    cat /tmp/agent-response.json 2>/dev/null && echo ""
    failed=$((failed + 1))
  fi
done

echo ""
echo "done: $deployed deployed, $failed failed"
[ "$failed" -eq 0 ] || exit 1
