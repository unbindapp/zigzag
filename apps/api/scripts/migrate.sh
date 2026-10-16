#!/usr/bin/env bash
# initdb refuses to run as root, so the embedded postgres runs as nobody when we are root.
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."
DIR=ent/migrate/migrations

if [ "$(id -u)" != "0" ]; then
	exec go run -mod=mod ./ent/migrate/main.go "$@"
fi

SCRATCH="$(mktemp -d -t zigzag-migrate.XXXXXX)"
trap 'chown -R root:root "$DIR"; chmod 755 "$DIR"; chmod 644 "$DIR"/*; rm -rf "$SCRATCH"' EXIT
chmod 777 "$SCRATCH" "$DIR"
chmod 666 "$DIR"/* 2>/dev/null || true
for d in "$PWD" $(p="$PWD"; while [ "$p" != "/" ]; do p="$(dirname "$p")"; echo "$p"; done); do chmod o+x "$d"; done

setpriv --reuid=65534 --regid=65534 --clear-groups env \
	HOME="$SCRATCH" GOCACHE="$SCRATCH/cache" GOMODCACHE="$(go env GOMODCACHE)" GOFLAGS=-mod=mod GOTOOLCHAIN=local PATH="$PATH" \
	go run -mod=mod ./ent/migrate/main.go "$@"
