#!/usr/bin/env bash
# Fails when apps/web/src/lib/server/client.gen.ts is out of sync with the API code.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMMITTED="$ROOT/apps/web/src/lib/server/client.gen.ts"
SPEC="$(mktemp -t zigzag-openapi.XXXXXX.yaml)"
GENERATED="$(mktemp -t zigzag-client.XXXXXX.ts)"
trap 'rm -f "$SPEC" "$GENERATED"' EXIT

(cd "$ROOT/apps/api" && go run ./cmd/openapi) >"$SPEC"
(cd "$ROOT/apps/web" && node ./scripts/generate-sdk/main.ts -i "$SPEC" -o "$GENERATED") >/dev/null

if ! diff -q "$COMMITTED" "$GENERATED" >/dev/null 2>&1; then
	echo "Web client types are out of sync with the API."
	echo "  Run make gen-web-types and commit apps/web/src/lib/server/client.gen.ts"
	exit 1
fi
