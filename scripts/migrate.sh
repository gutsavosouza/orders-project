#!/bin/bash
usage() {
    echo "Usage: $0 {orders|inventory} {up|down|status}" >&2
    exit 1
}

[ $# -ge 2 ] || usage

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
set -e

set -a
source "$SCRIPT_DIR/../.env"
set +a

case "$1" in
    orders)
        DB_URL="$ORDER_DATABASE_URL"
        ;;
    inventory)
        DB_URL="$INVENTORY_DATABASE_URL"
        ;;
    *)
        usage
        ;;
esac

MIGRATIONS_DIR="$SCRIPT_DIR/../db/migrations/$1"

case "$2" in
    up|down|status)
        goose -dir "$MIGRATIONS_DIR" postgres "$DB_URL" "$2"
        ;;
    *)
        usage
        ;;
esac
