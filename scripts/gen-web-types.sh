#!/usr/bin/env bash
# Builds the OpenAPI spec offline (no server) and feeds it to the web client generator.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SPEC="$(mktemp -t zigzag-openapi.XXXXXX.yaml)"
trap 'rm -f "$SPEC"' EXIT

echo "Generating OpenAPI spec from apps/api..."
(cd "$ROOT/apps/api" && go run ./cmd/openapi) >"$SPEC"

echo "Generating web client from spec..."
(cd "$ROOT/apps/web" && node ./scripts/generate-sdk/main.ts -i "$SPEC")

echo "Done."
