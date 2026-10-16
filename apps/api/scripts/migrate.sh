#!/usr/bin/env bash
# initdb refuses to run as root, so root diffs against a scratch database on the dev postgres container instead.
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

if [ "$(id -u)" != "0" ]; then
	exec go run -mod=mod ./ent/migrate/main.go "$@"
fi

for arg in "$@"; do
	if [ "$arg" = "-checksum" ]; then
		exec go run -mod=mod ./ent/migrate/main.go "$@"
	fi
done

CONTAINER="${MIGRATE_CONTAINER:-zigzag_postgres}"
PG_PORT="${POSTGRES_PORT:-54329}"
SCRATCH_DB="zigzag_migrate_$(od -An -N4 -tx4 /dev/urandom | tr -d ' ')"

psql_admin() {
	docker exec "$CONTAINER" psql -v ON_ERROR_STOP=1 -q -U postgres -d postgres -c "$1"
}

psql_admin "CREATE DATABASE $SCRATCH_DB"
trap 'psql_admin "DROP DATABASE IF EXISTS $SCRATCH_DB WITH (FORCE)"' EXIT

go run -mod=mod ./ent/migrate/main.go -database-url="postgres://postgres:postgres@127.0.0.1:$PG_PORT/$SCRATCH_DB?sslmode=disable" "$@"
