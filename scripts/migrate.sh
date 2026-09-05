#!/bin/bash
case "$1" in 
    up|down|status)
        ;;
    *)
        echo "Usage: $0 {up|down|status}" >&2
        exit 1
        ;;
esac

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
set -e

set -a
source "$SCRIPT_DIR/../.env"
set +a

goose -dir "$SCRIPT_DIR/../db/migrations" postgres "$DATABASE_URL" "$1"