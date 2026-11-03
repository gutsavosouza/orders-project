#!/bin/bash
# Usage: seed.sh [--reset]
#   --reset  also truncates processed_orders
#
# Runs psql inside the inventory-database container, so no host psql is needed.
# User and database come from the container's own POSTGRES_USER/POSTGRES_DB.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
set -e

PSQL_ARGS=()
case "$1" in
    "")
        ;;
    --reset)
        PSQL_ARGS=(-v reset=1)
        ;;
    *)
        echo "Usage: $0 [--reset]" >&2
        exit 1
        ;;
esac

docker exec -i inventory-database \
    sh -c 'exec psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" "$@" -f -' sh "${PSQL_ARGS[@]}" \
    < "$SCRIPT_DIR/seed_stock.sql"
